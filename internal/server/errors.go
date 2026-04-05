// internal/server/errors.go
//
// Package server defines custom HTTP error types.
// This allows us to return structured errors with HTTP status codes.
//
// حزمة server تعرّف أنواع أخطاء HTTP مخصصة.
// هذا يخليه يرجع أخطاء منظمة مع HTTP status codes.

package server

import "fmt"

// HTTPError represents an HTTP error with a status code and message.
//
// HTTPError يمثل خطأ HTTP مع status code ورسالة.
type HTTPError struct {
	StatusCode int    // HTTP status code (400, 404, 500, etc.)
	Message    string // Human-readable error message
}

// Error implements the error interface.
//
// Error تطبق واجهة error.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// Predefined errors:
// أخطاء معرّفة مسبقاً:

// BadRequest is a 400 Bad Request error.
var BadRequest = func(reason string) *HTTPError {
	return &HTTPError{400, "Bad Request: " + reason}
}

// MethodNotAllowed is a 405 Method Not Allowed error.
var MethodNotAllowed = func(method string) *HTTPError {
	return &HTTPError{405, "Method Not Allowed: " + method}
}

// RequestEntityTooLarge is a 413 Payload Too Large error.
var RequestEntityTooLarge = func(size int64, limit int64) *HTTPError {
	return &HTTPError{
		413,
		fmt.Sprintf("Request entity too large: %d bytes (limit: %d)", size, limit),
	}
}

// InternalServerError is a 500 Internal Server Error.
var InternalServerError = func(reason string) *HTTPError {
	return &HTTPError{500, "Internal Server Error: " + reason}
}

// RequestHeaderFieldsTooLarge is a 431 Request Header Fields Too Large.
var RequestHeaderFieldsTooLarge = func(size int) *HTTPError {
	return &HTTPError{
		431,
		fmt.Sprintf("Request headers too large: %d bytes (limit: 8KB)", size),
	}
}