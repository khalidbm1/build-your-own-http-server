package router

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/khalidbm1/build-your-own-http-server/internal/request"
	"github.com/khalidbm1/build-your-own-http-server/internal/response"
)

type HnadlerFunc func(*request.Request) *response.Response

type Router struct {
	staticDir string
}

func New() *Router {
	return &Router{
		staticDir: "./static",
	}
}

func (r *Router) Route(req *request.Request) *response.Response {
	// Only support GET requests for now
	// ما ندعم بس GET requests الحين
	if req.Method != "GET" && req.Method != "POST" {
		return response.NewResponse(405, "text/plain", "Method Not Allowed")
	}

	// Route to specific handlers
	// وجه لمعالجات معينة
	switch {
	case req.Path == "/":
		// Serve the landing page
		// قدم صفحة الهبوط
		return r.serveFile("index.html", "text/html")

	case req.Path == "/health":
		// Health check endpoint
		// نقطة فحص الصحة
		return response.NewResponse(200, "text/plain", "OK")

	case req.Path == "/api/info":
		// API endpoint: server info
		// نقطة API: معلومات السيرفر
		return response.NewResponse(200, "application/json",
			`{"server":"Build Your Own HTTP Server","version":"1.0","lang":"Go"}`)

	case strings.HasPrefix(req.Path, "/static/"):
		// Serve static files
		// قدم الملفات الثابتة
		filePath := req.Path[1:] // Remove leading /
		contentType := getContentType(filePath)
		return r.serveFile(filePath, contentType)

	default:
		// 404 Not Found
		return response.NewResponse(404, "text/plain", "File not found")
	}
}

// serveFile reads and serves a file
//
// serveFile تقرأ وتقدم ملف
func (r *Router) serveFile(path string, contentType string) *response.Response {
	// Security: prevent path traversal attacks (e.g., ../../etc/passwd)
	// أمان: منع هجمات path traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return response.NewResponse(403, "text/plain", "Forbidden")

	}
	// Security: prevent directory traversal attacks (e.g., /../etc/passwd)
	// Read the file
	// اقرأ الملف
	fullPath := filepath.Join(r.staticDir, cleanPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return response.NewResponse(404, "Error Reading File", err.Error())
	}

	return response.NewResponse(200, contentType, string(data))
}

// getContentType returns the MIME type for a file
//
// getContentType ترجع نوع MIME لملف
func getContentType(path string) string {
	if strings.HasSuffix(path, ".html") {
		return "text/html"
	} else if strings.HasSuffix(path, ".css") {
		return "text/css"
	} else if strings.HasSuffix(path, ".js") {
		return "application/javascript"
	} else if strings.HasSuffix(path, ".json") {
		return "application/json"
	} else if strings.HasSuffix(path, ".png") {
		return "image/png"
	} else if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
		return "image/jpeg"
	} else if strings.HasSuffix(path, ".gif") {
		return "image/gif"
	} else if strings.HasSuffix(path, ".svg") {
		return "image/svg+xml"
	}
	return "application/octet-stream"
}
