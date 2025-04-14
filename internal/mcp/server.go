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
		mcp.WithDescription("Executes a Cypher query against the Neo4j graph database. Use this for complex graph traversals or queries not covered by other specific tools."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The Cypher query string to execute. Use parameter placeholders like $paramName."),
		),
		mcp.WithObject("params",
			mcp.Description("Optional map of parameters to bind to the Cypher query. Keys should match placeholders in the query string (without the $)."),
		),
	)
	s.server.AddTool(queryTool, s.handleQueryTool)

	// Document tools
	createDocumentTool := mcp.NewTool("create_document",
		mcp.WithDescription("Creates a new 'Document' node in the knowledge graph. Useful for storing textual information like source file contents, documentation snippets, or research notes."),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("The title of the document (e.g., file name, article title)."),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The main textual content of the document."),
		),
		mcp.WithObject("metadata",
			mcp.Description("Optional map of key-value pairs for additional metadata (e.g., {'source_url': '...', 'author': '...'})."),
		),
	)
	s.server.AddTool(createDocumentTool, s.handleCreateDocumentTool)

	getDocumentTool := mcp.NewTool("get_document",
		mcp.WithDescription("Retrieves a 'Document' node from the knowledge graph using its unique ID."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The unique identifier (elementId) of the 'Document' node to retrieve."),
		),
	)
	s.server.AddTool(getDocumentTool, s.handleGetDocumentTool)

	searchDocumentsTool := mcp.NewTool("search_documents",
		mcp.WithDescription("Performs a text-based search across 'Document' nodes in the knowledge graph. (Note: Specific search implementation depends on the underlying graph store)."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The text query string to search for within document content or titles."),
		),
	)
	s.server.AddTool(searchDocumentsTool, s.handleSearchDocumentsTool)

	// Concept tools
	createConceptTool := mcp.NewTool("create_concept",
		mcp.WithDescription("Creates a new 'Concept' node in the knowledge graph. Concepts represent abstract ideas or entities."),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the concept."),
		),
		mcp.WithObject("properties",
			mcp.Description("Optional map of key-value pairs for additional properties describing the concept."),
		),
	)
	s.server.AddTool(createConceptTool, s.handleCreateConceptTool)

	getConceptTool := mcp.NewTool("get_concept",
		mcp.WithDescription("Retrieves a 'Concept' node from the knowledge graph using its unique ID."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The unique identifier (elementId) of the 'Concept' node to retrieve."),
		),
	)
	s.server.AddTool(getConceptTool, s.handleGetConceptTool)

	linkConceptsTool := mcp.NewTool("link_concepts",
		mcp.WithDescription("Creates a directed relationship between two existing 'Concept' nodes."),
		mcp.WithString("fromId",
			mcp.Required(),
			mcp.Description("The unique ID (elementId) of the source 'Concept' node."),
		),
		mcp.WithString("toId",
			mcp.Required(),
			mcp.Description("The unique ID (elementId) of the target 'Concept' node."),
		),
		mcp.WithString("relationshipType",
			mcp.Required(),
			mcp.Description("The type of the relationship (e.g., 'RELATED_TO', 'PART_OF')."),
		),
		mcp.WithObject("properties",
			mcp.Description("Optional map of key-value pairs for properties of the relationship itself."),
		),
	)
	s.server.AddTool(linkConceptsTool, s.handleLinkConceptsTool)

	// Legacy tools for backward compatibility (Consider deprecating or removing if not needed)
	createNodeTool := mcp.NewTool("create_node",
		mcp.WithDescription("[Legacy] Creates a generic node with a specified type (label) and properties. Prefer using specific tools like 'create_document', 'create_concept', or 'find_or_create_entity'."),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("The primary label for the new node."),
		),
		mcp.WithObject("properties",
			mcp.Required(),
			mcp.Description("Map of key-value pairs for the node's properties."),
		),
	)
	s.server.AddTool(createNodeTool, s.handleCreateNodeTool)

	getNodeTool := mcp.NewTool("get_node",
		mcp.WithDescription("[Legacy] Retrieves a generic node by its unique ID (elementId). Prefer using specific tools like 'get_document', 'get_concept', or 'get_entity_details'."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The unique identifier (elementId) of the node to retrieve."),
		),
	)
	s.server.AddTool(getNodeTool, s.handleGetNodeTool)

	createEdgeTool := mcp.NewTool("create_edge",
		mcp.WithDescription("[Legacy] Creates a directed relationship between two existing nodes identified by their IDs. Prefer using specific tools like 'link_concepts' or 'find_or_create_relationship'."),
		mcp.WithString("fromId",
			mcp.Required(),
			mcp.Description("The unique ID (elementId) of the source node."),
		),
		mcp.WithString("toId",
			mcp.Required(),
			mcp.Description("The unique ID (elementId) of the target node."),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("The type name for the relationship."),
		),
		mcp.WithObject("properties",
			mcp.Description("Optional map of key-value pairs for properties of the relationship itself."),
		),
	)
	s.server.AddTool(createEdgeTool, s.handleCreateEdgeTool)

	schemaTool := mcp.NewTool("upsert_schema",
		mcp.WithDescription("Applies schema definitions (like constraints and indexes) to the graph database. Accepts Cypher DDL statements."),
		mcp.WithString("schema",
			mcp.Required(),
			mcp.Description("A string containing one or more Cypher DDL statements (e.g., 'CREATE CONSTRAINT FOR (n:User) REQUIRE n.uuid IS UNIQUE;'). Statements can be separated by newlines or semicolons."),
		),
	)
	s.server.AddTool(schemaTool, s.handleSchemaTool)

	// --- Software Architecture Tools ---

	findOrCreateEntityTool := mcp.NewTool("find_or_create_entity",
		mcp.WithDescription("Idempotently finds a node (entity) based on its labels and identifying properties, or creates it if it doesn't exist. Updates properties on match. Use this to add software architecture elements like Functions, Classes, Files, etc., to the graph without creating duplicates. Automatically handles 'createdAt' and 'lastModifiedAt' timestamps."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels for the entity (e.g., ['Function', 'Go']). Must include at least one label. Order typically doesn't matter."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties used to uniquely identify the entity for matching (e.g., {'filePath': '/path/to/file.go', 'name': 'MyFunc'}). Must include at least one property. These properties should correspond to constraints/indexes for performance."),
		),
		mcp.WithObject("properties",
			mcp.Required(),
			mcp.Description("Map of all properties (including identifying ones and any others like 'description', 'source', 'tags', 'language', 'signature', etc.) to set on create or merge/update on match. 'lastModifiedAt' will always be updated."),
		),
	)
	s.server.AddTool(findOrCreateEntityTool, s.handleFindOrCreateEntityTool)

	findOrCreateRelationshipTool := mcp.NewTool("find_or_create_relationship",
		mcp.WithDescription("Idempotently finds or creates a directed relationship between two existing nodes (identified by labels and properties), merging provided properties on the relationship. Use this to represent connections like CALLS, DEPENDS_ON, IMPLEMENTS etc. Automatically handles 'createdAt' and 'lastModifiedAt' timestamps for the relationship."),
		mcp.WithArray("startNodeLabels",
			mcp.Required(),
			mcp.Description("List of labels for the starting node of the relationship."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("startNodeIdentifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the starting node."),
		),
		mcp.WithArray("endNodeLabels",
			mcp.Required(),
			mcp.Description("List of labels for the ending node of the relationship."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("endNodeIdentifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the ending node."),
		),
		mcp.WithString("relationshipType",
			mcp.Required(),
			mcp.Description("The type name for the relationship (e.g., 'CALLS', 'DEPENDS_ON'). Must be a valid Neo4j relationship type name."),
		),
		mcp.WithObject("properties",
			mcp.Description("Optional map of properties to set on create or merge/update on match for the relationship itself (e.g., {'lineNumber': 123} for a CALLS relationship). 'lastModifiedAt' will always be updated."),
		),
	)
	s.server.AddTool(findOrCreateRelationshipTool, s.handleFindOrCreateRelationshipTool)

	getEntityDetailsTool := mcp.NewTool("get_entity_details",
		mcp.WithDescription("Retrieves the full labels and properties of a specific entity identified by its labels and unique properties."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels used to find the entity (e.g., ['Function'])."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties used to uniquely identify the entity (e.g., {'filePath': '/path/to/file.go', 'name': 'MyFunc'})."),
		),
	)
	s.server.AddTool(getEntityDetailsTool, s.handleGetEntityDetailsTool)

	findNeighborsTool := mcp.NewTool("find_neighbors",
		mcp.WithDescription("Finds the direct neighbors (nodes connected by a single relationship) of a specific entity, up to a specified depth. Returns the central node details and a list of neighbors including relationship type and direction."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels for the central entity to find neighbors for."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the central entity."),
		),
		mcp.WithNumber("maxDepth",
			mcp.Description("Maximum relationship path depth to search for neighbors (e.g., 1 for direct neighbors, 2 for neighbors-of-neighbors). Defaults to 1 if not provided or invalid."),
		),
	)
	s.server.AddTool(findNeighborsTool, s.handleFindNeighborsTool)

	findDependenciesTool := mcp.NewTool("find_dependencies",
		mcp.WithDescription("Finds entities that the target entity depends on by following outgoing relationships (e.g., A depends on B if A -> B). Allows filtering by relationship types and specifying search depth."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels for the target entity whose dependencies are being sought."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the target entity."),
		),
		mcp.WithArray("relationshipTypes",
			mcp.Description("Optional list of specific relationship types to follow when searching for dependencies (e.g., ['DEPENDS_ON', 'CALLS']). If omitted or empty, all outgoing relationship types will be followed."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithNumber("maxDepth",
			mcp.Description("Maximum relationship path depth to search for dependencies (e.g., 1 for direct dependencies). Defaults to 1 if not provided or invalid."),
		),
	)
	s.server.AddTool(findDependenciesTool, s.handleFindDependenciesTool)

	findDependentsTool := mcp.NewTool("find_dependents",
		mcp.WithDescription("Finds entities that depend on the target entity by following incoming relationships (e.g., B depends on A if B -> A). Allows filtering by relationship types and specifying search depth. Useful for impact analysis."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels for the target entity whose dependents are being sought."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the target entity."),
		),
		mcp.WithArray("relationshipTypes",
			mcp.Description("Optional list of specific relationship types to follow when searching for dependents (e.g., ['DEPENDS_ON', 'CALLS']). If omitted or empty, all incoming relationship types will be followed."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithNumber("maxDepth",
			mcp.Description("Maximum relationship path depth to search for dependents (e.g., 1 for direct dependents). Defaults to 1 if not provided or invalid."),
		),
	)
	s.server.AddTool(findDependentsTool, s.handleFindDependentsTool)

	getEntitySubgraphTool := mcp.NewTool("get_entity_subgraph",
		mcp.WithDescription("Retrieves a subgraph containing nodes and relationships within a specified depth around a central entity. The result format is designed for easy conversion into visualisation formats like Mermaid diagrams. Requires the APOC plugin to be installed on the Neo4j server."),
		mcp.WithArray("labels",
			mcp.Required(),
			mcp.Description("List of labels for the central entity of the subgraph."),
			mcp.Items(map[string]interface{}{"type": "string"}),
		),
		mcp.WithObject("identifyingProperties",
			mcp.Required(),
			mcp.Description("Map of properties to uniquely identify the central entity."),
		),
		mcp.WithNumber("maxDepth",
			mcp.Description("Maximum relationship path depth to include in the subgraph (e.g., 1 for direct neighbors, 2 includes neighbors-of-neighbors). Defaults to 1 if not provided or invalid."),
		),
	)
	s.server.AddTool(getEntitySubgraphTool, s.handleGetEntitySubgraphTool)

	// --- Batch Operation Tools ---

	batchFindOrCreateEntitiesToolTool := mcp.NewTool("batch_find_or_create_entities",
		mcp.WithDescription("Creates or updates multiple entities in a single operation. This is significantly more efficient than making individual calls, especially when creating many related entities. Use this to add multiple software architecture elements like Functions, Classes, Files, etc., to the graph in one request."),
		mcp.WithArray("entities",
			mcp.Required(),
			mcp.Description("Array of entity definitions to create or update. Each entity follows the same structure as the input to find_or_create_entity."),
			mcp.Items(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"labels": map[string]interface{}{
						"type": "array",
						"description": "List of labels for the entity (e.g., ['Function', 'Go']). Must include at least one label.",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"identifyingProperties": map[string]interface{}{
						"type": "object",
						"description": "Map of properties used to uniquely identify the entity for matching.",
					},
					"properties": map[string]interface{}{
						"type": "object",
						"description": "Map of all properties to set on create or merge/update on match.",
					},
				},
				"required": []string{"labels", "identifyingProperties", "properties"},
			}),
		),
	)
	s.server.AddTool(batchFindOrCreateEntitiesToolTool, s.handleBatchFindOrCreateEntitiesToolTool)

	batchFindOrCreateRelationshipsToolTool := mcp.NewTool("batch_find_or_create_relationships",
		mcp.WithDescription("Creates or updates multiple relationships in a single operation. This is significantly more efficient than making individual calls, especially when creating many relationships between entities. Use this to add multiple connections like CALLS, DEPENDS_ON, IMPLEMENTS, etc., in one request."),
		mcp.WithArray("relationships",
			mcp.Required(),
			mcp.Description("Array of relationship definitions to create or update. Each relationship follows the same structure as the input to find_or_create_relationship."),
			mcp.Items(map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"startNodeLabels": map[string]interface{}{
						"type": "array",
						"description": "List of labels for the starting node of the relationship.",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"startNodeIdentifyingProperties": map[string]interface{}{
						"type": "object",
						"description": "Map of properties to uniquely identify the starting node.",
					},
					"endNodeLabels": map[string]interface{}{
						"type": "array",
						"description": "List of labels for the ending node of the relationship.",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"endNodeIdentifyingProperties": map[string]interface{}{
						"type": "object",
						"description": "Map of properties to uniquely identify the ending node.",
					},
					"relationshipType": map[string]interface{}{
						"type": "string",
						"description": "The type name for the relationship (e.g., 'CALLS', 'DEPENDS_ON').",
					},
					"properties": map[string]interface{}{
						"type": "object",
						"description": "Optional map of properties to set on the relationship.",
					},
				},
				"required": []string{"startNodeLabels", "startNodeIdentifyingProperties", "endNodeLabels", "endNodeIdentifyingProperties", "relationshipType"},
			}),
		),
	)
	s.server.AddTool(batchFindOrCreateRelationshipsToolTool, s.handleBatchFindOrCreateRelationshipsToolTool)
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

// handleBatchFindOrCreateEntitiesToolTool handles the batch_find_or_create_entities tool
func (s *Server) handleBatchFindOrCreateEntitiesToolTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse entities array
	entitiesArg, ok := request.Params.Arguments["entities"]
	if !ok {
		return nil, errors.New("entities array is required")
	}

	entitiesInterface, ok := entitiesArg.([]interface{})
	if !ok {
		return nil, errors.New("entities must be an array")
	}

	if len(entitiesInterface) == 0 {
		return nil, errors.New("at least one entity is required")
	}

	// Convert to EntityInput array
	inputs := make([]graph.EntityInput, len(entitiesInterface))
	for i, entityInterface := range entitiesInterface {
		entityMap, ok := entityInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("entity at index %d is not an object", i)
		}

		// Parse labels
		labelsArg, ok := entityMap["labels"]
		if !ok {
			return nil, fmt.Errorf("labels are required for entity at index %d", i)
		}
		labelsInterface, ok := labelsArg.([]interface{})
		if !ok {
			return nil, fmt.Errorf("labels must be an array of strings for entity at index %d", i)
		}

		inputs[i].Labels = make([]string, len(labelsInterface))
		for j, l := range labelsInterface {
			inputs[i].Labels[j], ok = l.(string)
			if !ok {
				return nil, fmt.Errorf("label item at index %d for entity at index %d is not a string", j, i)
			}
		}
		if len(inputs[i].Labels) == 0 {
			return nil, fmt.Errorf("at least one label is required for entity at index %d", i)
		}

		// Parse identifyingProperties
		idPropsArg, ok := entityMap["identifyingProperties"]
		if !ok {
			return nil, fmt.Errorf("identifyingProperties are required for entity at index %d", i)
		}
		inputs[i].IdentifyingProperties, ok = idPropsArg.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("identifyingProperties must be an object for entity at index %d", i)
		}
		if len(inputs[i].IdentifyingProperties) == 0 {
			return nil, fmt.Errorf("at least one identifying property is required for entity at index %d", i)
		}

		// Parse properties
		propsArg, ok := entityMap["properties"]
		if !ok {
			return nil, fmt.Errorf("properties are required for entity at index %d", i)
		}
		inputs[i].Properties, ok = propsArg.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("properties must be an object for entity at index %d", i)
		}
	}

	// Call graph store method
	results, individualErrors, err := s.graph.BatchFindOrCreateEntities(ctx, inputs)

	// Prepare response
	response := struct {
		Results         []graph.EntityDetails `json:"results"`
		IndividualErrors []string             `json:"individualErrors,omitempty"`
		Error           string                `json:"error,omitempty"`
	}{
		Results: results,
	}

	// Handle individual errors
	if individualErrors != nil {
		errorMessages := make([]string, 0)
		for i, err := range individualErrors {
			if err != nil {
				errorMessages = append(errorMessages, fmt.Sprintf("Error at index %d: %s", i, err.Error()))
			}
		}
		if len(errorMessages) > 0 {
			response.IndividualErrors = errorMessages
		}
	}

	// Handle overall error
	if err != nil {
		response.Error = err.Error()
	}

	// Return the results
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch results: %w", err)
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// handleBatchFindOrCreateRelationshipsToolTool handles the batch_find_or_create_relationships tool
func (s *Server) handleBatchFindOrCreateRelationshipsToolTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse relationships array
	relationshipsArg, ok := request.Params.Arguments["relationships"]
	if !ok {
		return nil, errors.New("relationships array is required")
	}

	relationshipsInterface, ok := relationshipsArg.([]interface{})
	if !ok {
		return nil, errors.New("relationships must be an array")
	}

	if len(relationshipsInterface) == 0 {
		return nil, errors.New("at least one relationship is required")
	}

	// Convert to RelationshipInput array
	inputs := make([]graph.RelationshipInput, len(relationshipsInterface))
	for i, relInterface := range relationshipsInterface {
		relMap, ok := relInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("relationship at index %d is not an object", i)
		}

		// Parse startNodeLabels
		startLabelsArg, ok := relMap["startNodeLabels"]
		if !ok {
			return nil, fmt.Errorf("startNodeLabels are required for relationship at index %d", i)
		}
		startLabelsInterface, ok := startLabelsArg.([]interface{})
		if !ok {
			return nil, fmt.Errorf("startNodeLabels must be an array of strings for relationship at index %d", i)
		}

		inputs[i].StartNodeLabels = make([]string, len(startLabelsInterface))
		for j, l := range startLabelsInterface {
			inputs[i].StartNodeLabels[j], ok = l.(string)
			if !ok {
				return nil, fmt.Errorf("startNodeLabel item at index %d for relationship at index %d is not a string", j, i)
			}
		}
		if len(inputs[i].StartNodeLabels) == 0 {
			return nil, fmt.Errorf("at least one startNodeLabel is required for relationship at index %d", i)
		}

		// Parse startNodeIdentifyingProperties
		startIdPropsArg, ok := relMap["startNodeIdentifyingProperties"]
		if !ok {
			return nil, fmt.Errorf("startNodeIdentifyingProperties are required for relationship at index %d", i)
		}
		inputs[i].StartNodeIdentifyingProperties, ok = startIdPropsArg.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("startNodeIdentifyingProperties must be an object for relationship at index %d", i)
		}
		if len(inputs[i].StartNodeIdentifyingProperties) == 0 {
			return nil, fmt.Errorf("at least one startNodeIdentifyingProperty is required for relationship at index %d", i)
		}

		// Parse endNodeLabels
		endLabelsArg, ok := relMap["endNodeLabels"]
		if !ok {
			return nil, fmt.Errorf("endNodeLabels are required for relationship at index %d", i)
		}
		endLabelsInterface, ok := endLabelsArg.([]interface{})
		if !ok {
			return nil, fmt.Errorf("endNodeLabels must be an array of strings for relationship at index %d", i)
		}

		inputs[i].EndNodeLabels = make([]string, len(endLabelsInterface))
		for j, l := range endLabelsInterface {
			inputs[i].EndNodeLabels[j], ok = l.(string)
			if !ok {
				return nil, fmt.Errorf("endNodeLabel item at index %d for relationship at index %d is not a string", j, i)
			}
		}
		if len(inputs[i].EndNodeLabels) == 0 {
			return nil, fmt.Errorf("at least one endNodeLabel is required for relationship at index %d", i)
		}

		// Parse endNodeIdentifyingProperties
		endIdPropsArg, ok := relMap["endNodeIdentifyingProperties"]
		if !ok {
			return nil, fmt.Errorf("endNodeIdentifyingProperties are required for relationship at index %d", i)
		}
		inputs[i].EndNodeIdentifyingProperties, ok = endIdPropsArg.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("endNodeIdentifyingProperties must be an object for relationship at index %d", i)
		}
		if len(inputs[i].EndNodeIdentifyingProperties) == 0 {
			return nil, fmt.Errorf("at least one endNodeIdentifyingProperty is required for relationship at index %d", i)
		}

		// Parse relationshipType
		relTypeArg, ok := relMap["relationshipType"]
		if !ok {
			return nil, fmt.Errorf("relationshipType is required for relationship at index %d", i)
		}
		inputs[i].RelationshipType, ok = relTypeArg.(string)
		if !ok || inputs[i].RelationshipType == "" {
			return nil, fmt.Errorf("relationshipType must be a non-empty string for relationship at index %d", i)
		}

		// Parse properties (optional)
		if propsArg, exists := relMap["properties"]; exists && propsArg != nil {
			inputs[i].Properties, ok = propsArg.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("properties must be an object for relationship at index %d", i)
			}
		} else {
			inputs[i].Properties = make(map[string]interface{}) // Ensure it's not nil
		}
	}

	// Call graph store method
	results, individualErrors, err := s.graph.BatchFindOrCreateRelationships(ctx, inputs)

	// Prepare response
	response := struct {
		Results         []map[string]interface{} `json:"results"`
		IndividualErrors []string                `json:"individualErrors,omitempty"`
		Error           string                   `json:"error,omitempty"`
	}{
		Results: results,
	}

	// Handle individual errors
	if individualErrors != nil {
		errorMessages := make([]string, 0)
		for i, err := range individualErrors {
			if err != nil {
				errorMessages = append(errorMessages, fmt.Sprintf("Error at index %d: %s", i, err.Error()))
			}
		}
		if len(errorMessages) > 0 {
			response.IndividualErrors = errorMessages
		}
	}

	// Handle overall error
	if err != nil {
		response.Error = err.Error()
	}

	// Return the results
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch results: %w", err)
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
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
