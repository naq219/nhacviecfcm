                              // CSS3 animations support check (function () { if (typeof document === "undefined") { return; // serverside rendering } let elem = document.createElement("div"); let hasAnimation = elem.style.animationName !== undefined; if (!hasAnimation) { document.documentElement.classList.add("no-animations"); } })(); // silent svelte console errors if (typeof window === "undefined") { window = {}; } window.Prism = window.Prism || {}; window.Prism.manual = true; if (typeof navigator !== "undefined" && navigator.userAgentData) { navigator.userAgentData.getHighEntropyValues(\[ "architecture", \]).then((ua) => { window.UA\_ARCHITECTURE = ua.architecture }); }                     Extend with Go - Record proxy - Docs - PocketBase

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

Extend with Go - Record proxy

Record proxy

The available [`core.Record` and its helpers](/docs/go-records) are usually the recommended way to interact with your data, but in case you want a typed access to your record fields you can create a helper struct that embeds [`core.BaseRecordProxy`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#BaseRecordProxy) _(which implements the `core.RecordProxy` interface)_ and define your collection fields as getters and setters.

By implementing the `core.RecordProxy` interface you can use your custom struct as part of a `RecordQuery` result like a regular record model. In addition, every DB change through the proxy struct will trigger the corresponding record validations and hooks. This ensures that other parts of your app, including 3rd party plugins, that don't know or use your custom struct will still work as expected.

Below is a sample `Article` record proxy implementation:

`// article.go package main import ( "github.com/pocketbase/pocketbase/core" "github.com/pocketbase/pocketbase/tools/types" ) // ensures that the Article struct satisfy the core.RecordProxy interface var _ core.RecordProxy = (*Article)(nil) type Article struct { core.BaseRecordProxy } func (a *Article) Title() string { return a.GetString("title") } func (a *Article) SetTitle(title string) { a.Set("title", title) } func (a *Article) Slug() string { return a.GetString("slug") } func (a *Article) SetSlug(slug string) { a.Set("slug", slug) } func (a *Article) Created() types.DateTime { return a.GetDateTime("created") } func (a *Article) Updated() types.DateTime { return a.GetDateTime("updated") }`

Accessing and modifying the proxy records is the same as for the regular records. Continuing with the above `Article` example:

`func FindArticleBySlug(app core.App, slug string) (*Article, error) { article := &Article{} err := app.RecordQuery("articles"). AndWhere(dbx.NewExp("LOWER(slug)={:slug}", dbx.Params{ "slug": strings.ToLower(slug), // case insensitive match })). Limit(1). One(article) if err != nil { return nil, err } return article, nil } ... article, err := FindArticleBySlug(app, "example") if err != nil { return err } // change the title article.SetTitle("Lorem ipsum...") // persist the change while also triggering the original record validations and hooks err = app.Save(article) if err != nil { return err }`

If you have an existing `*core.Record` value you can also load it into your proxy using the `SetProxyRecord` method:

`// fetch regular record record, err := app.FindRecordById("articles", "RECORD_ID") if err != nil { return err } // load into proxy article := &Article{} article.SetProxyRecord(record)`

* * *

[Prev: Miscellaneous](/docs/go-miscellaneous)

[FAQ](/faq) [Discussions](https://github.com/pocketbase/pocketbase/discussions) [Documentation](/docs)

[JavaScript SDK](https://github.com/pocketbase/js-sdk) [Dart SDK](https://github.com/pocketbase/dart-sdk)

Pocket**Base**

[](mailto:support)[](https://github.com/pocketbase/pocketbase)

© 2023-2025 Pocket**Base** The Gopher artwork is from [marcusolsson/gophers](https://github.com/marcusolsson/gophers)

Crafted by [**Gani**](https://gani.bg)

{ \_\_sveltekit\_jfyndd = { base: new URL("../..", location).pathname.slice(0, -1) }; const element = document.currentScript.parentElement; const data = \[null,null,null,null\]; Promise.all(\[ import("../../\_app/immutable/entry/start.HyAB6n-v.js"), import("../../\_app/immutable/entry/app.DExXzKO6.js") \]).then((\[kit, app\]) => { kit.start(app, element, { node\_ids: \[0, 2, 3, 31\], data, form: null, error: null }); }); }