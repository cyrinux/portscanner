package locker

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/cyrinux/grpcnmapscanner/config"
	"github.com/cyrinux/grpcnmapscanner/helpers"
	// "github.com/pkg/errors"
	// "github.com/cyrinux/grpcnmapscanner/database"
	"time"
)

type MyLocker interface {
	Obtain(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, key string) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Refresh(ctx context.Context, key string, ttl time.Duration) error
}

type redisLocker struct {
	locker *redislock.Client
	locks  map[string]*redislock.Lock
}

func CreateRedisLock(ctx context.Context, conf config.Config) MyLocker {
	redisClient := helpers.NewRedisClient(ctx, conf).Connect()
	locker := redislock.New(redisClient)
	return &redisLocker{locker: locker}
}

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

func (l *redisLocker) Release(ctx context.Context, key string) error {
	return l.locks[key].Release(ctx)
}

func (l *redisLocker) TTL(ctx context.Context, key string) (time.Duration, error) {
	return l.locks[key].TTL(ctx)
}

func (l *redisLocker) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	return l.locks[key].Refresh(ctx, ttl, nil)
}
