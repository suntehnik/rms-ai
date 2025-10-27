package init

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileManager handles file system operations for MCP server initialization,
// including directory creation, file writing, and permission management.
type FileManager struct{}

// NewFileManager creates a new FileManager instance.
func NewFileManager() *FileManager {
	return &FileManager{}
}

// EnsureConfigDirectory creates the configuration directory if it doesn't exist,
// setting appropriate permissions (0755) for directory access.
// The directory path is extracted from the provided config file path.
func (f *FileManager) EnsureConfigDirectory(configPath string) error {
	// Extract directory path from config file path
	configDir := filepath.Dir(configPath)

	// Check if directory already exists
	if info, err := os.Stat(configDir); err == nil {
		// Directory exists, check if it's actually a directory
		if !info.IsDir() {
			return fmt.Errorf("config path exists but is not a directory: %s", configDir)
		}
		// Directory exists and is valid, ensure proper permissions
		return f.setDirectoryPermissions(configDir)
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("failed to check config directory: %w", err)
	}

	// Directory doesn't exist, create it with proper permissions
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}

	return nil
}

// WriteConfig writes the configuration data to the specified file path
// with secure file permissions (0600) to protect sensitive information.
func (f *FileManager) WriteConfig(configPath string, configData []byte) error {
	// Write the configuration file with secure permissions
	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configPath, err)
	}

	return nil
}

// SetSecurePermissions sets secure file permissions (0600) on the specified file.
// This ensures only the owner can read and write the file, protecting sensitive data.
func (f *FileManager) SetSecurePermissions(filePath string) error {
	if err := os.Chmod(filePath, 0600); err != nil {
		return fmt.Errorf("failed to set secure permissions on %s: %w", filePath, err)
	}
	return nil
}

// ConfigExists checks if a configuration file exists at the specified path.
func (f *FileManager) ConfigExists(configPath string) bool {
	_, err := os.Stat(configPath)
	return err == nil
}

// GetFileInfo returns file information for the specified path, useful for
// checking file existence, permissions, and modification times.
func (f *FileManager) GetFileInfo(filePath string) (os.FileInfo, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}
	return info, nil
}

// ValidateConfigPath checks if the provided config path is valid and accessible.
// It verifies the parent directory exists or can be created.
func (f *FileManager) ValidateConfigPath(configPath string) error {
	// Check if path is absolute
	if !filepath.IsAbs(configPath) {
		return fmt.Errorf("config path must be absolute: %s", configPath)
	}

	// Extract and validate directory
	configDir := filepath.Dir(configPath)

	// Check if we can access/create the directory
	if err := f.EnsureConfigDirectory(configPath); err != nil {
		return fmt.Errorf("cannot access config directory: %w", err)
	}

	// Check if we have write permissions to the directory
	testFile := filepath.Join(configDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("no write permission to config directory %s: %w", configDir, err)
	}

	// Clean up test file
	os.Remove(testFile)

	return nil
}

// BackupExistingConfig creates a backup of an existing configuration file
// by copying it to a new file with a timestamp suffix. Returns the backup file path.
func (f *FileManager) BackupExistingConfig(configPath string) (string, error) {
	// Check if the original config file exists
	if !f.ConfigExists(configPath) {
		return "", fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup.%s", configPath, timestamp)

	// Read original file
	originalData, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read original config file: %w", err)
	}

	// Write backup file with same permissions as original
	originalInfo, err := f.GetFileInfo(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to get original file permissions: %w", err)
	}

	if err := os.WriteFile(backupPath, originalData, originalInfo.Mode()); err != nil {
		return "", fmt.Errorf("failed to create backup file %s: %w", backupPath, err)
	}

	return backupPath, nil
}

// RestoreFromBackup restores a configuration file from a backup.
// This is useful for recovery scenarios.
func (f *FileManager) RestoreFromBackup(configPath, backupPath string) error {
	// Check if backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// Read backup file
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Write to original config path with secure permissions
	if err := f.WriteConfig(configPath, backupData); err != nil {
		return fmt.Errorf("failed to restore config from backup: %w", err)
	}

	return nil
}

// ListBackups returns a list of backup files for the given config path.
// Backup files follow the pattern: configpath.backup.YYYYMMDD-HHMMSS
func (f *FileManager) ListBackups(configPath string) ([]string, error) {
	configDir := filepath.Dir(configPath)
	configFile := filepath.Base(configPath)

	// Read directory contents
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %w", err)
	}

	var backups []string
	backupPrefix := configFile + ".backup."

	// Find all backup files
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > len(backupPrefix) {
			if entry.Name()[:len(backupPrefix)] == backupPrefix {
				backups = append(backups, filepath.Join(configDir, entry.Name()))
			}
		}
	}

	return backups, nil
}

// CleanupOldBackups removes backup files older than the specified number of days.
// This helps prevent backup files from accumulating indefinitely.
func (f *FileManager) CleanupOldBackups(configPath string, maxAgeDays int) error {
	backups, err := f.ListBackups(configPath)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -maxAgeDays)
	var deletedCount int

	for _, backupPath := range backups {
		info, err := os.Stat(backupPath)
		if err != nil {
			continue // Skip files we can't stat
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(backupPath); err == nil {
				deletedCount++
			}
		}
	}

	return nil
}

// GetBackupInfo returns information about a backup file including its creation time.
func (f *FileManager) GetBackupInfo(backupPath string) (time.Time, error) {
	info, err := os.Stat(backupPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get backup info: %w", err)
	}

	return info.ModTime(), nil
}

// setDirectoryPermissions ensures the directory has proper permissions (0755).
func (f *FileManager) setDirectoryPermissions(dirPath string) error {
	if err := os.Chmod(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to set directory permissions on %s: %w", dirPath, err)
	}
	return nil
}
