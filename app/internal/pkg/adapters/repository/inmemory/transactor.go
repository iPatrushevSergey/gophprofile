package inmemory

import "context"

// Transactor is a no-op for in-memory storage.
type Transactor struct{}

// NewTransactor returns an in-memory transactor.
func NewTransactor() Transactor {
	return Transactor{}
}

// RunInTransaction runs fn on the same context without starting a transaction.
func (Transactor) RunInTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
