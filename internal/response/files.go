package response

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/khalidbm1/build-your-own-http-server/internal/request"
)

var mimeType = map[string]string{
	".html": "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".txt":  "text/plain",
	".json": "application/json",
	".ico":  "image/x-icon",
	".webp": "image/webp",
}

func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if contentType, ok := mimeType[ext]; ok {
		return contentType
	}
	return "application/octet-stream"
}

func FileServer(baseDir, urlPrefix string) func(*request.Request) *Response {
	return func(req *request.Request) *Response {
		relativePath := strings.TrimPrefix(req.Path, urlPrefix)
		relativePath = strings.TrimPrefix(relativePath, "/")

		if relativePath == "" {
			relativePath = "index.html"
		}

		cleanBase := filepath.Clean(baseDir)
		fullPath := filepath.Clean(filepath.Join(cleanBase, relativePath))

		if !strings.HasPrefix(fullPath, cleanBase+string(os.PathSeparator)) && fullPath != cleanBase {
			return Forbidden("Access Denied: path Traversal Attempt")
		}
		fileBytes, err := os.ReadFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				return NotFound("Not Found")
			}
			return InternalError(fmt.Sprintf("Error Reading File: %v", err))
		}
		resp := NewResponse(200)
		resp.SetHeader("Content-type", getMimeType(fullPath))
		resp.Body = string(fileBytes)
		return resp
	}
}

func Forbidden(message string) *Response {
	r := NewResponse(403)
	r.SetHeader("Content-Type", "text/plain; charset=utf-8")
	r.Body = message
	return r
}
