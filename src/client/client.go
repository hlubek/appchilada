package main

import (
	"flag"
	"net"
	"time"
	"strconv"
	"rand"
	"appchilada"
)

var port *int = flag.Int("port", 8686, "Server port")
var address *string = flag.String("address", "127.0.0.1", "Server address")

var randomLabels []string = []string{"Foo", "Bar", "Baz", "Blub"}

func main() {
	flag.Parse()

	ip := net.ParseIP(*address)
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   ip,
		Port: *port,
	})
	if err != nil {
		panic(err.String())
	}
	defer socket.Close()

	for {
		label := randomLabels[rand.Intn(len(randomLabels))]
		var message string
		if rand.Intn(2) < 1 {
			message = `{"type":` + strconv.Itoa(appchilada.EventTypeCount) + `,"name":"` + label + `","value":` + strconv.Itoa(rand.Intn(5) + 1) + `}`
		} else {
			message = `{"type":` + strconv.Itoa(appchilada.EventTypeTiming) + `,"name":"` + label + `","value":` + strconv.Itoa(rand.Intn(1000) + 10) + `}`
		}
		println(message)
		if _, err := socket.Write([]byte(message)); err != nil {
			println(err.String())
			break
		}
		time.Sleep(int64(rand.Intn(10) * 1e7))
	}
}
