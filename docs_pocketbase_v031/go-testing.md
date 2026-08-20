# PocketBase Go Testing

PocketBase provides several test mocks and stubs (e.g., `tests.TestApp`, `tests.ApiScenario`, `tests.MockMultipartData`) to help you write unit and integration tests for your application. You can find more details in the [`github.com/pocketbase/pocketbase/tests`](https://pkg.go.dev/github.com/pocketbase/pocketbase/tests) package.

Here is a simple example of how to write an integration test for a custom API route.

### 1. Setup

Let's say you have a custom API route `GET /my/hello` that requires superuser authentication.

```go
// main.go
package main

import (
    "log"
    "net/http"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/apis"
    "github.com/pocketbase/pocketbase/core"
)

func bindAppHooks(app core.App) {
    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        se.Router.GET("/my/hello", func(e *core.RequestEvent) error {
            return e.JSON(http.StatusOK, "Hello world!")
        }).Bind(apis.RequireSuperuserAuth())
        return se.Next()
    })
}

func main() {
    app := pocketbase.New()
    bindAppHooks(app)

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 2. Prepare Test Data

The easiest way to prepare test data is to start your application with a custom `test_pb_data` directory:

```sh
./pocketbase serve --dir="./test_pb_data" --automigrate=0
```

Then, use the PocketBase Dashboard to create your test collections and records. Once you are done, you can stop the server. You can also commit the `test_pb_data` directory to your repository.

### 3. Integration Test

To test the example endpoint, we want to ensure that it:
- Handles only GET requests.
- Can be accessed only by superusers.
- Returns the correct response body.

Below is a simple integration test for these cases, using the test data created in the previous step.

```go
// main_test.go
package main

import (
    "net/http"
    "testing"

    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/tests"
)

const testDataDir = "./test_pb_data"

// generateToken creates a new auth token for the specified user.
func generateToken(collectionNameOrId string, email string) (string, error) {
    app, err := tests.NewTestApp(testDataDir)
    if err != nil {
        return "", err
    }
    defer app.Cleanup()

    record, err := app.FindAuthRecordByEmail(collectionNameOrId, email)
    if err != nil {
        return "", err
    }

    return record.NewAuthToken()
}

func TestHelloEndpoint(t *testing.T) {
    recordToken, err := generateToken("users", "test@example.com")
    if err != nil {
        t.Fatal(err)
    }

    superuserToken, err := generateToken(core.CollectionNameSuperusers, "test@example.com")
    if err != nil {
        t.Fatal(err)
    }

    // setupTestApp creates a new test app instance and binds the app hooks.
    setupTestApp := func(t testing.TB) *tests.TestApp {
        testApp, err := tests.NewTestApp(testDataDir)
        if err != nil {
            t.Fatal(err)
        }
        bindAppHooks(testApp)
        return testApp
    }

    scenarios := []tests.ApiScenario{
        {
            Name:           "try with different http method, e.g. POST",
            Method:         http.MethodPost,
            URL:            "/my/hello",
            ExpectedStatus: 405,
            TestAppFactory: setupTestApp,
        },
        {
            Name:           "try as guest (aka. no Authorization header)",
            Method:         http.MethodGet,
            URL:            "/my/hello",
            ExpectedStatus: 401,
            TestAppFactory: setupTestApp,
        },
        {
            Name:   "try as authenticated app user",
            Method: http.MethodGet,
            URL:    "/my/hello",
            Headers: map[string]string{
                "Authorization": recordToken,
            },
            ExpectedStatus: 401,
            TestAppFactory: setupTestApp,
        },
        {
            Name:   "try as authenticated superuser",
            Method: http.MethodGet,
            URL:    "/my/hello",
            Headers: map[string]string{
                "Authorization": superuserToken,
            },
            ExpectedStatus:  200,
            ExpectedContent: []string{"Hello world!"},
            TestAppFactory:  setupTestApp,
        },
    }

    for _, scenario := range scenarios {
        scenario.Test(t)
    }
}
```