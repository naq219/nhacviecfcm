                              // CSS3 animations support check (function () { if (typeof document === "undefined") { return; // serverside rendering } let elem = document.createElement("div"); let hasAnimation = elem.style.animationName !== undefined; if (!hasAnimation) { document.documentElement.classList.add("no-animations"); } })(); // silent svelte console errors if (typeof window === "undefined") { window = {}; } window.Prism = window.Prism || {}; window.Prism.manual = true; if (typeof navigator !== "undefined" && navigator.userAgentData) { navigator.userAgentData.getHighEntropyValues(\[ "architecture", \]).then((ua) => { window.UA\_ARCHITECTURE = ua.architecture }); }                      Extend with Go - Realtime messaging - Docs - PocketBase

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

Extend with Go - Realtime messaging

Realtime messaging

By default PocketBase sends realtime events only for Record create/update/delete operations (_and for the OAuth2 auth redirect_), but you are free to send custom realtime messages to the connected clients via the [`app.SubscriptionsBroker()`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#BaseApp.SubscriptionsBroker) instance.

[`app.SubscriptionsBroker().Clients()`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/subscriptions#Broker.Clients) returns all connected [`subscriptions.Client`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/subscriptions#Client) indexed by their unique connection id.

[`app.SubscriptionsBroker().ChunkedClients(size)`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/subscriptions#Broker.ChunkedClients) is similar but returns the result as a chunked slice allowing you to split the iteration across several goroutines (usually combined with [`errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup) ).

The current auth record associated with a client could be accessed through `client.Get(apis.RealtimeClientAuthKey)`

Note that a single authenticated user could have more than one active realtime connection (aka. multiple clients). This could happen for example when opening the same app in different tabs, browsers, devices, etc.

Below you can find a minimal code sample that sends a JSON payload to all clients subscribed to the "example" topic:

`func notify(app core.App, subscription string, data any) error { rawData, err := json.Marshal(data) if err != nil { return err } message := subscriptions.Message{ Name: subscription, Data: rawData, } group := new(errgroup.Group) chunks := app.SubscriptionsBroker().ChunkedClients(300) for _, chunk := range chunks { group.Go(func() error { for _, client := range chunk { if !client.HasSubscription(subscription) { continue } client.Send(message) } return nil }) } return group.Wait() } err := notify(app, "example", map[string]any{"test": 123}) if err != nil { return err }`

From the client-side, users can listen to the custom subscription topic by doing something like:

JavaScript

Dart

`import PocketBase from 'pocketbase'; const pb = new PocketBase('http://127.0.0.1:8090'); ... await pb.realtime.subscribe('example', (e) => { console.log(e) })`

`import 'package:pocketbase/pocketbase.dart'; final pb = PocketBase('http://127.0.0.1:8090'); ... await pb.realtime.subscribe('example', (e) { print(e) })`

* * *

[Prev: Console commands](/docs/go-console-commands) [Next: Filesystem](/docs/go-filesystem)

[FAQ](/faq) [Discussions](https://github.com/pocketbase/pocketbase/discussions) [Documentation](/docs)

[JavaScript SDK](https://github.com/pocketbase/js-sdk) [Dart SDK](https://github.com/pocketbase/dart-sdk)

Pocket**Base**

[](mailto:support)[](https://github.com/pocketbase/pocketbase)

© 2023-2025 Pocket**Base** The Gopher artwork is from [marcusolsson/gophers](https://github.com/marcusolsson/gophers)

Crafted by [**Gani**](https://gani.bg)

{ \_\_sveltekit\_jfyndd = { base: new URL("../..", location).pathname.slice(0, -1) }; const element = document.currentScript.parentElement; const data = \[null,null,null,null\]; Promise.all(\[ import("../../\_app/immutable/entry/start.HyAB6n-v.js"), import("../../\_app/immutable/entry/app.DExXzKO6.js") \]).then((\[kit, app\]) => { kit.start(app, element, { node\_ids: \[0, 2, 3, 30\], data, form: null, error: null }); }); }