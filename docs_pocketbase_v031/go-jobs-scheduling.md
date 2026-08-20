# PocketBase Go Jobs Scheduling

This document explains how to schedule and manage periodic tasks (cron jobs) in PocketBase using Go.

## Overview

PocketBase has a built-in cron job scheduler that can be accessed via `app.Cron()`. This allows you to run tasks at specified intervals.

The scheduler starts automatically when the application serves. Jobs are registered using `app.Cron().Add()` or `app.Cron().MustAdd()`.

Each job requires:
- **id:** A unique identifier for the job.
- **cron expression:** A standard cron expression (e.g., `*/2 * * * *` for every 2 minutes).
- **handler:** The function to execute.

---

## Registering a Cron Job

Here is an example of a simple cron job that logs a message every two minutes:

```go
package main

import (
    "log"
    "github.com/pocketbase/pocketbase"
)

func main() {
    app := pocketbase.New()

    // Prints "Hello!" every 2 minutes
    app.Cron().MustAdd("hello", "*/2 * * * *", func() {
        log.Println("Hello!")
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

---

## Managing Cron Jobs

- **Remove a Job:** Use `app.Cron().Remove(id)` to stop a specific job.
- **View and Trigger:** Registered jobs can be viewed and manually triggered from the PocketBase Dashboard under _Settings > Crons_.

**Caution:** The application uses the same cron scheduler for system tasks (like log cleanup and backups). Avoid using `RemoveAll()` or `Stop()` on `app.Cron()`, as it can interfere with these system jobs.

For more advanced control, you can create your own independent cron instance with `cron.New()`.