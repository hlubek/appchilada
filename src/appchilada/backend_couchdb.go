package appchilada

import (
	"os"
	"fmt"
	"time"
	"json"
	"couch-go.googlecode.com/hg"
)

const designDocument = `
{
	"_id": "_design/appchilada",
	"language": "javascript",
	"views": {
		"counts": {
			"map":    "function(doc) {\n  if (!doc.Counts || !doc.Timings) return;\n  for(key in doc.Counts) {\n    emit([key, doc.Year, doc.Month, doc.Day, doc.Hour, doc.Minute, doc.Second], doc.Counts[key].Value);\n  }\n}",
			"reduce": "function(key, values, rereduce) {   \n    if (!rereduce) {\n        var length = values.length;\n        return [sum(values) / length, length];\n    } else {\n        var length = sum(values.map(function(v) {\n          return v[1];\n        }));\n        var avg = sum(values.map(function(v) {\n            return v[0] * (v[1] / length);\n        }));\n        return [avg, length]\n    }\n}\n"
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
	if id, _, err := db.InsertWith(m, "_design/appchilada"); err != nil {
		fmt.Printf("Error inserting design: %s\n", err)
	} else {
		fmt.Printf("Design inserted as %s\n", id)
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
	fmt.Printf("Inserted doc as: %s\n", id)
	return nil
}
