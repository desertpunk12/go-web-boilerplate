package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, keys ...interface{}) {
	m.Called(msg, keys)
}

func (m *MockLogger) Error(err error, msg string) {
	m.Called(err, msg)
}

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func TestLogin_Success(t *testing.T) {
	mockLogger := new(MockLogger)
	mockDB := new(MockDB)

	h := &Handler{
		Log: mockLogger,
		DB:  mockDB,
	}

	app := fiber.New()
	app.Post("/login", h.Login)

	params := LoginParams{
		Username: "testuser",
		Password: "password123",
}
	body, _ := json.Marshal(params)

	// Expectation: Info log should be called
	mockLogger.On("Info", "login attempt", []interface{}{"params", params}).Return()
	mockDB.On("Ping").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mockLogger.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestLogin_BindError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockDB := new(MockDB)

	h := &Handler{
		Log: mockLogger,
		DB:  mockDB,
	}

	app := fiber.New()
	app.Post("/login", h.Login)

	// Send invalid JSON
	body := []byte("{invalid-json}")

	// Expectation: Error log should be called
	mockLogger.On("Error", mock.Anything, "failed to bind body").Return()

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	// Fiber returns 400 Bad Request or 422 Unprocessable Entity for bind errors usually.
	// But in the code: return err. Fiber default error handler should handle it.
	// Let's assert code is not 200.
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)

	mockLogger.AssertExpectations(t)
}

func TestLogin_ProcessError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockDB := new(MockDB)

	h := &Handler{
		Log: mockLogger,
		DB:  mockDB,
	}

	app := fiber.New()
	app.Post("/login", h.Login)

	params := LoginParams{
		Username: "testuser", // valid json
		Password: "password123",
	}
	body, _ := json.Marshal(params)

	// Expectation: Info log should be called first
	mockLogger.On("Info", "login attempt", []interface{}{"params", params}).Return()
	// Then DB Ping fails
	mockDB.On("Ping").Return(assert.AnError)
	// Then Error log
	mockLogger.On("Error", assert.AnError, "login process failed").Return()

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockLogger.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}
