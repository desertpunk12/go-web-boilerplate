package handlers

import (
	"context"
	"web-boilerplate/internal/hr-api/repositories"
)

// MockQuerier embeds the interface to satisfy compiler,
// and adds function fields for test configuration
type MockQuerier struct {
	repositories.Querier
	GetUserByUsernameFunc func(ctx context.Context, username string) (repositories.User, error)
	// Add other mock functions here as needed for other tests
}

func (m *MockQuerier) GetUserByUsername(ctx context.Context, username string) (repositories.User, error) {
	if m.GetUserByUsernameFunc != nil {
		return m.GetUserByUsernameFunc(ctx, username)
	}
	return repositories.User{}, nil
}

// MockLogger implements interfaces.Logger
type MockLogger struct {
	InfoFunc  func(msg string, keys ...interface{})
	ErrorFunc func(err error, msg string)
}

func (m *MockLogger) Info(msg string, keys ...interface{}) {
	if m.InfoFunc != nil {
		m.InfoFunc(msg, keys...)
	}
}

func (m *MockLogger) Error(err error, msg string) {
	if m.ErrorFunc != nil {
		m.ErrorFunc(err, msg)
	}
}

// MockPool implements DBPool
type MockPool struct {
	PingFunc func(ctx context.Context) error
}

func (m *MockPool) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil // Default success
}
