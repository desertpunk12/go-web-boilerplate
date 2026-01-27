package middlewares

import (
	"github.com/gofiber/contrib/fiberzerolog"

	"github.com/rs/zerolog"

	"github.com/gofiber/fiber/v3"
)

func SetupLogger(app *fiber.App, log *zerolog.Logger) {
	// Middleware uses the passed logger instance
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: log,
		Fields: []string{fiberzerolog.FieldIPs, fiberzerolog.FieldLatency, fiberzerolog.FieldStatus,
			fiberzerolog.FieldMethod, fiberzerolog.FieldURL, fiberzerolog.FieldError},
	}))
}
