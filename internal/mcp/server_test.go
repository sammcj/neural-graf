package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/sammcj/mcp-graph/internal/mcp/mocks"
	"github.com/sammcj/mcp-graph/internal/service"
)

// TestHandleQueryTool tests the query_knowledge_graph tool handler
func TestHandleQueryTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test query
	query := "{ documents(func: type(Document)) { uid title content } }"
	params := map[string]interface{}{
		"limit": 10,
	}

	// Mock query results
	mockResults := []map[string]interface{}{
		{
			"uid":     "0x1",
			"title":   "Document 1",
			"content": "Content 1",
		},
		{
			"uid":     "0x2",
			"title":   "Document 2",
			"content": "Content 2",
		},
	}

	// Set up expectations
	mockGraph.EXPECT().Query(gomock.Any(), gomock.Eq(query), gomock.Eq(params)).Return(mockResults, nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"query":  query,
		"params": params,
	}

	// Call the handler
	result, err := server.handleQueryTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Use reflection to access the fields
	assert.Len(t, result.Content, 1, "Expected one content item")

	// Get the content item
	contentVal := reflect.ValueOf(result.Content[0])

	// Get the Type field
	typeField := contentVal.FieldByName("Type")
	assert.True(t, typeField.IsValid(), "Type field not found")
	assert.Equal(t, "text", typeField.String(), "Expected content type to be text")

	// Get the Text field
	textField := contentVal.FieldByName("Text")
	assert.True(t, textField.IsValid(), "Text field not found")
	resultText := textField.String()

	// Verify the result content
	var resultData []map[string]interface{}
	err = json.Unmarshal([]byte(resultText), &resultData)
	assert.NoError(t, err)
	assert.Len(t, resultData, 2)
	assert.Equal(t, "0x1", resultData[0]["uid"])
	assert.Equal(t, "Document 1", resultData[0]["title"])
	assert.Equal(t, "Content 1", resultData[0]["content"])
}

// Helper function to extract text from a CallToolResult
func getResultText(result *mcp.CallToolResult) string {
	contentVal := reflect.ValueOf(result.Content[0])
	textField := contentVal.FieldByName("Text")
	return textField.String()
}

// TestHandleCreateDocumentTool tests the create_document tool handler
func TestHandleCreateDocumentTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test document data
	title := "Test Document"
	content := "This is a test document"
	metadata := map[string]interface{}{
		"author": "Test Author",
		"tags":   []string{"test", "document"},
	}

	// Set up expectations
	mockService.EXPECT().CreateDocument(
		gomock.Any(),
		gomock.Eq(title),
		gomock.Eq(content),
		gomock.Eq(metadata),
	).Return("0x1", nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"title":    title,
		"content":  content,
		"metadata": metadata,
	}

	// Call the handler
	result, err := server.handleCreateDocumentTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get the result text
	resultText := getResultText(result)
	assert.Equal(t, `{"id":"0x1"}`, resultText)
}

// TestHandleGetDocumentTool tests the get_document tool handler
func TestHandleGetDocumentTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test document
	document := &service.Document{
		ID:      "0x1",
		Title:   "Test Document",
		Content: "This is a test document",
		Metadata: map[string]interface{}{
			"author": "Test Author",
			"tags":   []string{"test", "document"},
		},
	}

	// Set up expectations
	mockService.EXPECT().GetDocument(gomock.Any(), gomock.Eq("0x1")).Return(document, nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"id": "0x1",
	}

	// Call the handler
	result, err := server.handleGetDocumentTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get the result text
	resultText := getResultText(result)

	// Verify the result content
	var resultDoc map[string]interface{}
	err = json.Unmarshal([]byte(resultText), &resultDoc)
	assert.NoError(t, err)
	assert.Equal(t, "0x1", resultDoc["id"])
	assert.Equal(t, "Test Document", resultDoc["title"])
	assert.Equal(t, "This is a test document", resultDoc["content"])
	assert.Equal(t, "Test Author", resultDoc["metadata"].(map[string]interface{})["author"])
}

// TestHandleSearchDocumentsTool tests the search_documents tool handler
func TestHandleSearchDocumentsTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test search query
	query := "test"

	// Test search results
	documents := []*service.Document{
		{
			ID:      "0x1",
			Title:   "Test Document 1",
			Content: "This is test document 1",
		},
		{
			ID:      "0x2",
			Title:   "Test Document 2",
			Content: "This is test document 2",
		},
	}

	// Set up expectations
	mockService.EXPECT().SearchDocuments(gomock.Any(), gomock.Eq(query)).Return(documents, nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"query": query,
	}

	// Call the handler
	result, err := server.handleSearchDocumentsTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get the result text
	resultText := getResultText(result)

	// Verify the result content
	var resultDocs []map[string]interface{}
	err = json.Unmarshal([]byte(resultText), &resultDocs)
	assert.NoError(t, err)
	assert.Len(t, resultDocs, 2)
	assert.Equal(t, "0x1", resultDocs[0]["id"])
	assert.Equal(t, "Test Document 1", resultDocs[0]["title"])
	assert.Equal(t, "This is test document 1", resultDocs[0]["content"])
	assert.Equal(t, "0x2", resultDocs[1]["id"])
}

// TestHandleCreateConceptTool tests the create_concept tool handler
func TestHandleCreateConceptTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test concept data
	name := "Test Concept"
	properties := map[string]interface{}{
		"description": "A test concept",
		"category":    "test",
	}

	// Set up expectations
	mockService.EXPECT().CreateConcept(
		gomock.Any(),
		gomock.Eq(name),
		gomock.Eq(properties),
	).Return("0x1", nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"name":       name,
		"properties": properties,
	}

	// Call the handler
	result, err := server.handleCreateConceptTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get the result text
	resultText := getResultText(result)
	assert.Equal(t, `{"id":"0x1"}`, resultText)
}

// TestHandleGetConceptTool tests the get_concept tool handler
func TestHandleGetConceptTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test concept
	concept := &service.Concept{
		ID:   "0x1",
		Name: "Test Concept",
		Properties: map[string]interface{}{
			"description": "A test concept",
			"category":    "test",
		},
	}

	// Set up expectations
	mockService.EXPECT().GetConcept(gomock.Any(), gomock.Eq("0x1")).Return(concept, nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"id": "0x1",
	}

	// Call the handler
	result, err := server.handleGetConceptTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get the result text
	resultText := getResultText(result)

	// Verify the result content
	var resultConcept map[string]interface{}
	err = json.Unmarshal([]byte(resultText), &resultConcept)
	assert.NoError(t, err)
	assert.Equal(t, "0x1", resultConcept["id"])
	assert.Equal(t, "Test Concept", resultConcept["name"])
	assert.Equal(t, "A test concept", resultConcept["properties"].(map[string]interface{})["description"])
	assert.Equal(t, "test", resultConcept["properties"].(map[string]interface{})["category"])
}

// TestHandleLinkConceptsTool tests the link_concepts tool handler
func TestHandleLinkConceptsTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test link data
	fromID := "0x1"
	toID := "0x2"
	relationType := "RELATED_TO"
	properties := map[string]interface{}{
		"strength": 0.8,
		"source":   "user",
	}

	// Set up expectations
	mockService.EXPECT().LinkConcepts(
		gomock.Any(),
		gomock.Eq(fromID),
		gomock.Eq(toID),
		gomock.Eq(relationType),
		gomock.Eq(properties),
	).Return("0x3", nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"fromId":           fromID,
		"toId":             toID,
		"relationshipType": relationType,
		"properties":       properties,
	}

	// Call the handler
	result, err := server.handleLinkConceptsTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get the result text
	resultText := getResultText(result)
	assert.Equal(t, `{"id":"0x3"}`, resultText)
}

// TestHandleCreateNodeTool tests the create_node tool handler (legacy)
func TestHandleCreateNodeTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test node data
	nodeType := "Document"
	properties := map[string]interface{}{
		"title":   "Test Document",
		"content": "This is a test document",
	}

	// Set up expectations
	mockGraph.EXPECT().CreateNode(
		gomock.Any(),
		gomock.Eq(nodeType),
		gomock.Eq(properties),
	).Return("0x1", nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"type":       nodeType,
		"properties": properties,
	}

	// Call the handler
	result, err := server.handleCreateNodeTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get the result text
	resultText := getResultText(result)
	assert.Equal(t, `{"id":"0x1"}`, resultText)
}

// TestHandleGetNodeTool tests the get_node tool handler (legacy)
func TestHandleGetNodeTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test node data
	node := map[string]interface{}{
		"uid":     "0x1",
		"type":    "Document",
		"title":   "Test Document",
		"content": "This is a test document",
	}

	// Set up expectations
	mockGraph.EXPECT().GetNode(gomock.Any(), gomock.Eq("0x1")).Return(node, nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"id": "0x1",
	}

	// Call the handler
	result, err := server.handleGetNodeTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get the result text
	resultText := getResultText(result)

	// Verify the result content
	var resultNode map[string]interface{}
	err = json.Unmarshal([]byte(resultText), &resultNode)
	assert.NoError(t, err)
	assert.Equal(t, "0x1", resultNode["uid"])
	assert.Equal(t, "Document", resultNode["type"])
	assert.Equal(t, "Test Document", resultNode["title"])
	assert.Equal(t, "This is a test document", resultNode["content"])
}

// TestHandleCreateEdgeTool tests the create_edge tool handler (legacy)
func TestHandleCreateEdgeTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test edge data
	fromID := "0x1"
	toID := "0x2"
	edgeType := "RELATED_TO"
	properties := map[string]interface{}{
		"strength": 0.8,
	}

	// Set up expectations
	mockGraph.EXPECT().CreateEdge(
		gomock.Any(),
		gomock.Eq(fromID),
		gomock.Eq(toID),
		gomock.Eq(edgeType),
		gomock.Eq(properties),
	).Return("0x3", nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"fromId":     fromID,
		"toId":       toID,
		"type":       edgeType,
		"properties": properties,
	}

	// Call the handler
	result, err := server.handleCreateEdgeTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get the result text
	resultText := getResultText(result)
	assert.Equal(t, `{"id":"0x3"}`, resultText)
}

// TestHandleSchemaTool tests the upsert_schema tool handler
func TestHandleSchemaTool(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test schema
	schema := `
		type: string @index(exact) .
		title: string @index(fulltext, term) .
	`

	// Set up expectations
	mockGraph.EXPECT().UpsertSchema(gomock.Any(), gomock.Eq(schema)).Return(nil)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"schema": schema,
	}

	// Call the handler
	result, err := server.handleSchemaTool(context.Background(), request)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Get the result text
	resultText := getResultText(result)
	assert.Equal(t, `{"success":true}`, resultText)
}

// TestHandleCreateDocumentTool_Error tests the create_document tool handler with an error
func TestHandleCreateDocumentTool_Error(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test document data
	title := "Test Document"
	content := "This is a test document"
	metadata := map[string]interface{}{
		"author": "Test Author",
		"tags":   []string{"test", "document"},
	}

	// Set up expectations with error
	mockError := errors.New("failed to create document")
	mockService.EXPECT().CreateDocument(
		gomock.Any(),
		gomock.Eq(title),
		gomock.Eq(content),
		gomock.Eq(metadata),
	).Return("", mockError)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"title":    title,
		"content":  content,
		"metadata": metadata,
	}

	// Call the handler
	_, err := server.handleCreateDocumentTool(context.Background(), request)

	// Assert the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create document")
}

// TestHandleGetDocumentTool_Error tests the get_document tool handler with an error
func TestHandleGetDocumentTool_Error(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Set up expectations with error
	mockError := errors.New("document not found")
	mockService.EXPECT().GetDocument(gomock.Any(), gomock.Eq("0x1")).Return(nil, mockError)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"id": "0x1",
	}

	// Call the handler
	_, err := server.handleGetDocumentTool(context.Background(), request)

	// Assert the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get document")
}

// TestHandleCreateConceptTool_Error tests the create_concept tool handler with an error
func TestHandleCreateConceptTool_Error(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test concept data
	name := "Test Concept"
	properties := map[string]interface{}{
		"description": "A test concept",
		"category":    "test",
	}

	// Set up expectations with error
	mockError := errors.New("failed to create concept")
	mockService.EXPECT().CreateConcept(
		gomock.Any(),
		gomock.Eq(name),
		gomock.Eq(properties),
	).Return("", mockError)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"name":       name,
		"properties": properties,
	}

	// Call the handler
	_, err := server.handleCreateConceptTool(context.Background(), request)

	// Assert the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create concept")
}

// TestHandleSchemaTool_Error tests the upsert_schema tool handler with an error
func TestHandleSchemaTool_Error(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test schema
	schema := `
		type: string @index(exact) .
		title: string @index(fulltext, term) .
	`

	// Set up expectations with error
	mockError := errors.New("failed to update schema")
	mockGraph.EXPECT().UpsertSchema(gomock.Any(), gomock.Eq(schema)).Return(mockError)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"schema": schema,
	}

	// Call the handler
	_, err := server.handleSchemaTool(context.Background(), request)

	// Assert the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update schema")
}

// TestHandleQueryTool_MissingQuery tests the query_knowledge_graph tool handler with a missing query parameter
func TestHandleQueryTool_MissingQuery(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Create tool request with missing query
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"params": map[string]interface{}{
			"limit": 10,
		},
	}

	// Call the handler
	_, err := server.handleQueryTool(context.Background(), request)

	// Assert the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query must be a string")
}

// TestHandleCreateDocumentTool_MissingTitle tests the create_document tool handler with a missing title parameter
func TestHandleCreateDocumentTool_MissingTitle(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Create tool request with missing title
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"content": "This is a test document",
		"metadata": map[string]interface{}{
			"author": "Test Author",
		},
	}

	// Call the handler
	_, err := server.handleCreateDocumentTool(context.Background(), request)

	// Assert the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title must be a string")
}

// TestHandleCreateConceptTool_MissingName tests the create_concept tool handler with a missing name parameter
func TestHandleCreateConceptTool_MissingName(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Create tool request with missing name
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"properties": map[string]interface{}{
			"description": "A test concept",
		},
	}

	// Call the handler
	_, err := server.handleCreateConceptTool(context.Background(), request)

	// Assert the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name must be a string")
}

// TestHandleQueryTool_Error tests the query_knowledge_graph tool handler with an error
func TestHandleQueryTool_Error(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock graph store and service
	mockGraph := mocks.NewMockStore(ctrl)
	mockService := mocks.NewMockKnowledgeManager(ctrl)

	// Create MCP server with mocks
	server := &Server{
		graph:   mockGraph,
		service: mockService,
	}

	// Test query
	query := "{ documents(func: type(Document)) { uid title content } }"

	// Set up expectations with error
	mockError := errors.New("query failed")
	mockGraph.EXPECT().Query(gomock.Any(), gomock.Eq(query), gomock.Any()).Return(nil, mockError)

	// Create tool request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"query": query,
	}

	// Call the handler
	_, err := server.handleQueryTool(context.Background(), request)

	// Assert the error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query failed")
}
