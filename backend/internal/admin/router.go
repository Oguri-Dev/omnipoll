package admin

import (
	"net/http"
	"strings"
)

// Router provides simple path-based routing
type Router struct {
	routes []route
}

type route struct {
	pattern string
	handler http.HandlerFunc
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes: make([]route, 0),
	}
}

// HandleFunc registers a handler for a pattern
// Supports patterns like "/api/events" and "/api/events/"
func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	r.routes = append(r.routes, route{
		pattern: pattern,
		handler: handler,
	})
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Try exact match first
	for _, rt := range r.routes {
		if req.URL.Path == rt.pattern {
			rt.handler(w, req)
			return
		}
	}

	// Try prefix match with trailing slash handling
	for _, rt := range r.routes {
		if rt.pattern == "/api/events" && strings.HasPrefix(req.URL.Path, "/api/events") {
			// Check if it's a sub-path like /api/events/123
			if req.URL.Path == "/api/events" || req.URL.Path == "/api/events/" {
				rt.handler(w, req)
				return
			}
		}
	}

	// Check for event ID routes
	for _, rt := range r.routes {
		if strings.HasPrefix(rt.pattern, "/api/events/") && strings.HasPrefix(req.URL.Path, "/api/events/") {
			rt.handler(w, req)
			return
		}
	}

	http.NotFound(w, req)
}
