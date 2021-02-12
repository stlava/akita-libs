package spec_util

import (
	"github.com/golang/protobuf/proto"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
)

// Converts {"a": [1, 2], "b": [3]} to [{"a": 1, "b": 3}, {"a": 2, "b": 3}]
// Passing in a nil returns [nil].
func FlattenAlternatives(alts map[string][]*pb.Data) []map[string]*pb.Data {
	maps := []map[string]*pb.Data{nil}
	newMaps := []map[string]*pb.Data{}
	for k, vs := range alts {
		for _, v := range vs {
			for _, m := range maps {
				var newMap map[string]*pb.Data
				if m == nil {
					newMap = map[string]*pb.Data{k: v}
				} else {
					newMap = copyDataMap(m)
					newMap[k] = v
				}
				newMaps = append(newMaps, newMap)
			}
		}
		maps, newMaps = newMaps, maps // swap
		newMaps = newMaps[:0]         // clear
	}
	return maps
}

func copyDataMap(m map[string]*pb.Data) map[string]*pb.Data {
	newMap := make(map[string]*pb.Data)
	for k, v := range m {
		newMap[k] = proto.Clone(v).(*pb.Data)
	}
	return newMap
}
