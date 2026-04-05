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
	"io"
	"log"
	"net"
	"net/http"
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
	// Panic recovery: if anything panics, log it and send 500 response
	// Panic recovery: لو شي panics، سجل و ارسل استجابة 500
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[%s] 🔴 PANIC recovered: %v", conn.RemoteAddr(), r)
			errResp := response.NewResponse(500, "text/plain", "Internal Server Error")
			conn.Write(errResp.Bytes())
		}
	}()

	// Set read/write deadlines to prevent slow-client attacks
	// حط read/write deadlines عشان نمنع هجمات العميل البطيء
	conn.SetReadDeadline(time.Now().Add(s.ReadTimeout))
	conn.SetWriteDeadline(time.Now().Add(s.WriteTimeout))

	// Read raw bytes from the connection
	// اقرأ بايتات خام من الاتصال
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)

	// Handle read errors
	// تعامل مع أخطاء القراءة
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("[%s] ⏱ Read timeout after %v", conn.RemoteAddr(), s.ReadTimeout)
		} else if err == io.EOF {
			log.Printf("[%s] ↪ Client closed connection", conn.RemoteAddr())
		} else {
			log.Printf("[%s] ❌ Read error: %v", conn.RemoteAddr(), err)
		}
		return
	}

	rawRequest := string(buf[:n])
	log.Printf("[%s] 📨 Received %d bytes", conn.RemoteAddr(), n)

	// Parse the request with size validation
	// حلل الطلب مع تحقق من الحجم
	req, err := request.Parse(rawRequest)
	if err != nil {
		log.Printf("[%s] ❌ Parse error: %v", conn.RemoteAddr(), err)
		errResp := response.NewResponse(400, "text/plain", "Bad Request: "+err.Error())
		conn.Write(errResp.Bytes())
		return
	}

	// Validate the parsed request
	// تحقق من الطلب المحلل
	if err := req.Validate(); err != nil {
		log.Printf("[%s] ❌ Validation error: %s %s -> %v", conn.RemoteAddr(), req.Method, req.Path, err)

		// Determine appropriate status code
		// حدد HTTP status code المناسب
		statusCode := 400
		if req.Method != "" && !isMethodSupported(req.Method) {
			statusCode = 405
		}

		errResp := response.NewResponse(statusCode, "text/plain", "Bad Request: "+err.Error())
		conn.Write(errResp.Bytes())
		return
	}

	log.Printf("[%s] 📍 %s %s", conn.RemoteAddr(), req.Method, req.Path)

	// Route the request to a handler
	// وجه الطلب لمعالج
	resp := s.router.Route(req)

	// Write the response
	// اكتب الاستجابة
	_, err = conn.Write(resp.Bytes())
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("[%s] ⏱ Write timeout after %v", conn.RemoteAddr(), s.WriteTimeout)
		} else {
			log.Printf("[%s] ❌ Error writing response: %v", conn.RemoteAddr(), err)
		}
		return
	}

	log.Printf("[%s] ✓ %d %s", conn.RemoteAddr(), resp.StatusCode, http.StatusText(resp.StatusCode))
}

// isMethodSupported checks if a method is in the supported list
//
// isMethodSupported تتحقق إن method في القائمة المدعومة
func isMethodSupported(method string) bool {
	supported := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"HEAD":    true,
		"OPTIONS": true,
	}
	return supported[method]
}