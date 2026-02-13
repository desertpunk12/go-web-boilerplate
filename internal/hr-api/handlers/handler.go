package handlers

import (
	"web-boilerplate/internal/hr-api/db"
	"web-boilerplate/internal/hr-api/interfaces"
	"web-boilerplate/internal/hr-api/pkg/logger"
	"web-boilerplate/internal/hr-api/repositories"

	"github.com/rs/zerolog"
)

type Handler struct {
	Log  interfaces.Logger
	Repo repositories.Querier
	Pool interfaces.DBPool
}

func New(log *zerolog.Logger, dbInst *db.Database) *Handler {
	return &Handler{
		Log:  logger.NewZerologAdapter(log),
		Repo: repositories.New(dbInst.Pool),
		Pool: dbInst.Pool,
	}
}
