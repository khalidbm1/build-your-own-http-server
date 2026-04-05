package main

import (
	"flag"
	"log"
	"time"

	"github.com/khalidbm1/build-your-own-http-server/internal/server"
)

func main() {
	// Define command-line flags
	// عرّف command-line flags
	port := flag.String("port", "8080", "Port to listen on")
	maxConns := flag.Int("max-conns", 100, "Maximum concurrent connections")
	readTimeout := flag.Duration("read-timeout", 10*time.Second, "Read timeout")
	writeTimeout := flag.Duration("write-timeout", 10*time.Second, "Write timeout")
	flag.Parse()

	addr := ":" + *port

	log.Println("=== Build Your Own HTTP Server ===")
	log.Printf("Starting server on %s", addr)
	log.Printf("Max concurrent connections: %d", *maxConns)
	log.Printf("Read timeout: %v, Write timeout: %v", *readTimeout, *writeTimeout)

	srv := server.New(addr, *maxConns, *readTimeout, *writeTimeout)

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
