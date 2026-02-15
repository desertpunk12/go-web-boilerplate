package middlewares

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
	"web-boilerplate/internal/hr-api/config"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestProtected_Success(t *testing.T) {
	// Create a valid token using the same config.SECRET_KEY
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  "test-user-id",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.SECRET_KEY))
	assert.NoError(t, err)

	app := fiber.New()
	app.Get("/protected", Protected, func(c fiber.Ctx) error {
		claims := c.Locals("user")
		return c.JSON(fiber.Map{"user": claims})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", tokenString)

	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: 20 * time.Second,
	})
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var respBody map[string]any
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NotNil(t, respBody["user"])
}

func TestProtected_MissingToken(t *testing.T) {
	app := fiber.New()
	app.Get("/protected", Protected, func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)

	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: 20 * time.Second,
	})
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	var respBody map[string]any
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Unauthorized", respBody["message"])
}

func TestProtected_InvalidToken(t *testing.T) {
	app := fiber.New()
	app.Get("/protected", Protected, func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "invalid-token-string")

	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: 20 * time.Second,
	})
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	var respBody map[string]any
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Invalid or expired token", respBody["message"])
}

func TestProtected_WrongSecret(t *testing.T) {
	// Create a token with a different secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  "test-user-id",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	// Sign with a different secret than config.SECRET_KEY
	tokenString, err := token.SignedString([]byte("wrong-secret"))
	assert.NoError(t, err)

	app := fiber.New()
	app.Get("/protected", Protected, func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", tokenString)

	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: 20 * time.Second,
	})
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	var respBody map[string]any
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Invalid or expired token", respBody["message"])
}

func TestProtected_ExpiredToken(t *testing.T) {
	// Create an expired token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  "test-user-id",
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
	})

	tokenString, err := token.SignedString([]byte(config.SECRET_KEY))
	assert.NoError(t, err)

	app := fiber.New()
	app.Get("/protected", Protected, func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", tokenString)

	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: 20 * time.Second,
	})
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	var respBody map[string]any
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Invalid or expired token", respBody["message"])
}
