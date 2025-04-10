package neo4j

import (
	"context"
	"fmt"
	"strings"
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
