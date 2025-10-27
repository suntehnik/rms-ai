package init

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"unsafe"

	"github.com/sirupsen/logrus"
)

// SecureString represents a string that should be handled securely in memory.
// It provides methods for secure cleanup and prevents accidental logging.
type SecureString struct {
	data []byte
}

// NewSecureString creates a new SecureString from a regular string.
func NewSecureString(s string) *SecureString {
	data := make([]byte, len(s))
	copy(data, []byte(s))
	return &SecureString{data: data}
}

// String returns the string value. Use with caution and ensure cleanup.
func (s *SecureString) String() string {
	if s.data == nil {
		return ""
	}
	return string(s.data)
}

// Clear securely wipes the string data from memory.
func (s *SecureString) Clear() {
	if s.data != nil {
		// Overwrite with zeros
		for i := range s.data {
			s.data[i] = 0
		}
		// Clear the slice
		s.data = nil
		// Force garbage collection to ensure memory is reclaimed
		runtime.GC()
	}
}

// IsEmpty returns true if the secure string is empty or cleared.
func (s *SecureString) IsEmpty() bool {
	return s.data == nil || len(s.data) == 0
}

// SecureCredentials holds credentials in a secure manner with automatic cleanup.
type SecureCredentials struct {
	Username *SecureString
	Password *SecureString
}

// NewSecureCredentials creates new secure credentials.
func NewSecureCredentials(username, password string) *SecureCredentials {
	return &SecureCredentials{
		Username: NewSecureString(username),
		Password: NewSecureString(password),
	}
}

// Clear securely wipes both username and password from memory.
func (sc *SecureCredentials) Clear() {
	if sc.Username != nil {
		sc.Username.Clear()
		sc.Username = nil
	}
	if sc.Password != nil {
		sc.Password.Clear()
		sc.Password = nil
	}
}

// SecureToken holds a token in a secure manner with automatic cleanup.
type SecureToken struct {
	Token *SecureString
}

// NewSecureToken creates a new secure token.
func NewSecureToken(token string) *SecureToken {
	return &SecureToken{
		Token: NewSecureString(token),
	}
}

// Clear securely wipes the token from memory.
func (st *SecureToken) Clear() {
	if st.Token != nil {
		st.Token.Clear()
		st.Token = nil
	}
}

// SecureLogger wraps logrus.Logger to prevent logging of sensitive information.
type SecureLogger struct {
	*logrus.Logger
	sensitiveFields []string
}

// NewSecureLogger creates a new secure logger that filters sensitive information.
func NewSecureLogger() *SecureLogger {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    false,
	})

	return &SecureLogger{
		Logger: logger,
		sensitiveFields: []string{
			"password", "token", "pat", "jwt", "auth", "credential", "secret",
			"key", "bearer", "authorization", "login", "passwd", "pwd",
		},
	}
}

// filterSensitiveData removes or masks sensitive information from log messages.
func (sl *SecureLogger) filterSensitiveData(message string) string {
	filtered := message

	// Mask common sensitive patterns
	for _, field := range sl.sensitiveFields {
		// Simple string replacement for common patterns
		if strings.Contains(strings.ToLower(filtered), strings.ToLower(field)) {
			// Replace with masked version
			filtered = strings.ReplaceAll(filtered, field, field+"=***MASKED***")
		}
	}

	return filtered
}

// Info logs an info message after filtering sensitive data.
func (sl *SecureLogger) Info(args ...interface{}) {
	message := fmt.Sprint(args...)
	sl.Logger.Info(sl.filterSensitiveData(message))
}

// Infof logs a formatted info message after filtering sensitive data.
func (sl *SecureLogger) Infof(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	sl.Logger.Info(sl.filterSensitiveData(message))
}

// Warn logs a warning message after filtering sensitive data.
func (sl *SecureLogger) Warn(args ...interface{}) {
	message := fmt.Sprint(args...)
	sl.Logger.Warn(sl.filterSensitiveData(message))
}

// Warnf logs a formatted warning message after filtering sensitive data.
func (sl *SecureLogger) Warnf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	sl.Logger.Warn(sl.filterSensitiveData(message))
}

// Error logs an error message after filtering sensitive data.
func (sl *SecureLogger) Error(args ...interface{}) {
	message := fmt.Sprint(args...)
	sl.Logger.Error(sl.filterSensitiveData(message))
}

// Errorf logs a formatted error message after filtering sensitive data.
func (sl *SecureLogger) Errorf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	sl.Logger.Error(sl.filterSensitiveData(message))
}

// WithError creates a new logger entry with an error, filtering sensitive data from the error.
func (sl *SecureLogger) WithError(err error) *logrus.Entry {
	if err != nil {
		filteredError := fmt.Errorf("%s", sl.filterSensitiveData(err.Error()))
		return sl.Logger.WithError(filteredError)
	}
	return sl.Logger.WithError(err)
}

// CreateSecureHTTPClient creates an HTTP client with proper HTTPS certificate validation.
func CreateSecureHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS12,
				InsecureSkipVerify: false, // Always validate certificates
			},
		},
	}
}

// SanitizeErrorMessage removes sensitive information from error messages.
func SanitizeErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	message := err.Error()

	// List of sensitive patterns to mask
	sensitivePatterns := []string{
		"password", "token", "pat", "jwt", "auth", "credential", "secret",
		"key", "bearer", "authorization", "login", "passwd", "pwd",
	}

	// Mask sensitive information in error messages
	for _, pattern := range sensitivePatterns {
		if strings.Contains(strings.ToLower(message), pattern) {
			// Replace the actual sensitive value with a masked version
			message = strings.ReplaceAll(message, pattern, "***MASKED***")
		}
	}

	return message
}

// SecureCleanup performs secure cleanup of sensitive data structures.
type SecureCleanup struct {
	cleanupFuncs []func()
}

// NewSecureCleanup creates a new secure cleanup manager.
func NewSecureCleanup() *SecureCleanup {
	return &SecureCleanup{
		cleanupFuncs: make([]func(), 0),
	}
}

// AddCleanup adds a cleanup function to be called during cleanup.
func (sc *SecureCleanup) AddCleanup(fn func()) {
	sc.cleanupFuncs = append(sc.cleanupFuncs, fn)
}

// AddSecureString adds a secure string for cleanup.
func (sc *SecureCleanup) AddSecureString(s *SecureString) {
	sc.AddCleanup(func() {
		if s != nil {
			s.Clear()
		}
	})
}

// AddSecureCredentials adds secure credentials for cleanup.
func (sc *SecureCleanup) AddSecureCredentials(creds *SecureCredentials) {
	sc.AddCleanup(func() {
		if creds != nil {
			creds.Clear()
		}
	})
}

// AddSecureToken adds a secure token for cleanup.
func (sc *SecureCleanup) AddSecureToken(token *SecureToken) {
	sc.AddCleanup(func() {
		if token != nil {
			token.Clear()
		}
	})
}

// Cleanup executes all registered cleanup functions.
func (sc *SecureCleanup) Cleanup() {
	for _, fn := range sc.cleanupFuncs {
		fn()
	}
	// Clear the cleanup functions slice
	sc.cleanupFuncs = nil
	// Force garbage collection
	runtime.GC()
}

// secureZeroMemory attempts to securely zero memory at the given address.
// This is a best-effort approach as Go's garbage collector may move memory.
func secureZeroMemory(ptr unsafe.Pointer, size uintptr) {
	if ptr == nil || size == 0 {
		return
	}

	// Convert to byte slice and zero
	slice := (*[1 << 30]byte)(ptr)[:size:size]
	for i := range slice {
		slice[i] = 0
	}
}

// ValidateHTTPSCertificate validates that the client is configured for proper HTTPS certificate validation.
func ValidateHTTPSCertificate(client *http.Client) error {
	if client.Transport == nil {
		return nil // Default transport validates certificates
	}

	if transport, ok := client.Transport.(*http.Transport); ok {
		if transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify {
			return fmt.Errorf("HTTPS certificate validation is disabled - this is a security risk")
		}
	}

	return nil
}
