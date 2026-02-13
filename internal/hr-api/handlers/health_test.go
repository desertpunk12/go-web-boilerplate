package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"web-boilerplate/internal/hr-api/interfaces"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHealth_Success(t *testing.T) {
	mockPool := interfaces.NewMockDBPool(t)
	mockPool.EXPECT().Ping(context.Background()).Return(nil)

	mockLogger := interfaces.NewMockLogger(t)
	mockLogger.EXPECT().Info("health check passed", mock.Anything)

	h := Handler{
		Pool: mockPool,
		Log:  mockLogger,
	}

	app := fiber.New()
	app.Get("/health", h.Health)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestHealth_DBFailure(t *testing.T) {
	mockPool := interfaces.NewMockDBPool(t)
	mockPool.EXPECT().Ping(context.Background()).Return(assert.AnError)

	mockLogger := interfaces.NewMockLogger(t)
	mockLogger.EXPECT().Error(assert.AnError, "database ping failed")

	h := Handler{
		Pool: mockPool,
		Log:  mockLogger,
	}

	app := fiber.New()
	app.Get("/health", h.Health)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 500, resp.StatusCode)
}
