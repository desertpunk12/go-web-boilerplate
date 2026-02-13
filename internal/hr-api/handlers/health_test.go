package handlers

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestHealth_Success(t *testing.T) {
	mockPool := &MockPool{
		PingFunc: func(ctx context.Context) error {
			return nil
		},
	}
	h := &Handler{
		Log:  &MockLogger{},
		Pool: mockPool,
		Repo: &MockQuerier{},
	}

	app := fiber.New()
	app.Get("/health", h.Health)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}

func TestHealth_DBFailure(t *testing.T) {
	mockPool := &MockPool{
		PingFunc: func(ctx context.Context) error {
			return errors.New("db down")
		},
	}
	h := &Handler{
		Log:  &MockLogger{},
		Pool: mockPool,
		Repo: &MockQuerier{},
	}

	app := fiber.New()
	app.Get("/health", h.Health)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 500, resp.StatusCode)
}
