package middlewares

import (
	"os"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/rs/zerolog"
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
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"Origin", " Content-Type", " Accept", " Accept-Language", " Content-Length"},
	})
}

func SetupMiddlewareFiberZerolog(app *fiber.App) {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &logger,
	}))
}

func SetupMiddlewareLogger(app *fiber.App) {
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${ip} ${status} - ${latency} ${method} ${path} ${error}",
	}))
}
