package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/sammcj/mcp-graph/internal/graph"
)

// Server represents the MCP server for the knowledge graph
type Server struct {
	server *server.MCPServer
	graph  graph.Store
}

// NewServer creates a new MCP server
func NewServer(name, version string, graph graph.Store) *Server {
	s := server.NewMCPServer(
		name,
		version,
		server.WithResourceCapabilities(true, true),
	)

	return &Server{
		server: s,
		graph:  graph,
	}
}

// SetupTools configures the MCP tools
func (s *Server) SetupTools() {
	// Query tool
	queryTool := mcp.NewTool("query_knowledge_graph",
		mcp.WithDescription("Query the knowledge graph"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The GraphQL query to execute"),
		),
		mcp.WithObject("params",
			mcp.Description("Optional query parameters"),
		),
	)
	s.server.AddTool(queryTool, s.handleQueryTool)

	// Create node tool
	createNodeTool := mcp.NewTool("create_node",
		mcp.WithDescription("Create a new node in the knowledge graph"),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("The type of node to create"),
		),
		mcp.WithObject("properties",
			mcp.Required(),
			mcp.Description("Node properties"),
		),
	)
	s.server.AddTool(createNodeTool, s.handleCreateNodeTool)

	// Get node tool
	getNodeTool := mcp.NewTool("get_node",
		mcp.WithDescription("Get a node from the knowledge graph by ID"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The ID of the node to retrieve"),
		),
	)
	s.server.AddTool(getNodeTool, s.handleGetNodeTool)

	// Create edge tool
	createEdgeTool := mcp.NewTool("create_edge",
		mcp.WithDescription("Create a new edge between two nodes"),
		mcp.WithString("fromId",
			mcp.Required(),
			mcp.Description("The ID of the source node"),
		),
		mcp.WithString("toId",
			mcp.Required(),
			mcp.Description("The ID of the target node"),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("The type of relationship"),
		),
		mcp.WithObject("properties",
			mcp.Description("Edge properties"),
		),
	)
	s.server.AddTool(createEdgeTool, s.handleCreateEdgeTool)

	// Schema tool
	schemaTool := mcp.NewTool("upsert_schema",
		mcp.WithDescription("Update or create the graph schema"),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("The schema definition"),
		),
	)
	s.server.AddTool(schemaTool, s.handleSchemaTool)
}

// handleQueryTool handles the query_knowledge_graph tool
func (s *Server) handleQueryTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("query must be a string")
	}

	var params map[string]interface{}
	if paramsArg, ok := request.Params.Arguments["params"]; ok {
		if paramsArg != nil {
			params, ok = paramsArg.(map[string]interface{})
			if !ok {
				return nil, errors.New("params must be an object")
			}
		}
	}

	// Execute query against graph
	results, err := s.graph.Query(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// Format and return results
	resultJSON, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal results: %w", err)
	}
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// handleCreateNodeTool handles the create_node tool
func (s *Server) handleCreateNodeTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeType, ok := request.Params.Arguments["type"].(string)
	if !ok {
		return nil, errors.New("type must be a string")
	}

	propertiesArg, ok := request.Params.Arguments["properties"]
	if !ok {
		return nil, errors.New("properties are required")
	}

	properties, ok := propertiesArg.(map[string]interface{})
	if !ok {
		return nil, errors.New("properties must be an object")
	}

	// Create node
	id, err := s.graph.CreateNode(ctx, nodeType, properties)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	// Return the node ID
	return mcp.NewToolResultText(fmt.Sprintf(`{"id":"%s"}`, id)), nil
}

// handleGetNodeTool handles the get_node tool
func (s *Server) handleGetNodeTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, ok := request.Params.Arguments["id"].(string)
	if !ok {
		return nil, errors.New("id must be a string")
	}

	// Get node
	node, err := s.graph.GetNode(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Return the node
	nodeJSON, err := json.Marshal(node)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node: %w", err)
	}
	return mcp.NewToolResultText(string(nodeJSON)), nil
}

// handleCreateEdgeTool handles the create_edge tool
func (s *Server) handleCreateEdgeTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromID, ok := request.Params.Arguments["fromId"].(string)
	if !ok {
		return nil, errors.New("fromId must be a string")
	}

	toID, ok := request.Params.Arguments["toId"].(string)
	if !ok {
		return nil, errors.New("toId must be a string")
	}

	edgeType, ok := request.Params.Arguments["type"].(string)
	if !ok {
		return nil, errors.New("type must be a string")
	}

	var properties map[string]interface{}
	if propertiesArg, ok := request.Params.Arguments["properties"]; ok && propertiesArg != nil {
		properties, ok = propertiesArg.(map[string]interface{})
		if !ok {
			return nil, errors.New("properties must be an object")
		}
	}

	// Create edge
	id, err := s.graph.CreateEdge(ctx, fromID, toID, edgeType, properties)
	if err != nil {
		return nil, fmt.Errorf("failed to create edge: %w", err)
	}

	// Return the edge ID
	return mcp.NewToolResultText(fmt.Sprintf(`{"id":"%s"}`, id)), nil
}

// handleSchemaTool handles the upsert_schema tool
func (s *Server) handleSchemaTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	schema, ok := request.Params.Arguments["schema"].(string)
	if !ok {
		return nil, errors.New("schema must be a string")
	}

	// Update schema
	err := s.graph.UpsertSchema(ctx, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to update schema: %w", err)
	}

	// Return success
	return mcp.NewToolResultText(`{"success":true}`), nil
}

// ServeStdio serves the MCP server over stdio
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.server)
}

// ServeSSE serves the MCP server over SSE
func (s *Server) ServeSSE(addr string) error {
	sseServer := server.NewSSEServer(s.server)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: sseServer,
	}
	return httpServer.ListenAndServe()
}
