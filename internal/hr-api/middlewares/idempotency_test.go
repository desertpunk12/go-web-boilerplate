package middlewares

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/idempotency"
	"github.com/stretchr/testify/assert"
)

func TestIdempotency_DuplicateRequest(t *testing.T) {
	app := fiber.New()

	// Setup idempotency middleware
	SetupIdempotency(app)

	requestCount := 0

	// Add a test handler that tracks calls
	app.Post("/test", func(c fiber.Ctx) error {
		requestCount++
		return c.JSON(fiber.Map{"status": "success", "count": requestCount})
	})

	idempotencyKey := "550e8400-e29b-41d4-a716-446655440000"

	// First request - should succeed and call handler
	req1 := httptest.NewRequest("POST", "/test", nil)
	req1.Header.Set("X-Idempotency-Key", idempotencyKey)
	resp1, err := app.Test(req1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)
	assert.Equal(t, 1, requestCount, "Handler should be called once")

	// Second request with same key - should return cached response
	// Fiber's idempotency middleware returns 200 with cached response, not 409
	req2 := httptest.NewRequest("POST", "/test", nil)
	req2.Header.Set("X-Idempotency-Key", idempotencyKey)
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)
	assert.Equal(t, 1, requestCount, "Handler should NOT be called again - response from cache")
}

func TestIdempotency_DifferentKeys(t *testing.T) {
	app := fiber.New()

	// Setup idempotency middleware
	SetupIdempotency(app)

	// Add a test handler
	app.Post("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "success"})
	})

	// First request with key 1
	req1 := httptest.NewRequest("POST", "/test", nil)
	req1.Header.Set("X-Idempotency-Key", "550e8400-e29b-41d4-a716-446655440001")
	resp1, err := app.Test(req1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// Second request with key 2 - should succeed (different key)
	req2 := httptest.NewRequest("POST", "/test", nil)
	req2.Header.Set("X-Idempotency-Key", "550e8400-e29b-41d4-a716-446655440002")
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode)
}

func TestIdempotency_SafeMethodsSkipped(t *testing.T) {
	app := fiber.New()

	// Setup idempotency middleware
	SetupIdempotency(app)

	// Add a test handler
	app.Get("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "success"})
	})

	idempotencyKey := "550e8400-e29b-41d4-a716-446655440003"

	// GET requests should skip idempotency check (safe method)
	// Multiple GET requests with same key should all succeed
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Idempotency-Key", idempotencyKey)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode, "GET request %d should succeed", i+1)
	}
}

func TestIdempotency_InvalidKeyFormat(t *testing.T) {
	app := fiber.New()

	// Setup idempotency middleware with default validation
	SetupIdempotency(app)

	// Add a test handler
	app.Post("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "success"})
	})

	// Request with invalid key (not 36 chars) - should fail validation
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-Idempotency-Key", "invalid-key")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Should return error for invalid key format (Fiber returns 400 or 500)
	assert.NotEqual(t, 200, resp.StatusCode, "Invalid key should not succeed")
}

func TestIdempotency_NoKey(t *testing.T) {
	app := fiber.New()

	// Setup idempotency middleware
	SetupIdempotency(app)

	// Add a test handler
	app.Post("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "success"})
	})

	// Request without idempotency key - should succeed
	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestIdempotency_Expiration(t *testing.T) {
	app := fiber.New()

	// Setup idempotency middleware with very short lifetime for testing
	app.Use(idempotency.New(idempotency.Config{
		Lifetime: 100 * time.Millisecond, // Very short for testing
	}))

	// Add a test handler
	app.Post("/test", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "success"})
	})

	idempotencyKey := "550e8400-e29b-41d4-a716-446655440004"

	// First request - should succeed
	req1 := httptest.NewRequest("POST", "/test", nil)
	req1.Header.Set("X-Idempotency-Key", idempotencyKey)
	resp1, err := app.Test(req1)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp1.StatusCode)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Second request with same key after expiration - should succeed
	req2 := httptest.NewRequest("POST", "/test", nil)
	req2.Header.Set("X-Idempotency-Key", idempotencyKey)
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp2.StatusCode, "Request after expiration should succeed")
}
