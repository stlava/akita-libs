package akiuri

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	expected := URI{
		ServiceName: "my_service",
		ObjectName:  "foobar",
		ObjectType:  SPEC,
	}

	u, err := Parse("akita://my_service:spec:foobar")
	assert.NoError(t, err)
	assert.Equal(t, expected, u)
}

func TestString(t *testing.T) {
	u := URI{
		ServiceName: "my_service",
		ObjectName:  "foobar",
		ObjectType:  SPEC,
	}
	assert.Equal(t, "akita://my_service:spec:foobar", u.String())
}
