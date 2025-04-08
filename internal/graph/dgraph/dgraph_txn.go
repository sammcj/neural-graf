package dgraph

import (
	"context"

	"github.com/dgraph-io/dgo/v2/protos/api"
)

// DgraphTxn defines the interface for the Dgraph transaction
// This interface is used for mocking in tests
type DgraphTxn interface {
	// Mutate performs a mutation
	Mutate(ctx context.Context, mu *api.Mutation) (*api.Response, error)
	// Query performs a query
	Query(ctx context.Context, q string) (*api.Response, error)
	// QueryWithVars performs a query with variables
	QueryWithVars(ctx context.Context, q string, vars map[string]string) (*api.Response, error)
	// Discard discards the transaction
	Discard(ctx context.Context) error
	// Commit commits the transaction
	Commit(ctx context.Context) error
}
