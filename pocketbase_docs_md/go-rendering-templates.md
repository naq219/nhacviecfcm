                              // CSS3 animations support check (function () { if (typeof document === "undefined") { return; // serverside rendering } let elem = document.createElement("div"); let hasAnimation = elem.style.animationName !== undefined; if (!hasAnimation) { document.documentElement.classList.add("no-animations"); } })(); // silent svelte console errors if (typeof window === "undefined") { window = {}; } window.Prism = window.Prism || {}; window.Prism.manual = true; if (typeof navigator !== "undefined" && navigator.userAgentData) { navigator.userAgentData.getHighEntropyValues(\[ "architecture", \]).then((ua) => { window.UA\_ARCHITECTURE = ua.architecture }); }                       Extend with Go - Rendering templates - Docs - PocketBase

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

Extend with Go - Rendering templates

Rendering templates

### [Overview](#overview)

A common task when creating custom routes or emails is the need of generating HTML output.

There are plenty of Go template-engines available that you can use for this, but often for simple cases the Go standard library `html/template` package should work just fine.

To make it slightly easier to load template files concurrently and on the fly, PocketBase also provides a thin wrapper around the standard library in the [`github.com/pocketbase/pocketbase/tools/template`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/template) utility package.

`import "github.com/pocketbase/pocketbase/tools/template" data := map[string]any{"name": "John"} html, err := template.NewRegistry().LoadFiles( "views/base.html", "views/partial1.html", "views/partial2.html", ).Render(data)`

The general flow when working with composed and nested templates is that you create "base" template(s) that defines various placeholders using the `{{template "placeholderName" .}}` or `{{block "placeholderName" .}}default...{{end}}` actions.

Then in the partials, you define the content for those placeholders using the `{{define "placeholderName"}}custom...{{end}}` action.

The dot object (`.`) in the above represents the data passed to the templates via the `Render(data)` method.

By default the templates apply contextual (HTML, JS, CSS, URI) auto escaping so the generated template content should be injection-safe. To render raw/verbatim trusted content in the templates you can use the builtin `raw` function (e.g. `{{.content|raw}}`).

For more information about the template syntax please refer to the [_html/template_](https://pkg.go.dev/html/template#hdr-A_fuller_picture) and [_text/template_](https://pkg.go.dev/text/template) package godocs. **Another great resource is also the Hashicorp's [Learn Go Template Syntax](https://developer.hashicorp.com/nomad/tutorials/templates/go-template-syntax) tutorial.**

### [Example HTML page with layout](#example-html-page-with-layout)

Consider the following app directory structure:

`myapp/ views/ layout.html hello.html main.go`

We define the content for `layout.html` as:

`<!DOCTYPE html> <html lang="en"> <head> <title>{{block "title" .}}Default app title{{end}}</title> </head> <body> Header... {{block "body" .}} Default app body... {{end}} Footer... </body> </html>`

We define the content for `hello.html` as:

`{{define "title"}} Page 1 {{end}} {{define "body"}} <p>Hello from {{.name}}</p> {{end}}`

Then to output the final page, we'll register a custom `/hello/:name` route:

`// main.go package main import ( "log" "net/http" "github.com/pocketbase/pocketbase" "github.com/pocketbase/pocketbase/core" "github.com/pocketbase/pocketbase/tools/template" ) func main() { app := pocketbase.New() app.OnServe().BindFunc(func(se *core.ServeEvent) error { // this is safe to be used by multiple goroutines // (it acts as store for the parsed templates) registry := template.NewRegistry() se.Router.GET("/hello/{name}", func(e *core.RequestEvent) error { name := e.Request.PathValue("name") html, err := registry.LoadFiles( "views/layout.html", "views/hello.html", ).Render(map[string]any{ "name": name, }) if err != nil { // or redirect to a dedicated 404 HTML page return e.NotFoundError("", err) } return e.HTML(http.StatusOK, html) }) return se.Next() }) if err := app.Start(); err != nil { log.Fatal(err) } }`

* * *

[Prev: Sending emails](/docs/go-sending-emails) [Next: Console commands](/docs/go-console-commands)

[FAQ](/faq) [Discussions](https://github.com/pocketbase/pocketbase/discussions) [Documentation](/docs)

[JavaScript SDK](https://github.com/pocketbase/js-sdk) [Dart SDK](https://github.com/pocketbase/dart-sdk)

Pocket**Base**

[](mailto:support)[](https://github.com/pocketbase/pocketbase)

© 2023-2025 Pocket**Base** The Gopher artwork is from [marcusolsson/gophers](https://github.com/marcusolsson/gophers)

Crafted by [**Gani**](https://gani.bg)

{ \_\_sveltekit\_jfyndd = { base: new URL("../..", location).pathname.slice(0, -1) }; const element = document.currentScript.parentElement; const data = \[null,null,null,null\]; Promise.all(\[ import("../../\_app/immutable/entry/start.HyAB6n-v.js"), import("../../\_app/immutable/entry/app.DExXzKO6.js") \]).then((\[kit, app\]) => { kit.start(app, element, { node\_ids: \[0, 2, 3, 33\], data, form: null, error: null }); }); }