package spec_util

import (
	"fmt"
	"reflect"
	"sort"
)

type IterMapTuple struct {
	Key   interface{}
	Value interface{}
}

func IterMapInOrder(m interface{}) []IterMapTuple {
	mapValue := reflect.ValueOf(m)
	stringKeyMap := make(map[string]reflect.Value)
	for _, k := range mapValue.MapKeys() {
		stringKey := fmt.Sprintf("%v", k)
		stringKeyMap[stringKey] = k
	}

	stringKeys := []string{}
	for sk := range stringKeyMap {
		stringKeys = append(stringKeys, sk)
	}
	sort.Strings(stringKeys)

	tuples := make([]IterMapTuple, len(stringKeys))
	for i, sk := range stringKeys {
		tuples[i] = IterMapTuple{
			Key:   stringKeyMap[sk].Interface(),
			Value: mapValue.MapIndex(stringKeyMap[sk]).Interface(),
		}
	}
	return tuples
}
