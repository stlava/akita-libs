package spec_summary

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/akitasoftware/akita-ir/go/api_spec"
	"github.com/akitasoftware/akita-libs/test"
)

func TestSummarize(t *testing.T) {
	expected := &Summary{
		Authentications: map[string]struct{}{
			"BASIC": struct{}{},
		},
		HTTPMethods: map[string]struct{}{
			"POST": struct{}{},
		},
		Paths: map[string]struct{}{
			"/v1/projects/{arg3}": struct{}{},
		},
		Params: map[string]struct{}{
			"X-My-Header": struct{}{},
		},
		Properties: map[string]struct{}{
			"top-level-prop":       struct{}{},
			"my-special-prop":      struct{}{},
			"other-top-level-prop": struct{}{},
		},
		ResponseCodes: map[int32]struct{}{
			200: struct{}{},
		},
		DataFormats: map[string]struct{}{
			"rfc3339": struct{}{},
		},
		DataKinds: map[string]struct{}{},
	}

	m1 := test.LoadMethodFromFileOrDie("testdata/method1.pb.txt")
	assert.Equal(t, expected, Summarize(&pb.APISpec{Methods: []*pb.Method{m1}}))
}
