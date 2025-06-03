package middlewares

import "github.com/gofiber/fiber/v3"

func SetupMiddlewares(app *fiber.App) {
	// app.Use()
	SetupMiddlewaresEssentials(app)
}
