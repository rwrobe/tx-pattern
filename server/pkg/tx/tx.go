package tx

import (
	"context"
	"errors"
)

// Context is a marker interface for transaction context types.
type Context interface{}

// TransactionManager defines an interface for managing transactions.
type TransactionManager interface {
	Do(ctx context.Context, fn func(txCtx Context) error) error
}

var ErrInvalidTxContext = errors.New("invalid transaction context")
