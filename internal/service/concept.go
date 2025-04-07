package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/sammcj/mcp-graph/internal/graph"
)

// CreateConcept creates a new concept in the knowledge graph
func (s *Service) CreateConcept(ctx context.Context, name string, properties map[string]interface{}) (string, error) {
	// Create properties map
	conceptProps := map[string]interface{}{
		"name": name,
	}

	// Add additional properties if provided
	if properties != nil {
		for k, v := range properties {
			conceptProps[k] = v
		}
	}

	// Create concept node
	return s.graph.CreateNode(ctx, string(graph.NodeTypeConcept), conceptProps)
}

// GetConcept retrieves a concept by ID
func (s *Service) GetConcept(ctx context.Context, id string) (*Concept, error) {
	// Get node from graph
	node, err := s.graph.GetNode(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get concept: %w", err)
	}

	// Check node type
	nodeType, ok := node["type"].(string)
	if !ok || nodeType != string(graph.NodeTypeConcept) {
		return nil, errors.New("node is not a concept")
	}

	// Extract concept properties
	name, _ := node["name"].(string)

	// Extract other properties
	properties := make(map[string]interface{})
	for k, v := range node {
		// Skip internal properties
		if k == "uid" || k == "type" || k == "name" {
			continue
		}
		properties[k] = v
	}

	// Create concept
	return &Concept{
		ID:         id,
		Name:       name,
		Properties: properties,
	}, nil
}

// LinkConcepts creates a relationship between two concepts
func (s *Service) LinkConcepts(ctx context.Context, fromID, toID string, relationshipType string, properties map[string]interface{}) (string, error) {
	// Create edge
	return s.graph.CreateEdge(ctx, fromID, toID, relationshipType, properties)
}

// SearchConcepts searches for concepts matching the query
func (s *Service) SearchConcepts(ctx context.Context, query string) ([]*Concept, error) {
	// Create GraphQL query
	graphQuery := `
		{
			concepts(func: type(Concept)) @filter(anyoftext(name, "` + query + `")) {
				uid
				name
				expand(_all_) {
					uid
					name
				}
			}
		}
	`

	// Execute query
	results, err := s.graph.Query(ctx, graphQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results to concepts
	concepts := make([]*Concept, 0, len(results))
	for _, result := range results {
		id, _ := result["uid"].(string)
		name, _ := result["name"].(string)

		// Extract other properties
		properties := make(map[string]interface{})
		for k, v := range result {
			// Skip internal properties
			if k == "uid" || k == "type" || k == "name" {
				continue
			}
			properties[k] = v
		}

		concepts = append(concepts, &Concept{
			ID:         id,
			Name:       name,
			Properties: properties,
		})
	}

	return concepts, nil
}
