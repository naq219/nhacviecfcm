# PocketBase Go Event Hooks

This document outlines how to use event hooks in PocketBase with Go.

## Overview

Event hooks allow you to bind custom handlers to various application and model events.

- **Binding:** Use `app.On...().BindFunc(handler)` to attach a function to an event.
- **Chaining:** Multiple handlers can be bound and will be executed in the order they are registered.
- **Stopping Execution:** Return a non-nil error from a handler to stop the chain. `e.Next()` continues to the next handler.

---

## Application Hooks

### `OnBeforeServe`
Triggered before the HTTP server starts. Useful for attaching custom routes or middleware.

```go
app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
    // e.App
    // e.Router
    return nil
})
```

### `OnAfterBootstrap`
Triggered after the app bootstrap but before the `Serve` command.

```go
app.OnAfterBootstrap().Add(func(e *core.BootstrapEvent) error {
    // e.App
    return nil
})
```

### `OnTerminate`
Triggered on app termination (e.g., `SIGINT`).

```go
app.OnTerminate().Add(func(e *core.TerminateEvent) error {
    // e.App
    return nil
})
```

---

## Mailer Hooks

### `OnMailerBeforeAdminResetPasswordSend`
Triggered before sending a password reset email to an admin.

### `OnMailerBeforeRecordResetPasswordSend`
Triggered before sending a password reset email to a record.

### `OnMailerBeforeRecordVerificationSend`
Triggered before sending a verification email to a record.

### `OnMailerBeforeRecordChangeEmailSend`
Triggered before sending an email change confirmation.

---

## Realtime Hooks

### `OnRealtimeConnect`
Triggered when a new client connects to the realtime API.

### `OnRealtimeDisconnect`
Triggered when a client disconnects.

### `OnRealtimeSubscribe`
Triggered when a client subscribes to a topic.

---

## Record Model Hooks

These are proxies for the base model hooks, specific to records.

### Create Hooks
- `OnRecordBeforeCreateRequest`: Before an API create request.
- `OnRecordAfterCreateRequest`: After an API create request.
- `OnRecordBeforeCreate`: Before a record is created.
- `OnRecordAfterCreate`: After a record is created.

### Update Hooks
- `OnRecordBeforeUpdateRequest`: Before an API update request.
- `OnRecordAfterUpdateRequest`: After an API update request.
- `OnRecordBeforeUpdate`: Before a record is updated.
- `OnRecordAfterUpdate`: After a record is updated.

### Delete Hooks
- `OnRecordBeforeDeleteRequest`: Before an API delete request.
- `OnRecordAfterDeleteRequest`: After an API delete request.
- `OnRecordBeforeDelete`: Before a record is deleted.
- `OnRecordAfterDelete`: After a record is deleted.

### Other Record Hooks
- `OnRecordAuthRequest`: On successful API record authentication.
- `OnRecordAuthRefreshRequest`: On record auth refresh.
- `OnRecordAuthWithPasswordRequest`: On auth with password request.
- `OnRecordAuthWithOAuth2Request`: On OAuth2 sign-in/sign-up.
- `OnRecordRequestPasswordResetRequest`: On request password reset.
- `OnRecordConfirmPasswordResetRequest`: On confirm password reset.
- `OnRecordRequestVerificationRequest`: On request verification.
- `OnRecordConfirmVerificationRequest`: On confirm verification.
- `OnRecordRequestEmailChangeRequest`: On request email change.
- `OnRecordConfirmEmailChangeRequest`: On confirm email change.

---

## Collection Model Hooks

These are proxies for the base model hooks, specific to collections.

### Create Hooks
- `OnCollectionBeforeCreateRequest`: Before an API create request.
- `OnCollectionAfterCreateRequest`: After an API create request.
- `OnCollectionBeforeCreate`: Before a collection is created.
- `OnCollectionAfterCreate`: After a collection is created.

### Update Hooks
- `OnCollectionBeforeUpdateRequest`: Before an API update request.
- `OnCollectionAfterUpdateRequest`: After an API update request.
- `OnCollectionBeforeUpdate`: Before a collection is updated.
- `OnCollectionAfterUpdate`: After a collection is updated.

### Delete Hooks
- `OnCollectionBeforeDeleteRequest`: Before an API delete request.
- `OnCollectionAfterDeleteRequest`: After an API delete request.
- `OnCollectionBeforeDelete`: Before a collection is deleted.
- `OnCollectionAfterDelete`: After a collection is deleted.

---

## Base Model Hooks

These hooks apply to all models (Records, Collections, etc.).

### `OnModelValidate`
Called when a model is validated.

### Create Hooks
- `OnModelBeforeCreate`: Before model creation.
- `OnModelAfterCreate`: After model creation.

### Update Hooks
- `OnModelBeforeUpdate`: Before model update.
- `OnModelAfterUpdate`: After model update.

### Delete Hooks
- `OnModelBeforeDelete`: Before model deletion.
- `OnModelAfterDelete`: After model deletion.