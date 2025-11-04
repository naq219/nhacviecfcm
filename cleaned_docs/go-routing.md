# PocketBase Go Routing

PocketBase routing is built on top of the standard Go `net/http.ServeMux`. The router can be accessed via the `app.OnServe()` hook, allowing you to register custom endpoints and middlewares.

## Routes

### Registering New Routes

Every route has a path, a handler function, and optional middlewares.

**Example:**

```go
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    // Register "GET /hello/{name}" route (allowed for everyone)
    se.Router.GET("/hello/{name}", func(e *core.RequestEvent) error {
        name := e.Request.PathValue("name")
        return e.String(http.StatusOK, "Hello " + name)
    })

    // Register "POST /api/myapp/settings" route (allowed only for authenticated users)
    se.Router.POST("/api/myapp/settings", func(e *core.RequestEvent) error {
        // ...
        return e.JSON(http.StatusOK, map[string]bool{"success": true})
    }).Bind(apis.RequireAuth())

    return se.Next()
})
```

**Common Registration Methods:**

```go
se.Router.GET(path, action)
se.Router.POST(path, action)
se.Router.PUT(path, action)
se.Router.PATCH(path, action)
se.Router.DELETE(path, action)
se.Router.Any(pattern, action) // Handles any HTTP method
```

### Route Groups

You can group routes that share a common base path and middlewares.

**Example:**

```go
g := se.Router.Group("/api/myapp")
g.Bind(apis.RequireAuth()) // Group middleware

g.GET("", action1)
g.GET("/example/{id}", action2)
g.PATCH("/example/{id}", action3).BindFunc(/* custom route middleware */)

// Nested group
sub := g.Group("/sub")
sub.GET("/sub1", action4)
```

This registers the following endpoints, all requiring authentication:
- `GET /api/myapp`
- `GET /api/myapp/example/{id}`
- `PATCH /api/myapp/example/{id}`
- `GET /api/myapp/example/sub/sub1`

### Path Parameters and Matching

PocketBase follows the same pattern matching rules as `net/http.ServeMux`.

- **Path Parameters:** `{paramName}`
- **Wildcard Parameters:** `{paramName...}` (matches multiple path segments)
- **Trailing Slash:** A pattern ending in `/` matches any request starting with that path. To match the exact path with a trailing slash, use `{$}` at the end (e.g., `/static/{$}`).

**Note:** To avoid conflicts with system routes, prefix your API routes with a unique app name, like `/api/myapp/`.

### Handling Requests

Here are common operations within a route handler (`e` is `*core.RequestEvent`).

- **Reading Path Parameters:**
  ```go
  id := e.Request.PathValue("id")
  ```

- **Retrieving Auth State:**
  ```go
  authRecord := e.Auth
  isGuest := e.Auth == nil
  isSuperuser := e.HasSuperuserAuth()
  ```

- **Reading Query Parameters:**
  ```go
  search := e.Request.URL.Query().Get("search")
  ```

- **Reading Request Headers:**
  ```go
  token := e.Request.Header.Get("Some-Header")
  ```

- **Writing Response Headers:**
  ```go
  e.Response.Header().Set("Some-Header", "123")
  ```

- **Retrieving Uploaded Files:**
  ```go
  files, err := e.FindUploadedFiles("document")
  ```

- **Reading Request Body:**
  Use `e.BindBody()` to unmarshal the request body into a struct or map. Supported struct tags are `json`, `xml`, and `form`.

  ```go
  data := struct {
      Title       string `json:"title" form:"title"`
      Description string `json:"description" form:"description"`
  }{}
  if err := e.BindBody(&data); err != nil {
      return e.BadRequestError("Failed to read request data", err)
  }
  ```

- **Writing Response Body:**
  ```go
  e.JSON(http.StatusOK, data)
  e.String(http.StatusOK, "text")
  e.HTML(http.StatusOK, "<h1>html</h1>")
  e.Redirect(http.StatusTemporaryRedirect, "https://example.com")
  e.NoContent(http.StatusNoContent)
  ```

- **Reading Client IP:**
  ```go
  ip := e.RemoteIP() // Direct client IP
  realIP := e.RealIP() // Real client IP behind a proxy
  ```

- **Request Store:**
  Share data between middlewares and the route action for the duration of a single request.
  ```go
  e.Set("someKey", 123)
  val := e.Get("someKey").(int)
  ```

## Middlewares

Middlewares inspect, intercept, and filter route requests. They share the same signature as route actions and must call `e.Next()` to continue the execution chain.

### Registering Middlewares

Middlewares can be registered globally, on a group, or on a single route using `Bind` and `BindFunc`.

**Example (Global Middleware):**

```go
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    se.Router.BindFunc(func(e *core.RequestEvent) error {
        // ... middleware logic ...
        return e.Next()
    })
    return se.Next()
})
```

### Removing Middlewares

Use `Unbind(id)` to remove a middleware that has a non-empty `Id`.

```go
se.Router.GET("/B", action).Unbind("test")
```

### Built-in Middlewares

The `apis` package provides several useful middlewares:
- `apis.RequireGuestOnly()`
- `apis.RequireAuth(optCollectionNames...)`
- `apis.RequireSuperuserAuth()`
- `apis.RequireSuperuserOrOwnerAuth(ownerIdParam)`
- `apis.BodyLimit(limitBytes)`
- `apis.Gzip()`
- `apis.SkipSuccessActivityLog()`

### Default Global Middlewares

PocketBase registers several internal middlewares by default, including CORS, activity logging, panic recovery, auth token loading, and rate limiting. You can hook custom logic before or after these by adjusting middleware priority.

## Error Responses

Returned errors are converted to a generic `ApiError` to prevent leaking sensitive information. Use the provided helper methods to return formatted JSON error responses.

**Example:**

```go
return e.Error(500, "Something went wrong", validationData)
return e.BadRequestError("Invalid data", nil)
return e.NotFoundError("Resource not found", nil)
```

## Helpers

### Serving Static Directories

Use `apis.Static()` to serve content from a filesystem.

```go
se.Router.GET("/{path...}", apis.Static(os.DirFS("/path/to/public"), false))
```

### Auth Response

Use `apis.RecordAuthResponse()` to generate a standardized token and record data response.

```go
return apis.RecordAuthResponse(e, record, "phone", nil)
```

### Enriching Records

Use `apis.EnrichRecord()` and `apis.EnrichRecords()` to expand relations and manage email visibility.

```go
err := apis.EnrichRecords(e, records, "categories")
```

### Go `http.Handler` Wrappers

Use `apis.WrapStdHandler()` and `apis.WrapStdMiddleware()` to integrate standard Go handlers and middlewares.

## Sending Requests with SDKs

The official SDKs provide a `send()` method to communicate with your custom routes.

**JavaScript:**
```javascript
await pb.send("/hello", { query: { "abc": 123 } });
```

**Dart:**
```dart
await pb.send("/hello", query: { "abc": 123 });
```