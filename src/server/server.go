package main

import (
	// "fmt"
	"flag"
	"json"
	"log"
	"net"
	"http"
	"template"
	"os"
	"appchilada"
)

var port *int = flag.Int("port", 8686, "Listen port")
var address *string = flag.String("address", "0.0.0.0", "Listen address")

var backend appchilada.Backend

func main() {
	flag.Parse()

	socket := initializeSocket(*address, *port)
	defer socket.Close()

	backend = &appchilada.CouchDbBackend{
		Host:         "127.0.0.1",
		Port:         "5984",
		DatabaseName: "appchilada_test",
	}
	if err := backend.Open(); err != nil {
		log.Fatalf("Error opening backend: %v", err)
	} else {
		log.Printf("Opened backend")
	}

	go eventLoop(backend, socket)

	// Initialize HTTP server for frontend
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/show/", showHandler)
	if dir, err := os.Getwd(); err != nil {
		log.Fatalf("Cannot get cwd: %v", err)
	} else {
		dir = dir + string(os.PathSeparator) + "assets"
		log.Printf("Opening webserver in %s", dir)
		// Handle static files in /public served on /public (stripped prefix)
		http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir(dir))))
	}
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error running webserver: %v", err)
	} else {
		log.Printf("Opening webserver on :8080")
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	names, err := backend.Names()
	if err != nil {
		// TODO Output error in response
		log.Printf("Error getting names: %v", err)
		return
	}
	d := map[string]interface{}{
		"names": names,
	}

	t, err := template.ParseFile("resources/index.html")
	if err != nil {
		log.Printf("Error opening template: %v", err)
		return
	}
	if err := t.Execute(w, d); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func showHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/show/"):]
	results, err := backend.Read(name)
	if err != nil {
		// TODO Output error in response
		log.Printf("Error getting results for %s: %v", name, err)
		return
	}
	d := map[string]interface{}{
		"results": results,
	}

	t, err := template.ParseFile("resources/show.html")
	if err != nil {
		log.Printf("Error opening template: %v", err)
		return
	}
	if err := t.Execute(w, d); err != nil {
		log.Printf("Error executing template: %v", err)
	}
}

func eventLoop(backend appchilada.Backend, socket *net.UDPConn) {
	eventChan := make(chan appchilada.Event)
	// This is where all the aggregation is done
	go appchilada.Aggregator(eventChan, backend)

	buffer := make([]byte, 4096)
	for {
		if n, err := socket.Read(buffer); err != nil {
			log.Printf("Socket read error: %v", err)
		} else {
			handleMessage(eventChan, buffer[:n])
		}
	}
}

func initializeSocket(address string, port int) *net.UDPConn {
	ip := net.ParseIP(address)
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		log.Fatalf("Error opening socket: %v", err)
	} else {
		log.Printf("Starting appchilada server on udp://%s:%d", ip.String(), port)
	}
	return socket
}

// Handle a JSON encoded message and send an event to the channel
func handleMessage(eventChan chan appchilada.Event, message []byte) {
	var event appchilada.Event
	if err := json.Unmarshal(message, &event); err != nil {
		log.Printf("JSON decode error: %v", err)
		return
	}
	eventChan <- event
}
