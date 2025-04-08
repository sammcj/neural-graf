package dgraph

import (
	"context"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
)

// DgraphClient defines the interface for the Dgraph client
// This interface is used for mocking in tests
type DgraphClient interface {
	// NewTxn creates a new transaction
	NewTxn() DgraphTxn
	// NewReadOnlyTxn creates a new read-only transaction
	NewReadOnlyTxn() DgraphTxn
	// Alter runs schema operations
	Alter(ctx context.Context, op *api.Operation) error
}

// DgraphClientWrapper wraps the Dgraph client to implement the DgraphClient interface
type DgraphClientWrapper struct {
	client *dgo.Dgraph
}

// NewDgraphClientWrapper creates a new DgraphClientWrapper
func NewDgraphClientWrapper(client *dgo.Dgraph) *DgraphClientWrapper {
	return &DgraphClientWrapper{
		client: client,
	}
}

// DgraphTxnWrapper wraps the Dgraph transaction to implement the DgraphTxn interface
type DgraphTxnWrapper struct {
	txn *dgo.Txn
}

// Mutate performs a mutation
func (w *DgraphTxnWrapper) Mutate(ctx context.Context, mu *api.Mutation) (*api.Response, error) {
	return w.txn.Mutate(ctx, mu)
}

// Query performs a query
func (w *DgraphTxnWrapper) Query(ctx context.Context, q string) (*api.Response, error) {
	return w.txn.Query(ctx, q)
}

// QueryWithVars performs a query with variables
func (w *DgraphTxnWrapper) QueryWithVars(ctx context.Context, q string, vars map[string]string) (*api.Response, error) {
	return w.txn.QueryWithVars(ctx, q, vars)
}

// Discard discards the transaction
func (w *DgraphTxnWrapper) Discard(ctx context.Context) error {
	return w.txn.Discard(ctx)
}

// Commit commits the transaction
func (w *DgraphTxnWrapper) Commit(ctx context.Context) error {
	return w.txn.Commit(ctx)
}

// NewTxn creates a new transaction
func (w *DgraphClientWrapper) NewTxn() DgraphTxn {
	return &DgraphTxnWrapper{
		txn: w.client.NewTxn(),
	}
}

// NewReadOnlyTxn creates a new read-only transaction
func (w *DgraphClientWrapper) NewReadOnlyTxn() DgraphTxn {
	return &DgraphTxnWrapper{
		txn: w.client.NewReadOnlyTxn(),
	}
}

// Alter runs schema operations
func (w *DgraphClientWrapper) Alter(ctx context.Context, op *api.Operation) error {
	return w.client.Alter(ctx, op)
}
