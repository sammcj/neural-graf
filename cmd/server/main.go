package main

import (
	"context"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sammcj/mcp-graph/internal/api"
	"github.com/sammcj/mcp-graph/internal/config"
	"github.com/sammcj/mcp-graph/internal/graph/dgraph"
	"github.com/sammcj/mcp-graph/internal/mcp"
	"github.com/sammcj/mcp-graph/internal/service"
)

// conditionalLogger is a logger that only logs when enabled
type conditionalLogger struct {
	enabled bool
	logger  *log.Logger
}

// newConditionalLogger creates a new conditional logger
func newConditionalLogger(enabled bool) *conditionalLogger {
	var output io.Writer
	if enabled {
		output = os.Stderr
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

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a conditional logger that only logs in SSE mode
	logger := newConditionalLogger(cfg.MCP.UseSSE)

	// Initialize graph store
	graphStore, err := dgraph.NewDgraphStore(cfg.Dgraph.Address)
	if err != nil {
		log.Fatalf("Failed to connect to Dgraph: %v", err)
	}

	// Create knowledge manager service
	knowledgeService := service.NewService(graphStore)

	// Initialize schema
	logger.Println("Initialising knowledge graph schema...")
	if err := knowledgeService.InitialiseSchema(context.Background()); err != nil {
		logger.Printf("Warning: Failed to initialise schema: %v", err)
	}

	// Create API server
	apiServer := api.NewServer(
		cfg.API.Port,
		knowledgeService,
		graphStore,
	)

	// Enable API server logging only in SSE mode
	apiServer.EnableLogging(cfg.MCP.UseSSE)

	// Create MCP server
	mcpServer := mcp.NewServer(
		cfg.App.Name,
		cfg.App.Version,
		graphStore,
	)
	mcpServer.SetupTools()

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Start API server
	go func() {
		logger.Printf("Starting API server on port %d", cfg.API.Port)
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Printf("API server error: %v", err)
			cancel()
		}
	}()

	// Start MCP server
	if cfg.MCP.UseSSE {
		// Start SSE server in a goroutine
		go func() {
			logger.Printf("Starting MCP SSE server on %s", cfg.MCP.Address)
			if err := mcpServer.ServeSSE(cfg.MCP.Address); err != nil {
				logger.Printf("MCP SSE server error: %v", err)
				cancel()
			}
		}()
	} else {
		// Start stdio server - no logging in this mode
		go func() {
			if err := mcpServer.ServeStdio(); err != nil {
				// Only log fatal errors that cause the server to exit
				if cfg.MCP.UseSSE {
					log.Printf("MCP stdio server error: %v", err)
				}
				cancel()
			}
		}()
	}

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	logger.Println("Shutting down...")

	// Add a small delay to allow for graceful shutdown
	time.Sleep(100 * time.Millisecond)

	// Create a shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Shutdown.Timeout)
	defer shutdownCancel()

	// Shutdown API server
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		logger.Printf("API server shutdown error: %v", err)
	}

	logger.Println("Shutdown complete")
}
