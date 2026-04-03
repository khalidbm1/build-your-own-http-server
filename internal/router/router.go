package router

import (
	"fmt"

	"github.com/khalidbm1/build-your-own-http-server/internal/request"
	"github.com/khalidbm1/build-your-own-http-server/internal/response"
)

type HnadlerFunc func(*request.Request) *response.Response

type Route struct {
	Method  string
	Path    string
	Hnadler HnadlerFunc
}
type Router struct {
	routes []Route
}

func New() *Router {
	return &Router{
		routes: make([]Route, 0),
	}
}

func (r *Router) Handle(method, path string, handler HnadlerFunc) {
	for i, route := range r.routes {
		if route.Method == method && route.Path == path {
			r.routes[i].Hnadler = handler
			return
		}
	}
	r.routes = append(r.routes, Route{
		Method:  method,
		Path:    path,
		Hnadler: handler,
	})
}

// Lookup
func (r *Router) Lookup(method, path string) (HnadlerFunc, bool) {
	for _, route := range r.routes {
		if route.Method == method && route.Path == path {
			return route.Hnadler, true
		}
	}
	return nil, false
}

// String
func (r *Router) String() string {
	result := "registered routes:\n"
	for _, route := range r.routes {
		result += fmt.Sprintf("  %s %s\n", route.Method, route.Path)
	}
	return result
}
