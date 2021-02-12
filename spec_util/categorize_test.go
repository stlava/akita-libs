package spec_util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategorizeNonUTF8String(t *testing.T) {
	invalidUTF8 := "\xc3\x28"
	pv := CategorizeString(invalidUTF8)
	assert.IsType(t, []byte{}, pv.GoValue(), "expect invalid UTF-8 string to be treated as bytes")
}
