                              // CSS3 animations support check (function () { if (typeof document === "undefined") { return; // serverside rendering } let elem = document.createElement("div"); let hasAnimation = elem.style.animationName !== undefined; if (!hasAnimation) { document.documentElement.classList.add("no-animations"); } })(); // silent svelte console errors if (typeof window === "undefined") { window = {}; } window.Prism = window.Prism || {}; window.Prism.manual = true; if (typeof navigator !== "undefined" && navigator.userAgentData) { navigator.userAgentData.getHighEntropyValues(\[ "architecture", \]).then((ua) => { window.UA\_ARCHITECTURE = ua.architecture }); }                     Extend with Go - Jobs scheduling - Docs - PocketBase

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

Extend with Go - Jobs scheduling

Jobs scheduling

If you have tasks that need to be performed periodically, you could set up crontab-like jobs with the builtin `app.Cron()` _(it returns an app scoped [`cron.Cron`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/cron#Cron) value)_ .

The jobs scheduler is started automatically on app `serve`, so all you have to do is register a handler with [`app.Cron().Add(id, cronExpr, handler)`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/cron#Cron.Add) or [`app.Cron().MustAdd(id, cronExpr, handler)`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/cron#Cron.MustAdd) (_the latter panic if the cron expression is not valid_).

Each scheduled job runs in its own goroutine and must have:

*   **id** - identifier for the scheduled job; could be used to replace or remove an existing job
*   **cron expression** - e.g. `0 0 * * *` ( _supports numeric list, steps, ranges or macros_ )
*   **handler** - the function that will be executed every time when the job runs

Here is one minimal example:

`// main.go package main import ( "log" "github.com/pocketbase/pocketbase" ) func main() { app := pocketbase.New() // prints "Hello!" every 2 minutes app.Cron().MustAdd("hello", "*/2 * * * *", func() { log.Println("Hello!") }) if err := app.Start(); err != nil { log.Fatal(err) } }`

To remove already registered cron job you can call [`app.Cron().Remove(id)`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/cron#Cron.Remove)

All registered app level cron jobs can be also previewed and triggered from the _Dashboard > Settings > Crons_ section.

Keep in mind that the `app.Cron()` is also used for running the system scheduled jobs like the logs cleanup or auto backups (the jobs id is in the format `__pb*__`) and replacing these system jobs or calling `RemoveAll()`/`Stop()` could have unintended side-effects.

If you want more advanced control you can initialize your own cron instance independent from the application via `cron.New()`.

* * *

[Prev: Migrations](/docs/go-migrations) [Next: Sending emails](/docs/go-sending-emails)

[FAQ](/faq) [Discussions](https://github.com/pocketbase/pocketbase/discussions) [Documentation](/docs)

[JavaScript SDK](https://github.com/pocketbase/js-sdk) [Dart SDK](https://github.com/pocketbase/dart-sdk)

Pocket**Base**

[](mailto:support)[](https://github.com/pocketbase/pocketbase)

© 2023-2025 Pocket**Base** The Gopher artwork is from [marcusolsson/gophers](https://github.com/marcusolsson/gophers)

Crafted by [**Gani**](https://gani.bg)

{ \_\_sveltekit\_jfyndd = { base: new URL("../..", location).pathname.slice(0, -1) }; const element = document.currentScript.parentElement; const data = \[null,null,null,null\]; Promise.all(\[ import("../../\_app/immutable/entry/start.HyAB6n-v.js"), import("../../\_app/immutable/entry/app.DExXzKO6.js") \]).then((\[kit, app\]) => { kit.start(app, element, { node\_ids: \[0, 2, 3, 25\], data, form: null, error: null }); }); }