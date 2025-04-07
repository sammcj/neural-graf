package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
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
		return &mcp.CallToolResult{
			Content: message,
		}, nil
	})

	// Start the server
	log.Println("Starting MCP server...")
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Error serving MCP: %v", err)
	}
}
