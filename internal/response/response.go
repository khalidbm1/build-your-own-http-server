package response

import (
	"fmt"
	"strings"
)

type Response struct {
	StatusCode int
	// StatusText string
	ContentType string
	Headers     [][2]string
	Body        string
}

// var statusReasons = map[int]string{
// 	200: "OK",
// 	201: "Created",
// 	202: "Accepted",
// 	204: "No Content",
// 	301: "Moved Permanently",
// 	302: "Found",
// 	304: "Not Modified",
// 	400: "Bad Request",
// 	401: "Unauthorized",
// 	403: "Forbidden",
// 	404: "Not Found",
// 	405: "Method Not Allowed",
// 	406: "Not Acceptable",
// 	408: "Request Timeout",
// 	410: "Gone",
// 	411: "Length Required",
// 	412: "Precondition Failed",
// 	413: "Payload Too Large",
// 	414: "URI Too Long",
// 	415: "Unsupported Media Type",
// 	416: "Range Not Satisfiable",
// 	417: "Expectation Failed",
// 	500: "Internal Server Error",
// 	501: "Not Implemented",
// 	502: "Bad Gateway",
// 	503: "Service Unavailable",
// 	504: "Gateway Timeout",
// 	505: "HTTP Version Not Supported",
// }

func NewResponse(statusCode int, contentType, body string) *Response {
	return &Response{
		StatusCode:  statusCode,
		ContentType: contentType,
		Body:        body,
	}
}

func (r *Response) Bytes() []byte {
	statusText := reasonForStatus(r.StatusCode)
	statusLine := fmt.Sprintf("Http/1.1 %d %s\r\n", r.StatusCode, statusText)

	headers := fmt.Sprintf("Content-Type: %s\r\n", r.ContentType)
	headers += fmt.Sprintf("Content-Length: %d\r\n", len(r.Body))
	headers += "Connection: Close\r\n"

	response := statusLine + headers + "\r\n" + r.Body
	return []byte(response)
}

func StatusText(code int) string {
	statusTexts := map[int]string{
		100: "Continue",
		101: "Switching Protocols",
		200: "OK",
		201: "Created",
		204: "No Content",
		301: "Moved Permanently",
		302: "Found",
		304: "Not Modified",
		400: "Bad Request",
		401: "Unauthorized",
		403: "Forbidden",
		404: "Not Found",
		405: "Method Not Allowed",
		431: "Request Header Fields Too Large",
		500: "Internal Server Error",
		501: "Not Implemented",
		502: "Bad Gateway",
		503: "Service Unavailable",
	}
	if text, ok := statusTexts[code]; ok {
		return text
	}
	return "Unknown"
}

func reasonForStatus(code int) string {
	return StatusText(code)
}

func (r *Response) SetHeader(key, value string) {
	for i, h := range r.Headers {
		if strings.EqualFold(h[0], key) {
			r.Headers[i][1] = value
			return
		}
	}
	r.Headers = append(r.Headers, [2]string{key, value})
}

func (r *Response) Build() []byte {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.StatusCode, reasonForStatus(r.StatusCode)))

	if len(r.Body) > 0 {
		r.SetHeader("Content-Length", fmt.Sprintf("%d", len(r.Body)))
	}

	r.SetHeader("Connection", "Close")

	for _, h := range r.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", h[0], h[1]))
	}

	sb.WriteString("\r\n")
	sb.WriteString(r.Body)

	return []byte(sb.String())
}

// Text creates a 200 OK response with plain text body.
//
// Text يسوي استجابة 200 OK بجسم نص عادي.
func Text(body string) *Response {
	r := NewResponse(200, "Content-Type", "text/plain; charset=utf-8")
	r.Body = body
	return r
}

// HTML creates a 200 OK response with HTML body.
//
// HTML يسوي استجابة 200 OK بجسم HTML.
func HTML(body string) *Response {
	r := NewResponse(200, "Content-Type", "text/html; charset=utf-8")
	r.Body = body
	return r
}

// JSON creates a 200 OK response with JSON body.
//
// JSON يسوي استجابة 200 OK بجسم JSON.
func JSON(body string) *Response {
	r := NewResponse(200, "Content-Type", "application/json")
	r.Body = body
	return r
}

// NOTFOUND creates a 404 Not Found response.
//
// NOTFOUND يسوي استجابة 404 Not Found.
func NotFound(body string) *Response {
	r := NewResponse(404, "Content-Type", "text/plain; chrset=utf-8")
	r.Body = body
	return r
}

// BadRequest creates a 400 Bad Request response.
//
// BadRequest يسوي استجابة 400 Bad Request.
func BadRequest(message string) *Response {
	r := NewResponse(400, "Content-Type", "text/plain; charset=utf-8")
	r.Body = message
	return r
}

// InternalError creates a 500 Internal Server Error response.
//
// InternalError يسوي استجابة 500 Internal Server Error.
func InternalError(message string) *Response {
	r := NewResponse(500, "Content-Type", "text/plain; charset=utf-8")
	r.Body = message
	return r
}
