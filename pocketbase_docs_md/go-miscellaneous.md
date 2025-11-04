                              // CSS3 animations support check (function () { if (typeof document === "undefined") { return; // serverside rendering } let elem = document.createElement("div"); let hasAnimation = elem.style.animationName !== undefined; if (!hasAnimation) { document.documentElement.classList.add("no-animations"); } })(); // silent svelte console errors if (typeof window === "undefined") { window = {}; } window.Prism = window.Prism || {}; window.Prism.manual = true; if (typeof navigator !== "undefined" && navigator.userAgentData) { navigator.userAgentData.getHighEntropyValues(\[ "architecture", \]).then((ua) => { window.UA\_ARCHITECTURE = ua.architecture }); }                       Extend with Go - Miscellaneous - Docs - PocketBase

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

Extend with Go - Miscellaneous

Miscellaneous

### [app.Store()](#app-store)

[`app.Store()`](https://pkg.go.dev/github.com/pocketbase/pocketbase/core#BaseApp.Store) returns a concurrent-safe application memory store that you can use to store anything for the duration of the application process (e.g. cache, config flags, etc.).

You can find more details about the available store methods in the [`store.Store`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/store#Store) documentation but the most commonly used ones are `Get(key)`, `Set(key, value)` and `GetOrSet(key, setFunc)`.

`app.Store().Set("example", 123) v1 := app.Store().Get("example").(int) // 123 v2 := app.Store().GetOrSet("example2", func() any { // this setter is invoked only once unless "example2" is removed // (e.g. suitable for instantiating singletons) return 456 }).(int) // 456`

Keep in mind that the application store is also used internally usually with `pb*` prefixed keys (e.g. the collections cache is stored under the `pbAppCachedCollections` key) and changing these system keys or calling `RemoveAll()`/`Reset()` could have unintended side-effects.

If you want more advanced control you can initialize your own store independent from the application instance via `store.New[K, T](nil)`.

### [Security helpers](#security-helpers)

_Below are listed some of the most commonly used security helpers but you can find detailed documentation for all available methods in the [`security`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tools/security) subpackage._

##### [Generating random strings](#generating-random-strings)

`secret := security.RandomString(10) // e.g. a35Vdb10Z4 secret := security.RandomStringWithAlphabet(5, "1234567890") // e.g. 33215`

##### [Compare strings with constant time](#compare-strings-with-constant-time)

`isEqual := security.Equal(hash1, hash2)`

##### [AES Encrypt/Decrypt](#aes-encryptdecrypt)

`// must be random 32 characters string const key = "KaNom0KbaT2i0PoOfJQGd34R3NVf6cRQ" encrypted, err := security.Encrypt([]byte("test"), key) if err != nil { return err } decrypted := security.Decrypt(encrypted, key) // []byte("test")`

* * *

[Prev: Testing](/docs/go-testing) [Next: Record proxy](/docs/go-record-proxy)

[FAQ](/faq) [Discussions](https://github.com/pocketbase/pocketbase/discussions) [Documentation](/docs)

[JavaScript SDK](https://github.com/pocketbase/js-sdk) [Dart SDK](https://github.com/pocketbase/dart-sdk)

Pocket**Base**

[](mailto:support)[](https://github.com/pocketbase/pocketbase)

© 2023-2025 Pocket**Base** The Gopher artwork is from [marcusolsson/gophers](https://github.com/marcusolsson/gophers)

Crafted by [**Gani**](https://gani.bg)

{ \_\_sveltekit\_jfyndd = { base: new URL("../..", location).pathname.slice(0, -1) }; const element = document.currentScript.parentElement; const data = \[null,null,null,null\]; Promise.all(\[ import("../../\_app/immutable/entry/start.HyAB6n-v.js"), import("../../\_app/immutable/entry/app.DExXzKO6.js") \]).then((\[kit, app\]) => { kit.start(app, element, { node\_ids: \[0, 2, 3, 28\], data, form: null, error: null }); }); }