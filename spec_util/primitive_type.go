package spec_util

import pb "github.com/akitasoftware/akita-ir/go/api_spec"

func TypeOfPrimitive(p *pb.Primitive) string {
	if p.GetBoolValue() != nil {
		return "bool"
	} else if p.GetBytesValue() != nil {
		return "bytes"
	} else if p.GetDoubleValue() != nil {
		return "double"
	} else if p.GetFloatValue() != nil {
		return "float"
	} else if p.GetInt32Value() != nil {
		return "int32"
	} else if p.GetInt64Value() != nil {
		return "int64"
	} else if p.GetStringValue() != nil {
		return "string"
	} else if p.GetUint32Value() != nil {
		return "uint32"
	} else if p.GetUint64Value() != nil {
		return "uint64"
	}
	return ""
}
