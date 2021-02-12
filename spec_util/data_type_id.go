package spec_util

import (
	"encoding/base64"

	protohash "github.com/akitasoftware/objecthash-proto"
	"github.com/golang/glog"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

// Identifies the type of a protobuf Data message.
type DataTypeID string

const (
	invalidDataID = DataTypeID("invalid-data-type-id")
)

func DataToTypeID(d *pb.Data) DataTypeID {
	switch x := d.Value.(type) {
	case *pb.Data_Primitive:
		if x.Primitive.TypeHint != "" {
			// If a user type is provided, we use the user type as the ID for
			// this Data protobuf. If the API spec mistakenly assigns the same
			// user type to Data protobufs representing different types, the
			// sequence generator may generate a sequence with values that don't
			// match the arg types. The worker should still be able to
			// instantiate the sequence and run it but should see an error.
			return DataTypeID(x.Primitive.TypeHint)
		}
	}
	protoHasher := protohash.NewHasher(
		protohash.BasicHashFunction(protohash.FNV1A_128),
		protohash.IgnoreFieldName("akita_annotations"),
	)
	dataBytes, err := protoHasher.HashProto(d)
	if err != nil {
		glog.Errorf("Failed to convert Data to ID: %v", err)
		return invalidDataID
	}
	return DataTypeID(base64.URLEncoding.EncodeToString(dataBytes))
}
