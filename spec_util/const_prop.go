package spec_util

import (
	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

var (
	fullStructRef = &pb.StructRef{
		Ref: &pb.StructRef_FullStruct{&pb.StructRef_FullStructRef{}},
	}
	fullListRef = &pb.ListRef{
		Ref: &pb.ListRef_FullList{&pb.ListRef_FullListRef{}},
	}
	NoneData = &pb.Data{
		Value: &pb.Data_Optional{
			Optional: &pb.Optional{
				Value: &pb.Optional_None{
					None: &pb.None{},
				},
			},
		},
	}
)

// Propagates constants within prefix that are referred to by references in t.
// Mutates t and may mutate elements in prefix that are references to constant
// values themselves.
func PropagateConstants(prefix []*pb.MethodTemplate, t *pb.MethodTemplate) error {
	for argName, dataTemplate := range t.GetArgTemplates() {
		if err := propagateConstantsInDataTemplate(prefix, dataTemplate); err != nil {
			return errors.Wrapf(err, "failed to propagate constants to top-level arg %s", argName)
		}
	}
	return nil
}

func propagateConstantsInDataTemplate(prefix []*pb.MethodTemplate, dt *pb.DataTemplate) error {
	c := &constantPropagator{
		prefix: prefix,
	}
	_, err := c.modifyDataTemplateWithConst(dt)
	return err
}

type constantPropagator struct {
	prefix []*pb.MethodTemplate
}

// Modifies DataTemplate to replace references with constants. Returns true if
// the DataTemplate is a constant or has been converted to a constant.
func (c *constantPropagator) modifyDataTemplateWithConst(dt *pb.DataTemplate) (bool, error) {
	makeConst := func(constant *pb.Data) bool {
		if constant != nil {
			dt.ValueTemplate = &pb.DataTemplate_Value{
				Value: constant,
			}
			return true
		}
		return false
	}

	switch vt := dt.GetValueTemplate().(type) {
	case *pb.DataTemplate_StructTemplate:
		allFieldsConstant := true
		for _, fieldTemplate := range vt.StructTemplate.GetFieldTemplates() {
			if isConstant, err := c.modifyDataTemplateWithConst(fieldTemplate); err != nil {
				return false, err
			} else {
				allFieldsConstant = allFieldsConstant && isConstant
			}
		}
		if allFieldsConstant {
			// Convert this StructTemplate into a Data constant.
			constant, _ := c.structRefToConst(fullStructRef, vt.StructTemplate)
			return makeConst(constant), nil
		}
		return false, nil
	case *pb.DataTemplate_ListTemplate:
		allElemsConstant := true
		for _, elemTemplate := range vt.ListTemplate.GetElemTemplates() {
			if isConstant, err := c.modifyDataTemplateWithConst(elemTemplate); err != nil {
				return false, err
			} else {
				allElemsConstant = allElemsConstant && isConstant
			}
		}
		if allElemsConstant {
			// Convert this ListTemplate into a Data constant.
			constant, _ := c.listRefToConst(fullListRef, vt.ListTemplate)
			return makeConst(constant), nil
		}
		return false, nil
	case *pb.DataTemplate_Value:
		// Already a constant.
		return true, nil
	case *pb.DataTemplate_Ref:
		if constant, err := c.methodDataRefToConst(vt.Ref); err == nil {
			AddAnnotationToData(constant, vt.Ref.AkitaAnnotations)
			return makeConst(constant), nil
		} else {
			return false, errors.Wrap(err, "failed to fill DataTemplate with constant")
		}
	case *pb.DataTemplate_OptionalTemplate:
		isConstant, err := c.modifyDataTemplateWithConst(vt.OptionalTemplate.ValueTemplate)
		if err != nil {
			return makeConst(NoneData), nil
		}
		return isConstant, nil
	default:
		return false, errors.Errorf("Unknown ValueTemplate in DataTemplate of type %T", dt.GetValueTemplate())
	}
}

// May return nil *pb.Data if constant is not found.
func (c *constantPropagator) methodDataRefToConst(ref *pb.MethodDataRef) (*pb.Data, error) {
	if ref.MethodIndex < 0 || ref.MethodIndex >= int32(len(c.prefix)) {
		return nil, errors.Errorf("method index out of range (index=%d len=%d)", ref.MethodIndex, len(c.prefix))
	}
	t := c.prefix[ref.MethodIndex]

	switch r := ref.Ref.(type) {
	case *pb.MethodDataRef_ArgRef:
		if dt, ok := t.ArgTemplates[r.ArgRef.Key]; ok {
			if constant, err := c.dataRefToConst(r.ArgRef.DataRef, dt); err == nil {
				return constant, nil
			} else {
				return nil, errors.Wrap(err, "failed to convert DataRef to constant")
			}
		} else {
			return nil, errors.Errorf("reference non-existent arg %s", r.ArgRef.Key)
		}
	default:
		// We can't have constant in the response because we're dealing with
		// templates (i.e. before API calls are made).
		return nil, nil
	}
}

func (c *constantPropagator) dataRefToConst(ref *pb.DataRef, dt *pb.DataTemplate) (*pb.Data, error) {
	d, st, lt, err := c.extractValueTemplates(dt)
	if err != nil {
		return nil, errors.Wrap(err, "extractValueTemplates in dataRefToConst failed")
	}
	if d == nil {
		// Try to get constants inside templates.
		switch r := ref.GetValueRef().(type) {
		case *pb.DataRef_StructRef:
			if st == nil {
				return nil, errors.Errorf("no StructTemplate to match StructRef")
			}
			return c.structRefToConst(r.StructRef, st)
		case *pb.DataRef_ListRef:
			if lt == nil {
				return nil, errors.Errorf("no ListTemplate to match ListRef")
			}
			return c.listRefToConst(r.ListRef, lt)
		default:
			// No constant in this DataTemplate.
			return nil, nil
		}
	} else {
		// Already a constant. Extract the relevant bit of the constant.
		if resolved, err := GetDataRef(ref, d); err != nil {
			return nil, errors.Wrap(err, "GetDataRef for constant failed")
		} else {
			return resolved, nil
		}
	}
}

func (c *constantPropagator) structRefToConst(ref *pb.StructRef, st *pb.StructTemplate) (*pb.Data, error) {
	fieldTemplates := st.GetFieldTemplates()
	switch x := ref.Ref.(type) {
	case *pb.StructRef_FullStruct:
		constantFields := make(map[string]*pb.Data, len(fieldTemplates))
		for fieldName, fieldTemplate := range fieldTemplates {
			d, _, _, err := c.extractValueTemplates(fieldTemplate)
			if err != nil {
				return nil, errors.Wrap(err, "extractValueTemplates in structRefToConst failed")
			} else if d == nil {
				// The full struct contains non-constant values.
				return nil, nil
			} else {
				constantFields[fieldName] = d
			}
		}
		return &pb.Data{
			Value: &pb.Data_Struct{
				&pb.Struct{
					Fields: constantFields,
				},
			},
		}, nil
	case *pb.StructRef_FieldRef:
		fieldName := x.FieldRef.Key
		if fieldTemplate, ok := fieldTemplates[fieldName]; ok {
			return c.dataRefToConst(x.FieldRef.DataRef, fieldTemplate)
		} else {
			return nil, errors.Errorf("reference to missing struct field %s", fieldName)
		}
	}
	return nil, nil
}

func (c *constantPropagator) listRefToConst(ref *pb.ListRef, lt *pb.ListTemplate) (*pb.Data, error) {
	elemTemplates := lt.GetElemTemplates()
	switch x := ref.Ref.(type) {
	case *pb.ListRef_FullList:
		constantElems := make([]*pb.Data, len(elemTemplates))
		for i, elemTemplate := range elemTemplates {
			d, _, _, err := c.extractValueTemplates(elemTemplate)
			if err != nil {
				return nil, errors.Wrap(err, "extractValueTemplates in listRefToConst failed")
			} else if d == nil {
				// The full list contains non-constant values.
				return nil, nil
			} else {
				constantElems[i] = d
			}
		}
		return &pb.Data{
			Value: &pb.Data_List{
				&pb.List{
					Elems: constantElems,
				},
			},
		}, nil
	case *pb.ListRef_ElemRef:
		if x.ElemRef.Index < 0 || x.ElemRef.Index >= int32(len(elemTemplates)) {
			return nil, errors.Errorf("list ref elem out of range (index=%d len=%d)", x.ElemRef.Index, len(elemTemplates))
		}
		return c.dataRefToConst(x.ElemRef.DataRef, elemTemplates[x.ElemRef.Index])
	}
	return nil, nil
}

func (c *constantPropagator) extractValueTemplates(dt *pb.DataTemplate) (*pb.Data, *pb.StructTemplate, *pb.ListTemplate, error) {
	// Resolve nested references in DataTemplate to handle nested references (i.e.
	// A -> B -> C). This means that if this function returns nil for *pb.Data,
	// there is no constant being immediately referred to by the template.
	// TODO: This will modify data in c.prefix, but this is probably for the best
	// since it simplifies the prefix and saves us from having to do a copy.
	if _, err := c.modifyDataTemplateWithConst(dt); err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to convert nested refs in references DataTemplate to constant")
	}

	switch vt := dt.GetValueTemplate().(type) {
	case *pb.DataTemplate_StructTemplate:
		return nil, vt.StructTemplate, nil, nil
	case *pb.DataTemplate_ListTemplate:
		return nil, nil, vt.ListTemplate, nil
	case *pb.DataTemplate_Value:
		return vt.Value, nil, nil, nil
	}
	return nil, nil, nil, nil
}

func AddAnnotationToData(d *pb.Data, aa *pb.AkitaAnnotations) {
	// Annotations should only be provided for primitive data types
	if d != nil {
		if primitive, ok := d.Value.(*pb.Data_Primitive); ok {
			primitive.Primitive.AkitaAnnotations = aa
		}
	}
}
