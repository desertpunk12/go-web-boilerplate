package middlewares

import (
	"errors"
	"net/http/httptest"
	"testing"
	"web-boilerplate/internal/hr-api/interfaces"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRecoverMiddleware_Panic(t *testing.T) {
	mockLogger := interfaces.NewMockLogger(t)
	// Expect Info to be called with panic details
	mockLogger.EXPECT().Info("panic occurred", mock.Anything)

	app := fiber.New()
	SetupMiddlewareRecover(app, mockLogger)

	// Add a route that panics
	app.Get("/panic", func(c fiber.Ctx) error {
		panic(errors.New("test panic"))
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestRecoverMiddleware_NoPanic(t *testing.T) {
	mockLogger := interfaces.NewMockLogger(t)
	// Logger should not be called when no panic occurs

	app := fiber.New()
	SetupMiddlewareRecover(app, mockLogger)

	app.Get("/ok", func(c fiber.Ctx) error {
		return c.Status(200).SendString("ok")
	})

	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestRecoverMiddleware_StringPanic(t *testing.T) {
	mockLogger := interfaces.NewMockLogger(t)
	mockLogger.EXPECT().Info("panic occurred", mock.Anything)

	app := fiber.New()
	SetupMiddlewareRecover(app, mockLogger)

	app.Get("/panic-string", func(c fiber.Ctx) error {
		panic("string panic")
	})

	req := httptest.NewRequest("GET", "/panic-string", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}
