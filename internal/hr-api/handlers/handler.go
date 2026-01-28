package handlers

import (
	"web-boilerplate/internal/hr-api/db"
	"web-boilerplate/internal/hr-api/interfaces"
	"web-boilerplate/internal/hr-api/pkg/logger"
	"web-boilerplate/internal/hr-api/repositories"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Handler struct {
	Log  interfaces.Logger
	Repo *repositories.Queries
	Pool *pgxpool.Pool
}

func New(log *zerolog.Logger, dbInst *db.Database) *Handler {
	return &Handler{
		Log:  logger.NewZerologAdapter(log),
		Repo: repositories.New(dbInst.Pool),
		Pool: dbInst.Pool,
	}
}
