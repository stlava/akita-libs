package spec_util

import (
	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

// GetDataRef resolves a reference to the value being referenced. It treats
// references to a None value as an error. The caller may decide to ignore the
// error if the references are made by optional fields.
// TODO: Add unit tests for GetDataRef. Currently, unit test coverage comes from
// instantiator_test in worker, which is where we extract this piece of code
// from.
func GetDataRef(r *pb.DataRef, d *pb.Data) (*pb.Data, error) {
	dp, ds, dl := extractValue(d)

	switch x := r.ValueRef.(type) {
	case *pb.DataRef_PrimitiveRef:
		if dp == nil {
			return nil, errors.Errorf("GetDataRef invalid ref to primitive, instead got %v", d)
		}
		return &pb.Data{
			Value: &pb.Data_Primitive{dp},
		}, nil
	case *pb.DataRef_StructRef:
		if ds == nil {
			return nil, errors.Errorf("GetDataRef invalid ref to struct, instead got %v", d)
		}
		return getStructRef(x.StructRef, ds)
	case *pb.DataRef_ListRef:
		if dl == nil {
			return nil, errors.Errorf("GetDataRef invalid ref to list, instead got %v", d)
		}
		return getListRef(x.ListRef, dl)
	default:
		return nil, errors.Errorf("GetDataRef unsupported ValueRef type %T", x)
	}
}

func getStructRef(r *pb.StructRef, s *pb.Struct) (*pb.Data, error) {
	switch x := r.Ref.(type) {
	case *pb.StructRef_FullStruct:
		return &pb.Data{
			Value: &pb.Data_Struct{s},
		}, nil
	case *pb.StructRef_FieldRef:
		fieldName := x.FieldRef.Key
		if field, ok := s.Fields[fieldName]; ok {
			return GetDataRef(x.FieldRef.DataRef, field)
		} else {
			return nil, errors.Errorf("getStructRef missing field %v", fieldName)
		}
	default:
		return nil, errors.Errorf("getStructRef unsupported ref type %T", x)
	}
}

func getListRef(r *pb.ListRef, l *pb.List) (*pb.Data, error) {
	switch x := r.Ref.(type) {
	case *pb.ListRef_FullList:
		return &pb.Data{
			Value: &pb.Data_List{l},
		}, nil
	case *pb.ListRef_ElemRef:
		if x.ElemRef.Index < 0 || x.ElemRef.Index >= int32(len(l.Elems)) {
			return nil, errors.Errorf("getListRef out of bounds: index=%v len=%v", x.ElemRef.Index, len(l.Elems))
		}
		return GetDataRef(x.ElemRef.DataRef, l.Elems[x.ElemRef.Index])
	default:
		return nil, errors.Errorf("getListRef unsupported ref type %T", x)
	}
}

// Helper function to extract Primitive, Struct, or List from Data. All three
// return values may be nil in the case of None, or one of the three will be
// non-nil.
func extractValue(d *pb.Data) (*pb.Primitive, *pb.Struct, *pb.List) {
	var dp *pb.Primitive
	var ds *pb.Struct
	var dl *pb.List

	switch x := d.Value.(type) {
	case *pb.Data_Primitive:
		dp = x.Primitive
	case *pb.Data_Struct:
		ds = x.Struct
	case *pb.Data_List:
		dl = x.List
	case *pb.Data_Optional:
		if y, ok := x.Optional.Value.(*pb.Optional_Data); ok {
			return extractValue(y.Data)
		}
	}

	return dp, ds, dl
}
