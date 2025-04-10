package graph

import "context"

// Store defines the core knowledge graph operations
type Store interface {
	// Node operations
	CreateNode(ctx context.Context, nodeType string, properties map[string]interface{}) (string, error)
	GetNode(ctx context.Context, id string) (map[string]interface{}, error)
	UpdateNode(ctx context.Context, id string, properties map[string]interface{}) error
	DeleteNode(ctx context.Context, id string) error

	// Edge operations
	CreateEdge(ctx context.Context, fromID, toID, relationshipType string, properties map[string]interface{}) (string, error)
	GetEdge(ctx context.Context, id string) (map[string]interface{}, error)
	UpdateEdge(ctx context.Context, id string, properties map[string]interface{}) error
	DeleteEdge(ctx context.Context, id string) error

	// Query operations
	Query(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error)

	// Schema operations
	UpsertSchema(ctx context.Context, schema string) error

	// --- Software Architecture Specific Operations ---

	// FindOrCreateEntity finds an entity based on identifying properties or creates it if not found.
	// It merges the provided properties with any existing ones.
	// Returns the details of the found or created entity.
	FindOrCreateEntity(ctx context.Context, input EntityInput) (EntityDetails, error)

	// FindOrCreateRelationship finds a relationship or creates it if not found.
	// It merges the provided properties with any existing ones.
	// Returns the properties of the found or created relationship.
	FindOrCreateRelationship(ctx context.Context, input RelationshipInput) (map[string]interface{}, error)

	// GetEntityDetails retrieves the labels and properties of a specific entity.
	GetEntityDetails(ctx context.Context, labels []string, identifyingProperties map[string]interface{}) (EntityDetails, error)

	// FindNeighbors finds the direct neighbors of a given entity up to a specified depth (depth 1 for direct neighbors).
	FindNeighbors(ctx context.Context, labels []string, identifyingProperties map[string]interface{}, maxDepth int) (NeighborsResult, error)
}

// NodeType represents common node types in the knowledge graph
type NodeType string

// Common node types
const (
	NodeTypeDocument NodeType = "Document"
	NodeTypeConcept  NodeType = "Concept"
	NodeTypeEntity   NodeType = "Entity"
	NodeTypeEvent    NodeType = "Event"
)

// EdgeType represents common edge types in the knowledge graph
type EdgeType string

// Common edge types
const (
	EdgeTypeRelatedTo    EdgeType = "RELATED_TO"
	EdgeTypeContains     EdgeType = "CONTAINS"
	EdgeTypeReferencesTo EdgeType = "REFERENCES_TO"
	EdgeTypeCreatedBy    EdgeType = "CREATED_BY"
	EdgeTypeHasProperty  EdgeType = "HAS_PROPERTY"
)

// Node represents a node in the knowledge graph
type Node struct {
	ID         string                 `json:"id"`
	Type       NodeType               `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// Edge represents an edge in the knowledge graph
type Edge struct {
	ID         string                 `json:"id"`
	FromID     string                 `json:"fromId"`
	ToID       string                 `json:"toId"`
	Type       EdgeType               `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

// QueryResult represents a result from a graph query
type QueryResult struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}
