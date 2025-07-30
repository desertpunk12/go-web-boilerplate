package routes

import (
	"web-boilerplate/internal/hr-api/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App) {
	v1 := app.Group("/v1")

	v1.Get("/health", handlers.Health)

	v1.Post("/login", handlers.LoginHandler)
}
