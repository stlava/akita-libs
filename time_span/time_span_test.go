package time_span

import (
	"testing"
	"time"
)

func TestIntervalOverlap(t *testing.T) {
	testCases := []struct {
		start1          int // expressed in seconds for clarity
		end1            int
		start2          int
		end2            int
		overlapClosed   bool // expected value of overlap for interval type
		overlapHalfOpen bool
	}{
		{
			0, 10,
			20, 30,
			false,
			false,
		},
		{
			0, 10,
			5, 15,
			true,
			true,
		},
		{
			0, 10,
			0, 10,
			true,
			true,
		},
		{
			0, 10,
			10, 20,
			true,
			false,
		},
		{
			15, 15,
			10, 20,
			true,
			false,
		},
		{
			15, 0,
			10, 20,
			true,  // because closed constructor reverses the order
			false, // because half-open takes the arguments as-is
		},
		{
			15, 1,
			16, 0,
			true,  // [1,15] inside [0,16]
			false, // [15,1) not inside empty interval [16,0)
		},
	}

	for _, tc := range testCases {
		t.Logf("Test case: [%v, %v] and [%v, %v]", tc.start1, tc.end1, tc.start2, tc.end2)
		start1 := time.Date(2020, 1, 1, 0, 0, tc.start1, 0, time.UTC)
		end1 := time.Date(2020, 1, 1, 0, 0, tc.end1, 0, time.UTC)
		start2 := time.Date(2020, 1, 1, 0, 0, tc.start2, 0, time.UTC)
		end2 := time.Date(2020, 1, 1, 0, 0, tc.end2, 0, time.UTC)
		ci1 := NewClosedInterval(start1, end1)
		ci2 := NewClosedInterval(start2, end2)
		hoi1 := NewHalfOpenInterval(start1, end1)
		hoi2 := NewHalfOpenInterval(start2, end2)

		actual := ci1.Overlaps(ci2)
		if actual != tc.overlapClosed {
			t.Errorf("closed expected %v got %v", tc.overlapClosed, actual)
		}
		actual = ci2.Overlaps(ci1)
		if actual != tc.overlapClosed {
			t.Errorf("reversed closed expected %v got %v", tc.overlapClosed, actual)
		}

		actual = hoi1.Overlaps(hoi2)
		if actual != tc.overlapHalfOpen {
			t.Errorf("half-open expected %v got %v", tc.overlapHalfOpen, actual)
		}
		actual = hoi2.Overlaps(hoi1)
		if actual != tc.overlapHalfOpen {
			t.Errorf("reversed half-open expected %v got %v", tc.overlapHalfOpen, actual)
		}
	}
}

func TestIntervalContains(t *testing.T) {
	testCases := []struct {
		start1   int // expressed in seconds for clarity
		end1     int
		start2   int
		end2     int
		expected bool // only HO is implemented
	}{
		{
			0, 10,
			5, 15,
			false,
		},
		{
			5, 15,
			0, 10,
			false,
		},
		{
			0, 20,
			5, 10,
			true,
		},
		{
			5, 10,
			0, 20,
			false,
		},
		{
			0, 10,
			0, 10,
			true,
		},
		{
			0, 10,
			5, 10,
			true,
		},
		{
			5, 10,
			0, 10,
			false,
		},
		{
			0, 10,
			0, 5,
			true,
		},
		{
			0, 5,
			0, 10,
			false,
		},
		{
			10, 15,
			5, 0,
			true, // empty interval is contained in any nonempty interval
		},
		{
			15, 0,
			5, 0,
			false, // empty interval cannot contain anything
		},
	}

	for _, tc := range testCases {
		t.Logf("Test case: [%v, %v] and [%v, %v]", tc.start1, tc.end1, tc.start2, tc.end2)
		start1 := time.Date(2020, 1, 1, 0, 0, tc.start1, 0, time.UTC)
		end1 := time.Date(2020, 1, 1, 0, 0, tc.end1, 0, time.UTC)
		start2 := time.Date(2020, 1, 1, 0, 0, tc.start2, 0, time.UTC)
		end2 := time.Date(2020, 1, 1, 0, 0, tc.end2, 0, time.UTC)
		hoi1 := NewHalfOpenInterval(start1, end1)
		hoi2 := NewHalfOpenInterval(start2, end2)

		actual := hoi1.Contains(hoi2)
		if actual != tc.expected {
			t.Errorf("half-open expected %v got %v", tc.expected, actual)
		}
	}
}
