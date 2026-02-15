# AGENTS.md

This file contains guidelines for agentic coding assistants working in this repository.

## Build / Lint / Test Commands

### Running the Application
- `make be` - Run the API backend (`go run ./cmd/hrapp-api/main.go`)
- `make tg` - Run templ generate with watch (`templ generate --watch --proxy="http://localhost:4000"`)
- `make tcss` - Run tailwind CSS compilation with watch
- `make s` - Run all three (templ, tailwind, backend) in parallel

### Testing
- `go test ./...` - Run all tests in the project
- `go test ./path/to/package` - Run tests in a specific package
- `go test -run TestFunctionName ./path/to/package` - Run a single test function
- `go test -v ./path/to/package` - Run tests with verbose output
- `go test -cover ./path/to/package` - Run tests with coverage

### Code Generation
- `mockery --all` - Generate mocks using mockery (configured in .mockery.yml)
- `sqlc generate` - Generate database code from SQL queries (configured in sqlc.yaml)

## Code Style Guidelines

### Project Structure
- **cmd/** - Entry points for buildable programs (main.go files)
- **internal/** - Private, module-specific code (not reusable outside this project)
  - **internal/{module}/handlers/** - HTTP request handlers
  - **internal/{module}/repositories/** - Database access (sqlc-generated)
  - **internal/{module}/middlewares/** - Fiber middleware functions
  - **internal/{module}/config/** - Configuration loading and management
  - **internal/{module}/db/** - Database connection setup (PostgreSQL, Redis)
  - **internal/{module}/routes/** - Route definitions and setup
- **pkg/** - Public, reusable packages
- **shared/** - Common code shared across multiple services/modules
- **lambdas/** - AWS Lambda functions

### Imports
Order imports: standard library → third-party → internal/project. Group related imports together.

Example:
```go
import (
    "context"
    "time"

    "github.com/gofiber/fiber/v3"
    "github.com/rs/zerolog"

    "web-boilerplate/internal/hr-api/handlers"
    "web-boilerplate/shared/helpers"
)
```

### Naming Conventions
- **Functions/Methods**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase; use short names in short scopes (ctx, err, db, log)
- **Constants**: PascalCase or UPPER_CASE for global constants
- **Structs**: PascalCase (e.g., `Handler`, `LoginParams`)
- **Interfaces**: PascalCase, simple names (e.g., `Logger`, `DBPool`)
- **Files**: lowercase with underscores (e.g., `handler.go`, `login_test.go`)
- **Packages**: lowercase, single word, no underscores (e.g., `handlers`, `config`)

### Error Handling
- Always check and return errors; never ignore them
- Use `errors.Is()` to check specific errors
- Use `fmt.Errorf("context: %w", err)` to wrap errors
- Log errors before returning them in handlers
- Return appropriate Fiber error types from handlers (fiber.ErrBadRequest, fiber.ErrUnauthorized, etc.)

Example:
```go
if err != nil {
    h.Log.Error(err, "failed to perform operation")
    return fiber.ErrInternalServerError
}
```

### Logging
- Use the `Logger` interface defined in interfaces.go, not zerolog directly in handlers
- Methods: `Log.Info(msg string, keys ...interface{})` and `Log.Error(err error, msg string)`
- Use structured logging with key-value pairs
- Example: `h.Log.Info("login successful", "username", user.Username, "id", userID)`

### Handlers
- Handlers are methods on the `Handler` struct
- Signature: `func (h *Handler) FunctionName(c fiber.Ctx) error`
- Use `c.Bind().Body(&params)` to parse request bodies
- Return `c.JSON(response)` for JSON responses
- Return status with message for errors: `return fiber.ErrBadRequest`
- Use `c.Context()` to get context.Context for database calls

Example:
```go
func (h *Handler) Login(c fiber.Ctx) error {
    var params LoginParams
    if err := c.Bind().Body(&params); err != nil {
        h.Log.Error(err, "failed to bind body")
        return fiber.ErrBadRequest
    }
    return c.JSON(fiber.Map{"token": token})
}
```

### Middleware
- Middlewares are functions that accept and return `fiber.Ctx`
- Use `c.Next()` to pass control to next handler
- Use `c.Locals()` to store data for downstream handlers
- Example: `c.Locals("user", claims)`

### Testing
- Use `github.com/stretchr/testify/assert` and `github.com/stretchr/testify/mock`
- Generate mocks using mockery (configured in .mockery.yml)
- Test structure: Setup mocks → Create handler → Create Fiber app → Create request → Execute → Assert
- Use `httptest.NewRequest` to create test requests
- Use `app.Test(req)` to execute requests
- Test success path and error paths separately

Example:
```go
func TestLogin_Success(t *testing.T) {
    mockRepo := repositories.NewMockQuerier(t)
    mockRepo.EXPECT().GetUserByUsername(context.Background(), "testuser").Return(user, nil)
    mockLogger := interfaces.NewMockLogger(t)
    mockLogger.EXPECT().Info("login successful", mock.Anything)

    h := &Handler{Log: mockLogger, Repo: mockRepo}
    app := fiber.New()
    app.Post("/login", h.Login)

    body, _ := json.Marshal(map[string]string{"username": "testuser", "password": "password"})
    req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)

    assert.Equal(t, 200, resp.StatusCode)
}
```

### Database
- PostgreSQL with pgx/v5 driver
- Use sqlc to generate database code from SQL queries
- SQL queries in `internal/{module}/repositories/queries/`
- Migrations in `internal/{module}/repositories/migrations/`
- Generated code in `internal/{module}/repositories/`
- Use pgtype.UUID for UUID fields
- Use `context.Background()` for database calls in handlers

### Configuration
- Load config with `config.LoadAllConfig()` at application startup
- Environment variables loaded from `.env` or `cmd/{app}/.env`
- Configuration values are global variables in `config` package
- Default values defined in `config/globalvars.go`

### Security
- Never commit secrets (API keys, database URLs, etc.)
- Use bcrypt for password hashing via `shared/helpers.HashPass()`
- Use JWT tokens for authentication via `golang-jwt/jwt/v5`
- Validate JWT signing method in middleware
- Store secrets in environment variables, never in code

### Type Definitions
- Use struct tags for JSON serialization: `json:"fieldName"`
- Use inline comments for exported functions and types
- Keep parameter lists under 4 parameters; consider a config struct for more

### Framework
- Web framework: Fiber v3 (github.com/gofiber/fiber/v3)
- Templating: templ (github.com/a-h/templ)
- HTTP client responses use `fiber.Ctx` methods
- Routes are grouped by version (e.g., `v1 := app.Group("/v1")`)

### General Go Conventions
- Use `go fmt` to format code
- Use `goimports` to manage imports (mockery is configured to use goimports)
- Prefer explicit error checks over ignoring errors with `_`
- Keep functions small and focused on single responsibility
- Avoid package-level variables except for configuration
- Use `defer` for cleanup (database connections, file handles, etc.)

## Known Issues and Improvements

### 🔴 Critical Security Issues (Fix Immediately)
1. **Hardcoded JWT secret**: `middlewares/auth.go:27` contains `"your-secret-key"` - must use `config.SECRET_KEY`
2. **Weak default secret**: `config/globalvars.go` has default `SECRET_KEY = "qweasd123"` - reject in production
3. **CORS wildcard**: `ALLOWED_ORIGINS` defaults to empty which allows `"*"` - require explicit config in production
4. **No rate limiting**: Exposes API to brute force attacks
5. **No input sanitization**: Vulnerable to XSS attacks
6. **No CSRF protection**: Web forms vulnerable

### 🟠 High Priority Fixes
1. **Mixed Fiber versions**: hr-web uses `fiber/v2`, hr-api uses `fiber/v3.0.0-rc.3` - standardize on one version
2. **Invalid Go version**: `go.mod` specifies `1.25.5` which doesn't exist - update to `1.23` or `1.24`
3. **Protected middleware unused**: Exists in code but never attached to routes
4. **No panic recovery middleware**: Server crashes on unhandled panics
5. **No request ID middleware**: Makes debugging distributed requests difficult
6. **No migration tool**: Only raw SQL files - add goose or migrate
7. **Redis unused**: Connected but not used for caching or sessions
8. **No transaction examples**: Show how to use pgx transactions
9. **No standardized response format**: Each handler returns different structure
10. **No request validation**: Using struct tags but no validation layer
11. **No pagination pattern**: All list endpoints return all records
12. **Config as globals**: Replace with structured config, add validation
13. **Missing .env.example**: Document all required environment variables
14. **S3 config broken**: `internal/hr-api/db/s3.go` has empty vars that can't be set
15. **No refresh tokens**: JWT tokens have no refresh mechanism
16. **No password strength validation**: Accepts weak passwords
17. **No account lockout**: No brute force protection on login

### 🟡 Medium Priority Enhancements
1. **docker-compose.yml**: Add for local development environment
2. **Pre-commit hooks**: Run tests, lint, fmt before commit
3. **Air/livereload**: Auto-restart on code changes
4. **Makefile targets**: Add `make test`, `make lint`, `make coverage`
5. **Test database setup**: Use testcontainers or separate test DB
6. **Integration tests**: Test full request-response flows
7. **Repository tests**: Test sqlc-generated code
8. **Middleware tests**: Test auth, CORS, recovery
9. **OpenAPI/Swagger docs**: Add API documentation
10. **Custom error types**: Create domain-specific errors
11. **Error response formatting**: Consistent error structure
12. **Metrics collection**: Add Prometheus middleware
13. **Graceful shutdown**: Handle SIGTERM properly
14. **golangci-lint config**: Add `.golangci.yml`
15. **Code coverage reporting**: Add coverage thresholds
16. **Rate limiting**: Add middleware or Redis-based limiting
17. **CSRF protection**: For web forms
18. **Request timeout**: Per-route or global timeout handling
19. **Client-side validation**: Reduce unnecessary round trips
20. **Loading states**: Better UX on async operations
21. **Toast notifications**: User feedback system

### 🔵 Low Priority (Nice to Have)
1. **E2E tests**: With Playwright or similar
2. **Test fixtures**: Factories for test data
3. **Lambda deployment**: Serverless framework or SAM templates
4. **Lambda shared utilities**: Common patterns for lambdas
5. **Table-driven tests**: For handlers with multiple cases

### Tech Stack Notes
- **Web Framework**: Fiber v3 (API), Fiber v2 (web) - needs unification
- **Database**: PostgreSQL with pgx/v5 driver
- **ORM**: None - uses sqlc for type-safe SQL
- **Templating**: templ for HTML
- **Authentication**: JWT with HS256 signing
- **Logging**: zerolog with structured logging
- **Testing**: testify (assert + mock), mockery for mocks
