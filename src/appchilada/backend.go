package appchilada

import (
	"os"
	"time"
)

type Backend interface {
	Open() os.Error
	Store(m AggregateMap, t *time.Time) os.Error
	Read(name string) (data *Results, err os.Error)
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
