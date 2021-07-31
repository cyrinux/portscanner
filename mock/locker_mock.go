package mock

import (
	"context"
	"time"
)

// MockLocker is a lock mocker
type MockLocker struct {
	ObtainImpl func(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

// Obtain get a lock from a key and duration
func (l *MockLocker) Obtain(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return l.ObtainImpl(ctx, key, ttl)
}

// Release release the lock
func (l *MockLocker) Release(ctx context.Context, key string) error {
	return nil
}

// TTL return the life duration of the lock from the lock key
func (l *MockLocker) TTL(ctx context.Context, key string) (time.Duration, error) {
	now := time.Now()

	duration := now.Add(5 * time.Minute)

	ttl := duration.Sub(now)

	return ttl, nil
}

// Refresh add some time to the ttl life from the lock key and time to add
func (l *MockLocker) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}
