package spec_util

import (
	"bytes"
	"encoding/base64"

	protohash "github.com/akitasoftware/objecthash-proto"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

// Given 2 DataTemplates that may contain references to Data in the common
// prefix, return whether the 2 DataTemplates are equivalent.
// Equivalence is defined as identical (modulo non-fixed values) after constant
// propagation, and there's a bijection between the non-fixed values. We can't
// use bijection to define equivalence over fixed values because each fixed
// value has a different meaning. For example, in enum{WRITER, READER}, it would
// be wrong to say WRITER == READER. Note we always treat bool primitives as
// "fixed", so bijection does not apply to bool values.
func EquivalentDataTemplates(sharedPrefix []*pb.MethodTemplate, dt1 *pb.DataTemplate, dt2 *pb.DataTemplate) (bool, error) {
	chk := &equivChecker{
		sharedPrefix:     sharedPrefix,
		lToRPrimitiveMap: make(map[string]*pb.Primitive),
		rToLPrimitiveMap: make(map[string]*pb.Primitive),
	}
	return chk.equivalentDataTemplates(dt1, dt2)
}

type equivChecker struct {
	sharedPrefix []*pb.MethodTemplate

	// Maintains bijective mapping from hash(primitive_1) to primitive_2 and vice
	// versa.
	// TODO: We are assuming there are no hash collision. If collision happens,
	// we'll say that 2 equivalent templates are not equivalent.
	lToRPrimitiveMap map[string]*pb.Primitive
	rToLPrimitiveMap map[string]*pb.Primitive
}

func (chk *equivChecker) equivalentDataTemplates(dt1 *pb.DataTemplate, dt2 *pb.DataTemplate) (bool, error) {
	// Unroll DataTemplates to resolve top-level references.
	var err error
	dt1, err = unrollDataTemplate(chk.sharedPrefix, dt1)
	if err != nil {
		return false, errors.Wrapf(err, "unroll DataTemplate failed")
	}
	dt2, err = unrollDataTemplate(chk.sharedPrefix, dt2)
	if err != nil {
		return false, errors.Wrapf(err, "unroll DataTemplate failed")
	}

	switch vt1 := dt1.ValueTemplate.(type) {
	case *pb.DataTemplate_StructTemplate:
		if vt2, ok := dt2.ValueTemplate.(*pb.DataTemplate_StructTemplate); ok {
			return chk.equivalentStructTemplates(vt1.StructTemplate, vt2.StructTemplate)
		} else {
			return false, nil
		}
	case *pb.DataTemplate_ListTemplate:
		if vt2, ok := dt2.ValueTemplate.(*pb.DataTemplate_ListTemplate); ok {
			return chk.equivalentListTemplates(vt1.ListTemplate, vt2.ListTemplate)
		}
		return false, nil
	case *pb.DataTemplate_Value:
		if vt2, ok := dt2.ValueTemplate.(*pb.DataTemplate_Value); ok {
			return chk.equivalentData(vt1.Value, vt2.Value)
		}
		return false, nil
	case *pb.DataTemplate_Ref:
		if vt2, ok := dt2.ValueTemplate.(*pb.DataTemplate_Ref); ok {
			// Since we performed unrolling first, all remaining refs will only refer
			// to responses. Thus we can directly use proto.Equal to compare the refs.
			return proto.Equal(vt1.Ref, vt2.Ref), nil
		}
		return false, nil
	case *pb.DataTemplate_OptionalTemplate:
		if vt2, ok := dt2.ValueTemplate.(*pb.DataTemplate_OptionalTemplate); ok {
			return chk.equivalentDataTemplates(vt1.OptionalTemplate.ValueTemplate, vt2.OptionalTemplate.ValueTemplate)
		}
		return false, nil
	default:
		return false, errors.Errorf("unsupported value_template type %T", vt1)
	}
}

func (chk *equivChecker) equivalentStructTemplates(st1 *pb.StructTemplate, st2 *pb.StructTemplate) (bool, error) {
	for fieldName, ft1 := range st1.GetFieldTemplates() {
		if ft2, ok := st2.GetFieldTemplates()[fieldName]; ok {
			if eq, err := chk.equivalentDataTemplates(ft1, ft2); err != nil {
				return false, errors.Wrapf(err, "failed to compare struct template field %s", fieldName)
			} else if !eq {
				return false, nil
			}
		} else {
			return false, nil
		}
	}
	return true, nil
}

func (chk *equivChecker) equivalentListTemplates(lt1 *pb.ListTemplate, lt2 *pb.ListTemplate) (bool, error) {
	if len(lt1.GetElemTemplates()) != len(lt2.GetElemTemplates()) {
		return false, nil
	}
	for i, et1 := range lt1.GetElemTemplates() {
		if eq, err := chk.equivalentDataTemplates(et1, lt2.GetElemTemplates()[i]); err != nil {
			return false, errors.Wrapf(err, "failed to compare list template element %d", i)
		} else if !eq {
			return false, nil
		}
	}
	return true, nil
}

func (chk *equivChecker) equivalentData(d1 *pb.Data, d2 *pb.Data) (bool, error) {
	switch v1 := d1.Value.(type) {
	case *pb.Data_Primitive:
		if v2, ok := d2.Value.(*pb.Data_Primitive); ok {
			return chk.equivalentPrimitives(v1.Primitive, v2.Primitive)
		}
		return false, nil
	case *pb.Data_Struct:
		if v2, ok := d2.Value.(*pb.Data_Struct); ok {
			return chk.equivalentStructs(v1.Struct, v2.Struct)
		}
		return false, nil
	case *pb.Data_List:
		if v2, ok := d2.Value.(*pb.Data_List); ok {
			return chk.equivalentLists(v1.List, v2.List)
		}
		return false, nil
	case *pb.Data_Optional:
		if v2, ok := d2.Value.(*pb.Data_Optional); ok {
			return chk.equivalentOptionals(v1.Optional, v2.Optional)
		}
		return false, nil
	default:
		return false, errors.Errorf("unsupported data type %T", v1)
	}
	return false, nil
}

func (chk *equivChecker) equivalentPrimitives(p1 *pb.Primitive, p2 *pb.Primitive) (bool, error) {
	if p1.GetTypeHint() != p2.GetTypeHint() {
		return false, nil
	}

	isFixed := containsFixedValue(p1)
	if isFixed != containsFixedValue(p2) {
		return false, nil
	}
	if isFixed {
		return proto.Equal(p1, p2), nil
	} else {
		return chk.isBijective(p1, p2)
	}
}

func (chk *equivChecker) isBijective(p1 *pb.Primitive, p2 *pb.Primitive) (bool, error) {
	hash1, err := hashPrimitive(p1)
	if err != nil {
		return false, err
	}
	hash2, err := hashPrimitive(p2)
	if err != nil {
		return false, err
	}

	checkOrInsert := func(m map[string]*pb.Primitive, key string, val *pb.Primitive) bool {
		if v, ok := m[key]; ok {
			return proto.Equal(v, val)
		} else {
			m[key] = val
			return true
		}
	}
	return checkOrInsert(chk.lToRPrimitiveMap, hash1, p2) && checkOrInsert(chk.rToLPrimitiveMap, hash2, p1), nil
}

func (chk *equivChecker) equivalentStructs(s1 *pb.Struct, s2 *pb.Struct) (bool, error) {
	for fieldName, f1 := range s1.GetFields() {
		if f2, ok := s2.GetFields()[fieldName]; ok {
			if eq, err := chk.equivalentData(f1, f2); err != nil {
				return false, errors.Wrapf(err, "failed to compare struct field %s", fieldName)
			} else if !eq {
				return false, nil
			}
		} else {
			return false, nil
		}
	}
	return true, nil
}

func (chk *equivChecker) equivalentLists(l1 *pb.List, l2 *pb.List) (bool, error) {
	if len(l1.GetElems()) != len(l2.GetElems()) {
		return false, nil
	}
	for i, e1 := range l1.GetElems() {
		if eq, err := chk.equivalentData(e1, l2.GetElems()[i]); err != nil {
			return false, errors.Wrapf(err, "failed to compare list elem %d", i)
		} else if !eq {
			return false, nil
		}
	}
	return true, nil
}

func (chk *equivChecker) equivalentOptionals(o1 *pb.Optional, o2 *pb.Optional) (bool, error) {
	switch v1 := o1.Value.(type) {
	case *pb.Optional_Data:
		if v2, ok := o2.Value.(*pb.Optional_Data); ok {
			return chk.equivalentData(v1.Data, v2.Data)
		}
		return false, nil
	case *pb.Optional_None:
		_, ok := o2.Value.(*pb.Optional_None)
		return ok, nil
	default:
		return false, errors.Errorf("unsupported optional value type %T", v1)
	}
}

func containsFixedValue(p *pb.Primitive) bool {
	switch pv := p.Value.(type) {
	case *pb.Primitive_BoolValue:
		// We always consider bool as fixed because it's inherently limited to 2
		// values.
		return true
	case *pb.Primitive_BytesValue:
		v := pv.BytesValue.GetValue()
		for _, fv := range pv.BytesValue.GetType().GetFixedValues() {
			if bytes.Equal(fv, v) {
				return true
			}
		}
	case *pb.Primitive_StringValue:
		v := pv.StringValue.GetValue()
		for _, fv := range pv.StringValue.GetType().GetFixedValues() {
			if fv == v {
				return true
			}
		}
	case *pb.Primitive_Int32Value:
		v := pv.Int32Value.GetValue()
		for _, fv := range pv.Int32Value.GetType().GetFixedValues() {
			if fv == v {
				return true
			}
		}
	case *pb.Primitive_Int64Value:
		v := pv.Int64Value.GetValue()
		for _, fv := range pv.Int64Value.GetType().GetFixedValues() {
			if fv == v {
				return true
			}
		}
	case *pb.Primitive_Uint32Value:
		v := pv.Uint32Value.GetValue()
		for _, fv := range pv.Uint32Value.GetType().GetFixedValues() {
			if fv == v {
				return true
			}
		}
	case *pb.Primitive_Uint64Value:
		v := pv.Uint64Value.GetValue()
		for _, fv := range pv.Uint64Value.GetType().GetFixedValues() {
			if fv == v {
				return true
			}
		}
	case *pb.Primitive_DoubleValue:
		v := pv.DoubleValue.GetValue()
		for _, fv := range pv.DoubleValue.GetType().GetFixedValues() {
			if fv == v {
				return true
			}
		}
	case *pb.Primitive_FloatValue:
		v := pv.FloatValue.GetValue()
		for _, fv := range pv.FloatValue.GetType().GetFixedValues() {
			if fv == v {
				return true
			}
		}
	}
	return false
}

func hashPrimitive(p *pb.Primitive) (string, error) {
	// Use FNV1-a as the hash function since it's faster and we don't need
	// cryptographically secure hash.
	protoHasher := protohash.NewHasher(protohash.BasicHashFunction(protohash.FNV1A_128))
	if protoBytes, err := protoHasher.HashProto(p); err != nil {
		return "", errors.Wrap(err, "hashPrimitive failed")
	} else {
		return base64.StdEncoding.EncodeToString(protoBytes), nil
	}
}

// If DataTemplate is a reference, resolves it into one of
//  - template
//  - contant value
//  - ref to response
// Note that there could still be references nested inside templates (i.e.
// StructTemplate).
// This function is intended to be a helper function for checking data
// templates for equivalence, hence the preservation of response refs.
func unrollDataTemplate(prefix []*pb.MethodTemplate, dt *pb.DataTemplate) (*pb.DataTemplate, error) {
	switch vt := dt.ValueTemplate.(type) {
	case *pb.DataTemplate_Ref:
		return unrollMethodDataRef(prefix, dt, vt.Ref)
	default:
		return dt, nil
	}
}

func unrollMethodDataRef(prefix []*pb.MethodTemplate, dt *pb.DataTemplate, r *pb.MethodDataRef) (*pb.DataTemplate, error) {
	if r.GetMethodIndex() < 0 || r.GetMethodIndex() >= int32(len(prefix)) {
		return nil, errors.Errorf("unrollMethodDataRef index of out range index=%d len=%d", r.GetMethodIndex(), len(prefix))
	}

	d := prefix[r.GetMethodIndex()]
	switch ref := r.Ref.(type) {
	case *pb.MethodDataRef_ArgRef:
		namedRef := ref.ArgRef
		if arg, ok := d.GetArgTemplates()[namedRef.GetKey()]; ok {
			if t, err := unrollDataRef(prefix, arg, namedRef.GetDataRef()); err != nil {
				return nil, errors.Wrapf(err, "failed to resolve reference to arg %s", namedRef.GetKey())
			} else {
				return t, nil
			}
		} else {
			return nil, errors.Errorf("no such argument %s", namedRef.GetKey())
		}
	case *pb.MethodDataRef_ResponseRef:
		// Return DataTemplate directly to avoid memory allocation.
		return dt, nil
	default:
		return nil, errors.Errorf("unsuppported MethodDataRef type %T", ref)
	}
}

func unrollDataRef(prefix []*pb.MethodTemplate, dt *pb.DataTemplate, ref *pb.DataRef) (*pb.DataTemplate, error) {
	switch v := dt.ValueTemplate.(type) {
	case *pb.DataTemplate_Value:
		if data, err := GetDataRef(ref, v.Value); err != nil {
			return nil, errors.Wrapf(err, "failed to resolve reference into constant value")
		} else {
			// This extra memory allocation makes the interface nicer by only having
			// to worry about comparing DataTemplates. If this is too costly, we can
			// do more plumbing to return either a DataTemplate or a raw Data proto.
			return &pb.DataTemplate{
				ValueTemplate: &pb.DataTemplate_Value{data},
			}, nil
		}
	case *pb.DataTemplate_StructTemplate:
		if r, ok := ref.ValueRef.(*pb.DataRef_StructRef); ok {
			return unrollStructRef(prefix, dt, v.StructTemplate, r.StructRef)
		} else {
			return nil, errors.Errorf("got value_ref type %T for struct template", ref.ValueRef)
		}
	case *pb.DataTemplate_ListTemplate:
		if r, ok := ref.ValueRef.(*pb.DataRef_ListRef); ok {
			return unrollListRef(prefix, dt, v.ListTemplate, r.ListRef)
		} else {
			return nil, errors.Errorf("got value_ref type %T for list template", ref.ValueRef)
		}
	case *pb.DataTemplate_OptionalTemplate:
		return unrollDataRef(prefix, v.OptionalTemplate.ValueTemplate, ref)
	case *pb.DataTemplate_Ref:
		return unrollMethodDataRef(prefix, dt, v.Ref)
	default:
		return nil, errors.Errorf("unsupported value_template type %T", v)
	}
}

func unrollStructRef(prefix []*pb.MethodTemplate, dt *pb.DataTemplate, s *pb.StructTemplate, ref *pb.StructRef) (*pb.DataTemplate, error) {
	switch r := ref.Ref.(type) {
	case *pb.StructRef_FullStruct:
		return dt, nil
	case *pb.StructRef_FieldRef:
		if field, ok := s.GetFieldTemplates()[r.FieldRef.GetKey()]; ok {
			return unrollDataRef(prefix, field, r.FieldRef.GetDataRef())
		} else {
			return nil, errors.Errorf("StructTemplate does not contain field %s", r.FieldRef.GetKey())
		}
	default:
		return nil, errors.Errorf("unsupported StructRef type %T", r)
	}
}

func unrollListRef(prefix []*pb.MethodTemplate, dt *pb.DataTemplate, l *pb.ListTemplate, ref *pb.ListRef) (*pb.DataTemplate, error) {
	switch r := ref.Ref.(type) {
	case *pb.ListRef_FullList:
		return dt, nil
	case *pb.ListRef_ElemRef:
		if r.ElemRef.GetIndex() < 0 || r.ElemRef.GetIndex() >= int32(len(l.ElemTemplates)) {
			return nil, errors.Errorf("out of bounds on ListTemplate index=%d len=%d", r.ElemRef.GetIndex(), len(l.ElemTemplates))
		}
		return unrollDataRef(prefix, l.ElemTemplates[r.ElemRef.GetIndex()], r.ElemRef.GetDataRef())
	default:
		return nil, errors.Errorf("unsupported StructRef type %T", r)
	}
}
