package service

import (
	"context"
	"fmt"
)

// InitialiseSchema sets up the initial schema for the knowledge graph
func (s *Service) InitialiseSchema(ctx context.Context) error {
	// Define schema for Dgraph
	schema := `
		# Node type predicate
		type: string @index(exact) .

		# Document predicates
		title: string @index(fulltext, term) .
		content: string @index(fulltext) .
		metadata: json .

		# Concept predicates
		name: string @index(fulltext, term) .

		# Edge predicates
		RELATED_TO: uid @reverse .
		CONTAINS: uid @reverse .
		REFERENCES_TO: uid @reverse .
		CREATED_BY: uid @reverse .
		HAS_PROPERTY: uid @reverse .

		# Define types
		type Document {
			type
			title
			content
			metadata
			CONTAINS
			REFERENCES_TO
			CREATED_BY
		}

		type Concept {
			type
			name
			RELATED_TO
			HAS_PROPERTY
		}

		type Entity {
			type
			name
			properties
			RELATED_TO
		}

		type Event {
			type
			name
			date
			description
			RELATED_TO
		}
	`

	// Upsert schema
	if err := s.graph.UpsertSchema(ctx, schema); err != nil {
		return fmt.Errorf("failed to initialise schema: %w", err)
	}

	return nil
}
