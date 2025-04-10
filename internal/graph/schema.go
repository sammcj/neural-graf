package graph

import "time"

// --- Base & Common Structures ---

// BaseNode represents common properties for all graph nodes.
type BaseNode struct {
	ID             string    `json:"id"` // Unique identifier (often Neo4j internal ID or a generated UUID)
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Source         string    `json:"source"` // e.g., 'manual', 'static-analysis', 'agent-inference'
	Tags           []string  `json:"tags,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	LastModifiedAt time.Time `json:"lastModifiedAt"`
	Confidence     float64   `json:"confidence,omitempty"` // 0.0 - 1.0
	Status         string    `json:"status,omitempty"`     // e.g., 'stub', 'partially_analysed', 'fully_analysed', 'deprecated'
}

// BaseRelationship represents common properties for relationships.
// Specific relationship types might embed this or be represented directly.
type BaseRelationship struct {
	ID             string    `json:"id"` // Unique identifier (often Neo4j internal ID)
	Source         string    `json:"source"`
	CreatedAt      time.Time `json:"createdAt"`
	LastModifiedAt time.Time `json:"lastModifiedAt"`
	Confidence     float64   `json:"confidence,omitempty"`
}

// --- Node Type Definitions ---
// Using struct embedding for common properties.

type Application struct {
	BaseNode
	OwnerTeam string `json:"ownerTeam,omitempty"`
}

type Repository struct {
	BaseNode
	URL           string `json:"url"` // Required identifying property
	DefaultBranch string `json:"defaultBranch,omitempty"`
}

type Module struct {
	BaseNode
	Language string `json:"language,omitempty"`
	Version  string `json:"version,omitempty"`
	FilePath string `json:"filePath"` // Required identifying property (path to definition file, e.g., go.mod)
}

type Component struct {
	BaseNode
	Language  string `json:"language,omitempty"`
	FilePath  string `json:"filePath,omitempty"` // Primary path
	OwnerTeam string `json:"ownerTeam,omitempty"`
}

type Service struct {
	BaseNode
	Language            string `json:"language,omitempty"`
	FilePath            string `json:"filePath,omitempty"`
	OwnerTeam           string `json:"ownerTeam,omitempty"`
	APIEndpoint         string `json:"apiEndpoint,omitempty"`
	CommunicationProtocol string `json:"communicationProtocol,omitempty"`
}

type Library struct {
	BaseNode
	Language   string `json:"language,omitempty"`
	Version    string `json:"version,omitempty"` // Part of identifying properties
	GroupID    string `json:"groupId,omitempty"` // Part of identifying properties (e.g., Maven)
	ArtifactID string `json:"artifactId,omitempty"` // Part of identifying properties (e.g., Maven)
	Scope      string `json:"scope,omitempty"`    // e.g., 'compile', 'test'
}

type Class struct {
	BaseNode
	Language string `json:"language,omitempty"`
	FilePath string `json:"filePath"` // Part of identifying properties
	Visibility string `json:"visibility,omitempty"` // e.g., 'public', 'private'
}

type Interface struct {
	BaseNode
	Language string `json:"language,omitempty"`
	FilePath string `json:"filePath"` // Part of identifying properties
}

type Function struct {
	BaseNode
	Language   string   `json:"language,omitempty"`
	FilePath   string   `json:"filePath"` // Part of identifying properties
	Signature  string   `json:"signature,omitempty"` // Potentially identifying
	Parameters []string `json:"parameters,omitempty"`
	ReturnType string   `json:"returnType,omitempty"`
	Visibility string   `json:"visibility,omitempty"`
}

type File struct {
	BaseNode // Name might be the filename
	FilePath string `json:"filePath"` // Required identifying property (relative path from repo root)
	Language string `json:"language,omitempty"`
	Format   string `json:"format,omitempty"` // e.g., 'go', 'yaml', 'json'
}

// ConfigurationFile can reuse the File struct, potentially identified by label + filePath
// type ConfigurationFile File

type DataStore struct {
	BaseNode
	Type     string `json:"type"` // Required identifying property (e.g., 'postgresql', 'redis', 's3')
	Location string `json:"location,omitempty"` // Identifying property (e.g., connection string, hostname, bucket name)
}

type ExternalAPI struct {
	BaseNode
	EndpointURL      string `json:"endpointUrl,omitempty"` // Identifying property
	DocumentationURL string `json:"documentationUrl,omitempty"`
}

// --- Relationship Type Definitions (Conceptual) ---
// These types might not need specific structs if properties are handled directly
// in Cypher queries or a generic relationship struct is used.
// Properties specific to relationships can be added if needed.

type DependsOnRelationship struct {
	BaseRelationship
	VersionConstraint string `json:"versionConstraint,omitempty"`
	Scope             string `json:"scope,omitempty"`
}

type CallsRelationship struct {
	BaseRelationship
	LineNumber int  `json:"lineNumber,omitempty"`
	IsAsync    bool `json:"isAsync,omitempty"`
}

type CommunicatesWithRelationship struct {
	BaseRelationship
	Protocol string `json:"protocol,omitempty"`
	IsAsync  bool   `json:"isAsync,omitempty"`
}

type DefinedInRelationship struct {
	BaseRelationship
	StartLine int `json:"startLine,omitempty"`
	EndLine   int `json:"endLine,omitempty"`
}

// Other relationships like CONTAINS, IMPLEMENTS, USES, CONFIGURED_BY, PART_OF
// might use BaseRelationship or have no specific properties beyond the base ones.

// --- Helper Structures for Tool Inputs/Outputs ---

// EntityInput represents the data needed to find or create an entity.
type EntityInput struct {
	Labels               []string               `json:"labels"`                 // e.g., ["Function", "Go"]
	IdentifyingProperties map[string]interface{} `json:"identifyingProperties"` // Properties to match for MERGE (e.g., {"filePath": "/path/to/file.go", "name": "MyFunc"})
	Properties           map[string]interface{} `json:"properties"`            // All properties including identifying ones and others to set/update
}

// RelationshipInput represents the data needed to find or create a relationship.
type RelationshipInput struct {
	StartNodeLabels               []string               `json:"startNodeLabels"`
	StartNodeIdentifyingProperties map[string]interface{} `json:"startNodeIdentifyingProperties"`
	EndNodeLabels                 []string               `json:"endNodeLabels"`
	EndNodeIdentifyingProperties   map[string]interface{} `json:"endNodeIdentifyingProperties"`
	RelationshipType              string                 `json:"relationshipType"` // e.g., "CALLS"
	Properties                    map[string]interface{} `json:"properties"`       // Properties to set/update on the relationship
}

// EntityDetails represents the output for get_entity_details.
type EntityDetails struct {
	Labels     []string               `json:"labels"`
	Properties map[string]interface{} `json:"properties"`
}

// Neighbor represents a node connected to a central node.
type Neighbor struct {
	RelationshipType string                 `json:"relationshipType"` // Type of relationship connecting them
	Direction        string                 `json:"direction"`        // "incoming" or "outgoing"
	NodeLabels       []string               `json:"nodeLabels"`
	NodeProperties   map[string]interface{} `json:"nodeProperties"`
}

// NeighborsResult represents the output for find_neighbors.
type NeighborsResult struct {
	CentralNode EntityDetails `json:"centralNode"`
	Neighbors   []Neighbor    `json:"neighbors"`
}

// DependencyResult represents the output for find_dependencies/find_dependents.
type DependencyResult struct {
	TargetNode EntityDetails   `json:"targetNode"` // The node for which dependencies/dependents were found
	Results    []EntityDetails `json:"results"`    // List of dependencies or dependents
	Depth      int             `json:"depth"`      // The depth searched
	Direction  string          `json:"direction"`  // "dependencies" or "dependents"
}

// SubgraphNode represents a node within a subgraph result, simplified for visualisation.
type SubgraphNode struct {
	ID     string                 `json:"id"`     // Unique ID (e.g., elementId)
	Labels []string               `json:"labels"` // Node labels
	Name   string                 `json:"name"`   // Primary display name
	Props  map[string]interface{} `json:"props"`  // Selected properties for display
}

// SubgraphRelationship represents a relationship within a subgraph result.
type SubgraphRelationship struct {
	ID        string                 `json:"id"`        // Unique ID (e.g., elementId)
	StartNode string                 `json:"startNode"` // ID of the start node
	EndNode   string                 `json:"endNode"`   // ID of the end node
	Type      string                 `json:"type"`      // Relationship type
	Props     map[string]interface{} `json:"props"`     // Selected properties for display
}

// SubgraphResult represents the output for get_entity_subgraph, suitable for visualisation.
type SubgraphResult struct {
	Nodes         []SubgraphNode         `json:"nodes"`
	Relationships []SubgraphRelationship `json:"relationships"`
}
