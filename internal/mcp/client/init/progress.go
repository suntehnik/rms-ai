package init

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressIndicator manages animated progress display during long-running operations.
// It provides visual feedback to users during network requests and other time-consuming tasks.
type ProgressIndicator struct {
	message     string
	done        chan bool
	mu          sync.Mutex
	isRunning   bool
	currentStep int
	totalSteps  int
}

// ProgressStep represents a step in the initialization process with its status.
type ProgressStep struct {
	Name        string
	Description string
	Status      StepStatus
	StartTime   time.Time
	EndTime     time.Time
	Error       error
}

// StepStatus represents the current status of a progress step.
type StepStatus int

const (
	StepStatusPending StepStatus = iota
	StepStatusInProgress
	StepStatusCompleted
	StepStatusFailed
)

// String returns a human-readable representation of the step status.
func (s StepStatus) String() string {
	switch s {
	case StepStatusPending:
		return "⏳"
	case StepStatusInProgress:
		return "🔄"
	case StepStatusCompleted:
		return "✅"
	case StepStatusFailed:
		return "❌"
	default:
		return "❓"
	}
}

// ProgressTracker manages the overall progress of the initialization process.
type ProgressTracker struct {
	steps       []ProgressStep
	currentStep int
	mu          sync.RWMutex
}

// NewProgressIndicator creates a new progress indicator for animated display.
func NewProgressIndicator() *ProgressIndicator {
	return &ProgressIndicator{
		done: make(chan bool, 1),
	}
}

// NewProgressTracker creates a new progress tracker for managing initialization steps.
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		steps: []ProgressStep{
			{Name: "url_collection", Description: "Collecting server URL", Status: StepStatusPending},
			{Name: "connectivity_test", Description: "Testing server connectivity", Status: StepStatusPending},
			{Name: "credential_collection", Description: "Collecting user credentials", Status: StepStatusPending},
			{Name: "authentication", Description: "Authenticating with server", Status: StepStatusPending},
			{Name: "pat_generation", Description: "Generating Personal Access Token", Status: StepStatusPending},
			{Name: "config_generation", Description: "Generating configuration file", Status: StepStatusPending},
			{Name: "config_validation", Description: "Validating configuration", Status: StepStatusPending},
		},
	}
}

// Start begins displaying an animated progress indicator with the given message.
func (p *ProgressIndicator) Start(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		return // Already running
	}

	p.message = message
	p.isRunning = true
	go p.animate()
}

// Stop stops the progress indicator and clears the line.
func (p *ProgressIndicator) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isRunning {
		return // Not running
	}

	p.isRunning = false
	select {
	case p.done <- true:
	default:
		// Channel might be full, that's okay
	}

	// Clear the line
	fmt.Print("\r" + strings.Repeat(" ", len(p.message)+10) + "\r")
}

// UpdateMessage updates the progress message while the indicator is running.
func (p *ProgressIndicator) UpdateMessage(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.message = message
}

// animate displays the spinning animation with enhanced visual feedback.
func (p *ProgressIndicator) animate() {
	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	colors := []string{
		"\033[36m", // Cyan
		"\033[34m", // Blue
		"\033[35m", // Magenta
	}
	reset := "\033[0m"

	i := 0
	colorIndex := 0

	for {
		p.mu.Lock()
		if !p.isRunning {
			p.mu.Unlock()
			return
		}
		message := p.message
		p.mu.Unlock()

		select {
		case <-p.done:
			return
		default:
			// Create animated display with color
			color := colors[colorIndex%len(colors)]
			spinner := chars[i%len(chars)]

			fmt.Printf("\r%s%s%s %s", color, spinner, reset, message)

			time.Sleep(100 * time.Millisecond)
			i++

			// Change color every 10 frames for subtle animation
			if i%10 == 0 {
				colorIndex++
			}
		}
	}
}

// ShowProgress displays a progress indicator for the duration of the provided function.
func (p *ProgressIndicator) ShowProgress(message string, fn func() error) error {
	p.Start(message)
	defer p.Stop()

	return fn()
}

// ShowProgressWithTimeout displays a progress indicator with a timeout warning.
func (p *ProgressIndicator) ShowProgressWithTimeout(message string, timeout time.Duration, fn func() error) error {
	p.Start(message)
	defer p.Stop()

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	timeoutTimer := time.NewTimer(timeout)
	defer timeoutTimer.Stop()

	warningTimer := time.NewTimer(timeout / 2)
	defer warningTimer.Stop()

	for {
		select {
		case err := <-done:
			return err
		case <-warningTimer.C:
			p.UpdateMessage(message + " (taking longer than expected...)")
		case <-timeoutTimer.C:
			p.UpdateMessage(message + " (operation timed out)")
			return fmt.Errorf("operation timed out after %v", timeout)
		}
	}
}

// StartStep marks a step as in progress and updates the display.
func (t *ProgressTracker) StartStep(stepName string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i := range t.steps {
		if t.steps[i].Name == stepName {
			t.steps[i].Status = StepStatusInProgress
			t.steps[i].StartTime = time.Now()
			t.currentStep = i
			break
		}
	}

	t.displayProgress()
}

// CompleteStep marks a step as completed and updates the display.
func (t *ProgressTracker) CompleteStep(stepName string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i := range t.steps {
		if t.steps[i].Name == stepName {
			t.steps[i].Status = StepStatusCompleted
			t.steps[i].EndTime = time.Now()
			break
		}
	}

	t.displayProgress()
}

// FailStep marks a step as failed and updates the display.
func (t *ProgressTracker) FailStep(stepName string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i := range t.steps {
		if t.steps[i].Name == stepName {
			t.steps[i].Status = StepStatusFailed
			t.steps[i].EndTime = time.Now()
			t.steps[i].Error = err
			break
		}
	}

	t.displayProgress()
}

// displayProgress shows the current progress of all steps.
func (t *ProgressTracker) displayProgress() {
	fmt.Println("\n📋 Initialization Progress:")
	fmt.Println("==========================")

	for i, step := range t.steps {
		prefix := "   "
		if i == t.currentStep {
			prefix = "➤  "
		}

		duration := ""
		if !step.StartTime.IsZero() && !step.EndTime.IsZero() {
			duration = fmt.Sprintf(" (%.1fs)", step.EndTime.Sub(step.StartTime).Seconds())
		} else if !step.StartTime.IsZero() && step.Status == StepStatusInProgress {
			duration = fmt.Sprintf(" (%.1fs)", time.Since(step.StartTime).Seconds())
		}

		fmt.Printf("%s%s %s%s\n", prefix, step.Status.String(), step.Description, duration)

		if step.Status == StepStatusFailed && step.Error != nil {
			fmt.Printf("      Error: %v\n", step.Error)
		}
	}
	fmt.Println()
}

// GetOverallProgress returns the overall completion percentage.
func (t *ProgressTracker) GetOverallProgress() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	completed := 0
	for _, step := range t.steps {
		if step.Status == StepStatusCompleted {
			completed++
		}
	}

	return float64(completed) / float64(len(t.steps)) * 100
}

// DisplaySummary shows a summary of the initialization process.
func (t *ProgressTracker) DisplaySummary() {
	t.mu.RLock()
	defer t.mu.RUnlock()

	completed := 0
	failed := 0
	totalDuration := time.Duration(0)

	for _, step := range t.steps {
		switch step.Status {
		case StepStatusCompleted:
			completed++
			if !step.StartTime.IsZero() && !step.EndTime.IsZero() {
				totalDuration += step.EndTime.Sub(step.StartTime)
			}
		case StepStatusFailed:
			failed++
		}
	}

	fmt.Println("📊 Initialization Summary:")
	fmt.Println("=========================")
	fmt.Printf("✅ Completed steps: %d/%d\n", completed, len(t.steps))
	if failed > 0 {
		fmt.Printf("❌ Failed steps: %d\n", failed)
	}
	fmt.Printf("⏱️  Total time: %.1fs\n", totalDuration.Seconds())
	fmt.Printf("📈 Success rate: %.1f%%\n", float64(completed)/float64(len(t.steps))*100)
	fmt.Println()
}

// StatusMessage provides contextual status messages for different operations.
type StatusMessage struct {
	Operation string
	Messages  []string
	Tips      []string
}

// GetStatusMessage returns appropriate status messages for different operations.
func GetStatusMessage(operation string) StatusMessage {
	messages := map[string]StatusMessage{
		"url_collection": {
			Operation: "Server URL Collection",
			Messages: []string{
				"🌐 Collecting backend API server URL...",
				"🔍 Validating URL format and accessibility...",
				"✅ Server URL validated successfully",
			},
			Tips: []string{
				"Ensure the URL includes the protocol (http:// or https://)",
				"The server should be running and accessible",
				"Check firewall settings if connection fails",
			},
		},
		"connectivity_test": {
			Operation: "Server Connectivity Test",
			Messages: []string{
				"🔗 Testing connection to server...",
				"📡 Checking server health endpoint...",
				"✅ Server connectivity confirmed",
			},
			Tips: []string{
				"This verifies the server is running and reachable",
				"The /ready endpoint must be available",
				"Network connectivity is required",
			},
		},
		"credential_collection": {
			Operation: "Credential Collection",
			Messages: []string{
				"🔐 Collecting authentication credentials...",
				"👤 Validating username and password format...",
				"✅ Credentials collected securely",
			},
			Tips: []string{
				"Your password will not be displayed as you type",
				"Credentials are used only for token generation",
				"They are not stored in the configuration file",
			},
		},
		"authentication": {
			Operation: "Server Authentication",
			Messages: []string{
				"🔑 Authenticating with backend server...",
				"🎫 Requesting JWT authentication token...",
				"✅ Authentication successful",
			},
			Tips: []string{
				"This verifies your credentials with the server",
				"A temporary JWT token is generated for PAT creation",
				"The JWT token is not stored permanently",
			},
		},
		"pat_generation": {
			Operation: "Personal Access Token Generation",
			Messages: []string{
				"🎟️  Generating Personal Access Token...",
				"⏰ Setting 1-year expiration period...",
				"✅ PAT token generated successfully",
			},
			Tips: []string{
				"PAT tokens provide secure long-term access",
				"The token expires in 1 year from creation",
				"This token will be stored in your configuration",
			},
		},
		"config_generation": {
			Operation: "Configuration File Generation",
			Messages: []string{
				"📝 Generating configuration file...",
				"🔒 Setting secure file permissions...",
				"✅ Configuration saved successfully",
			},
			Tips: []string{
				"Configuration includes server URL and PAT token",
				"File permissions are set to owner-only (600)",
				"Backup is created if existing config exists",
			},
		},
		"config_validation": {
			Operation: "Configuration Validation",
			Messages: []string{
				"🔍 Validating generated configuration...",
				"🧪 Testing PAT token with MCP endpoint...",
				"✅ Configuration validated successfully",
			},
			Tips: []string{
				"This ensures the PAT token works correctly",
				"MCP protocol initialization is tested",
				"Configuration is ready for use",
			},
		},
	}

	if msg, exists := messages[operation]; exists {
		return msg
	}

	return StatusMessage{
		Operation: "Unknown Operation",
		Messages:  []string{"Processing..."},
		Tips:      []string{"Please wait while the operation completes"},
	}
}

// DisplayOperationStart shows the start of an operation with context.
func DisplayOperationStart(operation string) {
	msg := GetStatusMessage(operation)
	fmt.Printf("\n🚀 %s\n", msg.Operation)
	fmt.Println(strings.Repeat("-", len(msg.Operation)+4))

	if len(msg.Tips) > 0 {
		fmt.Println("💡 What's happening:")
		for _, tip := range msg.Tips {
			fmt.Printf("   • %s\n", tip)
		}
		fmt.Println()
	}
}

// DisplayOperationSuccess shows successful completion of an operation.
func DisplayOperationSuccess(operation string, details ...string) {
	msg := GetStatusMessage(operation)
	if len(msg.Messages) > 2 {
		fmt.Printf("%s\n", msg.Messages[2])
	} else {
		fmt.Printf("✅ %s completed successfully\n", msg.Operation)
	}

	for _, detail := range details {
		fmt.Printf("   %s\n", detail)
	}
	fmt.Println()
}

// DisplayOperationError shows an error during an operation.
func DisplayOperationError(operation string, err error) {
	msg := GetStatusMessage(operation)
	fmt.Printf("❌ %s failed: %v\n", msg.Operation, err)
	fmt.Println()
}
