package appchilada

import (
	"os"
	"time"
)

type Backend interface {
	Open() os.Error
	Store(m AggregateMap, t *time.Time) os.Error
	Read(name string, interval Interval) (data *Results, err os.Error)
	Names() (names []string, err os.Error)
}

type Results struct {
	Name string
	Rows []*Result
}

type Result struct {
	Time  *time.Time
	Value float64
}

type Interval struct {
	// Start time as timestamp
	Start int64
	// End time as timestamp
	End   int64
}

func (interval Interval) Seconds() int64 {
	return interval.End - interval.Start
}
