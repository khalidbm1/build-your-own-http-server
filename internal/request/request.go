package request

import (
	"fmt"
	"strconv"
	"strings"
)

type Request struct {
	Method        string
	Path          string
	Version       string
	Headers       map[string]string
	Body          string
	ContentLength int
}

const (
	MaxHeaderSize = 8 * 1024    // 8 KB
	MaxBodySize   = 1024 * 1024 // 1 MB
)

func (r *Request) String() string {
	return fmt.Sprintf("%s %s %s (Headers: %d, Body: %d bytes)",
		r.Method, r.Path, r.Version, len(r.Headers), len(r.Body))
}

func parseRequestLine(line string) (method, path, version string, err error) {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("malformed request line: %q", line)
	}
	method, path, version = parts[0], parts[1], parts[2]
	/*
				OR
				method = parts[0]   // GET, POST, PUT, DELETE, etc.
		    	path = parts[1]     // /about, /api/users, etc.
		    	version = parts[2]  // HTTP/1.1
	*/
	validMethods := map[string]bool{
		"GET":     true,
		"HEAD":    true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"CONNECT": true,
		"OPTIONS": true,
		"TRACE":   true,
		"PATCH":   true,
	}
	if !validMethods[method] {
		return "", "", "", fmt.Errorf("Unknown HTTP method: %q", method)
	}

	if version != "HTTP/1.1" && version != "HTTP/1.0" {
		return "", "", "", fmt.Errorf("Unsupported HTTP version: %q", version)
	}
	return method, path, version, nil
}

func parseHeaders(lines []string) (map[string]string, error) {
	headers := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}
		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			return nil, fmt.Errorf("malformed header line: %q", line)
		}
		key := strings.TrimSpace(line[:colonIndex])
		value := strings.TrimSpace(line[colonIndex+1:])
		headers[strings.ToLower(key)] = value
		// OR
		// headers[key] = value
	}
	return headers, nil
}

func extractBody(rawBody string, headers map[string]string) string {
	contentLengthStr, exists := headers["content-length"]
	if !exists {
		return ""
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil || contentLength <= 0 {
		return ""
	}

	if contentLength > len(rawBody) {
		return rawBody
	}
	return rawBody[:contentLength]
}

// Constants for request size limits
// ثوابت لحدود حجم الطلب
const (
	maxHeaderSize = 8 * 1024    // 8 KB
	maxBodySize   = 1024 * 1024 // 1 MB
)

func Parse(rawRequest string) (*Request, error) {
	// parts := strings.SplitN(rawRequest, "\r\n\r\n", 2)
	// headerSection := parts[0]
	// rawBody := ""
	if len(rawRequest) > maxHeaderSize+maxBodySize {
		return nil, fmt.Errorf("request too large")
	}

	parts := strings.Split(rawRequest, "\r\n\r\n")
	if len(parts) < 1 {
		return nil, fmt.Errorf("maleformed request: missing blank line")
	}

	headerSection := parts[0]
	bodySection := ""
	if len(parts) > 1 {
		bodySection = parts[1]
	}

	if len(headerSection) > maxHeaderSize {
		return nil,
			fmt.Errorf("headers too large: %d bytes (limit: %d)",
				len(headerSection), maxHeaderSize)
	}

	// Check 3: Body size limit
	// التحقق 3: حد حجم الجسم
	if len(bodySection) > maxBodySize {
		return nil,
			fmt.Errorf("body too large: %d bytes (limit: %d)",
				len(bodySection), maxBodySize)
	}

	// Parse request line, headers, body (existing code)
	// حلل سطر الطلب، الرؤوس، الجسم (الكود الموجود)

	// Parse request line
	// حلل سطر الطلب
	lines := strings.Split(headerSection, "\r\n")
	if len(lines) < 1 || lines[0] == "" {
		return nil, fmt.Errorf("missing request line")
	}

	requestLine := lines[0]
	parts = strings.Fields(requestLine)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: expected 3 parts, got %d", len(parts))
	}

	req := &Request{
		Method:  parts[0],
		Path:    parts[1],
		Version: parts[2],
		Headers: make(map[string]string),
		Body:    bodySection,
	}

	// Parse headers
	// حلل الرؤوس
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format: %s", line)
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		req.Headers[key] = value
	}

	// Parse Content-Length if present
	// حلل Content-Length لو موجود
	if contentLen, ok := req.Headers["content-length"]; ok {
		len, err := strconv.Atoi(contentLen)
		if err != nil {
			return nil, fmt.Errorf("invalid content-length: %s", contentLen)
		}
		req.ContentLength = len
	}

	return req, nil
}

// Validate checks if the request is valid according to HTTP/1.1 spec.
//
// Validate تتحقق إن الطلب صالح حسب HTTP/1.1 spec.
func (r *Request) Validate() error {
	// Check 1: Method must be present and supported
	// التحقق 1: Method لازم يكون موجود ومدعوم
	if r.Method == "" {
		return fmt.Errorf("missing HTTP method")
	}

	supportedMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"PATCH":   true,
		"DELETE":  true,
		"HEAD":    true,
		"OPTIONS": true,
	}

	if !supportedMethods[r.Method] {
		return fmt.Errorf("unsupported method: %s", r.Method)
	}

	// Check 2: Path must be present and start with /
	// التحقق 2: Path لازم يكون موجود ويبدأ بـ /
	if r.Path == "" || !strings.HasPrefix(r.Path, "/") {
		return fmt.Errorf("invalid path: must start with /, got %s", r.Path)
	}

	// Check 3: Host header is required in HTTP/1.1
	// التحقق 3: Host header مطلوب في HTTP/1.1
	if _, ok := r.Headers["host"]; !ok {
		return fmt.Errorf("missing required Host header")
	}

	// Check 4: Content-Length must be non-negative if present
	// التحقق 4: Content-Length لازم يكون موجب أو صفر لو موجود
	if contentLen, ok := r.Headers["content-length"]; ok {
		len, err := strconv.Atoi(contentLen)
		if err != nil || len < 0 {
			return fmt.Errorf("invalid content-length: %s", contentLen)
		}
	}

	return nil
}

func (r *Request) GetHeader(name string) string {
	return r.Headers[strings.ToLower(name)]
}
