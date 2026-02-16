package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/idempotency"
)

// SetupIdempotency configures and applies the idempotency middleware
// to prevent duplicate requests. By default, it skips safe methods
// (GET, HEAD, OPTIONS, TRACE) and only processes POST, PUT, DELETE.
//
// Usage:
//
//	middlewares.SetupIdempotency(app)
//
// The middleware checks for X-Idempotency-Key header (36-char UUID format)
// and returns 409 Conflict for duplicate requests within the lifetime window.
func SetupIdempotency(app *fiber.App) {
	app.Use(idempotency.New(idempotency.Config{
		Lifetime: 5 * time.Minute,
	}))
}
