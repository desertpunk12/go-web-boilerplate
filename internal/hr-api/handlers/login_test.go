package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"
	"web-boilerplate/internal/hr-api/repositories"
	"web-boilerplate/shared/helpers"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestLogin_Success(t *testing.T) {
	// 1. Setup Mock
	hashedPassword, _ := helpers.HashPass("password123")
	mockRepo := &MockQuerier{
		GetUserByUsernameFunc: func(ctx context.Context, username string) (repositories.User, error) {
			return repositories.User{
				ID:       pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4}, Valid: true},
				Username: "testuser",
				Password: hashedPassword,
			}, nil
		},
	}
	mockLogger := &MockLogger{}

	// 2. Initialize Handler
	h := &Handler{
		Log:  mockLogger,
		Repo: mockRepo,
		Pool: &MockPool{},
	}

	// 3. Setup Fiber app
	app := fiber.New()
	app.Post("/login", h.Login)

	// 4. Create Request
	payload := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 5. Execute Request
	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: 20 * time.Second,
	})
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	// 6. Assertions
	assert.Equal(t, 200, resp.StatusCode)

	var respBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NotEmpty(t, respBody["token"])
	assert.Equal(t, "01020304-0000-0000-0000-000000000000", respBody["id"])
}

func TestLogin_InvalidBody(t *testing.T) {
	h := &Handler{
		Log:  &MockLogger{},
		Repo: &MockQuerier{},
	}
	app := fiber.New()
	app.Post("/login", h.Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: 20 * time.Second,
	})
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 400, resp.StatusCode)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := &MockQuerier{
		GetUserByUsernameFunc: func(ctx context.Context, username string) (repositories.User, error) {
			return repositories.User{}, errors.New("user not found")
		},
	}
	h := &Handler{
		Log:  &MockLogger{},
		Repo: mockRepo,
	}
	app := fiber.New()
	app.Post("/login", h.Login)

	payload := map[string]string{
		"username": "unknown",
		"password": "password",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: 20 * time.Second,
	})
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 401, resp.StatusCode)
}

func TestLogin_InvalidPassword(t *testing.T) {
	hashedPassword, _ := helpers.HashPass("correct-password")
	mockRepo := &MockQuerier{
		GetUserByUsernameFunc: func(ctx context.Context, username string) (repositories.User, error) {
			return repositories.User{
				ID:       pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4}, Valid: true},
				Username: "testuser",
				Password: hashedPassword,
			}, nil
		},
	}
	h := &Handler{
		Log:  &MockLogger{},
		Repo: mockRepo,
	}
	app := fiber.New()
	app.Post("/login", h.Login)

	payload := map[string]string{
		"username": "testuser",
		"password": "wrong-password",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: 20 * time.Second,
	})
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 401, resp.StatusCode)
}

func TestLogin_DBError(t *testing.T) {
	mockRepo := &MockQuerier{
		GetUserByUsernameFunc: func(ctx context.Context, username string) (repositories.User, error) {
			return repositories.User{}, errors.New("db connection error")
		},
	}
	h := &Handler{
		Log:  &MockLogger{},
		Repo: mockRepo,
	}
	app := fiber.New()
	app.Post("/login", h.Login)

	payload := map[string]string{
		"username": "testuser",
		"password": "password",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: 20 * time.Second,
	})
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 401, resp.StatusCode)
}
