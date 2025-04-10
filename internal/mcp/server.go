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

	// --- Software Architecture Tools ---

	findOrCreateEntityTool := mcp.NewTool("find_or_create_entity",
		mcp.WithDescription("Idempotently finds or creates a node (entity) based on identifying properties, merging provided properties."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels for the entity (e.g., ['Function', 'Go']). At least one label is required."),
			mcp.Items(map[string]interface{}{"type": "string"}), // Corrected item schema
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties used to uniquely identify the entity for matching (e.g., {'filePath': '/path/to/file.go', 'name': 'MyFunc'}). At least one property is required."),
		),
		mcp.WithObject("properties",
			mcp.Required(),
			mcp.Description("Map of all properties (including identifying ones) to set on create or merge on match. 'lastModifiedAt' will be automatically set/updated."),
		),
	)
	s.server.AddTool(findOrCreateEntityTool, s.handleFindOrCreateEntityTool)

	findOrCreateRelationshipTool := mcp.NewTool("find_or_create_relationship",
		mcp.WithDescription("Idempotently finds or creates a relationship between two nodes, merging provided properties."),
		mcp.WithArray("startNodeLabels",
			mcp.Required(),
			mcp.Description("List of labels for the start node."),
			mcp.Items(map[string]interface{}{"type": "string"}), // Corrected item schema
		),
		mcp.WithObject("startNodeIdentifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the start node."),
		),
		mcp.WithArray("endNodeLabels",
			mcp.Required(),
			mcp.Description("List of labels for the end node."),
			mcp.Items(map[string]interface{}{"type": "string"}), // Corrected item schema
		),
		mcp.WithObject("endNodeIdentifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the end node."),
		),
		mcp.WithString("relationshipType",
			mcp.Required(),
			mcp.Description("The type of the relationship (e.g., 'CALLS', 'DEPENDS_ON')."),
		),
		mcp.WithObject("properties",
			mcp.Description("Map of properties to set on create or merge on match for the relationship. 'lastModifiedAt' will be automatically set/updated."),
		),
	)
	s.server.AddTool(findOrCreateRelationshipTool, s.handleFindOrCreateRelationshipTool)

	getEntityDetailsTool := mcp.NewTool("get_entity_details",
		mcp.WithDescription("Retrieves the labels and properties of a specific entity."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels to identify the entity."),
			mcp.Items(map[string]interface{}{"type": "string"}), // Corrected item schema
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the entity."),
		),
	)
	s.server.AddTool(getEntityDetailsTool, s.handleGetEntityDetailsTool)

	findNeighborsTool := mcp.NewTool("find_neighbors",
		mcp.WithDescription("Finds the neighbors of a specific entity up to a given depth."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels for the central entity."),
			mcp.Items(map[string]interface{}{"type": "string"}), // Corrected item schema
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the central entity."),
		),
		mcp.WithNumber("maxDepth", // Changed from mcp.WithInteger
			mcp.Description("Maximum depth to search for neighbors (default: 1)."),
			// Removed mcp.Default(1) - handled in Go code
		),
	)
	s.server.AddTool(findNeighborsTool, s.handleFindNeighborsTool)

	findDependenciesTool := mcp.NewTool("find_dependencies",
		mcp.WithDescription("Finds entities that the target entity depends on (outgoing relationships) up to a given depth."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels for the target entity."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the target entity."),
		),
		mcp.WithArray("relationshipTypes",
			mcp.Description("Optional list of relationship types to follow (e.g., ['DEPENDS_ON', 'CALLS']). If empty, follows all types."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithNumber("maxDepth",
			mcp.Description("Maximum depth to search for dependencies (default: 1)."),
		),
	)
	s.server.AddTool(findDependenciesTool, s.handleFindDependenciesTool)

	findDependentsTool := mcp.NewTool("find_dependents",
		mcp.WithDescription("Finds entities that depend on the target entity (incoming relationships) up to a given depth."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels for the target entity."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the target entity."),
		),
		mcp.WithArray("relationshipTypes",
			mcp.Description("Optional list of relationship types to follow (e.g., ['DEPENDS_ON', 'CALLS']). If empty, follows all types."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithNumber("maxDepth",
			mcp.Description("Maximum depth to search for dependents (default: 1)."),
		),
	)
	s.server.AddTool(findDependentsTool, s.handleFindDependentsTool)

	getEntitySubgraphTool := mcp.NewTool("get_entity_subgraph",
		mcp.WithDescription("Retrieves nodes and relationships around a central entity, suitable for visualisation."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels for the central entity."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the central entity."),
		),
		mcp.WithNumber("maxDepth",
			mcp.Description("Maximum depth to retrieve the subgraph (default: 1)."),
		),
	)
	s.server.AddTool(getEntitySubgraphTool, s.handleGetEntitySubgraphTool)
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

// --- Software Architecture Tool Handlers ---

// handleFindOrCreateEntityTool handles the find_or_create_entity tool
func (s *Server) handleFindOrCreateEntityTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input graph.EntityInput

	// Parse labels
	labelsArg, ok := request.Params.Arguments["labels"]
	if !ok {
		return nil, errors.New("labels are required")
	}
	labelsInterface, ok := labelsArg.([]interface{})
	if !ok {
		return nil, errors.New("labels must be an array of strings")
	}
	input.Labels = make([]string, len(labelsInterface))
	for i, l := range labelsInterface {
		input.Labels[i], ok = l.(string)
		if !ok {
			return nil, fmt.Errorf("label item at index %d is not a string", i)
		}
	}
	if len(input.Labels) == 0 {
		return nil, errors.New("at least one label is required")
	}

	// Parse identifyingProperties
	idPropsArg, ok := request.Params.Arguments["identifyingProperties"]
	if !ok {
		return nil, errors.New("identifyingProperties are required")
	}
	input.IdentifyingProperties, ok = idPropsArg.(map[string]interface{})
	if !ok {
		return nil, errors.New("identifyingProperties must be an object")
	}
	if len(input.IdentifyingProperties) == 0 {
		return nil, errors.New("at least one identifying property is required")
	}

	// Parse properties
	propsArg, ok := request.Params.Arguments["properties"]
	if !ok {
		return nil, errors.New("properties are required")
	}
	input.Properties, ok = propsArg.(map[string]interface{})
	if !ok {
		return nil, errors.New("properties must be an object")
	}

	// Call graph store method
	details, err := s.graph.FindOrCreateEntity(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create entity: %w", err)
	}

	// Return the entity details
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity details: %w", err)
	}
	return mcp.NewToolResultText(string(detailsJSON)), nil
}

// handleFindOrCreateRelationshipTool handles the find_or_create_relationship tool
func (s *Server) handleFindOrCreateRelationshipTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input graph.RelationshipInput
	var ok bool

	// Parse startNodeLabels
	startLabelsArg, ok := request.Params.Arguments["startNodeLabels"]
	if !ok { return nil, errors.New("startNodeLabels are required") }
	startLabelsInterface, ok := startLabelsArg.([]interface{})
	if !ok { return nil, errors.New("startNodeLabels must be an array of strings") }
	input.StartNodeLabels = make([]string, len(startLabelsInterface))
	for i, l := range startLabelsInterface {
		input.StartNodeLabels[i], ok = l.(string)
		if !ok { return nil, fmt.Errorf("startNodeLabel item at index %d is not a string", i) }
	}
	if len(input.StartNodeLabels) == 0 { return nil, errors.New("at least one startNodeLabel is required") }

	// Parse startNodeIdentifyingProperties
	startIdPropsArg, ok := request.Params.Arguments["startNodeIdentifyingProperties"]
	if !ok { return nil, errors.New("startNodeIdentifyingProperties are required") }
	input.StartNodeIdentifyingProperties, ok = startIdPropsArg.(map[string]interface{})
	if !ok { return nil, errors.New("startNodeIdentifyingProperties must be an object") }
	if len(input.StartNodeIdentifyingProperties) == 0 { return nil, errors.New("at least one startNodeIdentifyingProperty is required") }

	// Parse endNodeLabels
	endLabelsArg, ok := request.Params.Arguments["endNodeLabels"]
	if !ok { return nil, errors.New("endNodeLabels are required") }
	endLabelsInterface, ok := endLabelsArg.([]interface{})
	if !ok { return nil, errors.New("endNodeLabels must be an array of strings") }
	input.EndNodeLabels = make([]string, len(endLabelsInterface))
	for i, l := range endLabelsInterface {
		input.EndNodeLabels[i], ok = l.(string)
		if !ok { return nil, fmt.Errorf("endNodeLabel item at index %d is not a string", i) }
	}
	if len(input.EndNodeLabels) == 0 { return nil, errors.New("at least one endNodeLabel is required") }

	// Parse endNodeIdentifyingProperties
	endIdPropsArg, ok := request.Params.Arguments["endNodeIdentifyingProperties"]
	if !ok { return nil, errors.New("endNodeIdentifyingProperties are required") }
	input.EndNodeIdentifyingProperties, ok = endIdPropsArg.(map[string]interface{})
	if !ok { return nil, errors.New("endNodeIdentifyingProperties must be an object") }
	if len(input.EndNodeIdentifyingProperties) == 0 { return nil, errors.New("at least one endNodeIdentifyingProperty is required") }

	// Parse relationshipType
	relTypeArg, ok := request.Params.Arguments["relationshipType"]
	if !ok { return nil, errors.New("relationshipType is required") }
	input.RelationshipType, ok = relTypeArg.(string)
	if !ok || input.RelationshipType == "" { return nil, errors.New("relationshipType must be a non-empty string") }

	// Parse properties (optional)
	if propsArg, exists := request.Params.Arguments["properties"]; exists && propsArg != nil {
		input.Properties, ok = propsArg.(map[string]interface{})
		if !ok { return nil, errors.New("properties must be an object") }
	} else {
		input.Properties = make(map[string]interface{}) // Ensure it's not nil
	}


	// Call graph store method
	relProps, err := s.graph.FindOrCreateRelationship(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create relationship: %w", err)
	}

	// Return the relationship properties
	relPropsJSON, err := json.Marshal(relProps)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal relationship properties: %w", err)
	}
	return mcp.NewToolResultText(string(relPropsJSON)), nil
}

// handleGetEntityDetailsTool handles the get_entity_details tool
func (s *Server) handleGetEntityDetailsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var labels []string
	var identifyingProperties map[string]interface{}
	var ok bool

	// Parse labels
	labelsArg, ok := request.Params.Arguments["labels"]
	if !ok { return nil, errors.New("labels are required") }
	labelsInterface, ok := labelsArg.([]interface{})
	if !ok { return nil, errors.New("labels must be an array of strings") }
	labels = make([]string, len(labelsInterface))
	for i, l := range labelsInterface {
		labels[i], ok = l.(string)
		if !ok { return nil, fmt.Errorf("label item at index %d is not a string", i) }
	}
	if len(labels) == 0 { return nil, errors.New("at least one label is required") }

	// Parse identifyingProperties
	idPropsArg, ok := request.Params.Arguments["identifyingProperties"]
	if !ok { return nil, errors.New("identifyingProperties are required") }
	identifyingProperties, ok = idPropsArg.(map[string]interface{})
	if !ok { return nil, errors.New("identifyingProperties must be an object") }
	if len(identifyingProperties) == 0 { return nil, errors.New("at least one identifying property is required") }

	// Call graph store method
	details, err := s.graph.GetEntityDetails(ctx, labels, identifyingProperties)
	if err != nil {
		// Consider returning a structured error for "not found"
		return nil, fmt.Errorf("failed to get entity details: %w", err)
	}

	// Return the entity details
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity details: %w", err)
	}
	return mcp.NewToolResultText(string(detailsJSON)), nil
}

// handleFindNeighborsTool handles the find_neighbors tool
func (s *Server) handleFindNeighborsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var labels []string
	var identifyingProperties map[string]interface{}
	var maxDepth int = 1 // Default depth
	var ok bool

	// Parse labels
	labelsArg, ok := request.Params.Arguments["labels"]
	if !ok { return nil, errors.New("labels are required") }
	labelsInterface, ok := labelsArg.([]interface{})
	if !ok { return nil, errors.New("labels must be an array of strings") }
	labels = make([]string, len(labelsInterface))
	for i, l := range labelsInterface {
		labels[i], ok = l.(string)
		if !ok { return nil, fmt.Errorf("label item at index %d is not a string", i) }
	}
	if len(labels) == 0 { return nil, errors.New("at least one label is required") }

	// Parse identifyingProperties
	idPropsArg, ok := request.Params.Arguments["identifyingProperties"]
	if !ok { return nil, errors.New("identifyingProperties are required") }
	identifyingProperties, ok = idPropsArg.(map[string]interface{})
	if !ok { return nil, errors.New("identifyingProperties must be an object") }
	if len(identifyingProperties) == 0 { return nil, errors.New("at least one identifying property is required") }

	// Parse maxDepth (optional)
	if depthArg, exists := request.Params.Arguments["maxDepth"]; exists {
		depthFloat, ok := depthArg.(float64) // JSON numbers are often float64
		if !ok {
			return nil, errors.New("maxDepth must be an integer")
		}
		maxDepth = int(depthFloat)
		if maxDepth <= 0 {
			maxDepth = 1 // Ensure positive depth
		}
	}

	// Call graph store method
	neighborsResult, err := s.graph.FindNeighbors(ctx, labels, identifyingProperties, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to find neighbors: %w", err)
	}

	// Return the neighbors result
	resultJSON, err := json.Marshal(neighborsResult)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal neighbors result: %w", err)
	}
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// Helper function to parse common arguments for dependency/subgraph tools
func parseEntityLocatorArgs(request mcp.CallToolRequest) (labels []string, idProps map[string]interface{}, maxDepth int, err error) {
	var ok bool
	maxDepth = 1 // Default

	// Parse labels
	labelsArg, ok := request.Params.Arguments["labels"]
	if !ok { err = errors.New("labels are required"); return }
	labelsInterface, ok := labelsArg.([]interface{})
	if !ok { err = errors.New("labels must be an array of strings"); return }
	labels = make([]string, len(labelsInterface))
	for i, l := range labelsInterface {
		labels[i], ok = l.(string)
		if !ok { err = fmt.Errorf("label item at index %d is not a string", i); return }
	}
	if len(labels) == 0 { err = errors.New("at least one label is required"); return }

	// Parse identifyingProperties
	idPropsArg, ok := request.Params.Arguments["identifyingProperties"]
	if !ok { err = errors.New("identifyingProperties are required"); return }
	idProps, ok = idPropsArg.(map[string]interface{})
	if !ok { err = errors.New("identifyingProperties must be an object"); return }
	if len(idProps) == 0 { err = errors.New("at least one identifying property is required"); return }

	// Parse maxDepth (optional)
	if depthArg, exists := request.Params.Arguments["maxDepth"]; exists && depthArg != nil {
		depthFloat, ok := depthArg.(float64) // JSON numbers are often float64
		if !ok { err = errors.New("maxDepth must be a number"); return }
		maxDepth = int(depthFloat)
		if maxDepth <= 0 {
			maxDepth = 1 // Ensure positive depth
		}
	}
	return
}

// Helper function to parse optional relationship types
func parseOptionalRelationshipTypes(request mcp.CallToolRequest) ([]string, error) {
	var relTypes []string
	if relTypesArg, exists := request.Params.Arguments["relationshipTypes"]; exists && relTypesArg != nil {
		relTypesInterface, ok := relTypesArg.([]interface{})
		if !ok {
			return nil, errors.New("relationshipTypes must be an array of strings")
		}
		relTypes = make([]string, len(relTypesInterface))
		for i, rt := range relTypesInterface {
			relTypes[i], ok = rt.(string)
			if !ok {
				return nil, fmt.Errorf("relationshipType item at index %d is not a string", i)
			}
		}
	}
	return relTypes, nil
}


// handleFindDependenciesTool handles the find_dependencies tool
func (s *Server) handleFindDependenciesTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	labels, idProps, maxDepth, err := parseEntityLocatorArgs(request)
	if err != nil {
		return nil, err
	}
	relTypes, err := parseOptionalRelationshipTypes(request)
	if err != nil {
		return nil, err
	}

	// Call graph store method
	depResult, err := s.graph.FindDependencies(ctx, labels, idProps, relTypes, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to find dependencies: %w", err)
	}

	// Return the result
	resultJSON, err := json.Marshal(depResult)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dependency result: %w", err)
	}
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// handleFindDependentsTool handles the find_dependents tool
func (s *Server) handleFindDependentsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	labels, idProps, maxDepth, err := parseEntityLocatorArgs(request)
	if err != nil {
		return nil, err
	}
	relTypes, err := parseOptionalRelationshipTypes(request)
	if err != nil {
		return nil, err
	}

	// Call graph store method
	depResult, err := s.graph.FindDependents(ctx, labels, idProps, relTypes, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to find dependents: %w", err)
	}

	// Return the result
	resultJSON, err := json.Marshal(depResult)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dependency result: %w", err)
	}
	return mcp.NewToolResultText(string(resultJSON)), nil
}

// handleGetEntitySubgraphTool handles the get_entity_subgraph tool
func (s *Server) handleGetEntitySubgraphTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	labels, idProps, maxDepth, err := parseEntityLocatorArgs(request)
	if err != nil {
		return nil, err
	}

	// Call graph store method
	subgraphResult, err := s.graph.GetEntitySubgraph(ctx, labels, idProps, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity subgraph: %w", err)
	}

	// Return the result
	resultJSON, err := json.Marshal(subgraphResult)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subgraph result: %w", err)
	}
	return mcp.NewToolResultText(string(resultJSON)), nil
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
