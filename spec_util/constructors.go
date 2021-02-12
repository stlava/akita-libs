package spec_util

import (
	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/pkg/errors"

	"github.com/akitasoftware/akita-libs/pbhash"
)

func OneOf (data []*pb.Data, isConflict bool) (*pb.OneOf, error) {
	if len(data) == 0 {
		return &pb.OneOf{PotentialConflict: isConflict}, nil
	}
	options := make(map[string]*pb.Data, len(data))
	for _, option := range data {
		hash, err := pbhash.HashProto(option)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to hash data: %v", option)
		}
		options[hash] = option
	}
	return &pb.OneOf{
		Options:              options,
		PotentialConflict:    isConflict,
	}, nil
}
