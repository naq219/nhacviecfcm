                              // CSS3 animations support check (function () { if (typeof document === "undefined") { return; // serverside rendering } let elem = document.createElement("div"); let hasAnimation = elem.style.animationName !== undefined; if (!hasAnimation) { document.documentElement.classList.add("no-animations"); } })(); // silent svelte console errors if (typeof window === "undefined") { window = {}; } window.Prism = window.Prism || {}; window.Prism.manual = true; if (typeof navigator !== "undefined" && navigator.userAgentData) { navigator.userAgentData.getHighEntropyValues(\[ "architecture", \]).then((ua) => { window.UA\_ARCHITECTURE = ua.architecture }); }                       Extend with Go - Sending emails - Docs - PocketBase

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

Extend with Go - Sending emails

Sending emails

PocketBase provides a simple abstraction for sending emails via the `app.NewMailClient()` factory.

Depending on your configured mail settings (_Dashboard > Settings > Mail settings_) it will use the `sendmail` command or a SMTP client.

### [Send custom email](#send-custom-email)

You can send your own custom email from anywhere within the app (hooks, middlewares, routes, etc.) by using `app.NewMailClient().Send(message)`. Here is an example of sending a custom email after user registration:

`// main.go package main import ( "log" "net/mail" "github.com/pocketbase/pocketbase" "github.com/pocketbase/pocketbase/core" "github.com/pocketbase/pocketbase/tools/mailer" ) func main() { app := pocketbase.New() app.OnRecordCreateRequest("users").BindFunc(func(e *core.RecordRequestEvent) error { if err := e.Next(); err != nil { return err } message := &mailer.Message{ From: mail.Address{ Address: e.App.Settings().Meta.SenderAddress, Name: e.App.Settings().Meta.SenderName, }, To: []mail.Address{{Address: e.Record.Email()}}, Subject: "YOUR_SUBJECT...", HTML: "YOUR_HTML_BODY...", // bcc, cc, attachments and custom headers are also supported... } return e.App.NewMailClient().Send(message) }) if err := app.Start(); err != nil { log.Fatal(err) } }`

### [Overwrite system emails](#overwrite-system-emails)

If you want to overwrite the default system emails for forgotten password, verification, etc., you can adjust the default templates available from the _Dashboard > Collections > Edit collection > Options_ .

Alternatively, you can also apply individual changes by binding to one of the [mailer hooks](/docs/go-event-hooks/#mailer-hooks). Here is an example of appending a Record field value to the subject using the `OnMailerRecordPasswordResetSend` hook:

`// main.go package main import ( "log" "github.com/pocketbase/pocketbase" "github.com/pocketbase/pocketbase/core" ) func main() { app := pocketbase.New() app.OnMailerRecordPasswordResetSend("users").BindFunc(func(e *core.MailerRecordEvent) error { // modify the subject e.Message.Subject += (" " + e.Record.GetString("name")) return e.Next() }) if err := app.Start(); err != nil { log.Fatal(err) } }`

* * *

[Prev: Jobs scheduling](/docs/go-jobs-scheduling) [Next: Rendering templates](/docs/go-rendering-templates)

[FAQ](/faq) [Discussions](https://github.com/pocketbase/pocketbase/discussions) [Documentation](/docs)

[JavaScript SDK](https://github.com/pocketbase/js-sdk) [Dart SDK](https://github.com/pocketbase/dart-sdk)

Pocket**Base**

[](mailto:support)[](https://github.com/pocketbase/pocketbase)

© 2023-2025 Pocket**Base** The Gopher artwork is from [marcusolsson/gophers](https://github.com/marcusolsson/gophers)

Crafted by [**Gani**](https://gani.bg)

{ \_\_sveltekit\_jfyndd = { base: new URL("../..", location).pathname.slice(0, -1) }; const element = document.currentScript.parentElement; const data = \[null,null,null,null\]; Promise.all(\[ import("../../\_app/immutable/entry/start.HyAB6n-v.js"), import("../../\_app/immutable/entry/app.DExXzKO6.js") \]).then((\[kit, app\]) => { kit.start(app, element, { node\_ids: \[0, 2, 3, 35\], data, form: null, error: null }); }); }