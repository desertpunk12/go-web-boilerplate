package middlewares

import "github.com/gofiber/fiber/v3"

func SetupMiddlewares(app *fiber.App) {
	SetupIdempotency(app)
	SetupMiddlewaresEssentials(app)
}
