package routes

import (
	"web-boilerplate/internal/hr-api/db"
	"web-boilerplate/internal/hr-api/handlers"
	"web-boilerplate/internal/hr-api/middlewares"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func SetupRoutes(app *fiber.App, log *zerolog.Logger, db *db.Database) {
	v1 := app.Group("/v1")

	// Initialize handlers
	h := handlers.New(log, db)

	v1.Get("/health", h.Health)

	v1.Post("/login", h.Login)

	// Protected routes
	v1.Get("/me", middlewares.Protected, h.GetMe)
}
