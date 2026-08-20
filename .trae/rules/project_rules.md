# Project Rules for RemiAq - PocketBase + Go Worker

## Project Overview
RemiAq is a reminder notification system built with **Go 1.25** and **PocketBase v0.29+**. The system provides FCM-based notifications with lunar calendar support and background worker processing.

## Language and Documentation Rules

### Comments and Documentation
- Use **Vietnamese** for docstrings and comments in business logic
- Use **English** for technical comments and GoDoc documentation
- All exported functions, methods, types, interfaces, and structs **must** have GoDoc comments
- Comments should describe **what** the function does and **what** it returns, not **how** it's implemented
- Keep GoDoc comments concise (1-2 lines preferred)
- Start comments with the function/type name (GoDoc standard)

### Comment Examples
```go
// Get retrieves the system status by ID.
func (r *SystemStatusRepo) Get(ctx context.Context) (*models.SystemStatus, error) { ... }

// mapToSystemStatus converts raw DB rows to SystemStatus models.
// Used internally by Get() and IsWorkerEnabled().
func (r *SystemStatusRepo) mapToSystemStatus(raw dbx.NullStringMap) (*models.SystemStatus, error) { ... }
```

## Project Structure and Organization

### Directory Structure
```
cmd/
├── server/          # Main server application
└── worker/          # Background worker (if separate)

internal/
├── api/             # API layer (deprecated, use handlers/)
├── db/              # Database abstraction layer
├── handlers/        # HTTP request handlers
├── middleware/      # HTTP middleware
├── models/          # Data models and structs
├── repository/      # Repository interfaces and implementations
│   └── pocketbase/  # PocketBase-specific implementations
├── services/        # Business logic services
├── utils/           # Utility functions
└── worker/          # Background worker implementation

config/              # Configuration management
migrations/          # Database migrations
web/                 # Static web files (if any)
```

### Module and Package Naming
- Module name: `remiaq`
- Package imports use full module path: `remiaq/internal/models`
- Repository implementations in `internal/repository/pocketbase/`
- All models in `internal/models/`

## Database Access Layer Rules

### DBHelper Usage
- **NEVER** call `app.DB()` directly in repository code
- **ALWAYS** use `DBHelper` abstraction from `internal/db/`
- Use generic functions: `db.GetOne[T]()`, `db.GetAll[T]()`, `db.GetOneWithConfig[T]()`
- Use `helper.Exec()` for INSERT/UPDATE/DELETE operations
- Use `helper.Exists()` and `helper.Count()` for existence checks

### Database Mapping Rules
1. **Generic Mapping**
   - Always use `MapNullStringMapToStruct[T]` or `MapNullStringMapToStructWithConfig[T]`
   - Never manually map fields from `raw["field"].String` unless special handling required
   - Use `db` struct tags for field mapping: `db:"field_name"`

2. **Required Field Validation**
   - Use `RequiredFields` in mapper config for critical fields (ID, Updated, etc.)
   - Example:
   ```go
   cfg := &db.MapperConfig{
       RequiredFields: []string{"ID", "Updated"},
   }
   ```

3. **Custom Mappers**
   - Implement `CustomMapper` interface for complex field parsing
   - Use in `MapperConfig.CustomMappers` map

### Transaction Management
- Use `db.InTransaction()` for all transactional operations
- **NEVER** nest raw `app.RunInTransaction()` calls
- Example:
```go
err := db.InTransaction(app, func(tx *db.DBHelper) error {
    return tx.Exec("UPDATE musers SET name={:n} WHERE id={:id}", 
        dbx.Params{"n": "John", "id": 1})
})
```

## Repository Pattern Rules

### Interface Definition
- All repository interfaces defined in `internal/repository/interface.go`
- Use context.Context as first parameter in all methods
- Return pointers for single entities, slices for collections
- Use specific method names: `GetByID()`, `GetByUserID()`, `UpdateStatus()`

### Implementation Rules
- Implement repositories in `internal/repository/pocketbase/`
- Naming convention: `SystemStatusRepo`, `UserRepo`, `ReminderRepo`
- Always include interface compliance check:
```go
var _ repository.SystemStatusRepository = (*SystemStatusRepo)(nil)
```

### Repository Method Patterns
```go
// Single entity retrieval
func (r *SystemStatusRepo) Get(ctx context.Context) (*models.SystemStatus, error) {
    return db.GetOne[models.SystemStatus](r.helper,
        "SELECT * FROM system_status WHERE mid = {:mid}",
        dbx.Params{"mid": 1})
}

// Collection retrieval
func (r *ReminderRepo) GetByUserID(ctx context.Context, userID string) ([]*models.Reminder, error) {
    return db.GetAll[models.Reminder](r.helper,
        "SELECT * FROM reminders WHERE user_id = {:user_id}",
        dbx.Params{"user_id": userID})
}

// Update operations
func (r *SystemStatusRepo) EnableWorker(ctx context.Context) error {
    return r.helper.Exec(
        "UPDATE system_status SET worker_enabled = TRUE, updated = {:updated} WHERE mid = 1",
        dbx.Params{"updated": time.Now().UTC()})
}
```

## Model Definition Rules

### Struct Tags
- Use both `json` and `db` tags for all fields
- JSON tags for API serialization, DB tags for database mapping
- Example:
```go
type Reminder struct {
    ID          string    `json:"id" db:"id"`
    Title       string    `json:"title" db:"title"`
    Created     time.Time `json:"created" db:"created"`
}
```

### Field Naming Conventions
- Use PascalCase for struct fields
- Use snake_case for database columns
- Use camelCase for JSON fields
- Time fields should be `time.Time` type
- Use pointers for optional fields: `*time.Time`, `*string`

### Model Validation
- Implement `Validate()` method for models that need validation
- Return `*ValidationError` for validation failures
- Include business logic methods in models when appropriate

## Handler and API Rules

### Handler Structure
- Handlers in `internal/handlers/`
- One handler per domain entity
- Constructor pattern: `NewReminderHandler(service)`
- Methods should match HTTP verbs: `CreateReminder`, `GetReminder`, `UpdateReminder`

### Request/Response Patterns
- Always call `middleware.SetCORSHeaders(re)` first
- Use `utils.SendSuccess()` and `utils.SendError()` for responses
- Decode JSON requests using `json.NewDecoder(re.Request.Body).Decode()`
- Use context from request: `re.Request.Context()`

### Error Handling
```go
func (h *ReminderHandler) GetReminder(re *core.RequestEvent) error {
    middleware.SetCORSHeaders(re)
    
    id := re.Request.PathValue("id")
    reminder, err := h.service.GetByID(re.Request.Context(), id)
    if err != nil {
        return utils.SendError(re, 500, "Failed to get reminder", err)
    }
    
    return utils.SendSuccess(re, "", reminder)
}
```

## Service Layer Rules

### Service Structure
- Business logic in `internal/services/`
- Services coordinate between repositories and external services
- Constructor pattern with dependency injection
- Use interfaces for external dependencies (FCM, etc.)

### Service Method Patterns
- Accept context as first parameter
- Return domain models, not database models
- Handle business logic validation
- Coordinate multiple repository calls when needed

## Configuration Management

### Config Structure
- Configuration in `config/config.go`
- Use environment variables with fallbacks
- Struct-based configuration with `Load()` function
- Example:
```go
type Config struct {
    ServerAddr     string
    WorkerInterval int
    FCMCredentials string
    Environment    string
}
```

## Testing Rules

### Test File Organization
- Create `_test.go` files for every new Go file
- Use `testing` and `testify/assert` packages
- Group tests with `t.Run()` blocks
- Test file naming: `filename_test.go`

### Test Structure
```go
func TestReminderService_Create(t *testing.T) {
    t.Run("should create reminder successfully", func(t *testing.T) {
        // Arrange
        // Act  
        // Assert
    })
    
    t.Run("should return error for invalid data", func(t *testing.T) {
        // Test cases
    })
}
```

### Test Requirements
- Run `go test -v` after implementation
- All tests must pass before considering code complete
- Include both positive and negative test cases
- Mock external dependencies

## Worker and Background Processing

### Worker Structure
- Background workers in `internal/worker/`
- Use context for cancellation: `context.WithCancel()`
- Implement graceful shutdown
- Use configurable intervals from config

### Worker Pattern
```go
type Worker struct {
    repo     repository.SystemStatusRepository
    service  *services.ReminderService
    interval time.Duration
}

func (w *Worker) Start(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(w.interval)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                w.processReminders(ctx)
            }
        }
    }()
}
```

## Security and Best Practices

### Security Rules
- Never log or expose secrets/keys
- Never commit credentials to repository
- Use environment variables for sensitive configuration
- Validate all input data
- Use parameterized queries (handled by DBHelper)

### Error Handling
- Log errors with context information
- Return user-friendly error messages
- Use structured error types when appropriate
- Don't expose internal implementation details in API errors

### Performance Guidelines
- Use connection pooling (handled by PocketBase)
- Implement proper indexing in database
- Use pagination for large result sets
- Cache frequently accessed data when appropriate

## Code Quality Standards

### General Principles
- Follow Go idioms and conventions
- Keep functions small and focused
- Use meaningful variable and function names
- Avoid deep nesting (max 3-4 levels)
- Prefer composition over inheritance

### Import Organization
```go
import (
    // Standard library
    "context"
    "time"
    
    // Third-party packages
    "github.com/pocketbase/pocketbase"
    
    // Local packages
    "remiaq/internal/models"
    "remiaq/internal/repository"
)
```

### Dependency Management
- Use Go modules (`go.mod`)
- Pin specific versions for stability
- Regular dependency updates with testing
- Minimal external dependencies

## Development Workflow

### Code Generation Rules
- Generate corresponding `_test.go` for every new Go file
- Run tests after implementation: `go test -v`
- Continue only if all tests pass
- Use `go fmt` and `go vet` before commits

### Git Workflow
- Use meaningful commit messages
- Keep commits atomic and focused
- Test before pushing
- Use feature branches for new functionality

This comprehensive rule set ensures consistency, maintainability, and quality across the RemiAq project while leveraging the strengths of Go and PocketBase.

các table có created và updated không cần phải insert, pocketbase tự động insert cho rồi
chạy lệnh không dùng && để tránh lỗi, ví dụ cd d:\PROJECT\nhacviecfcm && go run test_parse.go sẽ lỗi trên windows

không cần chạy lệnh cd đến thư mục chính của nhacviecfcm project nữa vì cmd luôn ở đó 
để run project : go run ./cmd/server serve

#api để query db khi cần test: GET /api/rquery?query= {sql}
