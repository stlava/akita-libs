package spec_util

import (
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

// A "view" into RefMap that keeps track of the prefix MethodTemplates in order
// to avoid returning references that are not useful. A reference is defined as
// "unuseful" if it:
//   - Refers to yet another reference (we will also return the other
//     reference, so the new reference is redundant). RefMap may return
//     references to references in response values, so we should filter these
//     out.
//   - Refers to a skipped optional field
// This cuts down on the number of possible sequences we can generate for a
// given data type.
type RefMapView struct {
	m      RefMap
	prefix []*pb.MethodTemplate
}

func NewRefMapView(m RefMap, prefix []*pb.MethodTemplate) *RefMapView {
	return &RefMapView{
		m:      m,
		prefix: prefix,
	}
}

func (v *RefMapView) Copy() *RefMapView {
	n := &RefMapView{
		m:      v.m,
		prefix: make([]*pb.MethodTemplate, len(v.prefix)),
	}
	for i, t := range v.prefix {
		n.prefix[i] = t
	}
	return n
}

func (v *RefMapView) HasRefs(dID DataTypeID) bool {
	return v.m.HasRefs(dID)
}

func (v *RefMapView) GetDataRefs(dID DataTypeID, aa *pb.AkitaAnnotations) []*pb.MethodDataRef {
	refs := make([]*pb.MethodDataRef, 0)
	for _, ref := range v.m.GetDataRefs(dID, aa) {
		if isUsefulMethodDataRef(v.prefix[ref.GetMethodIndex()], ref) {
			refs = append(refs, ref)
		}
	}
	return refs
}

// Returns the set of data specs that can be filled using the data that's
// currently fillable. In most cases, if the arg is fillable, the returned
// result is identical. However, if the arg contains oneof fields, all fillable
// options are expanded and returned as separate results, where each result no
// longer has oneof fields.
func (v *RefMapView) GetFillableArgs(arg *pb.Data) []*pb.Data {
	// Note: we always consider arguments of prefix methods to be valid refs for
	// filling args. This is useful in scenarios such as setting a password on a
	// file and later using that password to read the file. The API won't echo
	// back your password in the result.
	if v.HasRefs(DataToTypeID(arg)) {
		return []*pb.Data{arg}
	}

	switch x := arg.Value.(type) {
	case *pb.Data_Primitive:
		if x.Primitive.GetAkitaAnnotations().GetIsFree() {
			return []*pb.Data{arg}
		}
		var fixedValuesLen int
		switch y := x.Primitive.Value.(type) {
		case *pb.Primitive_BoolValue:
			fixedValuesLen = len(y.BoolValue.GetType().GetFixedValues())
		case *pb.Primitive_BytesValue:
			fixedValuesLen = len(y.BytesValue.GetType().GetFixedValues())
		case *pb.Primitive_StringValue:
			fixedValuesLen = len(y.StringValue.GetType().GetFixedValues())
		case *pb.Primitive_Int32Value:
			fixedValuesLen = len(y.Int32Value.GetType().GetFixedValues())
		case *pb.Primitive_Int64Value:
			fixedValuesLen = len(y.Int64Value.GetType().GetFixedValues())
		case *pb.Primitive_Uint32Value:
			fixedValuesLen = len(y.Uint32Value.GetType().GetFixedValues())
		case *pb.Primitive_Uint64Value:
			fixedValuesLen = len(y.Uint64Value.GetType().GetFixedValues())
		case *pb.Primitive_DoubleValue:
			fixedValuesLen = len(y.DoubleValue.GetType().GetFixedValues())
		case *pb.Primitive_FloatValue:
			fixedValuesLen = len(y.FloatValue.GetType().GetFixedValues())
		}
		// Normally, having fixed values does not imply that the arg is fillable.
		// However, if there is only one fixed value, then the arg is effectively
		// free.
		if fixedValuesLen == 1 {
			return []*pb.Data{arg}
		}
		return nil
	case *pb.Data_Struct:
		// Map key to list of fillable data for the key.
		alts := make(map[string][]*pb.Data, len(x.Struct.Fields))
		for k, field := range x.Struct.Fields {
			f := v.GetFillableArgs(field)
			if f == nil {
				return nil
			}
			alts[k] = f
		}

		fillableArgs := FlattenAlternatives(alts)
		results := make([]*pb.Data, 0, len(fillableArgs))
		for _, fields := range fillableArgs {
			results = append(results, &pb.Data{
				Value: &pb.Data_Struct{&pb.Struct{Fields: fields}},
				Meta:  arg.Meta,
			})
		}
		return results
	case *pb.Data_List:
		if len(x.List.Elems) > 0 {
			results := []*pb.Data{}
			for _, f := range v.GetFillableArgs(x.List.Elems[0]) {
				results = append(results, &pb.Data{
					Value: &pb.Data_List{&pb.List{Elems: []*pb.Data{f}}},
					Meta:  arg.Meta,
				})
			}
			return results
		}
		glog.Errorf("Did not expect nil list %s", proto.MarshalTextString(arg))
		return nil
	case *pb.Data_Optional:
		return []*pb.Data{arg}
	case *pb.Data_Oneof:
		results := []*pb.Data{}
		for _, option := range x.Oneof.Options {
			results = append(results, v.GetFillableArgs(option)...)
		}
		for _, r := range results {
			r.Meta = MergeOneOfMeta(arg.Meta, r.Meta)
		}
		return results
	default:
		glog.Errorf("Unsupported data type in GetFillableArgs: %T", x)
		return nil
	}
}

func isUsefulMethodDataRef(mt *pb.MethodTemplate, ref *pb.MethodDataRef) bool {
	switch r := ref.Ref.(type) {
	case *pb.MethodDataRef_ArgRef:
		if arg, ok := mt.GetArgTemplates()[r.ArgRef.GetKey()]; ok {
			return isUsefulDataRef(arg, r.ArgRef.GetDataRef())
		} else {
			// The argument might be missing because it's optional and skipped.
			return false
		}
	case *pb.MethodDataRef_ResponseRef:
		// We don't know the response values at sequence generation time, and it's
		// impossible to have references in response, so this reference is always
		// useful.
		return true
	default:
		glog.Errorf("unrecognized ref type %T, default to useful ref", r)
		return true
	}
}

func isUsefulDataRef(d *pb.DataTemplate, ref *pb.DataRef) bool {
	switch x := d.ValueTemplate.(type) {
	case *pb.DataTemplate_StructTemplate:
		if r, ok := ref.ValueRef.(*pb.DataRef_StructRef); ok {
			return isUsefulStructRef(x.StructTemplate, r.StructRef)
		} else {
			glog.Errorf("got value_ref type %T for struct template, assuming non-useful ref", ref.ValueRef)
			return false
		}
	case *pb.DataTemplate_ListTemplate:
		if r, ok := ref.ValueRef.(*pb.DataRef_ListRef); ok {
			return isUsefulListRef(x.ListTemplate, r.ListRef)
		} else {
			glog.Errorf("got value_ref type %T for list template, assuming non-useful ref", ref.ValueRef)
			return false
		}
	case *pb.DataTemplate_Value:
		if resolved, err := GetDataRef(ref, x.Value); err != nil {
			glog.Errorf("GetDataRef failed, default to non-useful ref: %v", err)
			return false
		} else {
			if opt, isOptional := resolved.Value.(*pb.Data_Optional); isOptional {
				if _, isNone := opt.Optional.Value.(*pb.Optional_None); isNone {
					return false
				}
			}
		}
		return true
	case *pb.DataTemplate_Ref:
		return false
	case *pb.DataTemplate_OptionalTemplate:
		return isUsefulDataRef(x.OptionalTemplate.GetValueTemplate(), ref)
	default:
		glog.Errorf("unrecognized value_template type %T, defaulting to useful ref", x)
		return true
	}
}

func isUsefulStructRef(s *pb.StructTemplate, ref *pb.StructRef) bool {
	switch r := ref.Ref.(type) {
	case *pb.StructRef_FullStruct:
		// TODO: A full struct is only useful if all fields don't contain any refs.
		// Current implementation is a conservative estimate.
		return true
	case *pb.StructRef_FieldRef:
		if field, ok := s.GetFieldTemplates()[r.FieldRef.GetKey()]; ok {
			return isUsefulDataRef(field, r.FieldRef.GetDataRef())
		} else {
			// The struct field might be missing because it's optional and skipped.
			return false
		}
	default:
		glog.Errorf("unsupported StructRef type %T, default to useful ref", r)
		return true
	}
}

func isUsefulListRef(l *pb.ListTemplate, ref *pb.ListRef) bool {
	switch r := ref.Ref.(type) {
	case *pb.ListRef_FullList:
		// TODO: A full list is only useful if all elems don't contain any refs.
		// Current implementation is a conservative estimate.
		return true
	case *pb.ListRef_ElemRef:
		if r.ElemRef.GetIndex() < 0 || r.ElemRef.GetIndex() >= int32(len(l.ElemTemplates)) {
			// The list elem might be missing because it's optional and skipped.
			return false
		}
		return isUsefulDataRef(l.ElemTemplates[r.ElemRef.GetIndex()], r.ElemRef.GetDataRef())
	default:
		glog.Errorf("unsupported ListRef type %T, default to useful ref", r)
		return true
	}
}
