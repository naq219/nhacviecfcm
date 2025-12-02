package worker

import (
	"context"
)

// IsWorkerSystemEnabled checks if the worker is enabled in the system settings.
// It logs an error if the check fails.
// Returns true if enabled, false otherwise.
func IsWorkerSystemEnabled(ctx context.Context, sysRepo SystemStatusRepo, logError func(format string, v ...any)) bool {
	enabled, err := sysRepo.IsWorkerEnabled(ctx)
	if err != nil {
		logError("Worker: failed to check system status: %v", err)
		return false
	}
	return enabled
}
