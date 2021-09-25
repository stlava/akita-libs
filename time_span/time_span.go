package time_span

import "time"

// Return the earlier of the two times
func MinTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return a
	} else {
		return b
	}
}

// Return the kater of the two times
func MaxTime(a time.Time, b time.Time) time.Time {
	if a.Before(b) {
		return b
	} else {
		return a
	}
}

type BaseInterval struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// A closed interval of time.
type ClosedInterval BaseInterval

func NewClosedInterval(start time.Time, end time.Time) ClosedInterval {
	if start.After(end) {
		start, end = end, start
	}

	return ClosedInterval{
		Start: start,
		End:   end,
	}
}

func (span ClosedInterval) Duration() time.Duration {
	if span.End.After(span.Start) {
		return span.End.Sub(span.Start)
	} else {
		return time.Duration(0)
	}
}

// Determines whether the span includes the given query.
func (span ClosedInterval) Includes(query time.Time) bool {
	return span.Start.Before(query) && query.Before(span.End) || span.Start.Equal(query) || span.End.Equal(query)
}

// Extend an interval by "delta" in each direction (thus it gets 2*delta longer.)
func (t ClosedInterval) Expand(delta time.Duration) ClosedInterval {
	return ClosedInterval{
		Start: t.Start.Add(time.Duration(-1) * delta),
		End:   t.End.Add(delta),
	}
}

// Return the portion of this interval that lies entirely within another closed interval.
// Intersection with an open interval or half-open interval is not provided because the result might
// either be open or closed, and that's a pain to deal with (the point of splitting them out was
// to stop confusing the type of interval involved and be able to statically type-check it!)
func (t ClosedInterval) Intersect(t2 ClosedInterval) ClosedInterval {
	return ClosedInterval{
		Start: MaxTime(t.Start, t2.Start), // later of the two starts
		End:   MinTime(t.End, t2.End),     // earlier of the two ends
	}
}

// Determine whether there is an overlap between two closed intervals
func (t1 ClosedInterval) Overlaps(t2 ClosedInterval) bool {
	// There's some time C such that start1 <= C <= end1 and start2 <= C <= end2
	// iff start1 <= end2 AND start2 <= end1.
	return (t1.Start.Before(t2.End) || t1.Start.Equal(t2.End)) &&
		(t2.Start.Before(t1.End) || t2.Start.Equal(t1.End))
}

// TODO: overlap between open and closed?

// An open interval (start,end)
type OpenInterval BaseInterval

func NewOpenInterval(start time.Time, end time.Time) OpenInterval {
	if start.After(end) {
		start, end = end, start
	}

	return OpenInterval{
		Start: start,
		End:   end,
	}
}

// Extend an interval by "delta" in each direction (thus it gets 2*delta longer.)
func (t OpenInterval) Expand(delta time.Duration) OpenInterval {
	return OpenInterval{
		Start: t.Start.Add(time.Duration(-1) * delta),
		End:   t.End.Add(delta),
	}
}

// Return the portion of this interval that lies entirely within another open interval
func (t OpenInterval) Intersect(t2 OpenInterval) OpenInterval {
	return OpenInterval{
		Start: MaxTime(t.Start, t2.Start), // later of the two starts
		End:   MinTime(t.End, t2.End),     // earlier of the two ends
	}
}

func (t OpenInterval) Duration() time.Duration {
	if t.End.After(t.Start) {
		return t.End.Sub(t.Start)
	} else {
		return time.Duration(0)
	}
}

// Determines whether the interval includes the given query.
func (t OpenInterval) Includes(query time.Time) bool {
	return t.Start.Before(query) && query.Before(t.End)
}

// TODO: overlap between open and closed?

// Determine whether this interval overlaps with another open interval
// If one of them is empty, this should still return false.
func (t1 OpenInterval) Overlaps(t2 OpenInterval) bool {
	// Need C such that s1 < C < e1 and s2 < C < e2, so it is necessary
	// that s1 < e2 and s2 < e1, but also that both of them are nonempty.
	return t1.Start.Before(t2.End) && t2.Start.Before(t1.End) &&
		t1.Start.Before(t1.End) && t2.Start.Before(t2.End)
}

// An half-open interval [start,end)
type HalfOpenInterval BaseInterval

func NewHalfOpenInterval(start time.Time, end time.Time) HalfOpenInterval {
	return HalfOpenInterval{
		Start: start,
		End:   end,
	}
}

func (span HalfOpenInterval) Duration() time.Duration {
	if span.End.After(span.Start) {
		return span.End.Sub(span.Start)
	} else {
		return time.Duration(0)
	}
}

// Determines whether the interval includes the given query.
func (span HalfOpenInterval) Includes(query time.Time) bool {
	return (span.Start.Equal(query) || span.Start.Before(query)) && query.Before(span.End)
}

func (t1 HalfOpenInterval) Overlaps(t2 HalfOpenInterval) bool {
	// There's some time C such that start1 <= C < end1 and start2 <= C < end2
	// iff start1 < end2 AND start2 < end1.
	return t1.Start.Before(t2.End) && t2.Start.Before(t1.End)
}

// Returns true if t2 is entirely contained within t1
func (t1 HalfOpenInterval) Contains(t2 HalfOpenInterval) bool {
	return (t1.Start.Before(t2.Start) || t1.Start.Equal(t2.Start)) &&
		(t1.End.After(t2.End) || t1.End.Equal(t2.End))
}

// Extend an interval by "delta" in each direction (thus it gets 2*delta longer.)
func (t HalfOpenInterval) Expand(delta time.Duration) HalfOpenInterval {
	return HalfOpenInterval{
		Start: t.Start.Add(time.Duration(-1) * delta),
		End:   t.End.Add(delta),
	}
}

// Return the intersection of two half-open intervals.
// (i.e., trim this interval to lie entirely within another.)
// May result in an interval with end <= start, which should be treated as empty.
func (t HalfOpenInterval) Intersect(t2 HalfOpenInterval) HalfOpenInterval {
	return HalfOpenInterval{
		Start: MaxTime(t.Start, t2.Start), // later of the two starts
		End:   MinTime(t.End, t2.End),     // earlier of the two ends
	}
}

// Return the "union" of two half-open intervals, along with everything in between.
// TOOD: I don't know if this has a nice mathematical name.
func (t HalfOpenInterval) Combine(t2 HalfOpenInterval) HalfOpenInterval {
	return HalfOpenInterval{
		Start: MinTime(t.Start, t2.Start), // earlier of the two starts
		End:   MaxTime(t.End, t2.End),     // later of the two ends
	}
}

type TimeSpan = HalfOpenInterval

func NewTimeSpan(start time.Time, end time.Time) *TimeSpan {
	if start.After(end) {
		start, end = end, start
	}

	return &TimeSpan{
		Start: start,
		End:   end,
	}
}
