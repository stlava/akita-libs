package sampled_err

import (
	"fmt"
	"math/rand"
	"strings"
)

// Samples up to SampleCount errors.
type Errors struct {
	Samples     []error
	SampleCount int
	TotalCount  int
}

func (es Errors) Error() string {
	if es.TotalCount == 0 {
		return "no error"
	}

	strs := make([]string, 0, len(es.Samples))
	for _, e := range es.Samples {
		strs = append(strs, e.Error())
	}

	return fmt.Sprintf("encountered %d errors, showing %d samples: %s",
		es.TotalCount, es.SampleCount, strings.Join(strs, ","))
}

func (es *Errors) Add(e error) {
	es.TotalCount += 1
	if len(es.Samples) >= es.SampleCount {
		// Randomly replace one of the existing samples.
		if rand.Intn(2) == 1 {
			es.Samples[rand.Intn(len(es.Samples))] = e
		}
	} else {
		es.Samples = append(es.Samples, e)
	}
}
