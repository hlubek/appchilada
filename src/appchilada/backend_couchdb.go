package appchilada

import (
	"os"
	"time"
	"json"
	"log"
	"fmt"
	"couch-go.googlecode.com/hg"
)

const designDocument = `
{
	"_id": "_design/appchilada",
	"language": "javascript",
	"views": {
		"counts": {
			"map":    "function(doc) {\n  if (!doc.Counts || !doc.Timings) return;\n  for(key in doc.Counts) {\n    emit([key, doc.Year, doc.Month, doc.Day, doc.Hour, doc.Minute, doc.Second], doc.Counts[key].Value);\n  }\n}",
			"reduce": "function(keys, values, rereduce) {   \n    if (!rereduce) {\n        var length = values.length;\n        return [sum(values) / length, length];\n    } else {\n        var length = sum(values.map(function(v) {\n          return v[1];\n        }));\n        var avg = sum(values.map(function(v) {\n            return v[0] * (v[1] / length);\n        }));\n        return [avg, length];\n    }\n}\n"
		},
		"names": {
			"map":    "function(doc) {\n  if (!doc.Counts || !doc.Timings) return;\n  for(key in doc.Counts) {\n    emit(key, null);\n  }\n}",
			"reduce": "function(keys, values, rereduce) {   \n    return true;\n    }\n"
		}
	}
}
`

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
	Counts               map[string]*Count
	Timings              map[string]*Timing
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
	Value []float64
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

func (backend *CouchDbBackend) Read(name string) (data *Results, err os.Error) {
	results := &countRows{}
	// Add startkey, endkey and dynamic grouping, limit etc.
	opts := map[string]interface{}{
		"descending":  false,
		"group":       true,
		"group_level": 6,
		"limit":       24,
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
			t, err := time.Parse(time.RFC3339, fmt.Sprintf("%f-%f-%fT%f:%f", row.Key[1], row.Key[2], row.Key[3], row.Key[4], row.Key[5]))
			if err != nil {
				log.Printf("Error parsing time: %v", err)
			}
			data.Rows = append(data.Rows, &Result{Value: row.Value[0], Time: t})
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
		names = make([]string, 0, len(results.Rows))
		for _, row := range results.Rows {
			names = append(names, row.Key)
		}
	}
	return
}
