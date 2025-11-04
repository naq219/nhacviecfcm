# PocketBase Go Template Rendering

This document explains how to render HTML templates in PocketBase using Go, providing a safe and efficient way to generate dynamic content.

## Overview

While you can use any Go template engine, PocketBase provides a thin wrapper around the standard `html/template` package in the [`github.com/pocketbase/pocketbase/tools/template`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/template) utility. This wrapper makes it easier to load template files concurrently and on the fly.

### Key Features

- **Contextual Auto-Escaping**: Templates automatically escape HTML, JS, CSS, and URI content for injection safety.
- **Safe Raw Content**: Use the built-in `raw` function to render trusted content without escaping.
- **Concurrent Loading**: The registry is safe for use by multiple goroutines.

---

## Basic Template Usage

### Loading and Rendering Templates

```go
import "github.com/pocketbase/pocketbase/tools/template"

data := map[string]any{"name": "John"}
html, err := template.NewRegistry().LoadFiles(
    "views/base.html",
    "views/partial1.html",
    "views/partial2.html",
).Render(data)
```

---

## Template Composition

Templates are composed using a base template that defines placeholders, which are then filled by partials.

### Base Template (`layout.html`)

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <title>{{block "title" .}}Default app title{{end}}</title>
</head>
<body>
    Header...
    {{block "body" .}}
        Default app body...
    {{end}}
    Footer...
</body>
</html>
```

### Partial Template (`hello.html`)

```html
{{define "title"}}
    Page 1
{{end}}

{{define "body"}}
    <p>Hello from {{.name}}</p>
{{end}}
```

---

## Complete Example

### Directory Structure

```
myapp/
├── views/
│   ├── layout.html
│   └── hello.html
└── main.go
```

### `main.go`

```go
package main

import (
    "log"
    "net/http"
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/tools/template"
)

func main() {
    app := pocketbase.New()

    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        // Safe for concurrent use
        registry := template.NewRegistry()

        se.Router.GET("/hello/{name}", func(e *core.RequestEvent) error {
            name := e.Request.PathValue("name")

            html, err := registry.LoadFiles(
                "views/layout.html",
                "views/hello.html",
            ).Render(map[string]any{
                "name": name,
            })

            if err != nil {
                return e.NotFoundError("", err)
            }

            return e.HTML(http.StatusOK, html)
        })

        return se.Next()
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

---

## Additional Notes

- The dot (`.`) in templates represents the data passed to `Render(data)`.
- For more on template syntax, refer to the official Go documentation for [`html/template`](https://pkg.go.dev/html/template) and [`text/template`](https://pkg.go.dev/text/template).
- A helpful external resource is HashiCorp's [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax).