package request

import(
	"fmt"
	"strconv"
	"strings"
)

type Request struct {
	Method string
	Path string
	Version string
	Headers map[string]string
	Body string
}

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


func Parse(rawRequest string) (*Request, error) {
	parts := strings.SplitN(rawRequest, "\r\n\r\n", 2)
	headerSection := parts[0]
	rawBody := ""
	if len(parts) == 2 {
		rawBody = parts[1]
	}

	lines := strings.Split(headerSection, "\r\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty request")
	}

	method, path, version, err := parseRequestLine(lines[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse request line: %w", err)
	}

	headers, err := parseHeaders(lines[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to parse headers: %w", err)
	}

	body := extractBody(rawBody, headers)

	return &Request{
		Method:    method,
		Path:      path,
		Version:   version,
		Headers:   headers,
		Body:      body,
	}, nil
}