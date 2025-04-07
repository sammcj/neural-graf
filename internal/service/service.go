package service

import (
	"context"

	"github.com/sammcj/mcp-graph/internal/graph"
)

// KnowledgeManager defines the high-level operations for managing the knowledge graph
type KnowledgeManager interface {
	// Document operations
	CreateDocument(ctx context.Context, title, content string, metadata map[string]interface{}) (string, error)
	GetDocument(ctx context.Context, id string) (*Document, error)
	UpdateDocument(ctx context.Context, id, title, content string, metadata map[string]interface{}) error
	DeleteDocument(ctx context.Context, id string) error

	// Concept operations
	CreateConcept(ctx context.Context, name string, properties map[string]interface{}) (string, error)
	GetConcept(ctx context.Context, id string) (*Concept, error)
	LinkConcepts(ctx context.Context, fromID, toID string, relationshipType string, properties map[string]interface{}) (string, error)

	// Search operations
	SearchDocuments(ctx context.Context, query string) ([]*Document, error)
	SearchConcepts(ctx context.Context, query string) ([]*Concept, error)

	// Schema operations
	InitialiseSchema(ctx context.Context) error
}

// Document represents a document in the knowledge graph
type Document struct {
	ID       string                 `json:"id"`
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Concept represents a concept in the knowledge graph
type Concept struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Service implements the KnowledgeManager interface
type Service struct {
	graph graph.Store
}

// NewService creates a new knowledge manager service
func NewService(graph graph.Store) *Service {
	return &Service{
		graph: graph,
	}
}
