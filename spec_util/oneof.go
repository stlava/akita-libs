package spec_util

import (
	"github.com/pkg/errors"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

type ErrNoSuitableOneOf struct{}

func (e ErrNoSuitableOneOf) Error() string {
	return "no suitable oneof option"
}

// A function that creates an instantiated Data from a Data spec.
type OneOfInstantiator func(spec *pb.Data) (instantiated *pb.Data, err error)

// Given an OneOfInstantiator, this function feeds all oneof option specs to
// the factory and returns the "best" output. Currently, "best" is defined as
// having the most number of non-none fields.
// If a factory returns an error, its output is not considered. If all factories
// return error, this function returns ErrNoSuitableOneOf.
// containerMeta is the DataMeta of the Data object that contains the oneof.
//
// Rationale for the scoring to pick the "best" instantiation:
// It's possible that multiple oneof options are valid for instantiation. For
// example, if the spec is:
//
// oneof:
//   - struct{ "foo": optional bool } # option 1
//   - struct{ "bar": optional bool } # option 2
//
// If the raw data is `{"foo": true}`, technically both options are valid specs
// for the raw data because option 2 can just treat the "bar" field as null, but
// option 1 is clearly better.
func InstantiateOneOf(oneof *pb.OneOf, containerMeta *pb.DataMeta, factory OneOfInstantiator) (*pb.Data, error) {
	candidates := make([]*pb.Data, 0, len(oneof.GetOptions()))
	for _, option := range oneof.GetOptions() {
		if d, err := factory(option); err == nil {
			candidates = append(candidates, d)
		}
	}

	var bestCandidate *pb.Data
	if len(candidates) == 1 {
		bestCandidate = candidates[0]
	} else if len(candidates) > 0 {
		bestScore := -1
		bestIndex := -1
		for i, c := range candidates {
			if score, err := numNonNullFields(c); err != nil {
				return nil, errors.Wrap(err, "failed to score candidate")
			} else if score > bestScore {
				bestScore = score
				bestIndex = i
			}
		}
		bestCandidate = candidates[bestIndex]
	}

	if bestCandidate == nil {
		return nil, ErrNoSuitableOneOf{}
	}
	bestCandidate.Meta = MergeOneOfMeta(containerMeta, bestCandidate.Meta)
	return bestCandidate, nil
}

func numNonNullFields(d *pb.Data) (int, error) {
	switch v := d.Value.(type) {
	case *pb.Data_Primitive:
		return 1, nil
	case *pb.Data_Struct:
		sum := 0
		for k, f := range v.Struct.Fields {
			if c, err := numNonNullFields(f); err != nil {
				return -1, errors.Wrapf(err, "failed to count non-null fields in field %s", k)
			} else {
				sum += c
			}
		}
		return sum, nil
	case *pb.Data_List:
		sum := 0
		for i, e := range v.List.Elems {
			if c, err := numNonNullFields(e); err != nil {
				return -1, errors.Wrapf(err, "failed to count non-null fields in list element %d", i)
			} else {
				sum += c
			}
		}
		return sum, nil
	case *pb.Data_Optional:
		switch v := v.Optional.Value.(type) {
		case *pb.Optional_Data:
			return numNonNullFields(v.Data)
		case *pb.Optional_None:
			return 0, nil
		default:
			return -1, errors.Errorf("unsupported option type %T", v)
		}
	default:
		return -1, errors.Errorf("cannot cound fields in value type %T", v)
	}
}

// Merges DataMeta from an oneof option with that of its containing Data object.
func MergeOneOfMeta(containerMeta, optionMeta *pb.DataMeta) *pb.DataMeta {
	// Currently, the merge just prefers containerMeta since for HTTP specs we
	// don't use the per-option meta. This potentially needs to change when adding
	// gRPC support.
	if containerMeta == nil {
		return optionMeta
	}
	return containerMeta
}
