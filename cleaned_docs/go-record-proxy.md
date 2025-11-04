# PocketBase Go Record Proxy

This document explains how to create and use typed record proxies in PocketBase for cleaner, type-safe field access.

## Overview

While `core.Record` is the standard way to interact with data, you can create a typed struct that embeds `core.BaseRecordProxy` to define getters and setters for your collection fields. This approach implements the `core.RecordProxy` interface, allowing your custom struct to be used in `RecordQuery` results just like a regular record model.

Changes made through the proxy will trigger the corresponding record validations and hooks, ensuring data integrity and consistency.

---

## Implementing a Record Proxy

Below is an example of an `Article` record proxy with typed accessors for its fields.

### Go Example: `article.go`

```go
package main

import (
    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/tools/types"
)

// Ensure Article satisfies the core.RecordProxy interface.
var _ core.RecordProxy = (*Article)(nil)

type Article struct {
    core.BaseRecordProxy
}

func (a *Article) Title() string {
    return a.GetString("title")
}

func (a *Article) SetTitle(title string) {
    a.Set("title", title)
}

func (a *Article) Slug() string {
    return a.GetString("slug")
}

func (a *Article) SetSlug(slug string) {
    a.Set("slug", slug)
}

func (a *Article) Created() types.DateTime {
    return a.GetDateTime("created")
}

func (a *Article) Updated() types.DateTime {
    return a.GetDateTime("updated")
}
```

---

## Using the Record Proxy

Accessing and modifying proxy records is similar to working with regular records.

### Finding and Updating a Proxy Record

```go
import (
    "strings"
    "github.com/pocketbase/dbx"
    "github.com/pocketbase/pocketbase/core"
)

func FindArticleBySlug(app core.App, slug string) (*Article, error) {
    article := &Article{}
    err := app.RecordQuery("articles").
        AndWhere(dbx.NewExp("LOWER(slug)={:slug}", dbx.Params{"slug": strings.ToLower(slug)})).
        Limit(1).
        One(article)

    if err != nil {
        return nil, err
    }
    return article, nil
}

// Usage
article, err := FindArticleBySlug(app, "example")
if err != nil {
    return err
}

// Change the title
article.SetTitle("Lorem ipsum...")

// Persist the change, triggering validations and hooks
if err := app.Save(article); err != nil {
    return err
}
```

### Loading an Existing Record into a Proxy

If you have an existing `*core.Record`, you can load it into your proxy using `SetProxyRecord`.

```go
// Fetch a regular record
record, err := app.FindRecordById("articles", "RECORD_ID")
if err != nil {
    return err
}

// Load it into the proxy
article := &Article{}
article.SetProxyRecord(record)
```