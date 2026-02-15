package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"web-boilerplate/internal/hr-api/interfaces"
	"web-boilerplate/internal/hr-api/repositories"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetMe_Success(t *testing.T) {
	userID := uuid.UUID{1, 2, 3, 4}

	mockRepo := repositories.NewMockQuerier(t)
	mockRepo.EXPECT().GetUser(context.Background(), pgtype.UUID{Bytes: userID, Valid: true}).Return(repositories.User{
		ID:       pgtype.UUID{Bytes: userID, Valid: true},
		Name:     "Test User",
		Email:    "test@example.com",
		Username: "testuser",
		Password: "hashed-password",
	}, nil)

	mockLogger := interfaces.NewMockLogger(t)

	h := &Handler{
		Log:  mockLogger,
		Repo: mockRepo,
	}

	app := fiber.New()

	// Middleware that sets up the user claims (simulating the auth middleware)
	app.Use(func(c fiber.Ctx) error {
		c.Locals("user", map[string]interface{}{
			"id":  userID.String(),
			"exp": float64(9999999999), // far future
		})
		return c.Next()
	})

	app.Get("/me", h.GetMe)

	req := httptest.NewRequest("GET", "/me", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 200, resp.StatusCode)

	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "01020304-0000-0000-0000-000000000000", respBody["id"])
	assert.Equal(t, "Test User", respBody["name"])
	assert.Equal(t, "test@example.com", respBody["email"])
	assert.Equal(t, "testuser", respBody["username"])
	assert.Nil(t, respBody["password"])
}

func TestGetMe_NoUserClaims(t *testing.T) {
	mockLogger := interfaces.NewMockLogger(t)
	mockLogger.EXPECT().Error(nil, "failed to get user claims from context")

	h := &Handler{
		Log: mockLogger,
	}

	app := fiber.New()

	// Middleware that doesn't set user claims
	app.Use(func(c fiber.Ctx) error {
		return c.Next()
	})

	app.Get("/me", h.GetMe)

	req := httptest.NewRequest("GET", "/me", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 401, resp.StatusCode)
}

func TestGetMe_InvalidUserID(t *testing.T) {
	mockLogger := interfaces.NewMockLogger(t)
	mockLogger.EXPECT().Error(mock.Anything, "invalid user id format")

	h := &Handler{
		Log: mockLogger,
	}

	app := fiber.New()

	app.Use(func(c fiber.Ctx) error {
		c.Locals("user", map[string]interface{}{
			"id":  "invalid-uuid",
			"exp": float64(9999999999),
		})
		return c.Next()
	})

	app.Get("/me", h.GetMe)

	req := httptest.NewRequest("GET", "/me", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 401, resp.StatusCode)
}

func TestGetMe_UserNotFound(t *testing.T) {
	userID := uuid.UUID{1, 2, 3, 4}

	mockRepo := repositories.NewMockQuerier(t)
	mockRepo.EXPECT().GetUser(context.Background(), pgtype.UUID{Bytes: userID, Valid: true}).Return(repositories.User{}, assert.AnError)

	mockLogger := interfaces.NewMockLogger(t)
	mockLogger.EXPECT().Error(assert.AnError, "user not found")

	h := &Handler{
		Log:  mockLogger,
		Repo: mockRepo,
	}

	app := fiber.New()

	app.Use(func(c fiber.Ctx) error {
		c.Locals("user", map[string]interface{}{
			"id":  userID.String(),
			"exp": float64(9999999999),
		})
		return c.Next()
	})

	app.Get("/me", h.GetMe)

	req := httptest.NewRequest("GET", "/me", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 404, resp.StatusCode)
}

func TestGetMe_IDMissing(t *testing.T) {
	mockLogger := interfaces.NewMockLogger(t)
	mockLogger.EXPECT().Error(nil, "invalid user id in claims")

	h := &Handler{
		Log: mockLogger,
	}

	app := fiber.New()

	app.Use(func(c fiber.Ctx) error {
		c.Locals("user", map[string]interface{}{
			"exp": float64(9999999999),
		})
		return c.Next()
	})

	app.Get("/me", h.GetMe)

	req := httptest.NewRequest("GET", "/me", nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 401, resp.StatusCode)
}
