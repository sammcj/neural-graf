package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	useSSE := flag.Bool("sse", false, "Use SSE mode instead of stdio")
	sseAddr := flag.String("addr", ":8080", "Address for SSE server")
	flag.Parse()

	// Create a conditional logger that only logs in SSE mode
	logger := newConditionalLogger(*useSSE)

	// Create a new MCP server
	s := server.NewMCPServer(
		"example-server",
		"1.0.0",
	)

	// Create a simple tool
	echoTool := mcp.NewTool("echo",
		mcp.WithDescription("Echo the input"),
		mcp.WithString("message",
			mcp.Required(),
			mcp.Description("The message to echo"),
		),
	)

	// Add the tool to the server
	s.AddTool(echoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		message, ok := request.Params.Arguments["message"].(string)
		if !ok {
			return nil, fmt.Errorf("message must be a string")
		}

		// Return the message
		return mcp.NewToolResultText(message), nil
	})

	// Start the server
	logger.Println("Starting MCP server...")

	var err error
	if *useSSE {
		// Start SSE server
		logger.Printf("Starting MCP SSE server on %s", *sseAddr)
		sseServer := server.NewSSEServer(s)
		// Create an HTTP server with the SSE handler
		httpServer := &http.Server{
			Addr:    *sseAddr,
			Handler: sseServer,
		}
		err = httpServer.ListenAndServe()
	} else {
		// Start stdio server (no logging)
		err = server.ServeStdio(s)
	}

	// Only log fatal errors if in SSE mode
	if err != nil && *useSSE {
		log.Fatalf("Error serving MCP: %v", err)
	} else if err != nil {
		// Exit without logging in stdio mode
		os.Exit(1)
	}
}
