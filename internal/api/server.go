package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sammcj/mcp-graph/internal/graph"
	"github.com/sammcj/mcp-graph/internal/service"
)

// Logger defines the interface for logging
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// conditionalLogger is a logger that only logs when enabled
type conditionalLogger struct {
	enabled bool
	logger  *log.Logger
}

// newConditionalLogger creates a new conditional logger
func newConditionalLogger(enabled bool) *conditionalLogger {
	var output io.Writer
	if enabled {
		output = log.Writer()
	} else {
		output = ioutil.Discard
	}

	return &conditionalLogger{
		enabled: enabled,
		logger:  log.New(output, "", log.LstdFlags),
	}
}

// Printf logs a formatted message if enabled
func (l *conditionalLogger) Printf(format string, v ...interface{}) {
	if l.enabled {
		l.logger.Printf(format, v...)
	}
}

// Println logs a message if enabled
func (l *conditionalLogger) Println(v ...interface{}) {
	if l.enabled {
		l.logger.Println(v...)
	}
}

// Server represents the API server
type Server struct {
	router   *mux.Router
	server   *http.Server
	service  service.KnowledgeManager
	graph    graph.Store
	logger   Logger
}

// NewServer creates a new API server
func NewServer(port int, service service.KnowledgeManager, graph graph.Store) *Server {
	router := mux.NewRouter()

	// By default, disable logging (will be enabled in SSE mode)
	logger := newConditionalLogger(false)

	server := &Server{
		router:  router,
		service: service,
		graph:   graph,
		logger:  logger,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	// Set up routes
	server.setupRoutes()

	return server
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// API version prefix
	api := s.router.PathPrefix("/api/v1").Subrouter()

	// Document routes
	docs := api.PathPrefix("/documents").Subrouter()
	docs.HandleFunc("", s.createDocument).Methods(http.MethodPost)
	docs.HandleFunc("", s.searchDocuments).Methods(http.MethodGet)
	docs.HandleFunc("/{id}", s.getDocument).Methods(http.MethodGet)
	docs.HandleFunc("/{id}", s.updateDocument).Methods(http.MethodPut)
	docs.HandleFunc("/{id}", s.deleteDocument).Methods(http.MethodDelete)

	// Concept routes
	concepts := api.PathPrefix("/concepts").Subrouter()
	concepts.HandleFunc("", s.createConcept).Methods(http.MethodPost)
	concepts.HandleFunc("", s.searchConcepts).Methods(http.MethodGet)
	concepts.HandleFunc("/{id}", s.getConcept).Methods(http.MethodGet)
	concepts.HandleFunc("/link", s.linkConcepts).Methods(http.MethodPost)

	// Query route
	api.HandleFunc("/query", s.query).Methods(http.MethodPost)

	// Schema route
	api.HandleFunc("/schema", s.upsertSchema).Methods(http.MethodPost)

	// Add middleware
	api.Use(s.loggingMiddleware)
	api.Use(s.jsonContentTypeMiddleware)
}

// Start starts the API server
func (s *Server) Start() error {
	s.logger.Printf("Starting API server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// EnableLogging enables or disables logging
func (s *Server) EnableLogging(enabled bool) {
	if cl, ok := s.logger.(*conditionalLogger); ok {
		cl.enabled = enabled
	} else {
		// If not a conditionalLogger, replace with one
		s.logger = newConditionalLogger(enabled)
	}
}

// Shutdown gracefully shuts down the API server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// loggingMiddleware logs API requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// jsonContentTypeMiddleware sets the Content-Type header to application/json
func (s *Server) jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		// This is a server-side error, but we don't want to log it in non-SSE mode
		// The error will be reflected in the HTTP response
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Write(response)
}
