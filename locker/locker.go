package locker

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"time"
)

type redisLocker struct {
	locker *redislock.Client
	locks  map[string]*redislock.Lock
}

// MyLockerInterface create a mock locker interface
type MyLockerInterface interface {
	Obtain(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, key string) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Refresh(ctx context.Context, key string, ttl time.Duration) error
}

// CreateRedisLock create a redis lock
func CreateRedisLock(redisClient *redis.Client) MyLockerInterface {
	locker := redislock.New(redisClient)
	return &redisLocker{locker: locker, locks: map[string]*redislock.Lock{}}
}

// Obtain get a lock from a key and duration
func (l *redisLocker) Obtain(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lock, err := l.locker.Obtain(ctx, key, ttl, nil)

	ok := err == nil

	if err == redislock.ErrNotObtained {
		err = nil
	}

	if lock != nil {
		l.locks[key] = lock
	}

	return ok, err
}

// Release release the lock
func (l *redisLocker) Release(ctx context.Context, key string) error {
	return l.locks[key].Release(ctx)
}

// TTL return the life duration of the lock from the lock key
func (l *redisLocker) TTL(ctx context.Context, key string) (time.Duration, error) {
	return l.locks[key].TTL(ctx)
}

// Refresh add some time to the ttl life from the lock key and time to add
func (l *redisLocker) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	return l.locks[key].Refresh(ctx, ttl, nil)
}
