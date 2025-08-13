package customcommands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Scanner handles scanning directories for custom command files
type Scanner struct {
	directories []string
}

// NewScanner creates a new scanner with the specified directories
func NewScanner(directories ...string) *Scanner {
	return &Scanner{
		directories: directories,
	}
}

// NewDefaultScanner creates a scanner with default directories
func NewDefaultScanner() *Scanner {
	var directories []string

	// Add user-level commands directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		userCommandsDir := filepath.Join(homeDir, ".opencode", "commands")
		directories = append(directories, userCommandsDir)
	}

	// Add project-level commands directory
	if cwd, err := os.Getwd(); err == nil {
		projectCommandsDir := filepath.Join(cwd, ".opencode", "commands")
		directories = append(directories, projectCommandsDir)
	}

	return NewScanner(directories...)
}

// NewScannerWithProjectRoot creates a scanner with default directories using the specified project root
func NewScannerWithProjectRoot(projectRoot string) *Scanner {
	var directories []string

	// Add user-level commands directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		userCommandsDir := filepath.Join(homeDir, ".opencode", "commands")
		directories = append(directories, userCommandsDir)
	}

	// Add project-level commands directory using the provided project root
	if projectRoot != "" {
		projectCommandsDir := filepath.Join(projectRoot, ".opencode", "commands")
		directories = append(directories, projectCommandsDir)
	}

	return NewScanner(directories...)
}

// ScanCommands scans all configured directories for custom command files
func (s *Scanner) ScanCommands() ([]CustomCommand, []CommandLoadError) {
	var commands []CustomCommand
	var errors []CommandLoadError

	for _, dir := range s.directories {
		dirCommands, dirErrors := s.scanDirectory(dir)
		commands = append(commands, dirCommands...)
		errors = append(errors, dirErrors...)
	}

	return commands, errors
}

// scanDirectory scans a single directory for command files
func (s *Scanner) scanDirectory(dir string) ([]CustomCommand, []CommandLoadError) {
	var commands []CustomCommand
	var errors []CommandLoadError

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return commands, errors // Directory doesn't exist, return empty slices
	}

	// Walk through the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errors = append(errors, CommandLoadError{
				FilePath: path,
				Error:    fmt.Errorf("failed to access file: %w", err),
			})
			return nil // Continue processing other files
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			return nil
		}

		// Load the command
		command, err := LoadCustomCommand(path)
		if err != nil {
			errors = append(errors, CommandLoadError{
				FilePath: path,
				Error:    err,
			})
			return nil // Continue processing other files
		}

		commands = append(commands, *command)
		return nil
	})

	if err != nil {
		errors = append(errors, CommandLoadError{
			FilePath: dir,
			Error:    fmt.Errorf("failed to walk directory: %w", err),
		})
	}

	return commands, errors
}

// AddDirectory adds a directory to scan for commands
func (s *Scanner) AddDirectory(dir string) {
	s.directories = append(s.directories, dir)
}

// ScanCommandsWithCache scans all configured directories for custom command files with caching
func (s *Scanner) ScanCommandsWithCache(cachedModTimes map[string]time.Time) ([]CustomCommand, []CommandLoadError, map[string]time.Time) {
	var commands []CustomCommand
	var errors []CommandLoadError
	newModTimes := make(map[string]time.Time)

	for _, dir := range s.directories {
		dirCommands, dirErrors, dirModTimes := s.scanDirectoryWithCache(dir, cachedModTimes)
		commands = append(commands, dirCommands...)
		errors = append(errors, dirErrors...)

		// Merge modification times
		for path, modTime := range dirModTimes {
			newModTimes[path] = modTime
		}
	}

	return commands, errors, newModTimes
}

// scanDirectoryWithCache scans a single directory for command files with caching
func (s *Scanner) scanDirectoryWithCache(dir string, cachedModTimes map[string]time.Time) ([]CustomCommand, []CommandLoadError, map[string]time.Time) {
	var commands []CustomCommand
	var errors []CommandLoadError
	newModTimes := make(map[string]time.Time)

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return commands, errors, newModTimes // Directory doesn't exist, return empty slices
	}

	// Walk through the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errors = append(errors, CommandLoadError{
				FilePath: path,
				Error:    fmt.Errorf("failed to access file: %w", err),
			})
			return nil // Continue processing other files
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			return nil
		}

		// Record modification time
		newModTimes[path] = info.ModTime()

		// Check if file has been modified since last scan
		if cachedTime, exists := cachedModTimes[path]; exists && !info.ModTime().After(cachedTime) {
			// File hasn't changed, skip loading
			return nil
		}

		// Load the command
		command, err := LoadCustomCommand(path)
		if err != nil {
			errors = append(errors, CommandLoadError{
				FilePath: path,
				Error:    err,
			})
			return nil // Continue processing other files
		}

		commands = append(commands, *command)
		return nil
	})

	if err != nil {
		errors = append(errors, CommandLoadError{
			FilePath: dir,
			Error:    fmt.Errorf("failed to walk directory: %w", err),
		})
	}

	return commands, errors, newModTimes
}

// GetDirectories returns the list of directories being scanned
func (s *Scanner) GetDirectories() []string {
	return append([]string(nil), s.directories...) // Return a copy
}
