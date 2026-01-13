package admin

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/omnipoll/backend/internal/config"
	"github.com/omnipoll/backend/internal/poller"
)

// Server represents the admin HTTP server
type Server struct {
	server        *http.Server
	configManager *config.Manager
	worker        *poller.Worker
	staticFS      fs.FS
}

// NewServer creates a new admin server
func NewServer(cfg *config.Manager, worker *poller.Worker, staticFiles embed.FS) (*Server, error) {
	// Try to get the web/dist subdirectory for static files
	var staticFS fs.FS
	subFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		// If no embedded files, serve without static files
		staticFS = nil
	} else {
		staticFS = subFS
	}

	s := &Server{
		configManager: cfg,
		worker:        worker,
		staticFS:      staticFS,
	}

	return s, nil
}

// NewServerWithoutStatic creates a server without static file serving
func NewServerWithoutStatic(cfg *config.Manager, worker *poller.Worker) *Server {
	return &Server{
		configManager: cfg,
		worker:        worker,
		staticFS:      nil,
	}
}

// NewServerWithFilesystem creates a server serving static files from filesystem
func NewServerWithFilesystem(cfg *config.Manager, worker *poller.Worker, staticDir string) *Server {
	var staticFS fs.FS
	if _, err := os.Stat(staticDir); err == nil {
		staticFS = os.DirFS(staticDir)
	}
	return &Server{
		configManager: cfg,
		worker:        worker,
		staticFS:      staticFS,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	cfg := s.configManager.Get()
	addr := fmt.Sprintf("%s:%d", cfg.Admin.Host, cfg.Admin.Port)

	mux := http.NewServeMux()
	router := NewRouter()

	// API routes
	mux.HandleFunc("/api/status", s.withAuth(s.handleStatus))
	mux.HandleFunc("/api/config", s.withAuth(s.handleConfig))
	mux.HandleFunc("/api/worker/start", s.withAuth(s.handleWorkerStart))
	mux.HandleFunc("/api/worker/stop", s.withAuth(s.handleWorkerStop))
	mux.HandleFunc("/api/watermark/reset", s.withAuth(s.handleWatermarkReset))
	mux.HandleFunc("/api/test/sqlserver", s.withAuth(s.handleTestSQLServer))
	mux.HandleFunc("/api/test/mqtt", s.withAuth(s.handleTestMQTT))
	mux.HandleFunc("/api/test/mongodb", s.withAuth(s.handleTestMongoDB))
	mux.HandleFunc("/api/logs", s.withAuth(s.handleLogsImproved))
	
	// Events routes (using custom router for ID support)
	router.HandleFunc("/api/events", s.withAuth(s.handleEventsRoute))
	router.HandleFunc("/api/events/", s.withAuth(s.handleEventByID))
	router.HandleFunc("/api/events/batch", s.withAuth(s.handleEventsBatch))

	// Static files (frontend)
	if s.staticFS != nil {
		mux.Handle("/", http.FileServer(http.FS(s.staticFS)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "text/html")
				w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Omnipoll Admin</title></head>
<body>
<h1>Omnipoll Admin API</h1>
<p>Frontend not embedded. Run frontend dev server at port 3000.</p>
<h2>API Endpoints:</h2>
<ul>
<li>GET /api/status</li>
<li>GET/PUT /api/config</li>
<li>POST /api/worker/start</li>
<li>POST /api/worker/stop</li>
<li>POST /api/watermark/reset</li>
<li>POST /api/test/sqlserver</li>
<li>POST /api/test/mqtt</li>
<li>POST /api/test/mongodb</li>
<li>GET /api/logs</li>
<li>GET /api/events - List events with pagination</li>
<li>GET /api/events/:id - Get event by ID</li>
<li>PUT /api/events/:id - Update event</li>
<li>DELETE /api/events/:id - Delete event</li>
<li>DELETE /api/events/batch - Delete multiple events</li>
</ul>
</body>
</html>`))
				return
			}
			http.NotFound(w, r)
		})
	}

	// Combine mux and router
	handler := s.withCORS(s.withLogging(combineHandlers(mux, router)))

	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Admin server starting on http://%s", addr)
	return s.server.ListenAndServe()
}

// combineHandlers combines http.Handler and custom Router
func combineHandlers(mux *http.ServeMux, router *Router) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try router first for events API
		if r.URL.Path == "/api/events" || (len(r.URL.Path) > len("/api/events") && r.URL.Path[:len("/api/events")] == "/api/events") {
			router.ServeHTTP(w, r)
			return
		}
		// Fall back to mux for other routes
		mux.ServeHTTP(w, r)
	})
}

// Stop gracefully stops the server
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	cfg := s.configManager.Get()
	return fmt.Sprintf("%s:%d", cfg.Admin.Host, cfg.Admin.Port)
}
