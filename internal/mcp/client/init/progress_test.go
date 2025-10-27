package init

import (
	"errors"
	"testing"
	"time"
)

func TestProgressIndicator(t *testing.T) {
	indicator := NewProgressIndicator()

	// Test basic start/stop functionality
	indicator.Start("Testing progress...")
	time.Sleep(200 * time.Millisecond) // Let it animate briefly
	indicator.Stop()

	// Test ShowProgress with successful function
	err := indicator.ShowProgress("Testing function execution...", func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test ShowProgress with failing function
	testError := errors.New("test error")
	err = indicator.ShowProgress("Testing error handling...", func() error {
		time.Sleep(50 * time.Millisecond)
		return testError
	})
	if err != testError {
		t.Errorf("Expected test error, got: %v", err)
	}
}

func TestProgressTracker(t *testing.T) {
	tracker := NewProgressTracker()

	// Test initial state
	progress := tracker.GetOverallProgress()
	if progress != 0 {
		t.Errorf("Expected 0%% progress initially, got: %.1f%%", progress)
	}

	// Test starting and completing steps
	tracker.StartStep("url_collection")
	tracker.CompleteStep("url_collection")

	progress = tracker.GetOverallProgress()
	expectedProgress := float64(1) / float64(7) * 100 // 1 out of 7 steps
	if progress != expectedProgress {
		t.Errorf("Expected %.1f%% progress, got: %.1f%%", expectedProgress, progress)
	}

	// Test failing a step
	tracker.StartStep("connectivity_test")
	testError := errors.New("connection failed")
	tracker.FailStep("connectivity_test", testError)

	// Progress should still be the same since failed step doesn't count as completed
	progress = tracker.GetOverallProgress()
	if progress != expectedProgress {
		t.Errorf("Expected %.1f%% progress after failure, got: %.1f%%", expectedProgress, progress)
	}
}

func TestStatusMessage(t *testing.T) {
	// Test known operation
	msg := GetStatusMessage("url_collection")
	if msg.Operation != "Server URL Collection" {
		t.Errorf("Expected 'Server URL Collection', got: %s", msg.Operation)
	}
	if len(msg.Messages) == 0 {
		t.Error("Expected messages to be populated")
	}
	if len(msg.Tips) == 0 {
		t.Error("Expected tips to be populated")
	}

	// Test unknown operation
	msg = GetStatusMessage("unknown_operation")
	if msg.Operation != "Unknown Operation" {
		t.Errorf("Expected 'Unknown Operation', got: %s", msg.Operation)
	}
}

func TestProgressIndicatorTimeout(t *testing.T) {
	indicator := NewProgressIndicator()

	// Test timeout functionality
	start := time.Now()
	err := indicator.ShowProgressWithTimeout("Testing timeout...", 100*time.Millisecond, func() error {
		time.Sleep(200 * time.Millisecond) // Sleep longer than timeout
		return nil
	})

	elapsed := time.Since(start)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	if elapsed < 100*time.Millisecond {
		t.Errorf("Expected at least 100ms elapsed, got: %v", elapsed)
	}
}

func TestStepStatus(t *testing.T) {
	tests := []struct {
		status   StepStatus
		expected string
	}{
		{StepStatusPending, "⏳"},
		{StepStatusInProgress, "🔄"},
		{StepStatusCompleted, "✅"},
		{StepStatusFailed, "❌"},
	}

	for _, test := range tests {
		result := test.status.String()
		if result != test.expected {
			t.Errorf("Expected %s for status %d, got: %s", test.expected, test.status, result)
		}
	}
}
