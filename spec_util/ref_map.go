package spec_util

import (
	"sort"

	"github.com/golang/protobuf/proto"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

type methodIDHash string

func hashMethodID(id *pb.MethodID) methodIDHash {
	// TODO: Hash the method ID proto to avoid name collision, but that is
	// expensive.
	return methodIDHash(id.GetName())
}

// A read-only map that stores references to each method's argument and
// response values. We store MethodDataRef, but since the AkitaAnnotations
// inside each ref should come from the Data protobuf making the reference,
// AkitaAnnotations is filled in only in GetDataRefs
type RefMap map[DataTypeID][]*pb.MethodDataRef

// Given a method ID and a data type ID, returns whether there are references to
// the method's arguments or responses with the data type.
func (m RefMap) HasRefs(dID DataTypeID) bool {
	if len(m[dID]) > 0 {
		return true
	}
	return false
}

// Create MethodDataRefs for all data of the given type in the args/responses of
// any previous method.
func (m RefMap) GetDataRefs(dID DataTypeID, aa *pb.AkitaAnnotations) []*pb.MethodDataRef {
	results := []*pb.MethodDataRef{}
	for _, r := range m[dID] {
		// Instantiate a MethodDataRef now that we know the
		// AkitaAnnotations.
		newRef := proto.Clone(r).(*pb.MethodDataRef)
		newRef.AkitaAnnotations = aa
		results = append(results, newRef)
	}
	return results
}

func NewRefMap(methods []*pb.Method) RefMap {
	result := make(RefMap)
	for i, method := range methods {
		addDataMap(result, i, method.Args, true)
		// Extract only successful responses since we don't want to reference
		// any failed responses.
		addDataMap(result, i, HTTPSuccessResponses(method), false)
	}
	return result
}

func addDataMap(m RefMap, methodIndex int, dataMap map[string]*pb.Data, isArg bool) {
	// Get stable order for dataMap
	keys := DataMapKeys(dataMap)
	for _, name := range keys {
		data := dataMap[name]
		refsWithID := dataToDataRefs(isArg, data)
		for _, refWithID := range refsWithID {
			elem := &pb.MethodDataRef{
				MethodIndex: int32(methodIndex),
			}
			if isArg {
				elem.Ref = &pb.MethodDataRef_ArgRef{
					ArgRef: &pb.NamedDataRef{
						Key:     name,
						DataRef: refWithID.ref,
					},
				}
			} else {
				elem.Ref = &pb.MethodDataRef_ResponseRef{
					ResponseRef: &pb.NamedDataRef{
						Key:     name,
						DataRef: refWithID.ref,
					},
				}
			}

			dataID := refWithID.dID
			if _, ok := m[dataID]; ok {
				m[dataID] = append(m[dataID], elem)
			} else {
				m[dataID] = []*pb.MethodDataRef{elem}
			}
		}
	}
}

type dataRefWithTypeID struct {
	dID DataTypeID
	ref *pb.DataRef
}

func dataToDataRefs(isArg bool, d *pb.Data) []*dataRefWithTypeID {
	switch x := d.Value.(type) {
	case *pb.Data_Primitive:
		if isArg && !x.Primitive.GetAkitaAnnotations().GetIsFree() {
			// Don't generate references to non-free arguments.
			// Since we don't generate random values for non-free args, any non-free
			// arg must be filled in by a reference to a free arg or a response value.
			// Thus, we'll never need to reference the non-free arg itself - we can
			// just reference the source of its value.
			return nil
		}

		fixedValuesLen := 0
		primitiveRef := &pb.PrimitiveRef{}
		switch p := x.Primitive.Value.(type) {
		case *pb.Primitive_BoolValue:
			fixedValuesLen = len(p.BoolValue.GetType().GetFixedValues())
			primitiveRef.Type = &pb.PrimitiveRef_BoolType{p.BoolValue.GetType()}
		case *pb.Primitive_BytesValue:
			fixedValuesLen = len(p.BytesValue.GetType().GetFixedValues())
			primitiveRef.Type = &pb.PrimitiveRef_BytesType{p.BytesValue.GetType()}
		case *pb.Primitive_StringValue:
			fixedValuesLen = len(p.StringValue.GetType().GetFixedValues())
			primitiveRef.Type = &pb.PrimitiveRef_StringType{p.StringValue.GetType()}
		case *pb.Primitive_Int32Value:
			fixedValuesLen = len(p.Int32Value.GetType().GetFixedValues())
			primitiveRef.Type = &pb.PrimitiveRef_Int32Type{p.Int32Value.GetType()}
		case *pb.Primitive_Int64Value:
			fixedValuesLen = len(p.Int64Value.GetType().GetFixedValues())
			primitiveRef.Type = &pb.PrimitiveRef_Int64Type{p.Int64Value.GetType()}
		case *pb.Primitive_Uint32Value:
			fixedValuesLen = len(p.Uint32Value.GetType().GetFixedValues())
			primitiveRef.Type = &pb.PrimitiveRef_Uint32Type{p.Uint32Value.GetType()}
		case *pb.Primitive_Uint64Value:
			fixedValuesLen = len(p.Uint64Value.GetType().GetFixedValues())
			primitiveRef.Type = &pb.PrimitiveRef_Uint64Type{p.Uint64Value.GetType()}
		case *pb.Primitive_DoubleValue:
			fixedValuesLen = len(p.DoubleValue.GetType().GetFixedValues())
			primitiveRef.Type = &pb.PrimitiveRef_DoubleType{p.DoubleValue.GetType()}
		case *pb.Primitive_FloatValue:
			fixedValuesLen = len(p.FloatValue.GetType().GetFixedValues())
			primitiveRef.Type = &pb.PrimitiveRef_FloatType{p.FloatValue.GetType()}
		}

		if fixedValuesLen == 1 {
			// If there is only one fixed value, don't generate a reference since we
			// can always generate the one fixed value instead of referring to same
			// value in a previous arg/response. This cuts down on the number of
			// possible templates.
			return nil
		}

		return []*dataRefWithTypeID{
			&dataRefWithTypeID{
				dID: DataToTypeID(d),
				ref: &pb.DataRef{
					ValueRef: &pb.DataRef_PrimitiveRef{
						PrimitiveRef: primitiveRef,
					},
				},
			},
		}
	case *pb.Data_Struct:
		results := []*dataRefWithTypeID{}
		// Note: we don't build full struct references since we can just construct
		// the full struct using individual field references.
		fieldNames := DataMapKeys(x.Struct.Fields) // Get stable order for x.Struct.Fields
		for _, fieldName := range fieldNames {
			field := x.Struct.Fields[fieldName]
			for _, fieldRefWithID := range dataToDataRefs(isArg, field) {
				newRef := &pb.DataRef{
					ValueRef: &pb.DataRef_StructRef{
						StructRef: &pb.StructRef{
							Ref: &pb.StructRef_FieldRef{
								FieldRef: &pb.NamedDataRef{
									Key:     fieldName,
									DataRef: fieldRefWithID.ref,
								},
							},
						},
					},
				}
				results = append(results, &dataRefWithTypeID{
					dID: fieldRefWithID.dID,
					ref: newRef,
				})
			}
		}
		return results
	case *pb.Data_List:
		results := []*dataRefWithTypeID{}
		for i, elem := range x.List.Elems {
			for _, elemRefWithID := range dataToDataRefs(isArg, elem) {
				newRef := &pb.DataRef{
					ValueRef: &pb.DataRef_ListRef{
						ListRef: &pb.ListRef{
							Ref: &pb.ListRef_ElemRef{
								ElemRef: &pb.IndexedDataRef{
									// TODO: currently, we expect index to be always 0 because
									// we're operating on data types, not instantiated values, so
									// we don't know how many things there will be in the list.
									// https://app.asana.com/0/1131596442939587/1136785931761111
									Index:   int32(i),
									DataRef: elemRefWithID.ref,
								},
							},
						},
					},
				}
				results = append(results, &dataRefWithTypeID{
					dID: elemRefWithID.dID,
					ref: newRef,
				})
			}
		}
		return results
	case *pb.Data_Optional:
		switch y := x.Optional.Value.(type) {
		case *pb.Optional_Data:
			return dataToDataRefs(isArg, y.Data)
		default:
			return nil
		}
	default:
		return nil
	}
}

func DataMapKeys(fields map[string]*pb.Data) []string {
	keys := make([]string, len(fields))
	i := 0
	for k := range fields {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
