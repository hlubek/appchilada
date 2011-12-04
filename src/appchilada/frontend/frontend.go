package frontend

import (
	"appchilada"
	"http"
	"os"
	"log"
	"template"
	"time"
	"strconv"
)

var Development = false

// Initialize HTTP server for frontend
func ListenAndServeHttp(backend appchilada.Backend) os.Error {
	http.HandleFunc("/", indexHandler(backend))
	http.HandleFunc("/show/", showHandler(backend))

	if dir, err := os.Getwd(); err != nil {
		return err
	} else {
		dir = dir + string(os.PathSeparator) + "assets"
		log.Printf("Opening webserver in %s", dir)
		// Handle static files in /public served on /public (stripped prefix)
		http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir(dir))))
	}
	log.Printf("Opening webserver on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		return err
	}
	return nil
}

// Get a template by filename and refresh if in development mode
func getTemplateFunc(filename string) func() *template.Template {
	t, err := template.ParseFile(filename)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}
	return func() *template.Template {
		if Development {
			t, err = template.ParseFile(filename)
			if err != nil {
				log.Printf("Error parsing template: %v", err)
			}
		}
		return t
	}
}

func indexHandler(backend appchilada.Backend) func(http.ResponseWriter, *http.Request) {
	getTemplate := getTemplateFunc("resources/index.html")
	return func(w http.ResponseWriter, r *http.Request) {
		names, err := backend.Names()
		if err != nil {
			// TODO Output error in response
			log.Printf("Error getting names: %v", err)
			return
		}
		d := map[string]interface{}{
			"names": names,
		}
		if err := getTemplate().Execute(w, d); err != nil {
			log.Printf("Error executing template: %v", err)
		}
	}
}

func showHandler(backend appchilada.Backend) func(http.ResponseWriter, *http.Request) {
	getTemplate := getTemplateFunc("resources/show.html")
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		name := r.URL.Path[len("/show/"):]
		var start, end int64
		if startVal := r.Form.Get("start"); startVal != "" {
			start, _ = strconv.Atoi64(startVal)
		} else {
			// Default to last 24 hours
			start = time.Seconds() - 86400
		}
		if endVal := r.Form.Get("end"); endVal != "" {
			end, _ = strconv.Atoi64(endVal)
		} else {
			// Default to now
			end = time.Seconds()
		}
		results, err := backend.Read(name, appchilada.Interval{start, end})
		if err != nil {
			// TODO Output error in response
			log.Printf("Error getting results for %s: %v", name, err)
			return
		}
		if err := getTemplate().Execute(w, results); err != nil {
			log.Printf("Error executing template: %v", err)
		}
	}
}
