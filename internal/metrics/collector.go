package metrics

import (
	"time"
)

type Result struct {
	StatusCode int
	Latency    time.Duration
	Error      error
}

type Collector struct {
	// No mutex needed
	results []Result
}

func NewCollector() *Collector {
	return &Collector{
		// Preallocate capacity
		results: make([]Result, 0, 100000),
	}
}

// AddBatch appends a whole slice of results at once.
func (c *Collector) AddBatch(res []Result) {
	c.results = append(c.results, res...)
}