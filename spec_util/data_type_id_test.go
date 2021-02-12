package spec_util

import (
	"testing"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	wpb "github.com/golang/protobuf/ptypes/wrappers"
)

const (
	userType = `3c85b1d7-605c-46f9-b5e9-1b5e6b010181`
)

var (
	stringType = &pb.String{
		Type: &pb.StringType{
			Regex: "*",
		},
	}
	stringType2 = &pb.String{
		Type: &pb.StringType{
			FixedValues: []string{"foobar"},
		},
	}
)

// We should respect user_type field when create DataTypeID.
func TestTypeHint(t *testing.T) {
	d := &pb.Data{
		Value: &pb.Data_Primitive{
			&pb.Primitive{
				TypeHint: userType,
				Value:    &pb.Primitive_StringValue{stringType},
			},
		},
	}
	id := DataToTypeID(d)
	if id != DataTypeID(userType) {
		t.Errorf("expected %s, got %s", userType, id)
	}
}

func TestAkitaAnnotationsIgnored(t *testing.T) {
	d1 := &pb.Data{
		Value: &pb.Data_Primitive{
			&pb.Primitive{
				Value: &pb.Primitive_StringValue{stringType},
				AkitaAnnotations: &pb.AkitaAnnotations{
					IsFree: true,
				},
			},
		},
	}
	d2 := &pb.Data{
		Value: &pb.Data_Primitive{
			&pb.Primitive{
				Value: &pb.Primitive_StringValue{stringType},
				AkitaAnnotations: &pb.AkitaAnnotations{
					IsFree:      false,
					IsSensitive: true,
				},
			},
		},
	}

	id1 := DataToTypeID(d1)
	id2 := DataToTypeID(d2)
	if id1 == invalidDataID {
		t.Errorf("got invalid data ID")
	} else if id1 != id2 {
		t.Errorf("expected same DataTypeID, got different ones")
	}
}

func TestDifferentTypes(t *testing.T) {
	d1 := &pb.Data{
		Value: &pb.Data_Primitive{
			&pb.Primitive{
				Value: &pb.Primitive_StringValue{stringType},
			},
		},
	}
	d2 := &pb.Data{
		Value: &pb.Data_Primitive{
			&pb.Primitive{
				Value: &pb.Primitive_StringValue{stringType2},
			},
		},
	}

	if DataToTypeID(d1) == DataToTypeID(d2) {
		t.Errorf("expected different type IDs, got the same")
	}
}

func TestInt32MinMax(t *testing.T) {
	d := &pb.Data{
		Value: &pb.Data_Primitive{
			&pb.Primitive{
				Value: &pb.Primitive_Int32Value{
					&pb.Int32{
						Type: &pb.Int32Type{
							Min: &wpb.Int32Value{
								Value: 1,
							},
							Max: &wpb.Int32Value{
								Value: 100,
							},
						},
					},
				},
			},
		},
	}
	id := DataToTypeID(d)
	if id == invalidDataID {
		t.Error("Expected valid data ID for int32 with min and max, got invalidDataID")
	}
}
