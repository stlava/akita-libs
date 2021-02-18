package trackers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTrackerDomain(t *testing.T) {
	testCases := []struct {
		host     string
		expected bool
	}{
		{
			host:     "",
			expected: false,
		},
		{
			host:     "com",
			expected: false,
		},
		{
			host:     "akitasoftware.com",
			expected: false,
		},
		{
			host:     "www.akitasoftware.com",
			expected: false,
		},
		{
			host:     "segment.io",
			expected: true,
		},
		{
			host:     "segment.io:3000",
			expected: true,
		},
		{
			host:     "cdn.segment.io",
			expected: true,
		},
		{
			host:     "cdn.segment.io:3000",
			expected: true,
		},
	}

	for _, c := range testCases {
		assert.Equal(t, c.expected, IsTrackerDomain(c.host), c.host)
	}
}
