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

	eventChan := make(chan appchilada.Event)
	go appchilada.Aggregator(eventChan)

	buffer := make([]byte, 4096)
	for {
		if n, err := socket.Read(buffer); err != nil {
			fmt.Printf("Read error: %s", err.String())
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
		fmt.Printf("JSON decode error: %s", err.String())
		return
	}
	eventChan <- event
}