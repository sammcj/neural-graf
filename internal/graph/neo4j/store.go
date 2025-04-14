package neo4j

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

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
	query := fmt.Sprintf("CREATE (n:%s $props) RETURN n", nodeType)
	params := map[string]interface{}{
		"props": properties,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return "", fmt.Errorf("failed to create node: %w", err)
	}

	// Extract node
	if len(result.Records) == 0 {
		return "", fmt.Errorf("no node returned from node creation")
	}

	nodeVal, ok := result.Records[0].Get("n")
	if !ok {
		return "", fmt.Errorf("no node returned from node creation")
	}

	// Convert node to Neo4j node
	node, ok := nodeVal.(neo4j.Node)
	if !ok {
		return "", fmt.Errorf("failed to convert result to node")
	}

	// Return the element ID which is more reliable for retrieval
	return node.ElementId, nil
}

// GetNode retrieves a node by ID
func (s *Neo4jStore) GetNode(ctx context.Context, id string) (map[string]interface{}, error) {
	// Create Cypher query - use elementId for more reliable retrieval
	query := "MATCH (n) WHERE elementId(n) = $id RETURN n"
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
	props["id"] = node.ElementId
	props["labels"] = node.Labels

	return props, nil
}

// UpdateNode updates a node's properties
func (s *Neo4jStore) UpdateNode(ctx context.Context, id string, properties map[string]interface{}) error {
	// Create Cypher query - use elementId for more reliable retrieval
	query := "MATCH (n) WHERE elementId(n) = $id SET n += $props"
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
	// Create Cypher query - use elementId for more reliable retrieval
	query := "MATCH (n) WHERE elementId(n) = $id DETACH DELETE n"
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
	// Create Cypher query - use elementId for more reliable node lookup
	query := "MATCH (a), (b) WHERE elementId(a) = $fromID AND elementId(b) = $toID CREATE (a)-[r:" + relationshipType + " $props]->(b) RETURN r"
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

	// Extract relationship
	if len(result.Records) == 0 {
		return "", fmt.Errorf("no relationship returned from edge creation")
	}

	relVal, ok := result.Records[0].Get("r")
	if !ok {
		return "", fmt.Errorf("no relationship returned from edge creation")
	}

	// Convert to Neo4j relationship
	rel, ok := relVal.(neo4j.Relationship)
	if !ok {
		return "", fmt.Errorf("failed to convert result to relationship")
	}

	// Return the element ID which is more reliable for retrieval
	return rel.ElementId, nil
}

// GetEdge retrieves an edge by ID
func (s *Neo4jStore) GetEdge(ctx context.Context, id string) (map[string]interface{}, error) {
	// Create Cypher query - use elementId for more reliable retrieval
	query := "MATCH ()-[r]->() WHERE elementId(r) = $id RETURN r, startNode(r) as from, endNode(r) as to, type(r) as type"
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
	props["id"] = edge.ElementId
	props["fromId"] = fromNode.ElementId
	props["toId"] = toNode.ElementId
	props["type"] = relType

	return props, nil
}

// UpdateEdge updates an edge's properties
func (s *Neo4jStore) UpdateEdge(ctx context.Context, id string, properties map[string]interface{}) error {
	// Create Cypher query - use elementId for more reliable retrieval
	query := "MATCH ()-[r]->() WHERE elementId(r) = $id SET r += $props"
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
	// Create Cypher query - use elementId for more reliable retrieval
	query := "MATCH ()-[r]->() WHERE elementId(r) = $id DELETE r"
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

// --- Software Architecture Specific Operations ---

// FindOrCreateEntity finds an entity based on identifying properties or creates it if not found.
// It merges the provided properties with any existing ones.
func (s *Neo4jStore) FindOrCreateEntity(ctx context.Context, input graph.EntityInput) (graph.EntityDetails, error) {
	if len(input.Labels) == 0 {
		return graph.EntityDetails{}, fmt.Errorf("at least one label is required")
	}
	if len(input.IdentifyingProperties) == 0 {
		return graph.EntityDetails{}, fmt.Errorf("at least one identifying property is required")
	}

	// Build label string (e.g., :Label1:Label2)
	labelStr := ":" + strings.Join(input.Labels, ":")

	// Build identifying properties match string (e.g., {key1: $idProps.key1, key2: $idProps.key2})
	var idPropsParts []string
	for k := range input.IdentifyingProperties {
		// Ensure valid characters for property keys if necessary, though Neo4j is flexible
		idPropsParts = append(idPropsParts, fmt.Sprintf("%s: $idProps.%s", k, k))
	}
	idPropsMatchStr := "{" + strings.Join(idPropsParts, ", ") + "}"

	// Prepare all properties, ensuring timestamps are handled correctly
	allProps := make(map[string]interface{})
	for k, v := range input.Properties {
		allProps[k] = v
	}
	now := time.Now().UTC() // Use UTC for consistency
	// Ensure lastModifiedAt is always updated, even if present in input.Properties
	allProps["lastModifiedAt"] = now

	// Construct the MERGE query
	query := fmt.Sprintf(`
        MERGE (n%s %s)
        ON CREATE SET n = $allProps, n.createdAt = $now
        ON MATCH SET n += $allProps // Use += to merge properties, lastModifiedAt is updated via $allProps
        RETURN labels(n) as labels, properties(n) as props, elementId(n) as id
    `, labelStr, idPropsMatchStr)

	params := map[string]interface{}{
		"idProps":  input.IdentifyingProperties,
		"allProps": allProps,
		"now":      now,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return graph.EntityDetails{}, fmt.Errorf("failed to execute FindOrCreateEntity query: %w", err)
	}

	if len(result.Records) == 0 {
		return graph.EntityDetails{}, fmt.Errorf("no node returned from MERGE operation")
	}

	record := result.Records[0]

	// Extract labels
	labelsVal, labelsOk := record.Get("labels")
	if !labelsOk {
		return graph.EntityDetails{}, fmt.Errorf("failed to get labels from result")
	}
	labelsInterface, ok := labelsVal.([]interface{})
	if !ok {
		return graph.EntityDetails{}, fmt.Errorf("labels are not in expected format []interface{}")
	}
	labels := make([]string, len(labelsInterface))
	for i, l := range labelsInterface {
		labels[i], ok = l.(string)
		if !ok {
			return graph.EntityDetails{}, fmt.Errorf("label item is not a string")
		}
	}

	// Extract properties
	propsVal, propsOk := record.Get("props")
	if !propsOk {
		return graph.EntityDetails{}, fmt.Errorf("failed to get properties from result")
	}
	props, ok := propsVal.(map[string]interface{})
	if !ok {
		return graph.EntityDetails{}, fmt.Errorf("properties are not in expected format map[string]interface{}")
	}

	// Extract element ID
	idVal, idOk := record.Get("id")
	if !idOk {
		return graph.EntityDetails{}, fmt.Errorf("failed to get elementId from result")
	}
	idStr, ok := idVal.(string)
	if !ok {
		return graph.EntityDetails{}, fmt.Errorf("elementId is not a string")
	}
	// Add the element ID to the properties map, as Neo4j doesn't include it by default
	props["id"] = idStr

	// Convert Neo4j specific types (like time) within properties
	for k, v := range props {
		props[k] = convertNeo4jValue(v) // Reuse existing conversion logic
	}


	return graph.EntityDetails{
		Labels:     labels,
		Properties: props,
	}, nil
}

// FindOrCreateRelationship finds a relationship or creates it if not found.
// It merges the provided properties with any existing ones.
func (s *Neo4jStore) FindOrCreateRelationship(ctx context.Context, input graph.RelationshipInput) (map[string]interface{}, error) {
	if len(input.StartNodeLabels) == 0 || len(input.EndNodeLabels) == 0 {
		return nil, fmt.Errorf("start and end node labels are required")
	}
	if len(input.StartNodeIdentifyingProperties) == 0 || len(input.EndNodeIdentifyingProperties) == 0 {
		return nil, fmt.Errorf("start and end node identifying properties are required")
	}
	if input.RelationshipType == "" {
		return nil, fmt.Errorf("relationship type is required")
	}

	// Build start node match clause
	startLabelStr := ":" + strings.Join(input.StartNodeLabels, ":")
	var startIdPropsParts []string
	for k := range input.StartNodeIdentifyingProperties {
		startIdPropsParts = append(startIdPropsParts, fmt.Sprintf("%s: $startIdProps.%s", k, k))
	}
	startIdPropsMatchStr := "{" + strings.Join(startIdPropsParts, ", ") + "}"

	// Build end node match clause
	endLabelStr := ":" + strings.Join(input.EndNodeLabels, ":")
	var endIdPropsParts []string
	for k := range input.EndNodeIdentifyingProperties {
		endIdPropsParts = append(endIdPropsParts, fmt.Sprintf("%s: $endIdProps.%s", k, k))
	}
	endIdPropsMatchStr := "{" + strings.Join(endIdPropsParts, ", ") + "}"

	// Prepare relationship properties, ensuring timestamps are handled
	relProps := make(map[string]interface{})
	for k, v := range input.Properties {
		relProps[k] = v
	}
	now := time.Now().UTC()
	// Ensure lastModifiedAt is always updated
	relProps["lastModifiedAt"] = now

	// Construct the MERGE query for the relationship
	// Note: MERGE on relationships requires matching both start and end nodes first.
	query := fmt.Sprintf(`
        MATCH (start%s %s)
        MATCH (end%s %s)
        MERGE (start)-[r:%s]->(end)
        ON CREATE SET r = $relProps, r.createdAt = $now
        ON MATCH SET r += $relProps // Merge properties on match
        RETURN properties(r) as props, elementId(r) as id
    `, startLabelStr, startIdPropsMatchStr, endLabelStr, endIdPropsMatchStr, input.RelationshipType)

	params := map[string]interface{}{
		"startIdProps": input.StartNodeIdentifyingProperties,
		"endIdProps":   input.EndNodeIdentifyingProperties,
		"relProps":     relProps,
		"now":          now,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return nil, fmt.Errorf("failed to execute FindOrCreateRelationship query: %w", err)
	}

	if len(result.Records) == 0 {
		// This might happen if the start or end node doesn't exist.
		// Consider adding checks or returning a more specific error.
		return nil, fmt.Errorf("no relationship returned from MERGE operation (start or end node might not exist)")
	}

	record := result.Records[0]

	// Extract properties
	propsVal, propsOk := record.Get("props")
	if !propsOk {
		return nil, fmt.Errorf("failed to get properties from result")
	}
	props, ok := propsVal.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("relationship properties are not in expected format map[string]interface{}")
	}

	// Extract element ID
	idVal, idOk := record.Get("id")
	if !idOk {
		return nil, fmt.Errorf("failed to get elementId from result")
	}
	idStr, ok := idVal.(string)
	if !ok {
		return nil, fmt.Errorf("elementId is not a string")
	}
	// Add the element ID to the properties map
	props["id"] = idStr

	// Convert Neo4j specific types within properties
	for k, v := range props {
		props[k] = convertNeo4jValue(v)
	}

	return props, nil
}

// GetEntityDetails retrieves the labels and properties of a specific entity identified by its labels and unique properties.
func (s *Neo4jStore) GetEntityDetails(ctx context.Context, labels []string, identifyingProperties map[string]interface{}) (graph.EntityDetails, error) {
	if len(labels) == 0 {
		return graph.EntityDetails{}, fmt.Errorf("at least one label is required")
	}
	if len(identifyingProperties) == 0 {
		return graph.EntityDetails{}, fmt.Errorf("at least one identifying property is required")
	}

	// Build label string (e.g., :Label1:Label2)
	labelStr := ":" + strings.Join(labels, ":")

	// Build identifying properties match string (e.g., {key1: $idProps.key1, key2: $idProps.key2})
	var idPropsParts []string
	for k := range identifyingProperties {
		idPropsParts = append(idPropsParts, fmt.Sprintf("%s: $idProps.%s", k, k))
	}
	idPropsMatchStr := "{" + strings.Join(idPropsParts, ", ") + "}"

	// Construct the MATCH query
	query := fmt.Sprintf(`
        MATCH (n%s %s)
        RETURN labels(n) as labels, properties(n) as props, elementId(n) as id
        LIMIT 1 // Ensure only one node is returned
    `, labelStr, idPropsMatchStr)

	params := map[string]interface{}{
		"idProps": identifyingProperties,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return graph.EntityDetails{}, fmt.Errorf("failed to execute GetEntityDetails query: %w", err)
	}

	if len(result.Records) == 0 {
		// Node not found
		return graph.EntityDetails{}, fmt.Errorf("entity not found with labels %v and properties %v", labels, identifyingProperties) // Consider a more structured error
	}

	record := result.Records[0]

	// Extract labels
	labelsVal, labelsOk := record.Get("labels")
	if !labelsOk {
		return graph.EntityDetails{}, fmt.Errorf("failed to get labels from result")
	}
	labelsInterface, ok := labelsVal.([]interface{})
	if !ok {
		return graph.EntityDetails{}, fmt.Errorf("labels are not in expected format []interface{}")
	}
	foundLabels := make([]string, len(labelsInterface))
	for i, l := range labelsInterface {
		foundLabels[i], ok = l.(string)
		if !ok {
			return graph.EntityDetails{}, fmt.Errorf("label item is not a string")
		}
	}

	// Extract properties
	propsVal, propsOk := record.Get("props")
	if !propsOk {
		return graph.EntityDetails{}, fmt.Errorf("failed to get properties from result")
	}
	props, ok := propsVal.(map[string]interface{})
	if !ok {
		return graph.EntityDetails{}, fmt.Errorf("properties are not in expected format map[string]interface{}")
	}

	// Extract element ID
	idVal, idOk := record.Get("id")
	if !idOk {
		return graph.EntityDetails{}, fmt.Errorf("failed to get elementId from result")
	}
	idStr, ok := idVal.(string)
	if !ok {
		return graph.EntityDetails{}, fmt.Errorf("elementId is not a string")
	}
	// Add the element ID to the properties map
	props["id"] = idStr

	// Convert Neo4j specific types within properties
	for k, v := range props {
		props[k] = convertNeo4jValue(v)
	}

	return graph.EntityDetails{
		Labels:     foundLabels,
		Properties: props,
	}, nil
}

// FindNeighbors finds the direct neighbors of a given entity up to a specified depth.
func (s *Neo4jStore) FindNeighbors(ctx context.Context, labels []string, identifyingProperties map[string]interface{}, maxDepth int) (graph.NeighborsResult, error) {
	if len(labels) == 0 {
		return graph.NeighborsResult{}, fmt.Errorf("at least one label is required for the central node")
	}
	if len(identifyingProperties) == 0 {
		return graph.NeighborsResult{}, fmt.Errorf("at least one identifying property is required for the central node")
	}
	if maxDepth <= 0 {
		maxDepth = 1 // Default to direct neighbors if depth is invalid
	}

	// Build label string for the central node
	labelStr := ":" + strings.Join(labels, ":")

	// Build identifying properties match string for the central node
	var idPropsParts []string
	for k := range identifyingProperties {
		idPropsParts = append(idPropsParts, fmt.Sprintf("%s: $idProps.%s", k, k))
	}
	idPropsMatchStr := "{" + strings.Join(idPropsParts, ", ") + "}"

	// Construct the MATCH query for neighbors
	// This query finds the central node, then finds neighbors up to maxDepth
	// It returns the central node, the relationship, and the neighbor node
	query := fmt.Sprintf(`
        MATCH (center%s %s)
        CALL {
            WITH center
            MATCH path = (center)-[r*1..%d]-(neighbor)
            WHERE neighbor <> center // Avoid matching the center node as a neighbor
            RETURN center, r, neighbor, path
            LIMIT 100 // Add a reasonable limit to prevent excessive results
        }
        RETURN
            labels(center) as centerLabels,
            properties(center) as centerProps,
            elementId(center) as centerId,
            [rel in r | { type: type(rel), props: properties(rel), id: elementId(rel), startId: elementId(startNode(rel)), endId: elementId(endNode(rel)) }] as relationships,
            labels(neighbor) as neighborLabels,
            properties(neighbor) as neighborProps,
            elementId(neighbor) as neighborId
    `, labelStr, idPropsMatchStr, maxDepth)

	params := map[string]interface{}{
		"idProps": identifyingProperties,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return graph.NeighborsResult{}, fmt.Errorf("failed to execute FindNeighbors query: %w", err)
	}

	if len(result.Records) == 0 {
		// It's possible the node exists but has no neighbors, or the node doesn't exist.
		// Try getting just the central node to differentiate.
		centralNodeDetails, err := s.GetEntityDetails(ctx, labels, identifyingProperties)
		if err != nil {
			// Central node likely doesn't exist or there was another error
			return graph.NeighborsResult{}, fmt.Errorf("central node not found or error fetching details: %w", err)
		}
		// Central node exists but has no neighbors within the depth
		return graph.NeighborsResult{CentralNode: centralNodeDetails, Neighbors: []graph.Neighbor{}}, nil
	}

	// Process the first record to get central node details (should be the same across records)
	firstRecord := result.Records[0]
	centerLabelsVal, _ := firstRecord.Get("centerLabels")
	centerPropsVal, _ := firstRecord.Get("centerProps")
	centerIdVal, _ := firstRecord.Get("centerId")

	centerLabelsInterface, _ := centerLabelsVal.([]interface{})
	centerLabels := make([]string, len(centerLabelsInterface))
	for i, l := range centerLabelsInterface {
		centerLabels[i], _ = l.(string)
	}

	centerProps, _ := centerPropsVal.(map[string]interface{})
	centerId, _ := centerIdVal.(string)
	centerProps["id"] = centerId // Add element ID

	// Convert central node properties
	for k, v := range centerProps {
		centerProps[k] = convertNeo4jValue(v)
	}

	centralNodeResult := graph.EntityDetails{
		Labels:     centerLabels,
		Properties: centerProps,
	}

	// Process neighbors
	neighborsMap := make(map[string]graph.Neighbor) // Use map to deduplicate neighbors

	for _, record := range result.Records {
		relsVal, _ := record.Get("relationships")
		neighborLabelsVal, _ := record.Get("neighborLabels")
		neighborPropsVal, _ := record.Get("neighborProps")
		neighborIdVal, _ := record.Get("neighborId")

		neighborId, ok := neighborIdVal.(string)
		if !ok || neighborId == "" {
			continue // Skip if neighbor ID is invalid
		}

		// If neighbor already processed, skip (basic deduplication)
		if _, exists := neighborsMap[neighborId]; exists {
			continue
		}

		// Process relationship details (focus on the first relationship connecting to this neighbor for simplicity)
		relsInterface, _ := relsVal.([]interface{})
		var relType string
		var direction string
		if len(relsInterface) > 0 {
			firstRelMap, ok := relsInterface[0].(map[string]interface{})
			if ok {
				relType, _ = firstRelMap["type"].(string)
				startId, _ := firstRelMap["startId"].(string)
				if startId == centerId {
					direction = "outgoing"
				} else {
					direction = "incoming"
				}
			}
		}

		// Process neighbor labels
		neighborLabelsInterface, _ := neighborLabelsVal.([]interface{})
		neighborLabels := make([]string, len(neighborLabelsInterface))
		for i, l := range neighborLabelsInterface {
			neighborLabels[i], _ = l.(string)
		}

		// Process neighbor properties
		neighborProps, _ := neighborPropsVal.(map[string]interface{})
		neighborProps["id"] = neighborId // Add element ID

		// Convert neighbor properties
		for k, v := range neighborProps {
			neighborProps[k] = convertNeo4jValue(v)
		}

		neighborsMap[neighborId] = graph.Neighbor{
			RelationshipType: relType,
			Direction:        direction,
			NodeLabels:       neighborLabels,
			NodeProperties:   neighborProps,
		}
	}

	// Convert map to slice
	neighborsResult := make([]graph.Neighbor, 0, len(neighborsMap))
	for _, neighbor := range neighborsMap {
		neighborsResult = append(neighborsResult, neighbor)
	}

	return graph.NeighborsResult{
		CentralNode: centralNodeResult,
		Neighbors:   neighborsResult,
	}, nil
}

// buildRelationshipTypeFilter builds the relationship type part of a Cypher query.
// Example: ":REL1|:REL2" or "" if types are empty.
func buildRelationshipTypeFilter(types []string) string {
	if len(types) == 0 {
		return "" // Match any relationship type
	}
	var quotedTypes []string
	for _, t := range types {
		// Basic sanitization/quoting might be needed depending on allowed characters
		quotedTypes = append(quotedTypes, ":"+t)
	}
	return strings.Join(quotedTypes, "|")
}


// FindDependencies finds entities that the target entity depends on (outgoing relationships),
// following specified relationship types up to a certain depth.
func (s *Neo4jStore) FindDependencies(ctx context.Context, labels []string, identifyingProperties map[string]interface{}, relationshipTypes []string, maxDepth int) (graph.DependencyResult, error) {
	if len(labels) == 0 {
		return graph.DependencyResult{}, fmt.Errorf("at least one label is required for the target node")
	}
	if len(identifyingProperties) == 0 {
		return graph.DependencyResult{}, fmt.Errorf("at least one identifying property is required for the target node")
	}
	if maxDepth <= 0 {
		maxDepth = 1 // Default to depth 1 if invalid
	}

	// First, get the target node details to include in the result
	targetNodeDetails, err := s.GetEntityDetails(ctx, labels, identifyingProperties)
	if err != nil {
		return graph.DependencyResult{}, fmt.Errorf("target node not found or error fetching details: %w", err)
	}

	// Build label string for the target node
	labelStr := ":" + strings.Join(labels, ":")

	// Build identifying properties match string for the target node
	var idPropsParts []string
	for k := range identifyingProperties {
		idPropsParts = append(idPropsParts, fmt.Sprintf("%s: $idProps.%s", k, k))
	}
	idPropsMatchStr := "{" + strings.Join(idPropsParts, ", ") + "}"

	// Build relationship type filter string
	relTypeFilter := buildRelationshipTypeFilter(relationshipTypes)

	// Construct the MATCH query for dependencies (outgoing relationships)
	query := fmt.Sprintf(`
        MATCH (target%s %s)
        MATCH (target)-[r%s*1..%d]->(dependency)
        WHERE target <> dependency
        RETURN DISTINCT
            labels(dependency) as depLabels,
            properties(dependency) as depProps,
            elementId(dependency) as depId
        LIMIT 500 // Add a reasonable limit
    `, labelStr, idPropsMatchStr, relTypeFilter, maxDepth)

	params := map[string]interface{}{
		"idProps": identifyingProperties,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return graph.DependencyResult{}, fmt.Errorf("failed to execute FindDependencies query: %w", err)
	}

	// Process results
	dependencies := make([]graph.EntityDetails, 0, len(result.Records))
	for _, record := range result.Records {
		depLabelsVal, _ := record.Get("depLabels")
		depPropsVal, _ := record.Get("depProps")
		depIdVal, _ := record.Get("depId")

		depLabelsInterface, _ := depLabelsVal.([]interface{})
		depLabels := make([]string, len(depLabelsInterface))
		for i, l := range depLabelsInterface {
			depLabels[i], _ = l.(string)
		}

		depProps, _ := depPropsVal.(map[string]interface{})
		depId, _ := depIdVal.(string)
		depProps["id"] = depId // Add element ID

		// Convert properties
		for k, v := range depProps {
			depProps[k] = convertNeo4jValue(v)
		}

		dependencies = append(dependencies, graph.EntityDetails{
			Labels:     depLabels,
			Properties: depProps,
		})
	}

	return graph.DependencyResult{
		TargetNode: targetNodeDetails,
		Results:    dependencies,
		Depth:      maxDepth,
		Direction:  "dependencies",
	}, nil
}

// FindDependents finds entities that depend on the target entity (incoming relationships),
// following specified relationship types up to a certain depth.
func (s *Neo4jStore) FindDependents(ctx context.Context, labels []string, identifyingProperties map[string]interface{}, relationshipTypes []string, maxDepth int) (graph.DependencyResult, error) {
	if len(labels) == 0 {
		return graph.DependencyResult{}, fmt.Errorf("at least one label is required for the target node")
	}
	if len(identifyingProperties) == 0 {
		return graph.DependencyResult{}, fmt.Errorf("at least one identifying property is required for the target node")
	}
	if maxDepth <= 0 {
		maxDepth = 1 // Default to depth 1 if invalid
	}

	// First, get the target node details
	targetNodeDetails, err := s.GetEntityDetails(ctx, labels, identifyingProperties)
	if err != nil {
		return graph.DependencyResult{}, fmt.Errorf("target node not found or error fetching details: %w", err)
	}

	// Build label string for the target node
	labelStr := ":" + strings.Join(labels, ":")

	// Build identifying properties match string for the target node
	var idPropsParts []string
	for k := range identifyingProperties {
		idPropsParts = append(idPropsParts, fmt.Sprintf("%s: $idProps.%s", k, k))
	}
	idPropsMatchStr := "{" + strings.Join(idPropsParts, ", ") + "}"

	// Build relationship type filter string
	relTypeFilter := buildRelationshipTypeFilter(relationshipTypes)

	// Construct the MATCH query for dependents (incoming relationships)
	query := fmt.Sprintf(`
        MATCH (target%s %s)
        MATCH (dependent)-[r%s*1..%d]->(target)
        WHERE target <> dependent
        RETURN DISTINCT
            labels(dependent) as depLabels,
            properties(dependent) as depProps,
            elementId(dependent) as depId
        LIMIT 500 // Add a reasonable limit
    `, labelStr, idPropsMatchStr, relTypeFilter, maxDepth)

	params := map[string]interface{}{
		"idProps": identifyingProperties,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		return graph.DependencyResult{}, fmt.Errorf("failed to execute FindDependents query: %w", err)
	}

	// Process results
	dependents := make([]graph.EntityDetails, 0, len(result.Records))
	for _, record := range result.Records {
		depLabelsVal, _ := record.Get("depLabels")
		depPropsVal, _ := record.Get("depProps")
		depIdVal, _ := record.Get("depId")

		depLabelsInterface, _ := depLabelsVal.([]interface{})
		depLabels := make([]string, len(depLabelsInterface))
		for i, l := range depLabelsInterface {
			depLabels[i], _ = l.(string)
		}

		depProps, _ := depPropsVal.(map[string]interface{})
		depId, _ := depIdVal.(string)
		depProps["id"] = depId // Add element ID

		// Convert properties
		for k, v := range depProps {
			depProps[k] = convertNeo4jValue(v)
		}

		dependents = append(dependents, graph.EntityDetails{
			Labels:     depLabels,
			Properties: depProps,
		})
	}

	return graph.DependencyResult{
		TargetNode: targetNodeDetails,
		Results:    dependents,
		Depth:      maxDepth,
		Direction:  "dependents",
	}, nil
}

// BatchFindOrCreateEntities finds or creates multiple entities in a single operation.
// This is more efficient than making multiple individual calls.
// Returns details for all entities in the same order as the input array, along with any individual errors.
func (s *Neo4jStore) BatchFindOrCreateEntities(ctx context.Context, inputs []graph.EntityInput) ([]graph.EntityDetails, []error, error) {
	if len(inputs) == 0 {
		return nil, nil, fmt.Errorf("at least one entity input is required")
	}

	// Prepare results and errors slices with the same length as inputs
	results := make([]graph.EntityDetails, len(inputs))
	individualErrors := make([]error, len(inputs))

	// Process entities in parallel with a reasonable concurrency limit
	concurrencyLimit := 10
	if len(inputs) < concurrencyLimit {
		concurrencyLimit = len(inputs)
	}

	// Create a semaphore channel to limit concurrency
	sem := make(chan struct{}, concurrencyLimit)
	var wg sync.WaitGroup

	// Process each entity input
	for i, input := range inputs {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(index int, entityInput graph.EntityInput) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Use the existing FindOrCreateEntity method for each entity
			result, err := s.FindOrCreateEntity(ctx, entityInput)
			if err != nil {
				individualErrors[index] = fmt.Errorf("error processing entity at index %d: %w", index, err)
				return
			}

			results[index] = result
		}(i, input)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check if any individual operations failed
	var failedCount int
	for _, err := range individualErrors {
		if err != nil {
			failedCount++
		}
	}

	// Return results and individual errors
	if failedCount > 0 {
		return results, individualErrors, fmt.Errorf("%d out of %d entity operations failed", failedCount, len(inputs))
	}

	return results, individualErrors, nil
}

// BatchFindOrCreateRelationships finds or creates multiple relationships in a single operation.
// This is more efficient than making multiple individual calls.
// Returns properties for all relationships in the same order as the input array, along with any individual errors.
func (s *Neo4jStore) BatchFindOrCreateRelationships(ctx context.Context, inputs []graph.RelationshipInput) ([]map[string]interface{}, []error, error) {
	if len(inputs) == 0 {
		return nil, nil, fmt.Errorf("at least one relationship input is required")
	}

	// Prepare results and errors slices with the same length as inputs
	results := make([]map[string]interface{}, len(inputs))
	individualErrors := make([]error, len(inputs))

	// Process relationships in parallel with a reasonable concurrency limit
	concurrencyLimit := 10
	if len(inputs) < concurrencyLimit {
		concurrencyLimit = len(inputs)
	}

	// Create a semaphore channel to limit concurrency
	sem := make(chan struct{}, concurrencyLimit)
	var wg sync.WaitGroup

	// Process each relationship input
	for i, input := range inputs {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(index int, relInput graph.RelationshipInput) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Use the existing FindOrCreateRelationship method for each relationship
			result, err := s.FindOrCreateRelationship(ctx, relInput)
			if err != nil {
				individualErrors[index] = fmt.Errorf("error processing relationship at index %d: %w", index, err)
				return
			}

			results[index] = result
		}(i, input)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check if any individual operations failed
	var failedCount int
	for _, err := range individualErrors {
		if err != nil {
			failedCount++
		}
	}

	// Return results and individual errors
	if failedCount > 0 {
		return results, individualErrors, fmt.Errorf("%d out of %d relationship operations failed", failedCount, len(inputs))
	}

	return results, individualErrors, nil
}

// GetEntitySubgraph retrieves nodes and relationships around a central entity up to a specified depth,
// formatted suitably for visualisation tools like Mermaid.
func (s *Neo4jStore) GetEntitySubgraph(ctx context.Context, labels []string, identifyingProperties map[string]interface{}, maxDepth int) (graph.SubgraphResult, error) {
	if len(labels) == 0 {
		return graph.SubgraphResult{}, fmt.Errorf("at least one label is required for the target node")
	}
	if len(identifyingProperties) == 0 {
		return graph.SubgraphResult{}, fmt.Errorf("at least one identifying property is required for the target node")
	}
	if maxDepth <= 0 {
		maxDepth = 1 // Default to depth 1 if invalid
	}

	// Build label string for the target node
	labelStr := ":" + strings.Join(labels, ":")

	// Build identifying properties match string for the target node
	var idPropsParts []string
	for k := range identifyingProperties {
		idPropsParts = append(idPropsParts, fmt.Sprintf("%s: $idProps.%s", k, k))
	}
	idPropsMatchStr := "{" + strings.Join(idPropsParts, ", ") + "}"

	// Construct the MATCH query to get the subgraph
	// Uses apoc.path.subgraphAll for efficient subgraph retrieval
	// Returns distinct nodes and relationships within the specified depth
	query := fmt.Sprintf(`
        MATCH (target%s %s)
        CALL apoc.path.subgraphAll(target, {maxLevel: %d})
        YIELD nodes, relationships
        RETURN
            [node IN nodes | { id: elementId(node), labels: labels(node), name: node.name, props: properties(node) }] AS subgraphNodes,
            [rel IN relationships | { id: elementId(rel), startNode: elementId(startNode(rel)), endNode: elementId(endNode(rel)), type: type(rel), props: properties(rel) }] AS subgraphRels
    `, labelStr, idPropsMatchStr, maxDepth)

	params := map[string]interface{}{
		"idProps": identifyingProperties,
	}

	// Execute query
	result, err := neo4j.ExecuteQuery(ctx, s.driver, query, params, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(""))
	if err != nil {
		// Check if the error is due to APOC procedure not found
		if strings.Contains(err.Error(), "Unknown function 'apoc.path.subgraphAll'") {
			return graph.SubgraphResult{}, fmt.Errorf("failed to execute GetEntitySubgraph query: APOC procedures might not be installed or enabled on the Neo4j server. Error: %w", err)
		}
		return graph.SubgraphResult{}, fmt.Errorf("failed to execute GetEntitySubgraph query: %w", err)
	}

	if len(result.Records) == 0 {
		// This could mean the target node wasn't found, or APOC didn't return results
		// Check if the target node exists first
		_, err := s.GetEntityDetails(ctx, labels, identifyingProperties)
		if err != nil {
			return graph.SubgraphResult{}, fmt.Errorf("target node not found or error fetching details: %w", err)
		}
		// Node exists, but subgraph is empty (or APOC issue)
		return graph.SubgraphResult{Nodes: []graph.SubgraphNode{}, Relationships: []graph.SubgraphRelationship{}}, nil
	}

	record := result.Records[0]

	// Process nodes
	nodesVal, _ := record.Get("subgraphNodes")
	nodesInterface, _ := nodesVal.([]interface{})
	subgraphNodes := make([]graph.SubgraphNode, 0, len(nodesInterface))
	for _, nodeIntf := range nodesInterface {
		nodeMap, ok := nodeIntf.(map[string]interface{})
		if !ok { continue }

		nodeID, _ := nodeMap["id"].(string)
		nodeLabelsIntf, _ := nodeMap["labels"].([]interface{})
		nodeName, _ := nodeMap["name"].(string) // Assuming 'name' is the primary display property
		nodeProps, _ := nodeMap["props"].(map[string]interface{})

		nodeLabels := make([]string, len(nodeLabelsIntf))
		for i, l := range nodeLabelsIntf {
			nodeLabels[i], _ = l.(string)
		}

		// Convert properties, potentially selecting a subset for visualisation
		convertedProps := make(map[string]interface{})
		for k, v := range nodeProps {
			// Example: Only include certain properties or simplify complex ones
			if k != "id" && k != "createdAt" && k != "lastModifiedAt" { // Exclude some common ones
				convertedProps[k] = convertNeo4jValue(v)
			}
		}
		// Ensure ID is present if not already included in props map conversion
		if _, exists := convertedProps["id"]; !exists {
			convertedProps["id"] = nodeID
		}


		subgraphNodes = append(subgraphNodes, graph.SubgraphNode{
			ID:     nodeID,
			Labels: nodeLabels,
			Name:   nodeName,
			Props:  convertedProps,
		})
	}

	// Process relationships
	relsVal, _ := record.Get("subgraphRels")
	relsInterface, _ := relsVal.([]interface{})
	subgraphRels := make([]graph.SubgraphRelationship, 0, len(relsInterface))
	for _, relIntf := range relsInterface {
		relMap, ok := relIntf.(map[string]interface{})
		if !ok { continue }

		relID, _ := relMap["id"].(string)
		startNodeID, _ := relMap["startNode"].(string)
		endNodeID, _ := relMap["endNode"].(string)
		relType, _ := relMap["type"].(string)
		relProps, _ := relMap["props"].(map[string]interface{})

		// Convert properties
		convertedProps := make(map[string]interface{})
		for k, v := range relProps {
			if k != "id" && k != "createdAt" && k != "lastModifiedAt" {
				convertedProps[k] = convertNeo4jValue(v)
			}
		}
		if _, exists := convertedProps["id"]; !exists {
			convertedProps["id"] = relID
		}


		subgraphRels = append(subgraphRels, graph.SubgraphRelationship{
			ID:        relID,
			StartNode: startNodeID,
			EndNode:   endNodeID,
			Type:      relType,
			Props:     convertedProps,
		})
	}

	return graph.SubgraphResult{
		Nodes:         subgraphNodes,
		Relationships: subgraphRels,
	}, nil
}
