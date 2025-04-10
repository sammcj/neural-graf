package service

import (
	"context"
	"fmt"
)

// InitialiseSchema sets up the initial schema for the knowledge graph
func (s *Service) InitialiseSchema(ctx context.Context) error {
	// Define schema for Neo4j/Memgraph using Cypher syntax
	schema := `
		// Create indexes for node properties
		CREATE INDEX FOR (d:Document) ON (d.title);
		CREATE INDEX FOR (d:Document) ON (d.type);
		CREATE INDEX FOR (c:Concept) ON (c.name);
		CREATE INDEX FOR (c:Concept) ON (c.type);
		CREATE INDEX FOR (e:Entity) ON (e.name);
		CREATE INDEX FOR (e:Entity) ON (e.type);
		CREATE INDEX FOR (ev:Event) ON (ev.name);
		CREATE INDEX FOR (ev:Event) ON (ev.type);

		// Create constraints for unique IDs
		CREATE CONSTRAINT FOR (d:Document) REQUIRE d.id IS UNIQUE;
		CREATE CONSTRAINT FOR (c:Concept) REQUIRE c.id IS UNIQUE;
		CREATE CONSTRAINT FOR (e:Entity) REQUIRE e.id IS UNIQUE;
		CREATE CONSTRAINT FOR (ev:Event) REQUIRE ev.id IS UNIQUE;
	`

	// Upsert schema
	if err := s.graph.UpsertSchema(ctx, schema); err != nil {
		return fmt.Errorf("failed to initialise schema: %w", err)
	}

	return nil
}
