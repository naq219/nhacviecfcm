# PocketBase Go Logging

This document explains how to use the built-in logging functionality in PocketBase with Go.

## Overview

PocketBase provides a standard `slog.Logger` implementation available via `app.Logger()`. This logger writes logs to the database, which can be viewed in the PocketBase Dashboard under _Logs_.

To optimize performance, logs are batched and written with a debounce mechanism:
- 3 seconds after the last log write.
- When the batch size reaches 200 logs.
- Before the application terminates.

---

## Log Methods

The logger supports all standard `slog.Logger` methods.

### Basic Logging

- **Debug:** `app.Logger().Debug("Debug message", "key", "value")`
- **Info:** `app.Logger().Info("Info message", "key", "value")`
- **Warn:** `app.Logger().Warn("Warning message", "key", "value")`
- **Error:** `app.Logger().Error("Error message", "error", err)`

### Structured Logging

- **With:** `With(attrs...)` creates a new logger instance that includes the specified attributes in all subsequent logs.

  ```go
  l := app.Logger().With("total", 123)
  l.Info("message A") // Log data: {"total": 123}
  l.Info("message B", "name", "john") // Log data: {"total": 123, "name": "john"}
  ```

- **WithGroup:** `WithGroup(name)` groups all following log attributes under a specified name.

  ```go
  l := app.Logger().WithGroup("sub")
  l.Info("message A", "total", 123) // Log data: {"sub": {"total": 123}}
  ```

---

## Log Settings

Log settings, such as retention period, minimum log level, and IP logging, can be configured from the PocketBase Dashboard.

---

## Custom Log Queries

You can programmatically query stored logs using `app.LogQuery()`.

```go
logs := []*core.Log{}
err := app.LogQuery().
    AndWhere(dbx.In("level", -4, 0)). // Target debug and info logs
    AndWhere(dbx.NewExp("json_extract(data, '$.type') = 'request'")).
    OrderBy("created DESC").
    Limit(100).
    All(&logs)
```

---

## Intercepting Logs

You can intercept logs before they are written to the database by using the `OnModelCreate` hook for the `_logs` table.

```go
app.OnModelCreate(core.LogsTableName).BindFunc(func(e *core.ModelEvent) error {
    log := e.Model.(*core.Log)
    // Custom logic to process or forward the log
    fmt.Println(log.Message)
    return e.Next()
})
```