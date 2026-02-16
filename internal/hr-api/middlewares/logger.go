package middlewares

import (
	fiberzerolog "github.com/gofiber/contrib/v3/zerolog"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func SetupLogger(app *fiber.App, log *zerolog.Logger) {
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: log,
		GetLogger: func(c fiber.Ctx) zerolog.Logger {
			// Custom IP extraction logic
			ips := append(c.IPs(), c.IP(), c.Get("CF-Connecting-IP"))
			return log.With().Strs("ips", ips).Logger()
		},
		Fields: []string{
			fiberzerolog.FieldLatency,
			fiberzerolog.FieldStatus,
			fiberzerolog.FieldMethod,
			fiberzerolog.FieldURL,
			fiberzerolog.FieldError,
			fiberzerolog.FieldRequestID,
		},
	}))
}
