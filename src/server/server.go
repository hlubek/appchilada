package main

import (
	"fmt"
	"flag"
	"net"
	"json"
	"appchilada"
)

var port *int = flag.Int("port", 8686, "Listen port")
var address *string = flag.String("address", "0.0.0.0", "Listen address")

func main() {
	flag.Parse()

	socket := initializeSocket(*address, *port)
	defer socket.Close()

	backend := &appchilada.CouchDbBackend{
		Host: "127.0.0.1",
		Port: "5984",
		DatabaseName: "appchilada_test",
	}
	if err:= backend.Open(); err != nil {
		panic(err.String())
	}

	eventChan := make(chan appchilada.Event)
	// This is where all the aggregation is done
	go appchilada.Aggregator(eventChan, backend)

	buffer := make([]byte, 4096)
	for {
		if n, err := socket.Read(buffer); err != nil {
			fmt.Printf("Read error: %s\n", err.String())
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
		panic(err.String())
	} else {
		fmt.Printf("Starting server on %s:%d\n", ip.String(), port)
	}
	return socket
}

// Handle a JSON encoded message and send an event to the channel
func handleMessage(eventChan chan appchilada.Event, message []byte) {
	var event appchilada.Event
	if err := json.Unmarshal(message, &event); err != nil {
		fmt.Printf("JSON decode error: %s\n", err.String())
		return
	}
	eventChan <- event
}