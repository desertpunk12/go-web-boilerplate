package logger

import (
	"web-boilerplate/internal/hr-api/interfaces"

	"github.com/rs/zerolog"
)

type ZerologAdapter struct {
	logger *zerolog.Logger
}

func NewZerologAdapter(l *zerolog.Logger) interfaces.Logger {
	return &ZerologAdapter{logger: l}
}

func (z *ZerologAdapter) Info(msg string, keys ...interface{}) {
	event := z.logger.Info()
	if len(keys)%2 == 0 {
		for i := 0; i < len(keys); i += 2 {
			key, ok := keys[i].(string)
			if ok {
				event.Any(key, keys[i+1])
			}
		}
	}
	event.Msg(msg)
}

func (z *ZerologAdapter) Error(err error, msg string) {
	z.logger.Error().Err(err).Msg(msg)
}
