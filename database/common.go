package database

import (
	"github.com/go-redis/redis/v8"
)

func generateError(operation string, err error) (string, error) {
	if err == redis.Nil {
		return "", &OperationError{operation}
	}
	return "", &DownError{}
}
