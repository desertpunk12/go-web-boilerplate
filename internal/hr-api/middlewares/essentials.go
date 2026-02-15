package middlewares

import (
	"strings"
	"web-boilerplate/internal/hr-api/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
)

func SetupMiddlewaresEssentials(app *fiber.App) {
	SetupMiddlewareHelmet(app)
	SetupMiddlewareCompress(app)
}

func SetupMiddlewareCompress(app *fiber.App) {
	app.Use(compress.New())
}

func SetupMiddlewareHelmet(app *fiber.App) {
	app.Use(helmet.New())
}

func SetupMiddlewareCORS(app *fiber.App) {
	app.Use(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     strings.Split(config.ALLOWED_ORIGINS, ","),
		AllowHeaders:     []string{"Origin", " Content-Type", " Accept", " Accept-Language", " Content-Length"},
	})
}
