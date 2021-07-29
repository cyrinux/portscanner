package locker

import (
	"context"
	"time"
)

// MockLocker is a lock mocker
type MockLocker struct {
}

// CreateLock create a Mocklock
// return interface MyLocker
func CreateLock(MockClient *MockLocker) MyLocker {
	return &MockLocker{}
}

// Obtain get a lock from a key and duration
func (l *MockLocker) Obtain(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return true, nil
}

// Release release the lock
func (l *MockLocker) Release(ctx context.Context, key string) error {
	return nil
}

// TTL return the life duration of the lock from the lock key
func (l *MockLocker) TTL(ctx context.Context, key string) (time.Duration, error) {
	duration := time.Now().Sub(time.Now())
	return duration, nil
}

// Refresh add some time to the ttl life from the lock key and time to add
func (l *MockLocker) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}
