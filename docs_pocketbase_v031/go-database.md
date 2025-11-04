# Go - Database

## Overview
- Use `core.App` interface to interact with the database.
- `app.DB()` returns a `dbx.Builder` for running SQL statements.
- For `Record` and `Collection` models, see [Collection operations](/docs/go-collections) and [Record operations](/docs/go-records).

## Executing Queries

### `Execute()`
For non-retrieval queries (INSERT, UPDATE, DELETE).
```go
res, err := app.DB().NewQuery("DELETE FROM articles WHERE status = 'archived'").Execute()
```

### `One()`
Fetches a single row into a struct.
```go
import "github.com/pocketbase/pocketbase/tools/types"

type User struct {
    Id     string                `db:"id"`
    Status bool                  `db:"status"`
    Age    int                   `db:"age"`
    Roles  types.JSONArray[string] `db:"roles"`
}

user := User{}
err := app.DB().NewQuery("SELECT * FROM users WHERE id='1'").One(&user)
```

### `All()`
Fetches multiple rows into a slice of structs.
```go
users := []User{}
err := app.DB().NewQuery("SELECT * FROM users LIMIT 100").All(&users)
```

## Binding Parameters
Use named parameters `{:paramName}` to prevent SQL injection.
```go
import "github.com/pocketbase/dbx"

posts := []Post{}
err := app.DB().NewQuery("SELECT * FROM posts WHERE created >= {:from} AND created <= {:to}").
    Bind(dbx.Params{
        "from": "2023-06-25 00:00:00.000Z",
        "to":   "2023-06-28 23:59:59.999Z",
    }).All(&posts)
```

## Query Builder
Compose SQL statements programmatically.

### Basic Example
```go
users := []struct {
    Id    string `db:"id"`
    Email string `db:"email"`
}{}

app.DB().Select("id", "email").
    From("users").
    AndWhere(dbx.Like("email", "example.com")).
    Limit(100).
    OrderBy("created ASC").
    All(&users)
```

### Builder Methods
- `Select(...cols)`, `AndSelect(...cols)`, `Distinct(bool)`
- `From(...tables)`
- `Join(type, table, on)`, `InnerJoin(table, on)`, `LeftJoin(table, on)`, `RightJoin(table, on)`
- `Where(exp)`, `AndWhere(exp)`, `OrWhere(exp)`
- `OrderBy(...cols)`, `AndOrderBy(...cols)`
- `GroupBy(...cols)`, `AndGroupBy(...cols)`
- `Having(exp)`, `AndHaving(exp)`, `OrHaving(exp)`
- `Limit(int64)`
- `Offset(int64)`

### `dbx.Expression` Methods
- `dbx.NewExp(raw, optParams)`: Raw SQL fragment.
- `dbx.HashExp{k:v}`: Key-value equality conditions.
- `dbx.Not(exp)`: Negates an expression.
- `dbx.And(...exps)`: Joins expressions with `AND`.
- `dbx.Or(...exps)`: Joins expressions with `OR`.
- `dbx.In(col, ...values)`: `IN` condition.
- `dbx.NotIn(col, ...values)`: `NOT IN` condition.
- `dbx.Like(col, ...values)`: `LIKE` condition (AND).
- `dbx.NotLike(col, ...values)`: `NOT LIKE` condition (AND).
- `dbx.OrLike(col, ...values)`: `LIKE` condition (OR).
- `dbx.OrNotLike(col, ...values)`: `NOT LIKE` condition (OR).
- `dbx.Exists(exp)`: `EXISTS` subquery.
- `dbx.NotExists(exp)`: `NOT EXISTS` subquery.
- `dbx.Between(col, from, to)`: `BETWEEN` condition.
- `dbx.NotBetween(col, from, to)`: `NOT BETWEEN` condition.

## Transactions
Use `app.RunInTransaction(fn)` to execute queries in a transaction.

- Operations are persisted only if `fn` returns `nil`.
- It is safe to nest `RunInTransaction` calls.
- **Always use the `txApp` argument inside the transaction function.**
- Avoid slow tasks (e.g., sending emails) inside transactions.

```go
err := app.RunInTransaction(func(txApp core.App) error {
    // Update a record
    record, err := txApp.FindRecordById("articles", "RECORD_ID")
    if err != nil {
        return err
    }
    record.Set("status", "active")
    if err := txApp.Save(record); err != nil {
        return err
    }

    // Run a custom raw query
    rawQuery := "DELETE FROM articles WHERE status = 'pending'"
    if _, err := txApp.DB().NewQuery(rawQuery).Execute(); err != nil {
        return err
    }

    return nil
})
```