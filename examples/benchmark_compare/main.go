// package main implements an optimized, zero-allocation comparative write-throughput benchmark tool for:
// 1. Aerocast (UDP binary packets)
// 2. Redis Geo (TCP RESP - GEOADD commands)
// 3. Tile38 (TCP RESP - SET POINT commands)
package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	target := flag.String("target", "aerocast", "Target to benchmark: aerocast | redis | tile38")
	address := flag.String("addr", "127.0.0.1:9101", "Target address (e.g. 127.0.0.1:9101 for Aerocast, 127.0.0.1:6379 for Redis, 127.0.0.1:9851 for Tile38)")
	duration := flag.Duration("duration", 5*time.Second, "Duration of the benchmark run")
	workers := flag.Int("workers", 16, "Number of concurrent blasting goroutines")
	flag.Parse()

	fmt.Printf("🎯 Benchmarking [%s] against [%s] for %s with %d workers...\n", *target, *address, *duration, *workers)

	var (
		opsSent  uint64
		opsFail  uint64
		stopSig  int32
		wg       sync.WaitGroup
	)

	// Pre-generate coordinate values to avoid generating random floats on the hot path
	const pregenSize = 50000
	type pt struct {
		lat float64
		lng float64
		id  uint32
	}
	pts := make([]pt, pregenSize)
	for i := 0; i < pregenSize; i++ {
		pts[i] = pt{
			lat: rand.Float64()*180 - 90,
			lng: rand.Float64()*360 - 180,
			id:  uint32(i + 1),
		}
	}

	start := time.Now()

	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Establish socket connection
			var conn net.Conn
			var err error
			if *target == "aerocast" {
				conn, err = net.Dial("udp", *address)
			} else {
				conn, err = net.Dial("tcp", *address)
			}
			if err != nil {
				fmt.Printf("Worker %d Dial Error: %v\n", workerID, err)
				atomic.AddUint64(&opsFail, 1)
				return
			}
			defer conn.Close()

			// Start a background reader to drain TCP responses and prevent socket buffer choke
			if *target != "aerocast" {
				go func(c net.Conn) {
					tmp := make([]byte, 4096)
					for {
						_, err := c.Read(tmp)
						if err != nil {
							return
						}
					}
				}(conn)
			}

			// Reusable formatting buffers and buffered writers
			var buf []byte
			var writer *bufio.Writer
			if *target == "aerocast" {
				buf = make([]byte, 20)
			} else {
				writer = bufio.NewWriterSize(conn, 65536) // 64KB write buffer to group packets
				buf = make([]byte, 0, 128)
			}

			idx := workerID * (pregenSize / *workers)

			for atomic.LoadInt32(&stopSig) == 0 {
				p := pts[idx%pregenSize]
				idx++

				switch *target {
				case "aerocast":
					// Build 20-byte Aerocast UDP Binary packet
					binary.LittleEndian.PutUint32(buf[0:4], p.id)
					binary.LittleEndian.PutUint64(buf[4:12], math.Float64bits(p.lng))
					binary.LittleEndian.PutUint64(buf[12:20], math.Float64bits(p.lat))
					_, err = conn.Write(buf)

				case "redis":
					// Fast, zero-allocation binary RESP writer
					err = writeRedisCmd(writer, buf, p.lng, p.lat, p.id)

				case "tile38":
					// Fast, zero-allocation binary RESP writer
					err = writeTile38Cmd(writer, buf, p.id, p.lat, p.lng)
				}

				if err == nil {
					atomic.AddUint64(&opsSent, 1)
				} else {
					atomic.AddUint64(&opsFail, 1)
					break // Exit loop on write error
				}
			}

			// Flush any remaining buffered commands
			if writer != nil {
				_ = writer.Flush()
			}
		}(w)
	}

	// Wait for benchmark duration
	time.Sleep(*duration)
	atomic.StoreInt32(&stopSig, 1)
	wg.Wait()

	elapsed := time.Since(start).Seconds()
	totalOps := atomic.LoadUint64(&opsSent)
	failedOps := atomic.LoadUint64(&opsFail)

	fmt.Println("--------------------------------------------------")
	fmt.Printf("Elapsed:      %.2f seconds\n", elapsed)
	fmt.Printf("Success Ops:  %d\n", totalOps)
	fmt.Printf("Failed Ops:   %d\n", failedOps)
	fmt.Printf("Throughput:   %.2f ops/sec\n", float64(totalOps)/elapsed)
	fmt.Println("--------------------------------------------------")
}

func writeRedisCmd(w *bufio.Writer, tmpBuf []byte, lng, lat float64, id uint32) error {
	w.WriteString("*5\r\n$6\r\nGEOADD\r\n$5\r\nfleet\r\n")

	// Lng
	tmpBuf = tmpBuf[:0]
	tmpBuf = strconv.AppendFloat(tmpBuf, lng, 'f', 6, 64)
	w.WriteString("$")
	w.WriteString(strconv.Itoa(len(tmpBuf)))
	w.WriteString("\r\n")
	w.Write(tmpBuf)
	w.WriteString("\r\n")

	// Lat
	tmpBuf = tmpBuf[:0]
	tmpBuf = strconv.AppendFloat(tmpBuf, lat, 'f', 6, 64)
	w.WriteString("$")
	w.WriteString(strconv.Itoa(len(tmpBuf)))
	w.WriteString("\r\n")
	w.Write(tmpBuf)
	w.WriteString("\r\n")

	// ID: dev_<id>
	tmpBuf = tmpBuf[:0]
	tmpBuf = append(tmpBuf, "dev_"...)
	tmpBuf = strconv.AppendUint(tmpBuf, uint64(id), 10)
	w.WriteString("$")
	w.WriteString(strconv.Itoa(len(tmpBuf)))
	w.WriteString("\r\n")
	w.Write(tmpBuf)
	w.WriteString("\r\n")

	return nil
}

func writeTile38Cmd(w *bufio.Writer, tmpBuf []byte, id uint32, lat, lng float64) error {
	w.WriteString("*6\r\n$3\r\nSET\r\n$5\r\nfleet\r\n")

	// ID: dev_<id>
	tmpBuf = tmpBuf[:0]
	tmpBuf = append(tmpBuf, "dev_"...)
	tmpBuf = strconv.AppendUint(tmpBuf, uint64(id), 10)
	w.WriteString("$")
	w.WriteString(strconv.Itoa(len(tmpBuf)))
	w.WriteString("\r\n")
	w.Write(tmpBuf)
	w.WriteString("\r\n")

	w.WriteString("$5\r\nPOINT\r\n")

	// Lat
	tmpBuf = tmpBuf[:0]
	tmpBuf = strconv.AppendFloat(tmpBuf, lat, 'f', 6, 64)
	w.WriteString("$")
	w.WriteString(strconv.Itoa(len(tmpBuf)))
	w.WriteString("\r\n")
	w.Write(tmpBuf)
	w.WriteString("\r\n")

	// Lng
	tmpBuf = tmpBuf[:0]
	tmpBuf = strconv.AppendFloat(tmpBuf, lng, 'f', 6, 64)
	w.WriteString("$")
	w.WriteString(strconv.Itoa(len(tmpBuf)))
	w.WriteString("\r\n")
	w.Write(tmpBuf)
	w.WriteString("\r\n")

	return nil
}
