package memory

import (
	"context"
)

// Store provides an interface for memory storage operations.
type Store interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context) ([]string, error)
}
