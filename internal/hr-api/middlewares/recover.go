package middlewares

import (
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"web-boilerplate/internal/hr-api/interfaces"
)

// SetupMiddlewareRecover adds panic recovery middleware to the Fiber app
// It recovers from panics, logs them with stack trace, and returns a 500 error
func SetupMiddlewareRecover(app *fiber.App, log interfaces.Logger) {
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e any) {
			stack := debug.Stack()
			// Log panic details using Info to support structured key-value logging
			log.Info("panic occurred",
				"error", e,
				"path", c.Path(),
				"method", c.Method(),
				"ip", c.IP(),
				"stack", string(stack))
		},
	}))
}
