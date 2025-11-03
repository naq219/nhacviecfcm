# Project Rules for PocketBase + Go Worker


- Use Vietnamese for docstrings and comments.

## Commenting Rules for Go Code

### Public (Exported) Symbols
- Always comment on **exported functions, methods, types, interfaces, and structs**.
- Each comment must begin with the name of the function or type (GoDoc standard).
  - ✅ Example:
    ```go
    // Get retrieves the system status by ID.
    func (r *Repo) Get(ctx context.Context, id int) (*SystemStatus, error) { ... }
    ```
- Comments should describe **what the function does** and **what it returns**, not how it is implemented.
- Keep comments in **English** and concise (1–2 lines preferred).

### Private (Unexported) Symbols
- Comments are **not mandatory** for unexported (private) functions, methods, or structs.
- However, comments **must be added** when:
  1. The logic is **non-trivial** (complex data flow, reflection, concurrency, DB transaction, etc.).
  2. The behavior could be **misleading** or unexpected.
  3. The code implements a **workaround** or depends on a specific external library behavior.
- Comments for private methods should explain **why the code exists or behaves this way**, not just restate the code.

  - ✅ Example:
    ```go
    // mapToSystemStatus converts raw DB rows to SystemStatus models.
    // Used internally by Get() and IsWorkerEnabled().
    func (r *Repo) mapToSystemStatus(raw dbx.NullStringMap) (*models.SystemStatus, error) { ... }
    ```

### General Style
- Use complete sentences starting with a capital letter and ending with a period.
- Avoid redundant comments like “// SetWorkerEnabled sets the workerEnabled field”.
- If code is self-explanatory, omit the comment.

### Goal
These rules help AI and developers:
- Maintain consistent GoDoc quality.
- Focus on clarity and reasoning behind the code.
- Avoid unnecessary verbosity in obvious functions.


## General Principles
- The project is built with **Go 1.25** and **PocketBase v0.29+**
- The database access layer is implemented via **DBHelper** abstraction built on top of `github.com/pocketbase/dbx`
- All repository code must use `DBHelper` helper functions instead of raw queries.

---

## File Organization
- Common database helpers are defined in `internal/db/db.go`
- Repository interfaces are in `internal/repository/interface.go`
- Each PocketBase-based repository (system_status_repo.go, user_repo.go, etc.) must:
  - Use `db.GetOne[T]`, `db.GetAll[T]`, `db.Exec()`, `db.Exists()`, `db.InTransaction()`
  - Never call `app.DB()` directly.
- All model structs are under `internal/models/`

---

## Database Access Rules
1. **Generic Mapping**
   - Always use `MapNullStringMapToStruct[T]` or `MapNullStringMapToStructWithConfig[T]` for converting `dbx.NullStringMap` to structs.
   - Never manually map fields from `raw["field"].String` unless you need special handling.
2. **Transactions**
   - Use `helper.InTransaction(func(tx *db.DBHelper) error { ... })` for all transactional logic.
   - Do not nest raw `app.RunInTransaction()` calls.
3. **Validation**
   - Always use `RequiredFields` in mapper config to ensure critical fields (like ID, Updated) are not empty.

---

## Repository Implementation Rules
- Each repository must implement its interface from `internal/repository/interface.go`
- Naming convention:
  - `SystemStatusRepo`, `UserRepo`, `ReminderRepo`...
- Example pattern:
  ```go
  func (r *SystemStatusRepo) Get(ctx context.Context) (*models.SystemStatus, error) {
      return db.GetOne[models.SystemStatus](r.helper,
          "SELECT * FROM system_status WHERE mid = {:mid}",
          dbx.Params{"mid": 1})
  }

## Testing Rules
- For every new Go function generated, automatically create a corresponding `_test.go` file.
- Use the `testing` and `testify/assert` packages.
- Always include `t.Run()` blocks for grouped tests.
- When generating code examples, run `go test -v` after completion to ensure no compile or test errors.
- Continue coding only if tests pass.