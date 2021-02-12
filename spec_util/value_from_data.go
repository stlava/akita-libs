package spec_util

import (
	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

func ValueFromPrimitive(p *pb.Primitive) (interface{}, error) {
	switch v := p.GetValue().(type) {
	case *pb.Primitive_BoolValue:
		return v.BoolValue.Value, nil
	case *pb.Primitive_BytesValue:
		return v.BytesValue.Value, nil
	case *pb.Primitive_StringValue:
		return v.StringValue.Value, nil
	case *pb.Primitive_Int32Value:
		return v.Int32Value.Value, nil
	case *pb.Primitive_Int64Value:
		return v.Int64Value.Value, nil
	case *pb.Primitive_Uint32Value:
		return v.Uint32Value.Value, nil
	case *pb.Primitive_Uint64Value:
		return v.Uint64Value.Value, nil
	case *pb.Primitive_DoubleValue:
		return v.DoubleValue.Value, nil
	case *pb.Primitive_FloatValue:
		return v.FloatValue.Value, nil
	default:
		return nil, errors.Errorf("unsupported primitive type %T", v)
	}
}
