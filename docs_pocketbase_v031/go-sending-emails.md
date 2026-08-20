# PocketBase Go Sending Emails

PocketBase provides a simple abstraction for sending emails via the `app.NewMailClient()` factory. Depending on your mail settings (configured in the Dashboard), it will use either the `sendmail` command or an SMTP client.

## Sending Custom Emails

You can send custom emails from anywhere in your application (hooks, middlewares, routes, etc.) by using `app.NewMailClient().Send(message)`.

**Example: Sending an email after a new user registers**

```go
package main

import (
    "log"
    "net/mail"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/tools/mailer"
)

func main() {
    app := pocketbase.New()

    app.OnRecordCreateRequest("users").BindFunc(func(e *core.RecordRequestEvent) error {
        if err := e.Next(); err != nil {
            return err
        }

        message := &mailer.Message{
            From: mail.Address{
                Address: e.App.Settings().Meta.SenderAddress,
                Name:    e.App.Settings().Meta.SenderName,
            },
            To:      []mail.Address{{Address: e.Record.Email()}},
            Subject: "YOUR_SUBJECT...",
            HTML:    "YOUR_HTML_BODY...",
            // BCC, CC, attachments, and custom headers are also supported.
        }

        return e.App.NewMailClient().Send(message)
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## Overwriting System Emails

You can overwrite the default system emails (e.g., for password resets, verification) in two ways:

1.  **Dashboard:** Adjust the default templates in *Dashboard > Collections > Edit collection > Options*.
2.  **Mailer Hooks:** Apply individual changes by binding to a specific mailer hook. For more details, see the [Mailer Hooks documentation](/docs/go-event-hooks/#mailer-hooks).

**Example: Appending a record field to the password reset email subject**

```go
package main

import (
    "log"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)

func main() {
    app := pocketbase.New()

    app.OnMailerRecordPasswordResetSend("users").BindFunc(func(e *core.MailerRecordEvent) error {
        // Modify the subject
        e.Message.Subject += (" " + e.Record.GetString("name"))
        return e.Next()
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```