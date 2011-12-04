package main

import (
	// "fmt"
	"flag"
	"json"
	"log"
	"net"
	"appchilada"
	"appchilada/frontend"
)

var port *int = flag.Int("port", 8686, "Listen port")
var address *string = flag.String("address", "0.0.0.0", "Listen address")
var debug *bool = flag.Bool("debug", false, "Log debug messages")
var interval *int = flag.Int("interval", 10, "Event flush interval (in seconds)")

var backend appchilada.Backend

func main() {
	flag.Parse()

	// Initialize the UDP socket to listen to event messages
	socket := initializeSocket(*address, *port)
	defer socket.Close()

	// Initialize the backend
	// TODO Support multiple, configurable backends
	backend = &appchilada.CouchDbBackend{
		Host:         "127.0.0.1",
		Port:         "5984",
		DatabaseName: "appchilada_test",
	}
	if err := backend.Open(); err != nil {
		log.Fatalf("Error opening backend: %v", err)
	}

	go eventLoop(backend, socket)

	frontend.Development = true
	err := frontend.ListenAndServeHttp(backend)
	if err != nil {
		log.Fatal(err)
	}
}

func eventLoop(backend appchilada.Backend, socket *net.UDPConn) {
	eventChan := make(chan appchilada.Event)
	// This is where all the aggregation is done
	go appchilada.Aggregator(eventChan, backend, *interval)

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
