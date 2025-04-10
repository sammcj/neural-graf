package neo4j

import (
	"context"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/sammcj/mcp-graph/internal/graph"
)

// Neo4jStore implements the graph.Store interface using Neo4j
type Neo4jStore struct {
	driver neo4j.DriverWithContext
}

// Ensure Neo4jStore implements graph.Store
var _ graph.Store = (*Neo4jStore)(nil)

// NewNeo4jStore creates a new Neo4j store
func NewNeo4jStore(uri, username, password string) (*Neo4jStore, error) {
	// Create a Neo4j driver
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity
	ctx := context.Background()
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	return &Neo4jStore{
		driver: driver,
	}, nil
}

// Close closes the Neo4j driver
func (s *Neo4jStore) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

// CreateNode creates a new node in the graph
func (s *Neo4jStore) CreateNode(ctx context.Context, nodeType string, properties map[string]interface{}) (string, error) {
	// Add the type to properties if not already present
	if _, ok := properties["type"]; !ok {
		properties["type"] = nodeType
	}

	// Create Cypher query
	query := fmt.Sprintf("CREATE (n:%s $props) RETURN id(n) as id", nodeType)
	params := map[string]interface{}{
		"props": properties,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return "", fmt.Errorf("failed to create node: %w", err)
	}

	// Extract node ID
	if len(result.Records) == 0 {
		return "", fmt.Errorf("no ID returned from node creation")
	}

	id, ok := result.Records[0].Get("id")
	if !ok {
		return "", fmt.Errorf("no ID returned from node creation")
	}

	// Convert ID to string
	idStr := fmt.Sprintf("%v", id)
	return idStr, nil
}

// GetNode retrieves a node by ID
func (s *Neo4jStore) GetNode(ctx context.Context, id string) (map[string]interface{}, error) {
	// Create Cypher query
	query := "MATCH (n) WHERE id(n) = $id RETURN n"
	params := map[string]interface{}{
		"id": id,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// Extract node
	if len(result.Records) == 0 {
		return nil, fmt.Errorf("node not found: %s", id)
	}

	// Get node from record
	nodeVal, ok := result.Records[0].Get("n")
	if !ok {
		return nil, fmt.Errorf("node not found in record")
	}

	// Convert node to map
	node, ok := nodeVal.(neo4j.Node)
	if !ok {
		return nil, fmt.Errorf("failed to convert result to node")
	}

	// Convert properties to map
	props := node.Props
	props["id"] = id
	props["labels"] = node.Labels

	return props, nil
}

// UpdateNode updates a node's properties
func (s *Neo4jStore) UpdateNode(ctx context.Context, id string, properties map[string]interface{}) error {
	// Create Cypher query
	query := "MATCH (n) WHERE id(n) = $id SET n += $props"
	params := map[string]interface{}{
		"id":    id,
		"props": properties,
	}

	// Execute query
	_, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	return nil
}

// DeleteNode deletes a node by ID
func (s *Neo4jStore) DeleteNode(ctx context.Context, id string) error {
	// Create Cypher query
	query := "MATCH (n) WHERE id(n) = $id DETACH DELETE n"
	params := map[string]interface{}{
		"id": id,
	}

	// Execute query
	_, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	return nil
}

// CreateEdge creates a new edge between two nodes
func (s *Neo4jStore) CreateEdge(ctx context.Context, fromID, toID, relationshipType string, properties map[string]interface{}) (string, error) {
	// Create Cypher query
	query := "MATCH (a), (b) WHERE id(a) = $fromID AND id(b) = $toID CREATE (a)-[r:" + relationshipType + " $props]->(b) RETURN id(r) as id"
	params := map[string]interface{}{
		"fromID": fromID,
		"toID":   toID,
		"props":  properties,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return "", fmt.Errorf("failed to create edge: %w", err)
	}

	// Extract edge ID
	if len(result.Records) == 0 {
		return "", fmt.Errorf("no ID returned from edge creation")
	}

	id, ok := result.Records[0].Get("id")
	if !ok {
		return "", fmt.Errorf("no ID returned from edge creation")
	}

	// Convert ID to string
	idStr := fmt.Sprintf("%v", id)
	return idStr, nil
}

// GetEdge retrieves an edge by ID
func (s *Neo4jStore) GetEdge(ctx context.Context, id string) (map[string]interface{}, error) {
	// Create Cypher query
	query := "MATCH ()-[r]->() WHERE id(r) = $id RETURN r, startNode(r) as from, endNode(r) as to, type(r) as type"
	params := map[string]interface{}{
		"id": id,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return nil, fmt.Errorf("failed to get edge: %w", err)
	}

	// Extract edge
	if len(result.Records) == 0 {
		return nil, fmt.Errorf("edge not found: %s", id)
	}

	// Get edge from record
	record := result.Records[0]
	edgeVal, ok := record.Get("r")
	if !ok {
		return nil, fmt.Errorf("edge not found in record")
	}

	// Convert edge to map
	edge, ok := edgeVal.(neo4j.Relationship)
	if !ok {
		return nil, fmt.Errorf("failed to convert result to edge")
	}

	// Get from and to nodes
	fromVal, ok := record.Get("from")
	if !ok {
		return nil, fmt.Errorf("from node not found in record")
	}
	fromNode, ok := fromVal.(neo4j.Node)
	if !ok {
		return nil, fmt.Errorf("failed to convert from to node")
	}

	toVal, ok := record.Get("to")
	if !ok {
		return nil, fmt.Errorf("to node not found in record")
	}
	toNode, ok := toVal.(neo4j.Node)
	if !ok {
		return nil, fmt.Errorf("failed to convert to to node")
	}

	// Get relationship type
	relType, ok := record.Get("type")
	if !ok {
		return nil, fmt.Errorf("relationship type not found in record")
	}

	// Convert properties to map
	props := edge.Props
	props["id"] = id
	props["fromId"] = fmt.Sprintf("%v", fromNode.ElementId)
	props["toId"] = fmt.Sprintf("%v", toNode.ElementId)
	props["type"] = relType

	return props, nil
}

// UpdateEdge updates an edge's properties
func (s *Neo4jStore) UpdateEdge(ctx context.Context, id string, properties map[string]interface{}) error {
	// Create Cypher query
	query := "MATCH ()-[r]->() WHERE id(r) = $id SET r += $props"
	params := map[string]interface{}{
		"id":    id,
		"props": properties,
	}

	// Execute query
	_, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return fmt.Errorf("failed to update edge: %w", err)
	}

	return nil
}

// DeleteEdge deletes an edge by ID
func (s *Neo4jStore) DeleteEdge(ctx context.Context, id string) error {
	// Create Cypher query
	query := "MATCH ()-[r]->() WHERE id(r) = $id DELETE r"
	params := map[string]interface{}{
		"id": id,
	}

	// Execute query
	_, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return fmt.Errorf("failed to delete edge: %w", err)
	}

	return nil
}

// Query executes a custom query against the graph
func (s *Neo4jStore) Query(ctx context.Context, query string, params map[string]interface{}) ([]map[string]interface{}, error) {
	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Convert records to maps
	var results []map[string]interface{}
	for _, record := range result.Records {
		resultMap := make(map[string]interface{})
		for _, key := range record.Keys {
			value, _ := record.Get(key)
			resultMap[key] = convertNeo4jValue(value)
		}
		results = append(results, resultMap)
	}

	return results, nil
}

// UpsertSchema updates or creates the schema
func (s *Neo4jStore) UpsertSchema(ctx context.Context, schema string) error {
	// Split schema into individual statements
	lines := strings.Split(schema, "\n")
	var currentStatement strings.Builder

	// Process each line
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") {
			continue
		}

		// Add the line to the current statement
		currentStatement.WriteString(line)
		currentStatement.WriteString(" ")

		// If the line ends with a semicolon, execute the statement
		if strings.HasSuffix(trimmedLine, ";") {
			stmt := strings.TrimSpace(currentStatement.String())
			if stmt != "" {
				// Execute statement
				_, err := neo4j.ExecuteQuery(ctx, s.driver, stmt, nil, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
				if err != nil {
					return fmt.Errorf("failed to execute schema statement: %w", err)
				}
			}

			// Reset the current statement
			currentStatement.Reset()
		}
	}

	// Check if there's any remaining statement
	remainingStmt := strings.TrimSpace(currentStatement.String())
	if remainingStmt != "" {
		// Execute the remaining statement
		_, err := neo4j.ExecuteQuery(ctx, s.driver, remainingStmt, nil, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
		if err != nil {
			return fmt.Errorf("failed to execute schema statement: %w", err)
		}
	}

	return nil
}

// convertNeo4jValue converts Neo4j values to Go values
func convertNeo4jValue(value interface{}) interface{} {
	switch v := value.(type) {
	case neo4j.Node:
		node := make(map[string]interface{})
		node["id"] = v.ElementId
		node["labels"] = v.Labels
		for k, prop := range v.Props {
			node[k] = convertNeo4jValue(prop)
		}
		return node
	case neo4j.Relationship:
		rel := make(map[string]interface{})
		rel["id"] = v.ElementId
		rel["type"] = v.Type
		rel["startId"] = v.StartElementId
		rel["endId"] = v.EndElementId
		for k, prop := range v.Props {
			rel[k] = convertNeo4jValue(prop)
		}
		return rel
	case neo4j.Path:
		path := make(map[string]interface{})
		nodes := make([]interface{}, len(v.Nodes))
		for i, node := range v.Nodes {
			nodes[i] = convertNeo4jValue(node)
		}
		path["nodes"] = nodes
		rels := make([]interface{}, len(v.Relationships))
		for i, rel := range v.Relationships {
			rels[i] = convertNeo4jValue(rel)
		}
		path["relationships"] = rels
		return path
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertNeo4jValue(item)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, item := range v {
			result[k] = convertNeo4jValue(item)
		}
		return result
	case neo4j.Date:
		return v.Time()
	case neo4j.LocalDateTime:
		return v.Time()
	case neo4j.OffsetTime:
		return v.String()
	case neo4j.LocalTime:
		return v.String()
	default:
		return v
	}
}
