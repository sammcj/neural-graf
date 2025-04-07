package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sammcj/mcp-graph/internal/graph"
)

// CreateDocument creates a new document in the knowledge graph
func (s *Service) CreateDocument(ctx context.Context, title, content string, metadata map[string]interface{}) (string, error) {
	// Create properties map
	properties := map[string]interface{}{
		"title":   title,
		"content": content,
	}

	// Add metadata if provided
	if metadata != nil {
		properties["metadata"] = metadata
	}

	// Create document node
	return s.graph.CreateNode(ctx, string(graph.NodeTypeDocument), properties)
}

// GetDocument retrieves a document by ID
func (s *Service) GetDocument(ctx context.Context, id string) (*Document, error) {
	// Get node from graph
	node, err := s.graph.GetNode(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// Check node type
	nodeType, ok := node["type"].(string)
	if !ok || nodeType != string(graph.NodeTypeDocument) {
		return nil, errors.New("node is not a document")
	}

	// Extract document properties
	title, _ := node["title"].(string)
	content, _ := node["content"].(string)

	// Extract metadata if present
	var metadata map[string]interface{}
	if metadataVal, ok := node["metadata"]; ok {
		switch m := metadataVal.(type) {
		case map[string]interface{}:
			metadata = m
		case string:
			// If metadata is stored as a JSON string, unmarshal it
			if err := json.Unmarshal([]byte(m), &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}
	}

	// Create document
	return &Document{
		ID:       id,
		Title:    title,
		Content:  content,
		Metadata: metadata,
	}, nil
}

// UpdateDocument updates a document's properties
func (s *Service) UpdateDocument(ctx context.Context, id, title, content string, metadata map[string]interface{}) error {
	// Create properties map
	properties := map[string]interface{}{
		"title":   title,
		"content": content,
	}

	// Add metadata if provided
	if metadata != nil {
		properties["metadata"] = metadata
	}

	// Update node
	return s.graph.UpdateNode(ctx, id, properties)
}

// DeleteDocument deletes a document by ID
func (s *Service) DeleteDocument(ctx context.Context, id string) error {
	// Delete node
	return s.graph.DeleteNode(ctx, id)
}

// SearchDocuments searches for documents matching the query
func (s *Service) SearchDocuments(ctx context.Context, query string) ([]*Document, error) {
	// Create GraphQL query
	graphQuery := `
		{
			documents(func: type(Document)) @filter(anyoftext(title, content, "` + query + `")) {
				uid
				title
				content
				metadata
			}
		}
	`

	// Execute query
	results, err := s.graph.Query(ctx, graphQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results to documents
	documents := make([]*Document, 0, len(results))
	for _, result := range results {
		id, _ := result["uid"].(string)
		title, _ := result["title"].(string)
		content, _ := result["content"].(string)

		// Extract metadata if present
		var metadata map[string]interface{}
		if metadataVal, ok := result["metadata"]; ok {
			switch m := metadataVal.(type) {
			case map[string]interface{}:
				metadata = m
			case string:
				// If metadata is stored as a JSON string, unmarshal it
				if err := json.Unmarshal([]byte(m), &metadata); err != nil {
					return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
				}
			}
		}

		documents = append(documents, &Document{
			ID:       id,
			Title:    title,
			Content:  content,
			Metadata: metadata,
		})
	}

	return documents, nil
}
