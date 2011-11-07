package appchilada

import (
	"fmt"
	"time"
	"math"
)

const (
	EventTypeCount  = 0
	EventTypeTiming = 1
)

type Event struct {
	Type  int8
	Name  string
	Value int64
}

type Count struct {
	Value int64
}

type Timing struct {
	Sum   int64
	Count int64
	Min   int64
	Max   int64
}

type Aggregate interface {
	reduce(event *Event)
}

func (count *Count) reduce(event *Event) {
	count.Value += event.Value
}

func (timing *Timing) reduce(event *Event) {
	timing.Count++
	timing.Sum += event.Value
	if timing.Min > event.Value {
		timing.Min = event.Value
	}
	if timing.Max < event.Value {
		timing.Max = event.Value
	}
}

func (timing *Timing) Avg() float64 {
	return float64(timing.Sum) / float64(timing.Count)
}

type AggregateMap map[string][]Aggregate

func (m AggregateMap) AddEvent(event *Event) {
	if m[event.Name] == nil {
		m[event.Name] = make([]Aggregate, 2)
	}
	switch event.Type {
	case EventTypeCount:
		count := m[event.Name][EventTypeCount]
		if count == nil {
			count = &Count{}
			m[event.Name][EventTypeCount] = count
		}
		count.reduce(event)
	case EventTypeTiming:
		timing := m[event.Name][EventTypeTiming]
		if timing == nil {
			timing = &Timing{0, 0, math.MaxInt64, 0}
			m[event.Name][EventTypeTiming] = timing
		}
		timing.reduce(event)
	}
}

func (m AggregateMap) Counts() map[string]*Count {
	counts := make(map[string]*Count, len(m))
	for name, arr := range m {
		if arr[EventTypeCount] != nil {
			count, _ := arr[EventTypeCount].(*Count)
			counts[name] = count
		}
	}
	return counts
}

func (m AggregateMap) Timings() map[string]*Timing {
	timings := make(map[string]*Timing, len(m))
	for name, arr := range m {
		if arr[EventTypeTiming] != nil {
			timing, _ := arr[EventTypeTiming].(*Timing)
			timings[name] = timing
		}
	}
	return timings
}

func Aggregator(eventChan chan Event) {
	events := make([]Event, 0, 64)

	// Notify every 10 seconds
	timer := time.Tick(10 * 1e9)
	for {
		select {
		case event := <-eventChan:
			events = append(events, event)
		case _ = <-timer:
			fmt.Printf("Aggregating %d events\n", len(events))
			m := make(AggregateMap)
			for _, event := range events {
				m.AddEvent(&event)
			}
			// Print values for debugging
			for name, count := range m.Counts() {
				fmt.Printf("Count: %s=%d\n", name, count.Value)
			}
			for name, timing := range m.Timings() {
				fmt.Printf("Timer: %s=%f (Min: %d, Max: %d)\n", name, timing.Avg(), timing.Min, timing.Max)
			}
			events = events[0:0]
		}
	}
}
