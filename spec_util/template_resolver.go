package spec_util

import (
	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

// Extracts a constant value from the given template. Does not recursively
// resolve references. Use in conjunction with PropagateConstants to handle
// recursive references, though the sequence generator should never generate
// recursive references in the first place since that is wasteful.
// Treats none value as an error.
func ExtractValueFromTemplate(t *pb.DataTemplate, ref *pb.DataRef) (*pb.Data, error) {
	switch vt := t.GetValueTemplate().(type) {
	case *pb.DataTemplate_StructTemplate:
		if r, ok := ref.ValueRef.(*pb.DataRef_StructRef); !ok {
			return nil, errors.Errorf("struct template got reference type %T", ref.ValueRef)
		} else {
			return extractValueFromStructTemplate(vt.StructTemplate, r.StructRef)
		}
	case *pb.DataTemplate_ListTemplate:
		if r, ok := ref.ValueRef.(*pb.DataRef_ListRef); !ok {
			return nil, errors.Errorf("list template got reference type %T", ref.ValueRef)
		} else {
			return extractValueFromListTemplate(vt.ListTemplate, r.ListRef)
		}
	case *pb.DataTemplate_Value:
		return GetDataRef(ref, vt.Value)
	case *pb.DataTemplate_Ref:
		return nil, errors.Errorf("recursive references not supported")
	case *pb.DataTemplate_OptionalTemplate:
		if v, err := ExtractValueFromTemplate(vt.OptionalTemplate.ValueTemplate, ref); err != nil {
			return nil, errors.Wrap(err, "failed to extract value from optional template")
		} else {
			return v, nil
		}
	default:
		return nil, errors.Errorf("unsupported value template type %T", vt)
	}
}

func extractValueFromStructTemplate(st *pb.StructTemplate, structRef *pb.StructRef) (*pb.Data, error) {
	switch ref := structRef.GetRef().(type) {
	case *pb.StructRef_FullStruct:
		c := &constantPropagator{}
		if constant, err := c.structRefToConst(fullStructRef, st); err != nil {
			return nil, errors.Wrap(err, "full struct ref points to non-const struct")
		} else {
			return constant, nil
		}
	case *pb.StructRef_FieldRef:
		k := ref.FieldRef.GetKey()
		if fieldTemplate, hasField := st.GetFieldTemplates()[k]; !hasField {
			return nil, errors.Errorf("struct template does not have field %s", k)
		} else if v, err := ExtractValueFromTemplate(fieldTemplate, ref.FieldRef.DataRef); err != nil {
			return nil, errors.Wrapf(err, "failed to extract value from struct template key %s", k)
		} else {
			return v, nil
		}
	default:
		return nil, errors.Errorf("unsupported struct ref type %T", ref)
	}
}

func extractValueFromListTemplate(lt *pb.ListTemplate, listRef *pb.ListRef) (*pb.Data, error) {
	switch ref := listRef.GetRef().(type) {
	case *pb.ListRef_FullList:
		c := &constantPropagator{}
		if constant, err := c.listRefToConst(fullListRef, lt); err != nil {
			return nil, errors.Wrap(err, "full list ref points to non-const list")
		} else {
			return constant, nil
		}
	case *pb.ListRef_ElemRef:
		index := ref.ElemRef.GetIndex()
		if index < 0 || index >= int32(len(lt.GetElemTemplates())) {
			return nil, errors.Errorf("list elem ref index out of bounds index=%d len=%d", index, len(lt.GetElemTemplates()))
		}

		if v, err := ExtractValueFromTemplate(lt.GetElemTemplates()[index], ref.ElemRef.GetDataRef()); err != nil {
			return nil, errors.Wrapf(err, "failed to extract value from list template index %d", index)
		} else {
			return v, nil
		}
	default:
		return nil, errors.Errorf("unsupported list ref type %T", ref)
	}
}
