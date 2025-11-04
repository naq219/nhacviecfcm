                              // CSS3 animations support check (function () { if (typeof document === "undefined") { return; // serverside rendering } let elem = document.createElement("div"); let hasAnimation = elem.style.animationName !== undefined; if (!hasAnimation) { document.documentElement.classList.add("no-animations"); } })(); // silent svelte console errors if (typeof window === "undefined") { window = {}; } window.Prism = window.Prism || {}; window.Prism.manual = true; if (typeof navigator !== "undefined" && navigator.userAgentData) { navigator.userAgentData.getHighEntropyValues(\[ "architecture", \]).then((ua) => { window.UA\_ARCHITECTURE = ua.architecture }); }                        Extend with Go - Logging - Docs - PocketBase

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

Extend with Go - Logging

Logging

`app.Logger()` provides access to a standard `slog.Logger` implementation that writes any logs into the database so that they can be later explored from the PocketBase _Dashboard > Logs_ section.

For better performance and to minimize blocking on hot paths, logs are written with debounce and on batches:

*   3 seconds after the last debounced log write
*   when the batch threshold is reached (currently 200)
*   right before app termination to attempt saving everything from the existing logs queue

### [Log methods](#log-methods)

All standard [`slog.Logger`](https://pkg.go.dev/log/slog) methods are available but below is a list with some of the most notable ones.

##### [Debug(message, attrs...)](#debugmessage-attrs-)

`app.Logger().Debug("Debug message!") app.Logger().Debug( "Debug message with attributes!", "name", "John Doe", "id", 123, )`

##### [Info(message, attrs...)](#infomessage-attrs-)

`app.Logger().Info("Info message!") app.Logger().Info( "Info message with attributes!", "name", "John Doe", "id", 123, )`

##### [Warn(message, attrs...)](#warnmessage-attrs-)

`app.Logger().Warn("Warning message!") app.Logger().Warn( "Warning message with attributes!", "name", "John Doe", "id", 123, )`

##### [Error(message, attrs...)](#errormessage-attrs-)

`app.Logger().Error("Error message!") app.Logger().Error( "Error message with attributes!", "id", 123, "error", err, )`

##### [With(attrs...)](#withattrs-)

`With(atrs...)` creates a new local logger that will "inject" the specified attributes with each following log.

`l := app.Logger().With("total", 123) // results in log with data {"total": 123} l.Info("message A") // results in log with data {"total": 123, "name": "john"} l.Info("message B", "name", "john")`

##### [WithGroup(name)](#withgroupname)

`WithGroup(name)` creates a new local logger that wraps all logs attributes under the specified group name.

`l := app.Logger().WithGroup("sub") // results in log with data {"sub": { "total": 123 }} l.Info("message A", "total", 123)`

### [Logs settings](#logs-settings)

You can control various log settings like logs retention period, minimal log level, request IP logging, etc. from the logs settings panel:

![Logs settings screenshot](/images/screenshots/logs.png)

### [Custom log queries](#custom-log-queries)

The logs are usually meant to be filtered from the UI but if you want to programmatically retrieve and filter the stored logs you can make use of the [`app.LogQuery()`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#BaseApp.LogsQuery) query builder method. For example:

`logs := []*core.Log{} // see https://pocketbase.io/docs/go-database/#query-builder err := app.LogQuery(). // target only debug and info logs AndWhere(dbx.In("level", -4, 0). // the data column is serialized json object and could be anything AndWhere(dbx.NewExp("json_extract(data, '$.type') = 'request'")). OrderBy("created DESC"). Limit(100). All(&logs)`

### [Intercepting logs write](#intercepting-logs-write)

If you want to modify the log data before persisting in the database or to forward it to an external system, then you can listen for changes of the `_logs` table by attaching to the [base model hooks](/docs/go-event-hooks/#base-model-hooks). For example:

`app.OnModelCreate(core.LogsTableName).BindFunc(func(e *core.ModelEvent) error { l := e.Model.(*core.Log) fmt.Println(l.Id) fmt.Println(l.Created) fmt.Println(l.Level) fmt.Println(l.Message) fmt.Println(l.Data) return e.Next() })`

* * *

[Prev: Filesystem](/docs/go-filesystem) [Next: Testing](/docs/go-testing)

[FAQ](/faq) [Discussions](https://github.com/pocketbase/pocketbase/discussions) [Documentation](/docs)

[JavaScript SDK](https://github.com/pocketbase/js-sdk) [Dart SDK](https://github.com/pocketbase/dart-sdk)

Pocket**Base**

[](mailto:support)[](https://github.com/pocketbase/pocketbase)

© 2023-2025 Pocket**Base** The Gopher artwork is from [marcusolsson/gophers](https://github.com/marcusolsson/gophers)

Crafted by [**Gani**](https://gani.bg)

{ \_\_sveltekit\_jfyndd = { base: new URL("../..", location).pathname.slice(0, -1) }; const element = document.currentScript.parentElement; const data = \[null,null,null,null\]; Promise.all(\[ import("../../\_app/immutable/entry/start.HyAB6n-v.js"), import("../../\_app/immutable/entry/app.DExXzKO6.js") \]).then((\[kit, app\]) => { kit.start(app, element, { node\_ids: \[0, 2, 3, 26\], data, form: null, error: null }); }); }