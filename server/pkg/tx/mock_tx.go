package tx

import (
	"context"
	"sync"
)

// MockTxContext is a simple transaction context used by the mock manager.
// It's an empty struct but useful to distinguish transaction context values in tests.
type MockTxContext struct{}

// MockTransactionManager is a configurable mock implementation of TransactionManager
// intended for tests. It records calls and optionally allows overriding behavior
// via the DoFunc field.
type MockTransactionManager struct {
	// DoFunc, if set, will be invoked for Do calls. It receives the original
	// context and the function to execute and should return whatever error the
	// test wants to simulate.
	DoFunc func(ctx context.Context, fn func(txCtx Context) error) error

	mu      sync.Mutex
	Calls   int             // number of times Do was called
	LastCtx context.Context // the context passed to the last Do
	LastTx  Context         // the tx context passed to the last fn when using default behavior
}

// Do implements TransactionManager. If DoFunc is provided it is invoked
// (and the mock records the call). Otherwise the mock will create a
// *MockTxContext, record it in LastTx, run the provided fn with that
// context and return the fn's error.
func (m *MockTransactionManager) Do(ctx context.Context, fn func(txCtx Context) error) error {
	if m == nil {
		// behave as a no-op transaction manager
		return fn(nil)
	}

	m.mu.Lock()
	m.Calls++
	m.LastCtx = ctx
	m.mu.Unlock()

	if m.DoFunc != nil {
		return m.DoFunc(ctx, fn)
	}

	// default behavior: call fn with a MockTxContext and capture it
	tx := &MockTxContext{}
	m.mu.Lock()
	m.LastTx = tx
	m.mu.Unlock()

	return fn(tx)
}

// Reset clears recorded call metadata on the mock.
func (m *MockTransactionManager) Reset() {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = 0
	m.LastCtx = nil
	m.LastTx = nil
}

// NewMockTransactionManager returns a ready-to-use mock manager.
func NewMockTransactionManager() *MockTransactionManager {
	return &MockTransactionManager{}
}

// NoopTransactionManager is a trivial TransactionManager that simply runs
// the function without creating any transaction context (fn receives nil).
// Useful in tests where transactional behaviour is not required.
type NoopTransactionManager struct{}

func (NoopTransactionManager) Do(ctx context.Context, fn func(txCtx Context) error) error {
	return fn(nil)
}
