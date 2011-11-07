package appchilada_test

import (
	"appchilada"
	"testing"
)

var countEvents = []appchilada.Event{
	{appchilada.EventTypeCount, "test.foo", 2},
	{appchilada.EventTypeCount, "test.foo", 3},
	{appchilada.EventTypeCount, "test.bar", 1},
}

var timingEvents = []appchilada.Event{
	{appchilada.EventTypeTiming, "test.foo", 334.0},
	{appchilada.EventTypeTiming, "test.bar", 656.0},
	{appchilada.EventTypeTiming, "test.bar", 2434.0},
}

func TestAggregateMapAddCounts(t *testing.T) {
	m := make(appchilada.AggregateMap)
	for _, e := range countEvents {
		m.AddEvent(&e)
	}
	counts := m.Counts()
	if len(counts) != 2 {
		t.Errorf("Expected count to be of length %d, got %d", 2, len(counts))
	}
	count := counts["test.foo"]
	if count == nil || count.Value != 5 {
		t.Errorf("Expected count with value %d for 'test.foo', got %d", 5, count.Value)
	}
}

func TestAggregateMapAddTimings(t *testing.T) {
	m := make(appchilada.AggregateMap)
	for _, e := range timingEvents {
		m.AddEvent(&e)
	}
	timings := m.Timings()
	if len(timings) != 2 {
		t.Errorf("Expected count to be of length %d, got %d", 2, len(timings))
	}
	timing := timings["test.bar"]
	if timing == nil || timing.Count != 2 {
		t.Errorf("Expected timing with count %d for 'test.bar', got %d", 2, timing.Count)
	}
	if timing.Sum != 3090 {
		t.Errorf("Expected timing with sum %d for 'test.bar', got %d", 3090, timing.Sum)
	}
}
