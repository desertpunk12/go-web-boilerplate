# Client-Side Request Deduplication

Prevents accidental duplicate submissions across hr-web frontend and hr-api backend.

## Overview

This feature prevents duplicate requests caused by:
- Double-clicking submit buttons
- Form resubmission on slow networks
- Browser retry logic on poor connections
- Accidental page refreshes

## Architecture

### Frontend (hr-web)

**Location:** `assets/js/deduplication.js`

**Features:**
1. **Generate unique IDs**: UUID v4 format for each form submission
2. **Include headers**: `Idempotency-Key: <uuid>` in all requests (Fiber standard)
3. **Track pending requests**: In-memory Map with timestamps
4. **Loading states**: Disable submit button, show "Processing..." text
5. **Handle responses**:
   - 409 Conflict: Show duplicate message
   - 429 Too Many Requests: Show rate limit message
   - Other errors: Display error message
6. **Auto-cleanup**: Remove expired entries every 60 seconds

**Usage Example:**
```javascript
const dedup = new RequestDeduplication();
dedup.initCleanup();

await dedup.submitForm('/v1/login', formData, {
  formId: 'login-form',
  submitButtonId: 'submit-btn',
  loadingMessage: 'Logging in...',
  onSuccess: (data) => {
    document.cookie = `auth_token=${data.token}; path=/; max-age=86400`;
    window.location.href = '/home';
  },
  onConflict: (message) => {
    document.getElementById('error').textContent = message;
  },
  onError: (error) => {
    document.getElementById('error').textContent = error.message;
  }
});
```

### Backend (hr-api)

**Uses Fiber's built-in Idempotency Middleware**

**Location:** `internal/hr-api/middlewares/middlewares.go`

**Features:**
1. **Built-in middleware**: Uses `github.com/gofiber/fiber/v3/middleware/idempotency`
2. **RFC compliant**: Follows IETF RFC 7231 §4.2.2 for idempotent methods
3. **Automatic handling**: Checks `X-Idempotency-Key` header automatically
4. **Smart filtering**: Skips safe methods (GET, HEAD, OPTIONS, TRACE) by default
5. **In-memory cache**: Built-in caching with configurable lifetime
6. **Duplicate detection**: Returns 409 Conflict if ID found in window
7. **Helper functions**: `idempotency.IsFromCache()`, `idempotency.WasPutToCache()`

**Configuration:**
```go
import (
    "github.com/gofiber/fiber/v3/middleware/idempotency"
    "time"
)

// Global setup - applies to all non-safe methods (POST, PUT, DELETE)
app.Use(idempotency.New(idempotency.Config{
    Lifetime: 5 * time.Minute,
}))
```

**Default Config:**
- **Header**: `X-Idempotency-Key` (must be 36 chars, UUID format)
- **Lifetime**: 30 minutes (customized to 5 minutes)
- **Storage**: In-memory
- **Skip**: Safe methods (GET, HEAD, OPTIONS, TRACE)

**API Response for Duplicates:**
```json
{
  "error": "Duplicate request",
  "message": "This request is already being processed",
  "key": "uuid-v4-here"
}
```

## Configuration

### Time Window
- **Current**: 5 minutes
- **Location**: `idempotency.go` - `IdempotencyWindow` constant

### Cache Cleanup
- **Frequency**: Called periodically from application main
- **Retention**: Entries kept for 2x window (10 minutes) before cleanup
- **Function**: `CleanExpiredIdempotencyKeys()`

## Migration Path

### Phase 1: In-Memory Cache (Current)
- ✅ Implemented in `handlers/idempotency.go`
- ✅ Integrated into `handlers/login.go`
- ✅ Tests passing
- Works for single-instance deployments

### Phase 2: Redis Cache (Future)
When Redis is integrated:
1. Replace in-memory cache with Redis client
2. Use `SETEX` with TTL (5 minutes)
3. Allow cross-instance deduplication
4. Use `GET` and `SET` operations for idempotency

Example Redis implementation:
```go
func CheckIdempotency(key string) (bool, error) {
  exists, _ := redisClient.Exists(ctx, "idempotency:"+key)
  return exists, nil
}

func StoreIdempotency(key string) error {
  return redisClient.Set(ctx, "idempotency:"+key, "1", 5*time.Minute)
}
```

## Security Considerations

- **ID Format**: UUID v4 provides sufficient randomness
- **Time Window**: 5 minutes balances UX vs security
- **Client-Side**: Generates IDs - server validates (prevents spoofing)
- **Backend**: Only trusts IDs from legitimate clients
- **Cleanup**: Prevents memory exhaustion attacks

## Testing

### Unit Tests
- ✅ No key handling
- ✅ New key detection
- ✅ Within window detection
- ✅ Store functionality
- ✅ Cleanup functionality

### Integration Tests
- Test with browser network throttling
- Test double-click scenarios
- Test 409 response handling
- Test 429 response handling

## Benefits

1. **Prevents double-charging**: Critical for payment operations
2. **Better UX**: Shows loading states, prevents user confusion
3. **Reduced server load**: Avoids processing duplicate requests
4. **Request tracing**: Each unique ID can be tracked in logs
5. **Network resilience**: Clients can retry safely with same ID

## Implementation Details

### Frontend Files Modified:
- `assets/js/deduplication.js` - Request deduplication utility class
- `assets/css/deduplication.css` - Alert, loading, and form states
- `ui/layouts/base.templ` - Added deduplication.js script and CSS import
- `internal/hr-web/ui/pages/login.templ` - Added error/success/loading divs, form IDs

### Backend Files Modified:
- `internal/hr-api/handlers/idempotency.go` - In-memory cache implementation
- `internal/hr-api/handlers/idempotency_test.go` - Comprehensive test suite
- `internal/hr-api/handlers/login.go` - Check X-Idempotency-Key header, return 409 on duplicate
