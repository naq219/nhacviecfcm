# PocketBase Go Record Operations

This document provides a guide to performing record operations in PocketBase using Go, covering everything from creating and fetching records to handling transactions and authentication.

## Record Model

The primary way to interact with collection records is through the `core.Record` model. You can find detailed documentation in the [`core.Record`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#Record) package.

### Setting Field Values

Use `record.Set()` to assign a value to a field. You can also use modifiers for specific field types.

```go
// Set a single field value
record.Set("title", "example")

// Append to an existing value (e.g., for relations)
record.Set("users+", "USER_ID")

// Populate a record from a map
record.Load(data)
```

### Getting Field Values

Retrieve field values with type casting for safe access.

```go
// Get a value without casting
record.Get("someField")

// Get a value with type casting
record.GetBool("someField")
record.GetString("someField")
record.GetInt("someField")
record.GetFloat("someField")
record.GetDateTime("someField")

// Get expanded relation records
record.ExpandedOne("author")
record.ExpandedAll("categories")
```

---

## Fetching Records

PocketBase provides several helper methods to fetch records.

### Fetching a Single Record

These methods return `sql.ErrNoRows` if no record is found.

```go
// Find a record by its ID
record, err := app.FindRecordById("articles", "RECORD_ID")

// Find a record by a key-value pair
record, err := app.FindFirstRecordByData("articles", "slug", "test")

// Find a record using a filter expression
record, err := app.FindFirstRecordByFilter(
    "articles",
    "status = 'public' && category = {:category}",
    dbx.Params{"category": "news"},
)
```

### Fetching Multiple Records

These methods return an empty slice if no records are found.

```go
// Find records by their IDs
records, err := app.FindRecordsByIds("articles", []string{"ID1", "ID2"})

// Find all records matching a filter
records, err := app.FindAllRecords("articles", dbx.HashExp{"status": "pending"})

// Find records with pagination and sorting
records, err := app.FindRecordsByFilter(
    "articles",
    "status = 'public'",
    "-published", // Sort descending
    10, // Limit
    0,  // Offset
    nil,
)
```

### Custom Record Queries

For more complex queries, use `app.RecordQuery()` to create a custom query builder.

```go
func FindActiveArticles(app core.App) ([]*core.Record, error) {
    records := []*core.Record{}
    err := app.RecordQuery("articles").
        AndWhere(dbx.HashExp{"status": "active"}).
        OrderBy("published DESC").
        Limit(10).
        All(&records)
    return records, err
}
```

---

## Creating and Updating Records

### Creating a New Record

Create a new record instance and use `app.Save()` to persist it.

```go
collection, _ := app.FindCollectionByNameOrId("articles")
record := core.NewRecord(collection)
record.Set("title", "Lorem ipsum")

// To upload files, use filesystem.File instances
f, _ := filesystem.NewFileFromPath("/path/to/file.txt")
record.Set("document", f)

if err := app.Save(record); err != nil {
    // Handle error
}
```

### Updating an Existing Record

Fetch a record, modify its fields, and use `app.Save()` to apply the changes.

```go
record, _ := app.FindRecordById("articles", "RECORD_ID")
record.Set("title", "New Title")

// To remove a file, use the '-' suffix
record.Set("document-", "old_file.txt")

if err := app.Save(record); err != nil {
    // Handle error
}
```

---

## Transactions

To execute multiple database operations in a single transaction, use `app.RunInTransaction()`.

```go
app.RunInTransaction(func(txApp core.App) error {
    // Always use txApp inside the transaction
    record1 := core.NewRecord(collection)
    record1.Set("title", "First")
    if err := txApp.Save(record1); err != nil {
        return err // Rollback
    }

    record2 := core.NewRecord(collection)
    record2.Set("title", "Second")
    if err := txApp.Save(record2); err != nil {
        return err // Rollback
    }

    return nil // Commit
})
```

---

## Authentication and Security

### Generating and Validating Tokens

PocketBase uses JWTs for stateless authentication. You can generate different types of tokens for a record.

```go
// Generate an auth token
token, err := record.NewAuthToken()

// Generate a password reset token
token, err := record.NewPasswordResetToken()

// Validate a token
authRecord, err := app.FindAuthRecordByToken("YOUR_TOKEN", core.TokenTypeAuth)
```

### Checking Access Permissions

Use `app.CanAccessRecord()` to check if a request has permission to access a record based on its collection's API rules.

```go
canAccess, err := e.App.CanAccessRecord(record, requestInfo, record.Collection().ViewRule)
if !canAccess {
    return e.ForbiddenError("", err)
}
```