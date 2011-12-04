package appchilada

import (
	"os"
	"time"
	"json"
	"log"
	"couch-go.googlecode.com/hg"
)

const designDocument = `
{
	"_id": "_design/appchilada",
	"language": "javascript",
	"views": {
		"counts": {
			"map": "function(doc) {\n if (!doc.Counts || !doc.Timings) return;\n for(key in doc.Counts) {\n  emit([key, doc.Year, doc.Month, doc.Day, doc.Hour, doc.Minute, doc.Second], doc.Counts[key].Value);\n }\n}",
			"reduce": "_stats"
		},
		"names": {
			"map": "function(doc) {\n if (!doc.Counts || !doc.Timings) return;\n for(key in doc.Counts) {\n  emit(key, null);\n }\n}",
			"reduce": "function(keys, values, rereduce) {   \n    return true;\n    }\n"
		}
	}
}
`

const (
	minuteSeconds = 60
	hourSeconds   = 60 * minuteSeconds
	daySeconds    = 24 * hourSeconds
)

type CouchDbBackend struct {
	Host         string
	Port         string
	DatabaseName string
	db           couch.Database
}

type couchDbRecord struct {
	// Time of the aggregation
	Year                 int64
	Month, Day           int
	Hour, Minute, Second int
	// Aggregated counts
	Counts map[string]*Count
	// Aggregated timings
	Timings map[string]*Timing
}

func (backend *CouchDbBackend) Open() os.Error {
	db, err := couch.NewDatabase(backend.Host, backend.Port, backend.DatabaseName)
	if err != nil {
		return err
	}
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(designDocument), &m); err != nil {
		return err
	}
	if id, rev, err := db.InsertWith(m, "_design/appchilada"); err != nil {
		log.Printf("Error inserting design: %v", err)
	} else {
		log.Printf("Design inserted as %s / %s", id, rev)
	}

	backend.db = db
	return nil
}

func (backend *CouchDbBackend) Store(m AggregateMap, t *time.Time) os.Error {
	if len(m) == 0 {
		return nil
	}
	r := &couchDbRecord{t.Year, t.Month, t.Day, t.Hour, t.Minute, t.Second, m.Counts(), m.Timings()}
	id, _, err := backend.db.Insert(r)
	if err != nil {
		return err
	}
	log.Printf("Inserted doc as: %s", id)
	return nil
}

type countRow struct {
	Key   []interface{}
	Value map[string]float64
}

type countRows struct {
	Rows []countRow
}

type keyValueRow struct {
	Key   string
	Value interface{}
}

type keyValueRows struct {
	Rows []keyValueRow
}

// Get a Time instance from an array key
func parseTimeFromKey(key []interface{}) *time.Time {
	t := new(time.Time)
	switch len(key) {
	case 6:
		t.Second = int(key[5].(float64))
		fallthrough
	case 5:
		t.Minute = int(key[4].(float64))
		fallthrough
	case 4:
		t.Hour = int(key[3].(float64))
		fallthrough
	case 3:
		t.Day = int(key[2].(float64))
		fallthrough
	case 2:
		t.Month = int(key[1].(float64))
		fallthrough
	case 1:
		t.Year = int64(key[0].(float64))
		fallthrough
	}
	return t
}

func (backend *CouchDbBackend) Read(name string, interval Interval) (data *Results, err os.Error) {
	var groupingLevel int
	switch s := interval.Seconds(); {
	case s >= 365*daySeconds:
		// Group months
		groupingLevel = 3
	case s > 29*daySeconds:
		// Group days
		groupingLevel = 4
	case s >= daySeconds:
		// Group hours
		groupingLevel = 5
	case s >= hourSeconds:
		// Group minutes
		groupingLevel = 6
	default:
		// Group seconds
		groupingLevel = 7
	}
	// Calculate start and endkey from interval
	startTime := time.SecondsToLocalTime(interval.Start)
	endTime := time.SecondsToLocalTime(interval.End)
	startkey := []interface{}{name, startTime.Year}
	endkey := []interface{}{name, endTime.Year}
	switch {
	case groupingLevel >= 3:
		startkey = append(startkey, startTime.Month)
		endkey = append(endkey, endTime.Month)
		fallthrough
	case groupingLevel >= 4:
		startkey = append(startkey, startTime.Day)
		endkey = append(endkey, endTime.Day)
		fallthrough
	case groupingLevel >= 5:
		startkey = append(startkey, startTime.Hour)
		endkey = append(endkey, endTime.Hour)
		fallthrough
	case groupingLevel >= 6:
		startkey = append(startkey, startTime.Minute)
		endkey = append(endkey, endTime.Minute)
		fallthrough
	default:
		startkey = append(startkey, nil)
		endkey = append(endkey, "_")
	}

	results := &countRows{}
	// Add startkey, endkey and dynamic grouping, limit etc.
	opts := map[string]interface{}{
		"startkey":    startkey,
		"endkey":      endkey,
		"descending":  false,
		"group":       true,
		"group_level": groupingLevel,
		"limit":       1000,
	}
	err = backend.db.Query("_design/appchilada/_view/counts", opts, results)
	if err != nil {
		return nil, err
	} else {
		data = &Results{
			Name: name,
			Rows: make([]*Result, 0, len(results.Rows)),
		}
		for _, row := range results.Rows {
			if row.Key[0] != name {
				continue
			}
			data.Rows = append(data.Rows, &Result{Value: row.Value["sum"] / row.Value["count"], Time: parseTimeFromKey(row.Key[1:])})
		}
	}
	return
}

func (backend *CouchDbBackend) Names() (names []string, err os.Error) {
	results := &keyValueRows{}
	opts := map[string]interface{}{"group": true}
	err = backend.db.Query("_design/appchilada/_view/names", opts, results)
	if err != nil {
		return nil, err
	} else {
		names = make([]string, len(results.Rows))
		for i, row := range results.Rows {
			names[i] = row.Key
		}
	}
	return
}
