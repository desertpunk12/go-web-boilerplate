package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(logLevel string) *zerolog.Logger {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Create a new logger instance (not using global zerolog.Log)
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		With().
		Timestamp().
		Stack().
		Logger()

	return &logger
}
