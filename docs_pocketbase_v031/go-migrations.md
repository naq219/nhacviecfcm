# PocketBase Go Migrations

This document explains how to manage database migrations in PocketBase using Go.

## Overview

PocketBase includes a built-in utility for database and data migrations. Migrations are written as Go functions, allowing you to version your database schema, create collections, and initialize data programmatically. Since they are Go files, they are embedded directly into your final application binary.

---

## Quick Setup

### 1. Register the Migrate Command

First, register the `migrate` command in your `main.go` file.

```go
package main

import (
    "log"
    "strings"
    "os"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/plugins/migratecmd"

    // Import your migrations package
    // _ "yourpackage/migrations"
)

func main() {
    app := pocketbase.New()

    isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

    migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
        // Enable auto-creation of migration files during development
        Automigrate: isGoRun,
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Create a New Migration

Use the `migrate create` command to generate a new migration file.

```bash
go run . migrate create "your_migration_name"
```

This creates a file in the `migrations` directory with `up` and `down` functions.

```go
// migrations/timestamp_your_migration_name.go
package migrations

import (
    "github.com/pocketbase/pocketbase/core"
    m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
    m.Register(func(app core.App) error {
        // "up" migration logic
        return nil
    }, func(app core.App) error {
        // "down" migration logic (optional)
        return nil
    })
}
```

### 3. Load and Run Migrations

- **Load:** Import your `migrations` package in `main.go` to make the application aware of them.
- **Run:** New migrations are automatically applied when the server starts. You can also run them manually:
  - `go run . migrate up` - Apply new migrations.
  - `go run . migrate down [number]` - Revert the last N migrations.

---

## Collections Snapshot

Generate a full snapshot of your current collections schema with the `migrate collections` command. This is useful for versioning your entire database structure.

```bash
go run . migrate collections
```

By default, this preserves collections and fields not in the snapshot. To delete missing items, you can modify the generated file.

---

## Migration History

Applied migrations are tracked in the `_migrations` table. If you manually edit or squash migration files during development, you can sync the history table:

```bash
go run . migrate history-sync
```

---

## Examples

### Raw SQL

```go
m.Register(func(app core.App) error {
    _, err := app.DB().NewQuery("UPDATE articles SET status = 'pending' WHERE status = ''").Execute()
    return err
}, nil)
```

### Create a Collection

```go
m.Register(func(app core.App) error {
    collection := &models.Collection{
        Name:       "clients",
        Type:       models.CollectionTypeAuth,
        ListRule:   types.Pointer("id = @request.auth.id"),
        ViewRule:   types.Pointer("id = @request.auth.id"),
        Schema:     schema.NewSchema(
            &schema.SchemaField{
                Name:     "company",
                Type:     schema.FieldTypeText,
                Required: true,
            },
        ),
    }
    return app.Dao().SaveCollection(collection)
}, func(app core.App) error {
    // Revert logic
    return nil
})
```