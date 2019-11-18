package oas3

import (
	"context"
)

type contextKey int

const (
	itemKey contextKey = iota
)

// WithOperation puts the current operation into the current context
func WithOperation(ctx context.Context, op *Item) context.Context {
	return context.WithValue(ctx, itemKey, op)
}

// OperationFromContext returns the current operation from the context
func OperationFromContext(ctx context.Context) *Item {
	return ctx.Value(itemKey).(*Item)
}
