package spec_util

import (
	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

func IsPrimitive(d *pb.Data) bool {
	switch v := d.Value.(type) {
	case *pb.Data_Primitive:
		return true
	case *pb.Data_Optional:
		switch ov := v.Optional.Value.(type) {
		case *pb.Optional_Data:
			return IsPrimitive(ov.Data)
		case *pb.Optional_None:
			return false
		}
	case *pb.Data_Oneof:
		for _, d := range v.Oneof.Options {
			if !IsPrimitive(d) {
				return false
			}
		}
		return true
	}
	return false
}

func IsPrimitiveList(d *pb.Data) bool {
	switch v := d.Value.(type) {
	case *pb.Data_List:
		for _, elem := range v.List.Elems {
			if !IsPrimitive(elem) {
				return false
			}
		}
		return true
	}
	return false
}
