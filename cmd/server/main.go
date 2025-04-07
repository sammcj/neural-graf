package main

import (
	"context"
	"flag"
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

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize graph store
	graphStore, err := dgraph.NewDgraphStore(cfg.Dgraph.Address)
	if err != nil {
		log.Fatalf("Failed to connect to Dgraph: %v", err)
	}

	// Create knowledge manager service
	knowledgeService := service.NewService(graphStore)

	// Initialize schema
	log.Println("Initialising knowledge graph schema...")
	if err := knowledgeService.InitialiseSchema(context.Background()); err != nil {
		log.Printf("Warning: Failed to initialise schema: %v", err)
	}

	// Create API server
	apiServer := api.NewServer(
		cfg.API.Port,
		knowledgeService,
		graphStore,
	)

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
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Start API server
	go func() {
		log.Printf("Starting API server on port %d", cfg.API.Port)
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
			cancel()
		}
	}()

	// Start MCP server
	if cfg.MCP.UseSSE {
		// Start SSE server in a goroutine
		go func() {
			log.Printf("Starting MCP SSE server on %s", cfg.MCP.Address)
			if err := mcpServer.ServeSSE(cfg.MCP.Address); err != nil {
				log.Printf("MCP SSE server error: %v", err)
				cancel()
			}
		}()
	} else {
		// Start stdio server
		log.Println("Starting MCP stdio server")
		go func() {
			if err := mcpServer.ServeStdio(); err != nil {
				log.Printf("MCP stdio server error: %v", err)
				cancel()
			}
		}()
	}

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	log.Println("Shutting down...")

	// Add a small delay to allow for graceful shutdown
	time.Sleep(100 * time.Millisecond)

	// Create a shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Shutdown.Timeout)
	defer shutdownCancel()

	// Shutdown API server
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("API server shutdown error: %v", err)
	}

	log.Println("Shutdown complete")
}
