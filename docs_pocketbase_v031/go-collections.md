# Go - Collection Operations

## Fetch Collections

### Fetch Single Collection
```go
collection, err := app.FindCollectionByNameOrId("example")
```

### Fetch Multiple Collections
```go
allCollections, err := app.FindAllCollections()
authAndViewCollections, err := app.FindAllCollections(core.CollectionTypeAuth, core.CollectionTypeView)
```

### Custom Collection Query
```go
import (
    "github.com/pocketbase/dbx"
    "github.com/pocketbase/pocketbase/core"
)

func FindSystemCollections(app core.App) ([]*core.Collection, error) {
    collections := []*core.Collection{}
    err := app.CollectionQuery().
        AndWhere(dbx.HashExp{"system": true}).
        OrderBy("created DESC").
        All(&collections)
    if err != nil {
        return nil, err
    }
    return collections, nil
}
```

## Collection Properties
```go
Id string
Name string
Type string // "base", "view", "auth"
System bool // !prevent collection rename, deletion and rules change of internal collections like _superusers
Fields core.FieldsList
Indexes types.JSONArray[string]
Created types.DateTime
Updated types.DateTime

// CRUD rules
ListRule *string
ViewRule *string
CreateRule *string
UpdateRule *string
DeleteRule *string

// "view" type specific options
ViewQuery string

// "auth" type specific options
AuthRule *string
ManageRule *string
AuthAlert core.AuthAlertConfig
OAuth2 core.OAuth2Config
PasswordAuth core.PasswordAuthConfig
MFA core.MFAConfig
OTP core.OTPConfig
AuthToken core.TokenConfig
PasswordResetToken core.TokenConfig
EmailChangeToken core.TokenConfig
VerificationToken core.TokenConfig
FileToken core.TokenConfig
VerificationTemplate core.EmailTemplate
ResetPasswordTemplate core.EmailTemplate
ConfirmEmailChangeTemplate core.EmailTemplate
```

## Field Definitions
- `core.BoolField`
- `core.NumberField`
- `core.TextField`
- `core.EmailField`
- `core.URLField`
- `core.EditorField`
- `core.DateField`
- `core.AutodateField`
- `core.SelectField`
- `core.FileField`
- `core.RelationField`
- `core.JSONField`
- `core.GeoPointField`

## Create New Collection
```go
import (
    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/tools/types"
)

collection := core.NewBaseCollection("example")
// OR: core.NewAuthCollection("example")
// OR: core.NewViewCollection("example")

// Set rules
collection.ViewRule = types.Pointer("@request.auth.id != ''")
collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id")
collection.UpdateRule = types.Pointer(`
    @request.auth.id != '' &&
    user = @request.auth.id &&
    (@request.body.user:isset = false || @request.body.user = @request.auth.id)
`)

// Add text field
collection.Fields.Add(&core.TextField{
    Name: "title",
    Required: true,
    Max: 100,
})

// Add relation field
usersCollection, err := app.FindCollectionByNameOrId("users")
if err != nil {
    return err
}
collection.Fields.Add(&core.RelationField{
    Name: "user",
    Required: true,
    Max: 100,
    CascadeDelete: true,
    CollectionId: usersCollection.Id,
})

// Add autodate/timestamp fields (created/updated)
collection.Fields.Add(&core.AutodateField{
    Name: "created",
    OnCreate: true,
})
collection.Fields.Add(&core.AutodateField{
    Name: "updated",
    OnCreate: true,
    OnUpdate: true,
})

// Add index
collection.AddIndex("idx_example_user", true, "user", "")

// Validate and persist (use SaveNoValidate to skip fields validation)
err = app.Save(collection)
if err != nil {
    return err
}
```

## Update Existing Collection
```go
import (
    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/tools/types"
)

collection, err := app.FindCollectionByNameOrId("example")
if err != nil {
    return err
}

// Change rule
collection.DeleteRule = types.Pointer("@request.auth.id != ''")

// Add new editor field
collection.Fields.Add(&core.EditorField{
    Name: "description",
    Required: true,
})

// Change existing field (returns a pointer and direct modifications are allowed without the need of reinsert)
titleField := collection.Fields.GetByName("title").(*core.TextField)
titleField.Min = 10

// Add index
collection.AddIndex("idx_example_title", false, "title", "")

// Validate and persist (use SaveNoValidate to skip fields validation)
err = app.Save(collection)
if err != nil {
    return err
}
```

## Delete Collection
```go
collection, err := app.FindCollectionByNameOrId("example")
if err != nil {
    return err
}
err = app.Delete(collection)
if err != nil {
    return err
}
```