package interfaces

import "context"

type Logger interface {
	Info(msg string, keys ...interface{})
	Error(err error, msg string)
}

type RedisDB interface {
	Ping(ctx context.Context) error
}

type DBPool interface {
	Ping(ctx context.Context) error
}
