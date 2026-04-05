package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khalidbm1/build-your-own-http-server/internal/server"
)

func main() {
	// Define command-line flags
	// عرّف command-line flags
	port := flag.String("port", "8080", "Port to listen on (default: 8080)")
	addr := flag.String("addr", "", "Address to listen on (default: all interfaces)")
	maxConns := flag.Int("max-conns", 100, "Maximum concurrent connections (default: 100)")
	readTimeout := flag.Duration("read-timeout", 10*time.Second, "Read timeout (default: 10s)")
	writeTimeout := flag.Duration("write-timeout", 10*time.Second, "Write timeout (default: 10s)")
	staticDir := flag.String("static", "./static", "Static files directory")
	flag.Parse()

	// Construct full address
	// ابن العنوان الكامل
	listenAddr := *addr + ":" + *port
	if *addr == "" {
		listenAddr = ":" + *port
	}

	// Setup logging with structured format
	// اعدد logging مع صيغة منظمة
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	// Print startup banner
	// اطبع بانر البدء
	log.Println("╔════════════════════════════════════════════════════════════╗")
	log.Println("║       Build Your Own HTTP Server in Go — Stage 9 v1.0      ║")
	log.Println("╚════════════════════════════════════════════════════════════╝")
	log.Printf("Starting server on %s", listenAddr)
	log.Printf("Max concurrent connections: %d", *maxConns)
	log.Printf("Read timeout: %v | Write timeout: %v", *readTimeout, *writeTimeout)
	log.Printf("Static files directory: %s", *staticDir)
	log.Println("Press Ctrl+C to stop")
	log.Println("─────────────────────────────────────────────────────────────")

	// Create server instance
	// اسوي instance من السيرفر
	srv := server.New(listenAddr, *maxConns, *readTimeout, *writeTimeout)

	// Setup graceful shutdown
	// اعدد graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	// ابدأ السيرفر في goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	// انتظر signal قطع
	<-sigChan

	log.Println("\n─────────────────────────────────────────────────────────────")
	log.Println("Shutting down gracefully...")
	log.Println("Goodbye!")
}
