package dgraph

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sammcj/mcp-graph/internal/graph"
	"github.com/sammcj/mcp-graph/internal/graph/dgraph/dgraphtest"
)

// DgraphStore implements the graph.Store interface using Dgraph
type DgraphStore struct {
	client dgraphtest.DgraphClient
}

// Ensure DgraphStore implements graph.Store
var _ graph.Store = (*DgraphStore)(nil)

// NewDgraphStore creates a new Dgraph store
func NewDgraphStore(address string) (*DgraphStore, error) {
	// Create a gRPC connection
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Dgraph: %w", err)
	}

	// Create a Dgraph client
	dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	client := NewDgraphClientWrapper(dgraphClient)

	return &DgraphStore{
		client: client,
	}, nil
}

// NewDgraphStoreWithClient creates a new Dgraph store with a provided client
// This is useful for testing with a mock client
func NewDgraphStoreWithClient(client dgraphtest.DgraphClient) *DgraphStore {
	return &DgraphStore{
		client: client,
	}
}

// CreateNode creates a new node in the graph
func (s *DgraphStore) CreateNode(ctx context.Context, nodeType string, properties map[string]interface{}) (string, error) {
	txn := s.client.NewTxn()
	defer txn.Discard(ctx)

	// Add the type to properties
	properties["type"] = nodeType

	// Create mutation
	pb, err := json.Marshal(properties)
	if err != nil {
		return "", fmt.Errorf("failed to marshal properties: %w", err)
	}

	mu := &api.Mutation{
		SetJson:   pb,
		CommitNow: true,
	}

	// Execute mutation
	resp, err := txn.Mutate(ctx, mu)
	if err != nil {
		return "", fmt.Errorf("failed to create node: %w", err)
	}

	// Return UID of created node
	if len(resp.Uids) == 0 {
		return "", fmt.Errorf("no UID returned from node creation")
	}

	// Return the first UID (there should only be one)
	for _, uid := range resp.Uids {
		return uid, nil
	}

	return "", fmt.Errorf("no UID returned from node creation")
}

// GetNode retrieves a node by ID
func (s *DgraphStore) GetNode(ctx context.Context, id string) (map[string]interface{}, error) {
	txn := s.client.NewReadOnlyTxn()

	// Create query
	q := fmt.Sprintf(`
		{
			node(func: uid(%s)) {
				uid
				expand(_all_)
			}
		}
	`, id)

	// Execute query
	resp, err := txn.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Parse response
	var result struct {
		Node []map[string]interface{} `json:"node"`
	}

	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Node) == 0 {
		return nil, fmt.Errorf("node not found: %s", id)
	}

	return result.Node[0], nil
}

// UpdateNode updates a node's properties
func (s *DgraphStore) UpdateNode(ctx context.Context, id string, properties map[string]interface{}) error {
	txn := s.client.NewTxn()
	defer txn.Discard(ctx)

	// Add the UID to properties
	properties["uid"] = id

	// Create mutation
	pb, err := json.Marshal(properties)
	if err != nil {
		return fmt.Errorf("failed to marshal properties: %w", err)
	}

	mu := &api.Mutation{
		SetJson:   pb,
		CommitNow: true,
	}

	// Execute mutation
	_, err = txn.Mutate(ctx, mu)
	if err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	return nil
}

// --- Software Architecture Specific Operations ---

// FindOrCreateEntity finds an entity based on identifying properties or creates it if not found.
// Dgraph implementation needs careful handling of upserts.
func (s *DgraphStore) FindOrCreateEntity(ctx context.Context, input graph.EntityInput) (graph.EntityDetails, error) {
	// Placeholder implementation - Dgraph upserts require specific query logic
	return graph.EntityDetails{}, fmt.Errorf("FindOrCreateEntity not implemented for Dgraph")
}

// FindOrCreateRelationship finds a relationship or creates it if not found.
// Dgraph implementation needs careful handling of upserts.
func (s *DgraphStore) FindOrCreateRelationship(ctx context.Context, input graph.RelationshipInput) (map[string]interface{}, error) {
	// Placeholder implementation - Dgraph edge upserts are complex
	return nil, fmt.Errorf("FindOrCreateRelationship not implemented for Dgraph")
}

// GetEntityDetails retrieves the labels and properties of a specific entity.
// Dgraph doesn't have explicit labels like Neo4j, often relies on a 'type' predicate.
func (s *DgraphStore) GetEntityDetails(ctx context.Context, labels []string, identifyingProperties map[string]interface{}) (graph.EntityDetails, error) {
	// Placeholder implementation
	return graph.EntityDetails{}, fmt.Errorf("GetEntityDetails not implemented for Dgraph")
}

// FindNeighbors finds the direct neighbors of a given entity up to a specified depth.
func (s *DgraphStore) FindNeighbors(ctx context.Context, labels []string, identifyingProperties map[string]interface{}, maxDepth int) (graph.NeighborsResult, error) {
	// Placeholder implementation
	return graph.NeighborsResult{}, fmt.Errorf("FindNeighbors not implemented for Dgraph")
}

// DeleteNode deletes a node by ID
func (s *DgraphStore) DeleteNode(ctx context.Context, id string) error {
	txn := s.client.NewTxn()
	defer txn.Discard(ctx)

	// Create delete mutation
	mu := &api.Mutation{
		DelNquads: []byte(fmt.Sprintf(`<%s> * * .`, id)),
		CommitNow: true,
	}

	// Execute mutation
	_, err := txn.Mutate(ctx, mu)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	return nil
}

// CreateEdge creates a new edge between two nodes
func (s *DgraphStore) CreateEdge(ctx context.Context, fromID, toID, relationshipType string, properties map[string]interface{}) (string, error) {
	txn := s.client.NewTxn()
	defer txn.Discard(ctx)

	// Create edge data
	edgeData := map[string]interface{}{
		"uid":              fromID,
		relationshipType: map[string]interface{}{
			"uid": toID,
		},
	}

	// Add properties to the edge if provided
	if len(properties) > 0 {
		// In Dgraph, edge properties are represented as facets
		// This is a simplified implementation
		edgeData[relationshipType+"_props"] = properties
	}

	// Create mutation
	pb, err := json.Marshal(edgeData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal edge data: %w", err)
	}

	mu := &api.Mutation{
		SetJson:   pb,
		CommitNow: true,
	}

	// Execute mutation
	_, err = txn.Mutate(ctx, mu)
	if err != nil {
		return "", fmt.Errorf("failed to create edge: %w", err)
	}

	// In Dgraph, edges don't have their own UIDs
	// We return a composite ID for reference
	return fmt.Sprintf("%s-%s-%s", fromID, relationshipType, toID), nil
}

// GetEdge retrieves an edge by ID
func (s *DgraphStore) GetEdge(ctx context.Context, id string) (map[string]interface{}, error) {
	// In Dgraph, edges don't have their own UIDs
	// This is a simplified implementation that assumes the ID is in the format "fromID-relType-toID"
	return nil, fmt.Errorf("not implemented: edges in Dgraph don't have their own IDs")
}

// UpdateEdge updates an edge's properties
func (s *DgraphStore) UpdateEdge(ctx context.Context, id string, properties map[string]interface{}) error {
	// In Dgraph, edges don't have their own UIDs
	// This is a simplified implementation that assumes the ID is in the format "fromID-relType-toID"
	return fmt.Errorf("not implemented: edges in Dgraph don't have their own IDs")
}

// DeleteEdge deletes an edge by ID
func (s *DgraphStore) DeleteEdge(ctx context.Context, id string) error {
	// In Dgraph, edges don't have their own UIDs
	// This is a simplified implementation that assumes the ID is in the format "fromID-relType-toID"
	return fmt.Errorf("not implemented: edges in Dgraph don't have their own IDs")
}

// Query executes a custom query against the graph
func (s *DgraphStore) Query(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error) {
	txn := s.client.NewReadOnlyTxn()

	// Convert params to map[string]string
	stringParams := make(map[string]string)
	for k, v := range params {
		// Convert each value to string
		switch val := v.(type) {
		case string:
			stringParams[k] = val
		case fmt.Stringer:
			stringParams[k] = val.String()
		default:
			// Use JSON for complex types
			bytes, err := json.Marshal(val)
			if err != nil {
				return nil, fmt.Errorf("failed to convert parameter %s to string: %w", k, err)
			}
			stringParams[k] = string(bytes)
		}
	}

	// Execute query
	resp, err := txn.QueryWithVars(ctx, query, stringParams)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Parse response
	var result map[string][]map[string]interface{}
	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract results
	var results []map[string]interface{}
	for _, v := range result {
		results = append(results, v...)
	}

	return results, nil
}

// UpsertSchema updates or creates the schema
func (s *DgraphStore) UpsertSchema(ctx context.Context, schema string) error {
	op := &api.Operation{
		Schema: schema,
	}

	err := s.client.Alter(ctx, op)
	if err != nil {
		return fmt.Errorf("failed to update schema: %w", err)
	}

	return nil
}
