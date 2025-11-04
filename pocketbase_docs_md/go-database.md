                              // CSS3 animations support check (function () { if (typeof document === "undefined") { return; // serverside rendering } let elem = document.createElement("div"); let hasAnimation = elem.style.animationName !== undefined; if (!hasAnimation) { document.documentElement.classList.add("no-animations"); } })(); // silent svelte console errors if (typeof window === "undefined") { window = {}; } window.Prism = window.Prism || {}; window.Prism.manual = true; if (typeof navigator !== "undefined" && navigator.userAgentData) { navigator.userAgentData.getHighEntropyValues(\[ "architecture", \]).then((ua) => { window.UA\_ARCHITECTURE = ua.architecture }); }                        Extend with Go - Database - Docs - PocketBase

 [![PocketBase logo](/images/logo.svg) Pocket**Base** v0.31.0](/)

Go JavaScript

[FAQ](/faq) [](https://github.com/pocketbase/pocketbase "GitHub Repo")[Documentation](/docs)

[Introduction](/docs) [Going to production](/docs/going-to-production) [Web APIs reference](/docs/api-records)

[

Extend with  
**Go**

](/docs/go-overview)[

Extend with  
**JavaScript**

](/docs/js-overview)

[Go Overview](/docs/go-overview) [Go Event hooks](/docs/go-event-hooks) [Go Routing](/docs/go-routing) [Go Database](/docs/go-database) [Go Record operations](/docs/go-records) [Go Collection operations](/docs/go-collections) [Go Migrations](/docs/go-migrations) [Go Jobs scheduling](/docs/go-jobs-scheduling) [Go Sending emails](/docs/go-sending-emails) [Go Rendering templates](/docs/go-rendering-templates) [Go Console commands](/docs/go-console-commands) [Go Realtime messaging](/docs/go-realtime) [Go Filesystem](/docs/go-filesystem) [Go Logging](/docs/go-logging) [Go Testing](/docs/go-testing) [Go Miscellaneous](/docs/go-miscellaneous) [Go Record proxy](/docs/go-record-proxy)

[JS Overview](/docs/js-overview) [JS Event hooks](/docs/js-event-hooks) [JS Routing](/docs/js-routing) [JS Database](/docs/js-database) [JS Record operations](/docs/js-records) [JS Collection operations](/docs/js-collections) [JS Migrations](/docs/js-migrations) [JS Jobs scheduling](/docs/js-jobs-scheduling) [JS Sending emails](/docs/js-sending-emails) [JS Rendering templates](/docs/js-rendering-templates) [JS Console commands](/docs/js-console-commands) [JS Sending HTTP requests](/docs/js-sending-http-requests) [JS Realtime messaging](/docs/js-realtime) [JS Filesystem](/docs/js-filesystem) [JS Logging](/docs/js-logging) [JS Types reference](/jsvm/index.html)

Extend with Go - Database

Database

[`core.App`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#App) is the main interface to interact with the database.

`App.DB()` returns a `dbx.Builder` that can run all kinds of SQL statements, including raw queries.

Most of the common DB operations are listed below, but you can find further information in the [dbx package godoc](https://pkg.go.dev/github.com/pocketbase/dbx) .

For more details and examples how to interact with Record and Collection models programmatically you could also check [Collection operations](/docs/go-collections) and [Record operations](/docs/go-records) sections.

### [Executing queries](#executing-queries)

To execute DB queries you can start with the `NewQuery("...")` statement and then call one of:

*   `[Execute()](#execute)` \- for any query statement that is not meant to retrieve data:
    
    `res, err := app.DB(). NewQuery("DELETE FROM articles WHERE status = 'archived'"). Execute()`
    
*   `[One()](#execute-one)` \- to populate a single row into a struct:
    
    ``type User struct { Id string `db:"id" json:"id"` Status bool `db:"status" json:"status"` Age int `db:"age" json:"age"` Roles types.JSONArray[string] `db:"roles" json:"roles"` } user := User{} err := app.DB(). NewQuery("SELECT id, status, age, roles FROM users WHERE id=1"). One(&user)``
    
*   `[All()](#execute-all)` \- to populate multiple rows into a slice of structs:
    
    ``type User struct { Id string `db:"id" json:"id"` Status bool `db:"status" json:"status"` Age int `db:"age" json:"age"` Roles types.JSONArray[string] `db:"roles" json:"roles"` } users := []User{} err := app.DB(). NewQuery("SELECT id, status, age, roles FROM users LIMIT 100"). All(&users)``
    

### [Binding parameters](#binding-parameters)

To prevent SQL injection attacks, you should use named parameters for any expression value that comes from user input. This could be done using the named `{:paramName}` placeholders in your SQL statement and then define the parameter values for the query with `Bind(params)`. For example:

``type Post struct { Name string `db:"name" json:"name"` Created types.DateTime `db:"created" json:"created"` } posts := []Post{} err := app.DB(). NewQuery("SELECT name, created FROM posts WHERE created >= {:from} and created <= {:to}"). Bind(dbx.Params{ "from": "2023-06-25 00:00:00.000Z", "to": "2023-06-28 23:59:59.999Z", }). All(&posts)``

### [Query builder](#query-builder)

Instead of writing plain SQLs, you can also compose SQL statements programmatically using the db query builder.  
Every SQL keyword has a corresponding query building method. For example, `SELECT` corresponds to `Select()`, `FROM` corresponds to `From()`, `WHERE` corresponds to `Where()`, and so on.

``users := []struct { Id string `db:"id" json:"id"` Email string `db:"email" json:"email"` }{} app.DB(). Select("id", "email"). From("users"). AndWhere(dbx.Like("email", "example.com")). Limit(100). OrderBy("created ASC"). All(&users)``

##### [Select(), AndSelect(), Distinct()](#select-andselect-distinct)

The `Select(...cols)` method initializes a `SELECT` query builder. It accepts a list of the column names to be selected.  
To add additional columns to an existing select query, you can call `AndSelect()`.  
To select distinct rows, you can call `Distinct(true)`.

`app.DB(). Select("id", "avatar as image"). AndSelect("(firstName || ' ' || lastName) as fullName"). Distinct(true) ...`

##### [From()](#from)

The `From(...tables)` method specifies which tables to select from (plain table names are automatically quoted).

`app.DB(). Select("table1.id", "table2.name"). From("table1", "table2") ...`

##### [Join()](#join)

The `Join(type, table, on)` method specifies a `JOIN` clause. It takes 3 parameters:

*   `type` - join type string like `INNER JOIN`, `LEFT JOIN`, etc.
*   `table` - the name of the table to be joined
*   `on` - optional `dbx.Expression` as an `ON` clause

For convenience, you can also use the shortcuts `InnerJoin(table, on)`, `LeftJoin(table, on)`, `RightJoin(table, on)` to specify `INNER JOIN`, `LEFT JOIN` and `RIGHT JOIN`, respectively.

`app.DB(). Select("users.*"). From("users"). InnerJoin("profiles", dbx.NewExp("profiles.user_id = users.id")). Join("FULL OUTER JOIN", "department", dbx.NewExp("department.id = {:id}", dbx.Params{ "id": "someId" })) ...`

##### [Where(), AndWhere(), OrWhere()](#where-andwhere-orwhere)

The `Where(exp)` method specifies the `WHERE` condition of the query.  
You can also use `AndWhere(exp)` or `OrWhere(exp)` to append additional one or more conditions to an existing `WHERE` clause.  
Each where condition accepts a single `dbx.Expression` (see below for full list).

`/* SELECT users.* FROM users WHERE id = "someId" AND status = "public" AND name like "%john%" OR ( role = "manager" AND fullTime IS TRUE AND experience > 10 ) */ app.DB(). Select("users.*"). From("users"). Where(dbx.NewExp("id = {:id}", dbx.Params{ "id": "someId" })). AndWhere(dbx.HashExp{"status": "public"}). AndWhere(dbx.Like("name", "john")). OrWhere(dbx.And( dbx.HashExp{ "role": "manager", "fullTime": true, }, dbx.NewExp("experience > {:exp}", dbx.Params{ "exp": 10 }) )) ...`

The following `dbx.Expression` methods are available:

*   `[dbx.NewExp(raw, optParams)](#dbx-newexpraw-optparams)`  
    Generates an expression with the specified raw query fragment. Use the `optParams` to bind `dbx.Params` to the expression.
    
    `dbx.NewExp("status = 'public'") dbx.NewExp("total > {:min} AND total < {:max}", dbx.Params{ "min": 10, "max": 30 })`
    
*   `[dbx.HashExp{k:v}](#dbx-hashexpkv)`  
    Generates a hash expression from a map whose keys are DB column names which need to be filtered according to the corresponding values.
    
    `// slug = "example" AND active IS TRUE AND tags in ("tag1", "tag2", "tag3") AND parent IS NULL dbx.HashExp{ "slug": "example", "active": true, "tags": []any{"tag1", "tag2", "tag3"}, "parent": nil, }`
    
*   `[dbx.Not(exp)](#dbx-notexp)`  
    Negates a single expression by wrapping it with `NOT()`.
    
    `// NOT(status = 1) dbx.Not(dbx.NewExp("status = 1"))`
    
*   `[dbx.And(...exps)](#dbx-and-exps)`  
    Creates a new expression by concatenating the specified ones with `AND`.
    
    `// (status = 1 AND username like "%john%") dbx.And( dbx.NewExp("status = 1"), dbx.Like("username", "john"), )`
    
*   `[dbx.Or(...exps)](#dbx-or-exps)`  
    Creates a new expression by concatenating the specified ones with `OR`.
    
    `// (status = 1 OR username like "%john%") dbx.Or( dbx.NewExp("status = 1"), dbx.Like("username", "john") )`
    
*   `[dbx.In(col, ...values)](#dbx-incol-values)`  
    Generates an `IN` expression for the specified column and the list of allowed values.
    
    `// status IN ("public", "reviewed") dbx.In("status", "public", "reviewed")`
    
*   `[dbx.NotIn(col, ...values)](#dbx-notincol-values)`  
    Generates an `NOT IN` expression for the specified column and the list of allowed values.
    
    `// status NOT IN ("public", "reviewed") dbx.NotIn("status", "public", "reviewed")`
    
*   `[dbx.Like(col, ...values)](#dbx-likecol-values)`  
    Generates a `LIKE` expression for the specified column and the possible strings that the column should be like. If multiple values are present, the column should be like **all** of them.  
    By default, each value will be surrounded by _"%"_ to enable partial matching. Special characters like _"%"_, _"\\"_, _"\_"_ will also be properly escaped. You may call `Escape(...pairs)` and/or `Match(left, right)` to change the default behavior.
    
    `// name LIKE "%test1%" AND name LIKE "%test2%" dbx.Like("name", "test1", "test2") // name LIKE "test1%" dbx.Like("name", "test1").Match(false, true)`
    
*   `[dbx.NotLike(col, ...values)](#dbx-notlikecol-values)`  
    Generates a `NOT LIKE` expression in similar manner as `Like()`.
    
    `// name NOT LIKE "%test1%" AND name NOT LIKE "%test2%" dbx.NotLike("name", "test1", "test2") // name NOT LIKE "test1%" dbx.NotLike("name", "test1").Match(false, true)`
    
*   `[dbx.OrLike(col, ...values)](#dbx-orlikecol-values)`  
    This is similar to `Like()` except that the column must be one of the provided values, aka. multiple values are concatenated with `OR` instead of `AND`.
    
    `// name LIKE "%test1%" OR name LIKE "%test2%" dbx.OrLike("name", "test1", "test2") // name LIKE "test1%" OR name LIKE "test2%" dbx.OrLike("name", "test1", "test2").Match(false, true)`
    
*   `[dbx.OrNotLike(col, ...values)](#dbx-ornotlikecol-values)`  
    This is similar to `NotLike()` except that the column must not be one of the provided values, aka. multiple values are concatenated with `OR` instead of `AND`.
    
    `// name NOT LIKE "%test1%" OR name NOT LIKE "%test2%" dbx.OrNotLike("name", "test1", "test2") // name NOT LIKE "test1%" OR name NOT LIKE "test2%" dbx.OrNotLike("name", "test1", "test2").Match(false, true)`
    
*   `[dbx.Exists(exp)](#dbx-existsexp)`  
    Prefix with `EXISTS` the specified expression (usually a subquery).
    
    `// EXISTS (SELECT 1 FROM users WHERE status = 'active') dbx.Exists(dbx.NewExp("SELECT 1 FROM users WHERE status = 'active'"))`
    
*   `[dbx.NotExists(exp)](#dbx-notexistsexp)`  
    Prefix with `NOT EXISTS` the specified expression (usually a subquery).
    
    `// NOT EXISTS (SELECT 1 FROM users WHERE status = 'active') dbx.NotExists(dbx.NewExp("SELECT 1 FROM users WHERE status = 'active'"))`
    
*   `[dbx.Between(col, from, to)](#dbx-betweencol-from-to)`  
    Generates a `BETWEEN` expression with the specified range.
    
    `// age BETWEEN 3 and 99 dbx.Between("age", 3, 99)`
    
*   `[dbx.NotBetween(col, from, to)](#dbx-notbetweencol-from-to)`  
    Generates a `NOT BETWEEN` expression with the specified range.
    
    `// age NOT BETWEEN 3 and 99 dbx.NotBetween("age", 3, 99)`
    

##### [OrderBy(), AndOrderBy()](#orderby-andorderby)

The `OrderBy(...cols)` specifies the `ORDER BY` clause of the query.  
A column name can contain _"ASC"_ or _"DESC"_ to indicate its ordering direction.  
You can also use `AndOrderBy(...cols)` to append additional columns to an existing `ORDER BY` clause.

`app.DB(). Select("users.*"). From("users"). OrderBy("created ASC", "updated DESC"). AndOrderBy("title ASC") ...`

##### [GroupBy(), AndGroupBy()](#groupby-andgroupby)

The `GroupBy(...cols)` specifies the `GROUP BY` clause of the query.  
You can also use `AndGroupBy(...cols)` to append additional columns to an existing `GROUP BY` clause.

`app.DB(). Select("users.*"). From("users"). GroupBy("department", "level") ...`

##### [Having(), AndHaving(), OrHaving()](#having-andhaving-orhaving)

The `Having(exp)` specifies the `HAVING` clause of the query.  
Similarly to `Where(exp)`, it accept a single `dbx.Expression` (see all available expressions listed above).  
You can also use `AndHaving(exp)` or `OrHaving(exp)` to append additional one or more conditions to an existing `HAVING` clause.

`app.DB(). Select("users.*"). From("users"). GroupBy("department", "level"). Having(dbx.NewExp("sum(level) > {:sum}", dbx.Params{ sum: 10 })) ...`

##### [Limit()](#limit)

The `Limit(number)` method specifies the `LIMIT` clause of the query.

`app.DB(). Select("users.*"). From("users"). Limit(30) ...`

##### [Offset()](#offset)

The `Offset(number)` method specifies the `OFFSET` clause of the query. Usually used together with `Limit(number)`.

`app.DB(). Select("users.*"). From("users"). Offset(5). Limit(30) ...`

### [Transaction](#transaction)

To execute multiple queries in a transaction you can use [`app.RunInTransaction(fn)`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#BaseApp.RunInTransaction) .

The DB operations are persisted only if the transaction returns `nil`.

It is safe to nest `RunInTransaction` calls as long as you use the callback's `txApp` argument.

Inside the transaction function always use its `txApp` argument and not the original `app` instance because we allow only a single writer/transaction at a time and it could result in a deadlock.

To avoid performance issues, try to minimize slow/long running tasks such as sending emails, connecting to external services, etc. as part of the transaction.

`err := app.RunInTransaction(func(txApp core.App) error { // update a record record, err := txApp.FindRecordById("articles", "RECORD_ID") if err != nil { return err } record.Set("status", "active") if err := txApp.Save(record); err != nil { return err } // run a custom raw query (doesn't fire event hooks) rawQuery := "DELETE FROM articles WHERE status = 'pending'" if _, err := txApp.DB().NewQuery(rawQuery).Execute(); err != nil { return err } return nil })`

* * *

[Prev: Routing](/docs/go-routing) [Next: Record operations](/docs/go-records)

[FAQ](/faq) [Discussions](https://github.com/pocketbase/pocketbase/discussions) [Documentation](/docs)

[JavaScript SDK](https://github.com/pocketbase/js-sdk) [Dart SDK](https://github.com/pocketbase/dart-sdk)

Pocket**Base**

[](mailto:support)[](https://github.com/pocketbase/pocketbase)

© 2023-2025 Pocket**Base** The Gopher artwork is from [marcusolsson/gophers](https://github.com/marcusolsson/gophers)

Crafted by [**Gani**](https://gani.bg)

{ \_\_sveltekit\_jfyndd = { base: new URL("../..", location).pathname.slice(0, -1) }; const element = document.currentScript.parentElement; const data = \[null,null,null,null\]; Promise.all(\[ import("../../\_app/immutable/entry/start.HyAB6n-v.js"), import("../../\_app/immutable/entry/app.DExXzKO6.js") \]).then((\[kit, app\]) => { kit.start(app, element, { node\_ids: \[0, 2, 3, 22\], data, form: null, error: null }); }); }