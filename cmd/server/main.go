package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sammcj/mcp-graph/internal/config"
	"github.com/sammcj/mcp-graph/internal/graph/dgraph"
	"github.com/sammcj/mcp-graph/internal/mcp"
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

	// Allow some time for graceful shutdown
	log.Println("Shutting down...")
	time.Sleep(cfg.Shutdown.Timeout)
	log.Println("Shutdown complete")
}
