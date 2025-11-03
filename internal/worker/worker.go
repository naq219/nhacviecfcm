package worker

import (
    "context"
    "log"
    "time"

    "remiaq/internal/repository"
    "remiaq/internal/services"
)

// Worker periodically processes due reminders when enabled in system_status.
type Worker struct {
    sysRepo         repository.SystemStatusRepository
    reminderService *services.ReminderService
    interval        time.Duration
}

// NewWorker creates a new Worker.
func NewWorker(sysRepo repository.SystemStatusRepository, reminderService *services.ReminderService, interval time.Duration) *Worker {
    return &Worker{
        sysRepo:         sysRepo,
        reminderService: reminderService,
        interval:        interval,
    }
}

// Start launches the background loop. It stops when ctx is cancelled.
func (w *Worker) Start(ctx context.Context) {
    if w == nil {
        return
    }

    // Minimum safety interval
    if w.interval <= 0 {
        w.interval = time.Minute
    }

    ticker := time.NewTicker(w.interval)
    go func() {
        defer ticker.Stop()

        log.Printf("Worker started (interval=%s)", w.interval.String())

        for {
            select {
            case <-ticker.C:
                w.runOnce(ctx)
            case <-ctx.Done():
                log.Println("Worker stopped")
                return
            }
        }
    }()
}

// runOnce performs a single worker cycle.
func (w *Worker) runOnce(ctx context.Context) {
    // Check if worker is enabled
    enabled, err := w.sysRepo.IsWorkerEnabled(ctx)
    if err != nil {
        log.Printf("Worker: failed to check system status: %v", err)
        return
    }
    if !enabled {
        // Silently skip when disabled
        return
    }

    // Process reminders
    if err := w.reminderService.ProcessDueReminders(ctx); err != nil {
        // On system-level errors, disable worker and record error
        log.Printf("Worker: disabling due to error: %v", err)
        _ = w.sysRepo.DisableWorker(ctx, err.Error())
        return
    }

    // Clear previous error if any
    _ = w.sysRepo.ClearError(ctx)
}