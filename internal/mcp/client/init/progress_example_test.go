package init

import (
	"fmt"
	"testing"
	"time"
)

// TestProgressIndicatorDemo demonstrates the progress indicator functionality
// This test shows how the progress indicators work in practice
func TestProgressIndicatorDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping progress indicator demo in short mode")
	}

	fmt.Println("\n🚀 Progress Indicator Demo")
	fmt.Println("==========================")

	// Create progress tracker and indicator
	tracker := NewProgressTracker()
	indicator := NewProgressIndicator()

	// Simulate the initialization process
	steps := []struct {
		name        string
		description string
		duration    time.Duration
		shouldFail  bool
	}{
		{"url_collection", "Collecting server URL", 800 * time.Millisecond, false},
		{"connectivity_test", "Testing server connectivity", 1200 * time.Millisecond, false},
		{"credential_collection", "Collecting user credentials", 600 * time.Millisecond, false},
		{"authentication", "Authenticating with server", 1000 * time.Millisecond, false},
		{"pat_generation", "Generating Personal Access Token", 900 * time.Millisecond, false},
		{"config_generation", "Generating configuration file", 400 * time.Millisecond, false},
		{"config_validation", "Validating configuration", 700 * time.Millisecond, false},
	}

	for _, step := range steps {
		// Start the step
		tracker.StartStep(step.name)
		DisplayOperationStart(step.name)

		// Simulate work with progress indicator
		message := fmt.Sprintf("🔄 %s...", step.description)
		err := indicator.ShowProgress(message, func() error {
			time.Sleep(step.duration)
			if step.shouldFail {
				return fmt.Errorf("simulated failure")
			}
			return nil
		})

		if err != nil {
			tracker.FailStep(step.name, err)
			DisplayOperationError(step.name, err)
		} else {
			tracker.CompleteStep(step.name)
			DisplayOperationSuccess(step.name, fmt.Sprintf("Completed in %.1fs", step.duration.Seconds()))
		}

		// Small pause between steps
		time.Sleep(200 * time.Millisecond)
	}

	// Display final summary
	tracker.DisplaySummary()

	fmt.Println("✅ Demo completed successfully!")
	fmt.Printf("📊 Overall progress: %.1f%%\n", tracker.GetOverallProgress())
}

// TestProgressIndicatorWithTimeout demonstrates timeout functionality
func TestProgressIndicatorWithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout demo in short mode")
	}

	fmt.Println("\n⏰ Progress Indicator Timeout Demo")
	fmt.Println("==================================")

	indicator := NewProgressIndicator()

	// Test successful operation within timeout
	fmt.Println("Testing successful operation...")
	err := indicator.ShowProgressWithTimeout("🔄 Quick operation...", 2*time.Second, func() error {
		time.Sleep(500 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	fmt.Println("✅ Operation completed within timeout")

	time.Sleep(500 * time.Millisecond)

	// Test operation that times out
	fmt.Println("\nTesting timeout scenario...")
	err = indicator.ShowProgressWithTimeout("🔄 Slow operation...", 1*time.Second, func() error {
		time.Sleep(2 * time.Second) // This will timeout
		return nil
	})
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	fmt.Printf("⏰ Operation timed out as expected: %v\n", err)
}

// TestProgressMessages demonstrates different status messages
func TestProgressMessages(t *testing.T) {
	fmt.Println("\n💬 Progress Messages Demo")
	fmt.Println("=========================")

	operations := []string{
		"url_collection",
		"connectivity_test",
		"credential_collection",
		"authentication",
		"pat_generation",
		"config_generation",
		"config_validation",
	}

	for _, op := range operations {
		msg := GetStatusMessage(op)
		fmt.Printf("\n🔧 %s:\n", msg.Operation)
		for i, tip := range msg.Tips {
			fmt.Printf("   %d. %s\n", i+1, tip)
		}
	}
}
