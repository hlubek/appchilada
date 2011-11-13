package appchilada

import (
	"os"
	"time"
)

type Backend interface {
	Open() os.Error
	Store(m AggregateMap, t *time.Time) os.Error
}
