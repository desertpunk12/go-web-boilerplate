package interfaces

import "context"

type Logger interface {
	Info(msg string, keys ...interface{})
	Error(err error, msg string)
}

type DB interface {
	Ping() error
	// Add other methods as needed here
}

type RedisDB interface {
	Ping(ctx context.Context) error
}
