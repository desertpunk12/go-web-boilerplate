package middlewares

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

// SetupMiddlewareRequestID adds request ID generation and propagation
// Uses Fiber's built-in requestid middleware which:
// - Generates unique request ID if X-Request-ID header is not present
// - Propagates existing X-Request-ID header if provided
// - Stores request ID in Fiber context for use in handlers
// - Adds X-Request-ID to response headers
// The zerolog middleware automatically includes request ID in all logs via FieldRequestId
func SetupMiddlewareRequestID(app *fiber.App) {
	app.Use(requestid.New())
}
