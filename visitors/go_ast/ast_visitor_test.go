package go_ast

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/akitasoftware/akita-libs/visitors"
)

type Counter struct {
	count int
}

func countInt(c visitors.Context, v interface{}, x interface{}) (visitors.Context, bool) {
	counter := v.(*Counter)
	if _, ok := x.(int); ok {
		counter.count++
	}
	return c, true
}

func TestApply(t *testing.T) {
	counter := new(Counter)
	vm := visitors.NewVisitorManager(visitors.NewContext(), counter, countInt)
	d := []int{1, 2, 3}
	Apply(PREORDER, vm, d)
	assert.Equal(t, 3, counter.count)
}
