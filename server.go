package main

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// Server is the web app server.
type Server struct {
	logger   *zerolog.Logger
	listener net.Listener
	http     http.Server
	assets   fs.FS
}

// ResponseRecorder is a custom ResponseWriter that captures the status code.
type ResponseRecorder struct {
	http.ResponseWriter
	Status int
}

// NewServer creates a new Server.
func NewServer(logger *zerolog.Logger, listener net.Listener, assets fs.FS) *Server {
	server := &Server{
		logger:   logger,
		listener: listener,
		assets:   assets,
	}

	r := mux.NewRouter()

	// Add middleware.
	r.Use(mux.CORSMethodMiddleware(r))

	// Add logging middleware.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a response recorder to capture the status code.
			rr := NewResponseRecorder(w)

			// Call the next handler in the chain.
			next.ServeHTTP(rr, r)

			// Log information about the incoming request and response status.
			logger.Info().Str("method", r.Method).Str("path", r.URL.Path).Int("status", rr.Status).Msg("Request received")
		})
	})

	NewHTTPRouter(r)

	r.PathPrefix("/").HandlerFunc(server.appHandler)

	server.http = http.Server{
		Handler: r,
	}

	return server
}

// Run starts the server that host webapp and api endpoints.
func (server *Server) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	var group errgroup.Group

	group.Go(func() error {
		<-ctx.Done()
		return server.http.Shutdown(context.Background())
	})
	group.Go(func() error {
		defer cancel()
		err := server.http.Serve(server.listener)
		if err == context.Canceled || errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return err
	})

	return group.Wait()
}

// Close closes server and underlying listener.
func (server *Server) Close() error {
	return server.http.Close()
}

// appHandler is web app http handler function.
func (server *Server) appHandler(w http.ResponseWriter, r *http.Request) {
	staticServer := http.FileServer(http.FS(server.assets))
	header := w.Header()

	if contentType, ok := commonContentType(path.Ext(r.URL.Path)); ok {
		header.Set("Content-Type", contentType)
	}

	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Referrer-Policy", "same-origin")

	staticServer.ServeHTTP(w, r)
}

func commonContentType(ext string) (string, bool) {
	ext = strings.ToLower(ext)
	mime, ok := commonTypes[ext]
	return mime, ok
}

var commonTypes = map[string]string{
	".css":   "text/css; charset=utf-8",
	".gif":   "image/gif",
	".htm":   "text/html; charset=utf-8",
	".html":  "text/html; charset=utf-8",
	".jpeg":  "image/jpeg",
	".jpg":   "image/jpeg",
	".js":    "application/javascript",
	".mjs":   "application/javascript",
	".otf":   "font/otf",
	".pdf":   "application/pdf",
	".png":   "image/png",
	".svg":   "image/svg+xml",
	".ttf":   "font/ttf",
	".wasm":  "application/wasm",
	".webp":  "image/webp",
	".xml":   "text/xml; charset=utf-8",
	".sfnt":  "font/sfnt",
	".woff":  "font/woff",
	".woff2": "font/woff2",
}

// NewResponseRecorder creates a new ResponseRecorder.
func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		Status:         http.StatusOK, // Default to 200 OK.
	}
}

// WriteHeader captures the status code.
func (r *ResponseRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}
