package storage

import (
	"context"
	"errors"
)

var ErrNotImplemented = errors.New("storage not implemented")

// DB is a placeholder for a future PostgreSQL connection.
// It allows the rest of the app to compile without wiring DB yet.
type DB struct{}

func Open(ctx context.Context, dsn string) (*DB, error) {
	_ = ctx
	_ = dsn
	return nil, ErrNotImplemented
}
