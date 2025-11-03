package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"remiaq/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock SystemStatusRepository
type MockSystemStatusRepository struct {
	mock.Mock
}

func (m *MockSystemStatusRepository) Get(ctx context.Context) (*models.SystemStatus, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SystemStatus), args.Error(1)
}

func (m *MockSystemStatusRepository) IsWorkerEnabled(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockSystemStatusRepository) EnableWorker(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSystemStatusRepository) DisableWorker(ctx context.Context, errorMsg string) error {
	args := m.Called(ctx, errorMsg)
	return args.Error(0)
}

func (m *MockSystemStatusRepository) UpdateError(ctx context.Context, errorMsg string) error {
	args := m.Called(ctx, errorMsg)
	return args.Error(0)
}

func (m *MockSystemStatusRepository) ClearError(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Mock ReminderService
type MockReminderService struct {
	mock.Mock
}

func (m *MockReminderService) ProcessDueReminders(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewWorker(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	interval := time.Minute

	worker := NewWorker(mockSysRepo, mockReminderService, interval)

	assert.NotNil(t, worker)
	assert.Equal(t, mockSysRepo, worker.sysRepo)
	assert.Equal(t, mockReminderService, worker.reminderService)
	assert.Equal(t, interval, worker.interval)
}

func TestWorker_Start_NilWorker(t *testing.T) {
	var worker *Worker
	ctx := context.Background()

	// Should not panic
	assert.NotPanics(t, func() {
		worker.Start(ctx)
	})
}

func TestWorker_Start_ZeroInterval(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, 0)
	
	// Should set minimum interval to 1 minute
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	worker.Start(ctx)
	
	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)
	
	assert.Equal(t, time.Minute, worker.interval)
}

func TestWorker_Start_ContextCancellation(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, 10*time.Millisecond)
	
	// Mock worker disabled to avoid processing
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(false, nil).Maybe()
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start worker
	worker.Start(ctx)
	
	// Give it a moment to start
	time.Sleep(5 * time.Millisecond)
	
	// Cancel context
	cancel()
	
	// Give it a moment to stop
	time.Sleep(20 * time.Millisecond)
	
	// Worker should have stopped gracefully
	// This test mainly ensures no panic occurs
	assert.True(t, true) // If we reach here, no panic occurred
}

func TestWorker_runOnce_WorkerDisabled(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, time.Minute)
	
	// Mock worker disabled
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(false, nil)
	
	ctx := context.Background()
	worker.runOnce(ctx)
	
	// Should check if worker is enabled
	mockSysRepo.AssertExpectations(t)
	
	// Should not call ProcessDueReminders
	mockReminderService.AssertNotCalled(t, "ProcessDueReminders")
}

func TestWorker_runOnce_WorkerEnabled_Success(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, time.Minute)
	
	// Mock worker enabled
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(true, nil)
	
	// Mock successful processing
	mockReminderService.On("ProcessDueReminders", mock.Anything).Return(nil)
	
	// Mock clear error
	mockSysRepo.On("ClearError", mock.Anything).Return(nil)
	
	ctx := context.Background()
	worker.runOnce(ctx)
	
	// Verify all calls
	mockSysRepo.AssertExpectations(t)
	mockReminderService.AssertExpectations(t)
}

func TestWorker_runOnce_WorkerEnabled_ProcessingError(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, time.Minute)
	
	// Mock worker enabled
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(true, nil)
	
	// Mock processing error
	processingError := errors.New("FCM service unavailable")
	mockReminderService.On("ProcessDueReminders", mock.Anything).Return(processingError)
	
	// Mock disable worker
	mockSysRepo.On("DisableWorker", mock.Anything, "FCM service unavailable").Return(nil)
	
	ctx := context.Background()
	worker.runOnce(ctx)
	
	// Verify all calls
	mockSysRepo.AssertExpectations(t)
	mockReminderService.AssertExpectations(t)
}

func TestWorker_runOnce_SystemStatusError(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, time.Minute)
	
	// Mock system status error
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(false, errors.New("database connection failed"))
	
	ctx := context.Background()
	worker.runOnce(ctx)
	
	// Should check if worker is enabled
	mockSysRepo.AssertExpectations(t)
	
	// Should not call ProcessDueReminders
	mockReminderService.AssertNotCalled(t, "ProcessDueReminders")
}

func TestWorker_runOnce_DisableWorkerError(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, time.Minute)
	
	// Mock worker enabled
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(true, nil)
	
	// Mock processing error
	processingError := errors.New("FCM service unavailable")
	mockReminderService.On("ProcessDueReminders", mock.Anything).Return(processingError)
	
	// Mock disable worker error (should not crash)
	mockSysRepo.On("DisableWorker", mock.Anything, "FCM service unavailable").Return(errors.New("failed to disable"))
	
	ctx := context.Background()
	
	// Should not panic
	assert.NotPanics(t, func() {
		worker.runOnce(ctx)
	})
	
	// Verify all calls
	mockSysRepo.AssertExpectations(t)
	mockReminderService.AssertExpectations(t)
}

func TestWorker_runOnce_ClearErrorFails(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, time.Minute)
	
	// Mock worker enabled
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(true, nil)
	
	// Mock successful processing
	mockReminderService.On("ProcessDueReminders", mock.Anything).Return(nil)
	
	// Mock clear error fails (should not crash)
	mockSysRepo.On("ClearError", mock.Anything).Return(errors.New("failed to clear error"))
	
	ctx := context.Background()
	
	// Should not panic
	assert.NotPanics(t, func() {
		worker.runOnce(ctx)
	})
	
	// Verify all calls
	mockSysRepo.AssertExpectations(t)
	mockReminderService.AssertExpectations(t)
}

// Integration test - worker runs multiple cycles
func TestWorker_Integration_MultipleCycles(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, 20*time.Millisecond)
	
	// Track number of calls
	var callCount int
	var mu sync.Mutex
	
	// Mock worker enabled for first 3 calls, then disabled
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(true, nil).Times(3)
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(false, nil).Maybe()
	
	// Mock successful processing
	mockReminderService.On("ProcessDueReminders", mock.Anything).Run(func(args mock.Arguments) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}).Return(nil).Times(3)
	
	// Mock clear error
	mockSysRepo.On("ClearError", mock.Anything).Return(nil).Times(3)
	
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	worker.Start(ctx)
	
	// Wait for context to timeout
	<-ctx.Done()
	
	// Give it a moment to finish
	time.Sleep(10 * time.Millisecond)
	
	mu.Lock()
	actualCallCount := callCount
	mu.Unlock()
	
	// Should have processed at least 2-3 times
	assert.GreaterOrEqual(t, actualCallCount, 2)
	assert.LessOrEqual(t, actualCallCount, 4) // Allow some variance due to timing
}

// Test worker behavior when system status changes during runtime
func TestWorker_Integration_StatusChange(t *testing.T) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, 15*time.Millisecond)
	
	// Track calls
	var enabledCalls, disabledCalls int
	var mu sync.Mutex
	
	// Mock worker enabled for first 2 calls, then disabled
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(true, nil).Times(2).Run(func(args mock.Arguments) {
		mu.Lock()
		enabledCalls++
		mu.Unlock()
	})
	
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(false, nil).Run(func(args mock.Arguments) {
		mu.Lock()
		disabledCalls++
		mu.Unlock()
	}).Maybe()
	
	// Mock successful processing for enabled calls
	mockReminderService.On("ProcessDueReminders", mock.Anything).Return(nil).Times(2)
	
	// Mock clear error for enabled calls
	mockSysRepo.On("ClearError", mock.Anything).Return(nil).Times(2)
	
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	
	worker.Start(ctx)
	
	// Wait for context to timeout
	<-ctx.Done()
	
	// Give it a moment to finish
	time.Sleep(10 * time.Millisecond)
	
	mu.Lock()
	actualEnabledCalls := enabledCalls
	actualDisabledCalls := disabledCalls
	mu.Unlock()
	
	// Should have made enabled calls and then disabled calls
	assert.Equal(t, 2, actualEnabledCalls)
	assert.GreaterOrEqual(t, actualDisabledCalls, 1)
}

// Benchmark tests
func BenchmarkWorker_runOnce_Enabled(b *testing.B) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, time.Minute)
	
	// Mock worker enabled - use Maybe() to allow unlimited calls
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(true, nil).Maybe()
	mockReminderService.On("ProcessDueReminders", mock.Anything).Return(nil).Maybe()
	mockSysRepo.On("ClearError", mock.Anything).Return(nil).Maybe()
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		worker.runOnce(ctx)
	}
}

func BenchmarkWorker_runOnce_Disabled(b *testing.B) {
	mockSysRepo := &MockSystemStatusRepository{}
	mockReminderService := &MockReminderService{}
	
	worker := NewWorker(mockSysRepo, mockReminderService, time.Minute)
	
	// Mock worker disabled - use Maybe() to allow unlimited calls
	mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(false, nil).Maybe()
	
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		worker.runOnce(ctx)
	}
}

// Test edge cases
func TestWorker_EdgeCases(t *testing.T) {
	t.Run("negative interval should be corrected", func(t *testing.T) {
		mockSysRepo := &MockSystemStatusRepository{}
		mockReminderService := &MockReminderService{}
		
		// Mock worker disabled to avoid processing
		mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(false, nil).Maybe()
		
		worker := NewWorker(mockSysRepo, mockReminderService, -time.Hour)
		
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		worker.Start(ctx)
		
		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)
		
		assert.Equal(t, time.Minute, worker.interval)
	})
	
	t.Run("zero interval should be corrected", func(t *testing.T) {
		mockSysRepo := &MockSystemStatusRepository{}
		mockReminderService := &MockReminderService{}
		
		// Mock worker disabled to avoid processing
		mockSysRepo.On("IsWorkerEnabled", mock.Anything).Return(false, nil).Maybe()
		
		worker := NewWorker(mockSysRepo, mockReminderService, 0)
		
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		worker.Start(ctx)
		
		// Give it a moment to start
		time.Sleep(10 * time.Millisecond)
		
		assert.Equal(t, time.Minute, worker.interval)
	})
}