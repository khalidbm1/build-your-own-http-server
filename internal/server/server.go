// internal/server/server.go
//
// Package server implements a TCP listener with concurrent request handling.
// This version uses goroutines to handle multiple clients simultaneously,
// with connection limits and timeouts to prevent resource exhaustion.
//
// حزمة server تطبق مستمع TCP مع معالجة طلبات متزامنة.
// هالإصدار يستخدم goroutines عشان يتعامل مع عملاء متعددين في نفس الوقت،
// مع حدود اتصالات و timeouts عشان نمنع استنزاف الموارد.

package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/khalidbm1/build-your-own-http-server/internal/request"
	"github.com/khalidbm1/build-your-own-http-server/internal/response"
	"github.com/khalidbm1/build-your-own-http-server/internal/router"
)

// Server holds the configuration for our HTTP server.
//
// Server يحتوي إعدادات خادم HTTP بتاعنا.
type Server struct {
	// Addr is the address to listen on (e.g., ":8080")
	// Addr هو العنوان اللي نستمع عليه (مثل ":8080")
	Addr string

	// MaxConns is the maximum number of concurrent connections allowed.
	// Additional connections will wait in the accept loop.
	// MaxConns هو الحد الأقصى من الاتصالات المتزامنة المسموحة.
	// الاتصالات الإضافية بتنتظر في حلقة القبول.
	MaxConns int

	// ReadTimeout is the maximum time allowed for reading a request.
	// ReadTimeout هو أقصى وقت مسموح لقراءة الطلب.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum time allowed for writing a response.
	// WriteTimeout هو أقصى وقت مسموح لكتابة الاستجابة.
	WriteTimeout time.Duration

	// semaphore is a buffered channel used to limit concurrent connections.
	// It's an implementation of the semaphore pattern using Go channels.
	// semaphore هي قناة ممسوحة تُستخدم عشان تحدد الاتصالات المتزامنة.
	// هي تطبيق لنمط semaphore باستخدام قنوات Go.
	semaphore chan struct{}

	// router handles URL routing.
	// router يتعامل مع توجيه URL.
	router *router.Router
}

// New creates a new Server with the given configuration.
//
// New يسوي Server جديد بالإعدادات المعطاة.
func New(addr string, maxConns int, readTimeout, writeTimeout time.Duration) *Server {
	return &Server{
		Addr:         addr,
		MaxConns:     maxConns,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		semaphore:    make(chan struct{}, maxConns),
		router:       router.New(),
	}
}

// Start begins listening for TCP connections and serving them concurrently.
// It blocks forever (or until an error occurs).
//
// Start يبدأ الاستماع لاتصالات TCP وخدمتها بشكل متزامن.
// يوقف للأبد (أو لين يصير خطأ).
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	log.Printf("Server listening on %s (max %d concurrent connections)", s.Addr, s.MaxConns)

	// Accept loop: wait for a connection, spawn a goroutine to handle it.
	// حلقة القبول: انتظر اتصال، طلق goroutine عشان يتعامل معاه.
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		// Acquire a semaphore slot. If MaxConns is reached, this will block.
		// خذ slot من semaphore. لو MaxConns تم الوصول له، هذا بيوقف.
		s.semaphore <- struct{}{}

		// Launch a goroutine to handle this connection.
		// Each goroutine operates on its own conn, so no race conditions.
		//
		// طلق goroutine عشان يتعامل مع هالاتصال.
		// كل goroutine يعمل على conn خاص فيه، فبدون race conditions.
		go func(conn net.Conn) {
			defer func() {
				// Always release the semaphore when done, even if there's a panic.
				// دائماً خلي semaphore لما تخلص، حتى لو في panic.
				<-s.semaphore
				conn.Close()
			}()

			s.handleConnection(conn)
		}(conn)
	}
}

// handleConnection reads a request from the client, routes it, and sends a response.
// It's now called concurrently for multiple clients.
//
// handleConnection يقرأ طلب من العميل، يوجهه، ويرسل استجابة.
// الحين يُستدعى بشكل متزامن لعملاء متعددين.
func (s *Server) handleConnection(conn net.Conn) {
	// Set read/write deadlines to prevent slow-client attacks.
	// If a read doesn't complete in ReadTimeout, it will error.
	//
	// حط read/write deadlines عشان نمنع هجمات العميل البطيء.
	// لو القراءة ما تنتهي في ReadTimeout، بتخطئ.
	_ = conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
	_ = conn.SetWriteDeadline(time.Now().Add(s.WriteTimeout))

	// Read raw bytes from the connection
	// اقرأ بايتات خام من الاتصال
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		// Timeout or connection error — log it and return
		// Timeout أو خطأ اتصال — سجله وارجع
		log.Printf("Error reading from %s: %v", conn.RemoteAddr(), err)
		return
	}

	rawRequest := string(buf[:n])
	log.Printf("[%s] Received request (%d bytes)", conn.RemoteAddr(), n)

	// Parse the raw bytes into a Request struct
	// حلل البايتات الخام لـ Request struct
	req, err := request.Parse(rawRequest)

	if err != nil {
		log.Printf("[%s] Parse error: %v", conn.RemoteAddr(), err)

		errResp := response.NewResponse(400)
		errResp.SetHeader("Content-Type", "text/plain")
		errResp.Body = "Bad Request"

		_, _ = conn.Write(errResp.Build())
		return
	}

	log.Printf("[%s] %s %s", conn.RemoteAddr(), req.Method, req.Path)

	handler, ok := s.router.Lookup(req.Method, req.Path)
	if !ok {
		notFoundResp := response.NewResponse(404)
		notFoundResp.SetHeader("Content-Type", "text/plain")
		notFoundResp.Body = "Not Found"

		_, _ = conn.Write(notFoundResp.Build())
		return
	}

	resp := handler(req)
	// Route the request to a handler
	// وجه الطلب لمعالج

	// Write the response back to the client
	// اكتب الاستجابة مباشرة إلى العميل
	if resp == nil {
		errResp := response.NewResponse(400)
		errResp.SetHeader("Content-Type", "text/plain")
		errResp.Body = "Internal Server Error"
	}

	_, err = conn.Write(resp.Build())
	if err != nil {
		log.Printf("[%s] Error writing response: %v", conn.RemoteAddr(), err)
		return
	}
	log.Printf("[%s] Sent response (%d bytes)", conn.RemoteAddr(), len(resp.Body))
}
