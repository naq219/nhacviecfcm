                              // CSS3 animations support check (function () { if (typeof document === "undefined") { return; // serverside rendering } let elem = document.createElement("div"); let hasAnimation = elem.style.animationName !== undefined; if (!hasAnimation) { document.documentElement.classList.add("no-animations"); } })(); // silent svelte console errors if (typeof window === "undefined") { window = {}; } window.Prism = window.Prism || {}; window.Prism.manual = true; if (typeof navigator !== "undefined" && navigator.userAgentData) { navigator.userAgentData.getHighEntropyValues(\[ "architecture", \]).then((ua) => { window.UA\_ARCHITECTURE = ua.architecture }); }                       Extend with Go - Testing - Docs - PocketBase

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

Extend with Go - Testing

Testing

PocketBase exposes several test mocks and stubs (eg. `tests.TestApp`, `tests.ApiScenario`, `tests.MockMultipartData`, etc.) to help you write unit and integration tests for your app.

You could find more information in the [`github.com/pocketbase/pocketbase/tests`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tests) sub package, but here is a simple example.

### [1\. Setup](#1-setup)

Let's say that we have a custom API route `GET /my/hello` that requires superuser authentication:

`// main.go package main import ( "log" "net/http" "github.com/pocketbase/pocketbase" "github.com/pocketbase/pocketbase/apis" "github.com/pocketbase/pocketbase/core" ) func bindAppHooks(app core.App) { app.OnServe().BindFunc(func(se *core.ServeEvent) error { se.Router.Get("/my/hello", func(e *core.RequestEvent) error { return e.JSON(http.StatusOK, "Hello world!") }).Bind(apis.RequireSuperuserAuth()) return se.Next() }) } func main() { app := pocketbase.New() bindAppHooks(app) if err := app.Start(); err != nil { log.Fatal(err) } }`

### [2\. Prepare test data](#2-prepare-test-data)

Now we have to prepare our test/mock data. There are several ways you can approach this, but the easiest one would be to start your application with a custom `test_pb_data` directory, e.g.:

`./pocketbase serve --dir="./test_pb_data" --automigrate=0`

Go to your browser and create the test data via the Dashboard (both collections and records). Once completed you can stop the server (you could also commit `test_pb_data` to your repo).

### [3\. Integration test](#3-integration-test)

To test the example endpoint, we want to:

*   ensure it handles only GET requests
*   ensure that it can be accessed only by superusers
*   check if the response body is properly set

Below is a simple integration test for the above test cases. We'll also use the test data created in the previous step.

`// main_test.go package main import ( "net/http" "testing" "github.com/pocketbase/pocketbase/core" "github.com/pocketbase/pocketbase/tests" ) const testDataDir = "./test_pb_data" func generateToken(collectionNameOrId string, email string) (string, error) { app, err := tests.NewTestApp(testDataDir) if err != nil { return "", err } defer app.Cleanup() record, err := app.FindAuthRecordByEmail(collectionNameOrId, email) if err != nil { return "", err } return record.NewAuthToken() } func TestHelloEndpoint(t *testing.T) { recordToken, err := generateToken("users", "test@example.com") if err != nil { t.Fatal(err) } superuserToken, err := generateToken(core.CollectionNameSuperusers, "test@example.com") if err != nil { t.Fatal(err) } // set up the test ApiScenario app instance setupTestApp := func(t testing.TB) *tests.TestApp { testApp, err := tests.NewTestApp(testDataDir) if err != nil { t.Fatal(err) } // no need to cleanup since scenario.Test() will do that for us // defer testApp.Cleanup() bindAppHooks(testApp) return testApp } scenarios := []tests.ApiScenario{ { Name: "try with different http method, e.g. POST", Method: http.MethodPost, URL: "/my/hello", ExpectedStatus: 405, ExpectedContent: []string{"\"data\":{}"}, TestAppFactory: setupTestApp, }, { Name: "try as guest (aka. no Authorization header)", Method: http.MethodGet, URL: "/my/hello", ExpectedStatus: 401, ExpectedContent: []string{"\"data\":{}"}, TestAppFactory: setupTestApp, }, { Name: "try as authenticated app user", Method: http.MethodGet, URL: "/my/hello", Headers: map[string]string{ "Authorization": recordToken, }, ExpectedStatus: 401, ExpectedContent: []string{"\"data\":{}"}, TestAppFactory: setupTestApp, }, { Name: "try as authenticated superuser", Method: http.MethodGet, URL: "/my/hello", Headers: map[string]string{ "Authorization": superuserToken, }, ExpectedStatus: 200, ExpectedContent: []string{"Hello world!"}, TestAppFactory: setupTestApp, }, } for _, scenario := range scenarios { scenario.Test(t) } }`

* * *

[Prev: Logging](/docs/go-logging) [Next: Miscellaneous](/docs/go-miscellaneous)

[FAQ](/faq) [Discussions](https://github.com/pocketbase/pocketbase/discussions) [Documentation](/docs)

[JavaScript SDK](https://github.com/pocketbase/js-sdk) [Dart SDK](https://github.com/pocketbase/dart-sdk)

Pocket**Base**

[](mailto:support)[](https://github.com/pocketbase/pocketbase)

© 2023-2025 Pocket**Base** The Gopher artwork is from [marcusolsson/gophers](https://github.com/marcusolsson/gophers)

Crafted by [**Gani**](https://gani.bg)

{ \_\_sveltekit\_jfyndd = { base: new URL("../..", location).pathname.slice(0, -1) }; const element = document.currentScript.parentElement; const data = \[null,null,null,null\]; Promise.all(\[ import("../../\_app/immutable/entry/start.HyAB6n-v.js"), import("../../\_app/immutable/entry/app.DExXzKO6.js") \]).then((\[kit, app\]) => { kit.start(app, element, { node\_ids: \[0, 2, 3, 36\], data, form: null, error: null }); }); }