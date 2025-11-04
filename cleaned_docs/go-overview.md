# PocketBase Go Overview

This document provides an overview of how to extend PocketBase with Go.

## Getting Started

PocketBase can be used as a Go package, allowing you to build a custom, portable application with your own business logic.

### Minimal Example

1.  **Install Go 1.23+**.
2.  Create a new project with a `main.go` file.

    ```go
    package main

    import (
        "log"
        "os"

        "github.com/pocketbase/pocketbase"
        "github.com/pocketbase/pocketbase/apis"
        "github.com/pocketbase/pocketbase/core"
    )

    func main() {
        app := pocketbase.New()

        app.OnServe().Bind(func(e *core.ServeEvent) error {
            // Serves static files from the provided public dir (if exists)
            e.Router.GET("/*", apis.Static(os.DirFS("./pb_public")))
            return nil
        })

        if err := app.Start(); err != nil {
            log.Fatal(err)
        }
    }
    ```

3.  Initialize dependencies:

    ```bash
    go mod init myapp && go mod tidy
    ```

4.  Run the application:

    ```bash
    go run . serve
    ```

5.  Build a statically linked executable:

    ```bash
    go build
    ./myapp serve
    ```

---

## Custom SQLite Driver

While the built-in pure Go SQLite driver (`modernc.org/sqlite`) is recommended, you can use a custom SQLite driver for advanced features like ICU or FTS5. This may require CGO.

To use a custom driver, define a `DBConnect` function in your app configuration. This function will be called for both the main database (`pb_data/data.db`) and the auxiliary database (`pb_data/auxiliary.db`).

### Example with `mattn/go-sqlite3`

```go
package main

import (
    "database/sql"
    "log"

    "github.com/mattn/go-sqlite3"
    "github.com/pocketbase/dbx"
    "github.com/pocketbase/pocketbase"
)

func init() {
    sql.Register("pb_sqlite3", &sqlite3.SQLiteDriver{
        ConnectHook: func(conn *sqlite3.SQLiteConn) error {
            _, err := conn.Exec(`
                PRAGMA busy_timeout = 10000;
                PRAGMA journal_mode = WAL;
                PRAGMA synchronous = NORMAL;
                PRAGMA foreign_keys = ON;
            `, nil)
            return err
        },
    })
    dbx.BuilderFuncMap["pb_sqlite3"] = dbx.BuilderFuncMap["sqlite3"]
}

func main() {
    app := pocketbase.NewWithConfig(pocketbase.Config{
        DBConnect: func(dbPath string) (*dbx.DB, error) {
            return dbx.Open("pb_sqlite3", dbPath)
        },
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### Example with `ncruces/go-sqlite3`

```go
package main

import (
    "log"

    "github.com/pocketbase/dbx"
    "github.com/pocketbase/pocketbase"
    _ "github.com/ncruces/go-sqlite3/driver"
    _ "github.com/ncruces/go-sqlite3/embed"
)

func main() {
    app := pocketbase.NewWithConfig(pocketbase.Config{
        DBConnect: func(dbPath string) (*dbx.DB, error) {
            const pragmas = "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)"
            return dbx.Open("sqlite3", "file:"+dbPath+pragmas)
        },
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```