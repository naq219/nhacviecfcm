                              // CSS3 animations support check (function () { if (typeof document === "undefined") { return; // serverside rendering } let elem = document.createElement("div"); let hasAnimation = elem.style.animationName !== undefined; if (!hasAnimation) { document.documentElement.classList.add("no-animations"); } })(); // silent svelte console errors if (typeof window === "undefined") { window = {}; } window.Prism = window.Prism || {}; window.Prism.manual = true; if (typeof navigator !== "undefined" && navigator.userAgentData) { navigator.userAgentData.getHighEntropyValues(\[ "architecture", \]).then((ua) => { window.UA\_ARCHITECTURE = ua.architecture }); }                        Extend with Go - Collection operations - Docs - PocketBase

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

Extend with Go - Collection operations

Collection operations

Collections are usually managed via the Dashboard interface, but there are some situations where you may want to create or edit a collection programmatically (usually as part of a [DB migration](/docs/go-migrations)). You can find all available Collection related operations and methods in [`core.App`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#App) and [`core.Collection`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#Collection) , but below are listed some of the most common ones:

### [Fetch collections](#fetch-collections)

##### [Fetch single collection](#fetch-single-collection)

All single collection retrieval methods return `nil` and `sql.ErrNoRows` error if no collection is found.

`collection, err := app.FindCollectionByNameOrId("example")`

##### [Fetch multiple collections](#fetch-multiple-collections)

All multiple collections retrieval methods return empty slice and `nil` error if no collections are found.

`allCollections, err := app.FindAllCollections() authAndViewCollections, err := app.FindAllCollections(core.CollectionTypeAuth, core.CollectionTypeView)`

##### [Custom collection query](#custom-collection-query)

In addition to the above query helpers, you can also create custom Collection queries using [`CollectionQuery()`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#CollectionQuery) method. It returns a SELECT DB builder that can be used with the same methods described in the [Database guide](/docs/go-database).

`import ( "github.com/pocketbase/dbx" "github.com/pocketbase/pocketbase/core" ) ... func FindSystemCollections(app core.App) ([]*core.Collection, error) { collections := []*core.Collection{} err := app.CollectionQuery(). AndWhere(dbx.HashExp{"system": true}). OrderBy("created DESC"). All(&collections) if err != nil { return nil, err } return collections, nil }`

### [Collection properties](#collection-properties)

`Id string Name string Type string // "base", "view", "auth" System bool // !prevent collection rename, deletion and rules change of internal collections like _superusers Fields core.FieldsList Indexes types.JSONArray[string] Created types.DateTime Updated types.DateTime // CRUD rules ListRule *string ViewRule *string CreateRule *string UpdateRule *string DeleteRule *string // "view" type specific options // (see https://github.com/pocketbase/pocketbase/blob/master/core/collection_model_view_options.go) ViewQuery string // "auth" type specific options // (see https://github.com/pocketbase/pocketbase/blob/master/core/collection_model_auth_options.go) AuthRule *string ManageRule *string AuthAlert core.AuthAlertConfig OAuth2 core.OAuth2Config PasswordAuth core.PasswordAuthConfig MFA core.MFAConfig OTP core.OTPConfig AuthToken core.TokenConfig PasswordResetToken core.TokenConfig EmailChangeToken core.TokenConfig VerificationToken core.TokenConfig FileToken core.TokenConfig VerificationTemplate core.EmailTemplate ResetPasswordTemplate core.EmailTemplate ConfirmEmailChangeTemplate core.EmailTemplate`

### [Field definitions](#field-definitions)

*   [`core.BoolField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#BoolField)
*   [`core.NumberField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#NumberField)
*   [`core.TextField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#TextField)
*   [`core.EmailField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#EmailField)
*   [`core.URLField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#URLField)
*   [`core.EditorField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#EditorField)
*   [`core.DateField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#DateField)
*   [`core.AutodateField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#AutodateField)
*   [`core.SelectField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#SelectField)
*   [`core.FileField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#FileField)
*   [`core.RelationField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#RelationField)
*   [`core.JSONField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#JSONField)
*   [`core.GeoPointField`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#GeoPointField)

### [Create new collection](#create-new-collection)

``import ( "github.com/pocketbase/pocketbase/core" "github.com/pocketbase/pocketbase/tools/types" ) ... // core.NewAuthCollection("example") // core.NewViewCollection("example") collection := core.NewBaseCollection("example") // set rules collection.ViewRule = types.Pointer("@request.auth.id != ''") collection.CreateRule = types.Pointer("@request.auth.id != '' && @request.body.user = @request.auth.id") collection.UpdateRule = types.Pointer(` @request.auth.id != '' && user = @request.auth.id && (@request.body.user:isset = false || @request.body.user = @request.auth.id) `) // add text field collection.Fields.Add(&core.TextField{ Name: "title", Required: true, Max: 100, }) // add relation field usersCollection, err := app.FindCollectionByNameOrId("users") if err != nil { return err } collection.Fields.Add(&core.RelationField{ Name: "user", Required: true, Max: 100, CascadeDelete: true, CollectionId: usersCollection.Id, }) // add autodate/timestamp fields (created/updated) collection.Fields.Add(&core.AutodateField{ Name: "created", OnCreate: true, }) collection.Fields.Add(&core.AutodateField{ Name: "updated", OnCreate: true, OnUpdate: true, }) // or: collection.Indexes = []string{"CREATE UNIQUE INDEX idx_example_user ON example (user)"} collection.AddIndex("idx_example_user", true, "user", "") // validate and persist // (use SaveNoValidate to skip fields validation) err = app.Save(collection) if err != nil { return err }``

### [Update existing collection](#update-existing-collection)

`import ( "github.com/pocketbase/pocketbase/core" "github.com/pocketbase/pocketbase/tools/types" ) ... collection, err := app.FindCollectionByNameOrId("example") if err != nil { return err } // change rule collection.DeleteRule = types.Pointer("@request.auth.id != ''") // add new editor field collection.Fields.Add(&core.EditorField{ Name: "description", Required: true, }) // change existing field // (returns a pointer and direct modifications are allowed without the need of reinsert) titleField := collection.Fields.GetByName("title").(*core.TextField) titleField.Min = 10 // or: collection.Indexes = append(collection.Indexes, "CREATE INDEX idx_example_title ON example (title)") collection.AddIndex("idx_example_title", false, "title", "") // validate and persist // (use SaveNoValidate to skip fields validation) err = app.Save(collection) if err != nil { return err }`

### [Delete collection](#delete-collection)

`collection, err := app.FindCollectionByNameOrId("example") if err != nil { return err } err = app.Delete(collection) if err != nil { return err }`

* * *

[Prev: Record operations](/docs/go-records) [Next: Migrations](/docs/go-migrations)

[FAQ](/faq) [Discussions](https://github.com/pocketbase/pocketbase/discussions) [Documentation](/docs)

[JavaScript SDK](https://github.com/pocketbase/js-sdk) [Dart SDK](https://github.com/pocketbase/dart-sdk)

Pocket**Base**

[](mailto:support)[](https://github.com/pocketbase/pocketbase)

© 2023-2025 Pocket**Base** The Gopher artwork is from [marcusolsson/gophers](https://github.com/marcusolsson/gophers)

Crafted by [**Gani**](https://gani.bg)

{ \_\_sveltekit\_jfyndd = { base: new URL("../..", location).pathname.slice(0, -1) }; const element = document.currentScript.parentElement; const data = \[null,null,null,null\]; Promise.all(\[ import("../../\_app/immutable/entry/start.HyAB6n-v.js"), import("../../\_app/immutable/entry/app.DExXzKO6.js") \]).then((\[kit, app\]) => { kit.start(app, element, { node\_ids: \[0, 2, 3, 20\], data, form: null, error: null }); }); }