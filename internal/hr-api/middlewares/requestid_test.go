package middlewares

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	app := fiber.New()
	SetupMiddlewareRequestID(app)

	app.Get("/test", func(c fiber.Ctx) error {
		// Retrieve request ID from context using FromContext
		rid := requestid.FromContext(c)
		assert.NotEmpty(t, rid)
		return c.JSON(fiber.Map{"request_id": rid})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify response header contains request ID
	requestID := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "X-Request-ID header should be set")
}

func TestRequestIDMiddleware_PropagatesExistingID(t *testing.T) {
	app := fiber.New()
	SetupMiddlewareRequestID(app)

	app.Get("/test", func(c fiber.Ctx) error {
		rid := requestid.FromContext(c)
		return c.JSON(fiber.Map{"request_id": rid})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// Set existing request ID in header
	customID := "custom-request-id-12345"
	req.Header.Set("X-Request-ID", customID)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify the custom ID is propagated
	requestID := resp.Header.Get("X-Request-ID")
	assert.Equal(t, customID, requestID, "Should propagate existing request ID")
}

func TestRequestIDMiddleware_IDInResponse(t *testing.T) {
	app := fiber.New()
	SetupMiddlewareRequestID(app)

	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify X-Request-ID is always in response
	requestID := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, requestID, "X-Request-ID should always be in response headers")
}

func TestRequestIDMiddleware_PropagatesMultipleRequests(t *testing.T) {
	app := fiber.New()
	SetupMiddlewareRequestID(app)

	app.Get("/test", func(c fiber.Ctx) error {
		rid := requestid.FromContext(c)
		return c.JSON(fiber.Map{"request_id": rid})
	})

	// Make multiple requests and verify different IDs
	req1 := httptest.NewRequest("GET", "/test", nil)
	resp1, _ := app.Test(req1)
	id1 := resp1.Header.Get("X-Request-ID")

	req2 := httptest.NewRequest("GET", "/test", nil)
	resp2, _ := app.Test(req2)
	id2 := resp2.Header.Get("X-Request-ID")

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2, "Different requests should have different IDs")
}
