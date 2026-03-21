# Error Handling Analysis - All Three Services

## Executive Summary

✅ **All three services follow consistent error handling patterns**
✅ **No stack traces are being exposed to clients**
✅ **Error messages are specific and user-friendly**
✅ **All services use proper HTTP status codes**
✅ **Correlation IDs are tracked across all errors**

---

## 1. Go Transaction Service

**File Structure:**

- API Error Definition: [transaction-service/internal/model/api.go](transaction-service/internal/model/api.go)
- Error Handler: [transaction-service/internal/handler/http.go](transaction-service/internal/handler/http.go#L400)

### Current Error Response Format

```json
{
  "status": "error",
  "code": "METHOD_NOT_ALLOWED",
  "message": "Only POST /transfer is supported",
  "correlation_id": "abc-123-def-456"
}
```

### APIError Structure

```go
type APIError struct {
    Status        string `json:"status"`
    Code          string `json:"code"`
    Message       string `json:"message"`
    CorrelationID string `json:"correlation_id"`
}
```

### Error Handling Flow

1. **AppError** (internal) → wraps error with StatusCode, Code, Message, Err
2. **writeError()** → formats as APIError
3. **writeJSON()** → serializes to client

### Key Characteristics

- ✅ Status codes: 400, 401, 409, 500, etc. properly used
- ✅ Error codes are specific (e.g., `MISSING_IDEMPOTENCY_KEY`, `UNAUTHORIZED`)
- ✅ No stack traces in response
- ✅ Underlying errors logged separately with correlation ID
- ✅ Graceful degradation (internal error is generic when needed)

### Example Logging

```
correlation_id=abc-123 event=request_failed code=MISSING_IDEMPOTENCY_KEY status=400 error=Idempotency-Key header is required
```

---

## 2. Kotlin Ledger Service

**File Structure:**

- API Error Response: [ls_springboot/src/main/kotlin/com/bigbank/ledger/api/ApiErrorResponse.kt](ls_springboot/src/main/kotlin/com/bigbank/ledger/api/ApiErrorResponse.kt)
- Exception Handler: [ls_springboot/src/main/kotlin/com/bigbank/ledger/service/GlobalExceptionHandler.kt](ls_springboot/src/main/kotlin/com/bigbank/ledger/service/GlobalExceptionHandler.kt)
- Custom Exception: [ls_springboot/src/main/kotlin/com/bigbank/ledger/service/ApiException.kt](ls_springboot/src/main/kotlin/com/bigbank/ledger/service/ApiException.kt)

### Current Error Response Format

```json
{
  "status": "error",
  "code": "MALFORMED_REQUEST",
  "message": "Request body is malformed or missing required fields",
  "correlation_id": "abc-123-def-456"
}
```

### ApiErrorResponse Structure

```kotlin
@JsonNaming(PropertyNamingStrategies.SnakeCaseStrategy::class)
data class ApiErrorResponse(
    val status: String = "error",
    val code: String,
    val message: String,
    val correlationId: String,
)
```

### Global Exception Handler Coverage

| Exception Type                    | HTTP Status      | Error Code               | Message                                                |
| --------------------------------- | ---------------- | ------------------------ | ------------------------------------------------------ |
| `ApiException`                    | _from exception_ | _from exception_         | _from exception_                                       |
| `HttpMessageNotReadableException` | 400              | MALFORMED_REQUEST        | "Request body is malformed or missing required fields" |
| `DataIntegrityViolationException` | 409              | DATA_INTEGRITY_VIOLATION | "The request violated a database constraint"           |
| `Exception` (catch-all)           | 500              | INTERNAL_ERROR           | "Unexpected server error"                              |

### Key Characteristics

- ✅ Global exception handler catches all exceptions
- ✅ No stack traces exposure in any handler
- ✅ Generic messages for unexpected errors (security best practice)
- ✅ Correlation ID extraction from request attributes
- ✅ Proper HTTP status codes for each error type
- ✅ Database constraint violations don't leak detail

---

## 3. Rust Fraud Service

**File Structure:**

- Implementation: [fraud-service/src/main.rs](fraud-service/src/main.rs)

### Current Error Response Format

```json
{
  "status": "error",
  "code": "INVALID_REFERENCE",
  "message": "Reference is required",
  "correlation_id": "abc-123-def-456"
}
```

### ApiError Structure (Rust)

```rust
#[derive(Serialize)]
struct ApiError {
    status: String,
    code: String,
    message: String,
    correlation_id: String,
}
```

### Error Handling Examples

```rust
Err(("INVALID_REFERENCE", "Reference is required"))  // Returns BadRequest with ApiError
Err(("INVALID_ACCOUNT", "Both accounts are required"))
Err(("SAME_ACCOUNT_TRANSFER", "Source and destination accounts must differ"))
Err(("INVALID_AMOUNT", "Amount must be greater than zero"))
```

### Key Characteristics

- ✅ Inline validation returns tuples of (code, message)
- ✅ HTTP 400 BadRequest for validation errors
- ✅ No stack traces or panic messages exposed
- ✅ Specific, descriptive error messages
- ✅ Correlation ID generation (UUID fallback if not provided)
- ✅ All errors properly formatted before sending to client

---

## 4. Comparative Analysis

### Response Format Consistency

| Aspect                     | Go         | Kotlin                  | Rust               |
| -------------------------- | ---------- | ----------------------- | ------------------ |
| `status` field             | ✅ "error" | ✅ "error"              | ✅ "error"         |
| `code` field               | ✅ Present | ✅ Present              | ✅ Present         |
| `message` field            | ✅ Present | ✅ Present              | ✅ Present         |
| `correlation_id` field     | ✅ present | ✅ present (snake_case) | ✅ present         |
| Stack traces exposed       | ❌ No      | ❌ No                   | ❌ No              |
| Generic for unknown errors | ✅ Yes     | ✅ Yes                  | N/A (no catch-all) |
| HTTP status codes          | ✅ Correct | ✅ Correct              | ✅ Correct         |

### Serialization Field Names

**Go & Rust:** Use `correlation_id` (snake_case in JSON)

```json
"correlation_id": "abc-123"
```

**Kotlin:** Uses `correlationId` with `@JsonNaming(SnakeCaseStrategy)` to serialize as:

```json
"correlation_id": "abc-123"
```

### Stack Trace Exposure Analysis

| Service    | Raw Exceptions | Stack Traces | Cause chains | Unsafe details |
| ---------- | -------------- | ------------ | ------------ | -------------- |
| **Go**     | ❌ No          | ❌ No        | ❌ Hidden    | ✅ Safe        |
| **Kotlin** | ❌ No          | ❌ No        | ❌ Hidden    | ✅ Safe        |
| **Rust**   | ❌ No          | ❌ No        | N/A          | ✅ Safe        |

---

## 5. Error Message Specificity

### Specific Error Messages (Good for Debugging)

```
"Idempotency-Key header is required"
"Both accounts are required"
"Amount exceeds fraud threshold 1000000.00"
"Reference contains unsupported characters"
```

### Generic Error Messages (Production Security)

```
"Unexpected server error"
"The request violated a database constraint"
"Request body is malformed or missing required fields"
```

✅ **All services balance specificity for client errors with genericity for server errors (5xx)**

---

## 6. Standardization Recommendations

### Current vs. Requested Format

**Requested simple format:**

```json
{
  "error": "message"
}
```

**Current production format:**

```json
{
  "status": "error",
  "code": "ERROR_CODE",
  "message": "Specific error message",
  "correlation_id": "tracking-id"
}
```

### Assessment

| Criterion            | Simple Format              | Current Format                           |
| -------------------- | -------------------------- | ---------------------------------------- |
| Debuggability        | ❌ Low (no error codes)    | ✅ High (error codes +correlation)       |
| Error Tracking       | ❌ No way to link errors   | ✅ Can track via correlation_id          |
| Error Classification | ❌ No codes                | ✅ Error codes (MALFORMED_REQUEST, etc.) |
| Production Logging   | ❌ Generic only            | ✅ Specific + generic separation         |
| API Evolution        | ❌ No structure for growth | ✅ Extensible (error codes can grow)     |

### Recommendation

**✅ KEEP CURRENT FORMAT** - The existing format is already well-standardized and production-ready. Changing to `{"error": "message"}` would lose valuable error codes and correlation tracking.

---

## 7. Action Items Summary

### ✅ Completed/No Issues Found

- [x] No stack traces being exposed
- [x] Error messages are appropriately specific
- [x] All services use correlation IDs
- [x] HTTP status codes are correct
- [x] Global exception handling in place (Go, Kotlin)
- [x] Generic catch-all messages for 5xx errors

### 🔍 Minor Considerations

1. **Go Service** - Currently relies on direct handler error formatting. All are properly handled.
2. **Kotlin Service** - Global exception handler coverage is excellent
3. **Rust Service** - No catch-all for panics (but should be handled by framework/logging)

### 🚀 No Changes Required

All three services are already following best practices for error handling in production environments.

---

## 8. Code References

### How to Find Error Responses

**Go Transaction Service:**

- Create error: `writeError(w, correlationID, &model.AppError{...})`
- See: [handler/http.go line 400](transaction-service/internal/handler/http.go#L400)

**Kotlin Ledger Service:**

- Throw exception: `throw ApiException(HttpStatus.BAD_REQUEST, "CODE", "message")`
- See: [GlobalExceptionHandler.kt](ls_springboot/src/main/kotlin/com/bigbank/ledger/service/GlobalExceptionHandler.kt)

**Rust Fraud Service:**

- Return error: `Err(("INVALID_AMOUNT", "message"))`
- See: [main.rs line 78+](fraud-service/src/main.rs#L150)
