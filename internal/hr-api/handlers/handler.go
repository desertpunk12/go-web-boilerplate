package handlers

import (
	"web-boilerplate/internal/hr-api/interfaces"
	"web-boilerplate/internal/hr-api/pkg/logger"

	"github.com/rs/zerolog"
)

type Handler struct {
	Log interfaces.Logger
	DB  interfaces.DB
}

func New(log *zerolog.Logger, db interfaces.DB) *Handler {
	return &Handler{
		Log: logger.NewZerologAdapter(log),
		DB:  db,
	}
}
