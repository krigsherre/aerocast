package main

import (
	encbin "encoding/binary"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	aerobin "github.com/krigsherre/aerocast/pkg/binary"
)

func main() {
	target := flag.String("target", "127.0.0.1:9101", "UDP target address")
	workers := flag.Int("workers", 16, "Number of concurrent blasting goroutines")
	wsTarget := flag.String("ws-target", "ws://127.0.0.1:9100/ws", "WebSocket target URL")
	wsClients := flag.Int("ws-clients", 100, "Number of WebSocket clients to connect")
	wsRadius := flag.Float64("ws-radius", 5000000.0, "Radius for WS clients (meters)")
	maxEntities := flag.Int("max-entities", 100000, "Maximum number of unique entity IDs to simulate")
	flag.Parse()

	addr, err := net.ResolveUDPAddr("udp", *target)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Starting loadtest against UDP %s with %d workers...\n", *target, *workers)
	fmt.Printf("Starting %d WS clients against %s (Radius: %.0fm)...\n", *wsClients, *wsTarget, *wsRadius)

	var wsReceived uint64

	for i := 0; i < *wsClients; i++ {
		go func(clientID int) {
			lat := rand.Float64()*180 - 90
			lng := rand.Float64()*360 - 180

			url := fmt.Sprintf("%s?lat=%f&lng=%f&radius=%f", *wsTarget, lat, lng, *wsRadius)

			conn, _, err := websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				fmt.Printf("WS Dial error: %v\n", err)
				return
			}
			defer conn.Close()

			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
				atomic.AddUint64(&wsReceived, 1)
			}
		}(i)
	}

	var sent uint64

	for i := 0; i < *workers; i++ {
		go func(workerID int) {
			conn, err := net.DialUDP("udp", nil, addr)
			if err != nil {
				panic(err)
			}
			defer conn.Close()

			buf := make([]byte, 20)
			for {
				id := uint32(rand.Intn(*maxEntities) + 1)
				encbin.LittleEndian.PutUint32(buf[0:4], id)
				coord := aerobin.CoordPacket{
					Lat: float64(rand.Float64()*180 - 90),
					Lng: float64(rand.Float64()*360 - 180),
				}
				aerobin.EncodeCoord(buf[4:20], &coord)

				_, err := conn.Write(buf)
				if err == nil {
					atomic.AddUint64(&sent, 1)
				} else {
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	ticker := time.NewTicker(time.Second)
	var lastSent uint64
	var lastReceived uint64

	for range ticker.C {
		currentSent := atomic.LoadUint64(&sent)
		currentReceived := atomic.LoadUint64(&wsReceived)

		sentRate := currentSent - lastSent
		recvRate := currentReceived - lastReceived

		lastSent = currentSent
		lastReceived = currentReceived

		fmt.Printf("Throughput: %d UDP pkts/sec sent | %d WS msgs/sec received\n", sentRate, recvRate)
	}
}
