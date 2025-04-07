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
	"github.com/sammcj/mcp-graph/internal/service"
)

// Server represents the MCP server for the knowledge graph
type Server struct {
	server   *server.MCPServer
	graph    graph.Store
	service  service.KnowledgeManager
}

// NewServer creates a new MCP server
func NewServer(name, version string, graph graph.Store) *Server {
	s := server.NewMCPServer(
		name,
		version,
		server.WithResourceCapabilities(true, true),
	)

	return &Server{
		server:  s,
		graph:   graph,
		service: service.NewService(graph),
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

	// Document tools
	createDocumentTool := mcp.NewTool("create_document",
		mcp.WithDescription("Create a new document in the knowledge graph"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("The document title"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The document content"),
		),
		mcp.WithObject("metadata",
			mcp.Description("Optional document metadata"),
		),
	)
	s.server.AddTool(createDocumentTool, s.handleCreateDocumentTool)

	getDocumentTool := mcp.NewTool("get_document",
		mcp.WithDescription("Get a document from the knowledge graph by ID"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The ID of the document to retrieve"),
		),
	)
	s.server.AddTool(getDocumentTool, s.handleGetDocumentTool)

	searchDocumentsTool := mcp.NewTool("search_documents",
		mcp.WithDescription("Search for documents in the knowledge graph"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The search query"),
		),
	)
	s.server.AddTool(searchDocumentsTool, s.handleSearchDocumentsTool)

	// Concept tools
	createConceptTool := mcp.NewTool("create_concept",
		mcp.WithDescription("Create a new concept in the knowledge graph"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The concept name"),
		),
		mcp.WithObject("properties",
			mcp.Description("Optional concept properties"),
		),
	)
	s.server.AddTool(createConceptTool, s.handleCreateConceptTool)

	getConceptTool := mcp.NewTool("get_concept",
		mcp.WithDescription("Get a concept from the knowledge graph by ID"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The ID of the concept to retrieve"),
		),
	)
	s.server.AddTool(getConceptTool, s.handleGetConceptTool)

	linkConceptsTool := mcp.NewTool("link_concepts",
		mcp.WithDescription("Create a relationship between two concepts"),
		mcp.WithString("fromId",
			mcp.Required(),
			mcp.Description("The ID of the source concept"),
		),
		mcp.WithString("toId",
			mcp.Required(),
			mcp.Description("The ID of the target concept"),
		),
		mcp.WithString("relationshipType",
			mcp.Required(),
			mcp.Description("The type of relationship"),
		),
		mcp.WithObject("properties",
			mcp.Description("Optional relationship properties"),
		),
	)
	s.server.AddTool(linkConceptsTool, s.handleLinkConceptsTool)

	// Legacy tools for backward compatibility
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

	getNodeTool := mcp.NewTool("get_node",
		mcp.WithDescription("Get a node from the knowledge graph by ID"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The ID of the node to retrieve"),
		),
	)
	s.server.AddTool(getNodeTool, s.handleGetNodeTool)

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

// handleCreateDocumentTool handles the create_document tool
func (s *Server) handleCreateDocumentTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, ok := request.Params.Arguments["title"].(string)
	if !ok {
		return nil, errors.New("title must be a string")
	}

	content, ok := request.Params.Arguments["content"].(string)
	if !ok {
		return nil, errors.New("content must be a string")
	}

	var metadata map[string]interface{}
	if metadataArg, ok := request.Params.Arguments["metadata"]; ok && metadataArg != nil {
		metadata, ok = metadataArg.(map[string]interface{})
		if !ok {
			return nil, errors.New("metadata must be an object")
		}
	}

	// Create document
	id, err := s.service.CreateDocument(ctx, title, content, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Return the document ID
	return mcp.NewToolResultText(fmt.Sprintf(`{"id":"%s"}`, id)), nil
}

// handleGetDocumentTool handles the get_document tool
func (s *Server) handleGetDocumentTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, ok := request.Params.Arguments["id"].(string)
	if !ok {
		return nil, errors.New("id must be a string")
	}

	// Get document
	doc, err := s.service.GetDocument(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// Return the document
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}
	return mcp.NewToolResultText(string(docJSON)), nil
}

// handleSearchDocumentsTool handles the search_documents tool
func (s *Server) handleSearchDocumentsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, ok := request.Params.Arguments["query"].(string)
	if !ok {
		return nil, errors.New("query must be a string")
	}

	// Search documents
	docs, err := s.service.SearchDocuments(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}

	// Return the documents
	docsJSON, err := json.Marshal(docs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal documents: %w", err)
	}
	return mcp.NewToolResultText(string(docsJSON)), nil
}

// handleCreateConceptTool handles the create_concept tool
func (s *Server) handleCreateConceptTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return nil, errors.New("name must be a string")
	}

	var properties map[string]interface{}
	if propertiesArg, ok := request.Params.Arguments["properties"]; ok && propertiesArg != nil {
		properties, ok = propertiesArg.(map[string]interface{})
		if !ok {
			return nil, errors.New("properties must be an object")
		}
	}

	// Create concept
	id, err := s.service.CreateConcept(ctx, name, properties)
	if err != nil {
		return nil, fmt.Errorf("failed to create concept: %w", err)
	}

	// Return the concept ID
	return mcp.NewToolResultText(fmt.Sprintf(`{"id":"%s"}`, id)), nil
}

// handleGetConceptTool handles the get_concept tool
func (s *Server) handleGetConceptTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, ok := request.Params.Arguments["id"].(string)
	if !ok {
		return nil, errors.New("id must be a string")
	}

	// Get concept
	concept, err := s.service.GetConcept(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get concept: %w", err)
	}

	// Return the concept
	conceptJSON, err := json.Marshal(concept)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal concept: %w", err)
	}
	return mcp.NewToolResultText(string(conceptJSON)), nil
}

// handleLinkConceptsTool handles the link_concepts tool
func (s *Server) handleLinkConceptsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromID, ok := request.Params.Arguments["fromId"].(string)
	if !ok {
		return nil, errors.New("fromId must be a string")
	}

	toID, ok := request.Params.Arguments["toId"].(string)
	if !ok {
		return nil, errors.New("toId must be a string")
	}

	relationshipType, ok := request.Params.Arguments["relationshipType"].(string)
	if !ok {
		return nil, errors.New("relationshipType must be a string")
	}

	var properties map[string]interface{}
	if propertiesArg, ok := request.Params.Arguments["properties"]; ok && propertiesArg != nil {
		properties, ok = propertiesArg.(map[string]interface{})
		if !ok {
			return nil, errors.New("properties must be an object")
		}
	}

	// Link concepts
	id, err := s.service.LinkConcepts(ctx, fromID, toID, relationshipType, properties)
	if err != nil {
		return nil, fmt.Errorf("failed to link concepts: %w", err)
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
