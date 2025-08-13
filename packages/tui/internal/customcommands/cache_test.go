package customcommands

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLazyLoading(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "customcommands_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test command file
	commandContent := `---
name: test-command
description: A test command
trigger: ["test"]
---

This is a test command.
`
	commandFile := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(commandFile, []byte(commandContent), 0644); err != nil {
		t.Fatalf("Failed to write test command file: %v", err)
	}

	// Create scanner and registry
	scanner := NewScanner(tempDir)
	registry := NewRegistryWithScanner(nil, scanner)

	// First access should trigger loading
	commands := registry.GetCustomCommands()
	if len(commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(commands))
	}

	// Verify the command was loaded correctly
	if commands[0].Name != "test-command" {
		t.Errorf("Expected command name 'test-command', got '%s'", commands[0].Name)
	}

	// Second access should use cache (no file system access)
	commands2 := registry.GetCustomCommands()
	if len(commands2) != 1 {
		t.Errorf("Expected 1 command from cache, got %d", len(commands2))
	}
}

func TestCacheInvalidation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "customcommands_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test command file
	commandContent := `---
name: test-command
description: A test command
trigger: ["test"]
---

This is a test command.
`
	commandFile := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(commandFile, []byte(commandContent), 0644); err != nil {
		t.Fatalf("Failed to write test command file: %v", err)
	}

	// Create scanner and registry
	scanner := NewScanner(tempDir)
	registry := NewRegistryWithScanner(nil, scanner)

	// First access should trigger loading
	commands := registry.GetCustomCommands()
	if len(commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(commands))
	}

	// Wait a bit to ensure different modification time
	time.Sleep(10 * time.Millisecond)

	// Modify the file
	newContent := `---
name: updated-command
description: An updated test command
trigger: ["updated"]
---

This is an updated test command.
`
	if err := os.WriteFile(commandFile, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to update test command file: %v", err)
	}

	// Force refresh
	registry.RefreshCommands()

	// Should get updated command
	commands = registry.GetCustomCommands()
	if len(commands) != 1 {
		t.Errorf("Expected 1 command after update, got %d", len(commands))
	}

	if commands[0].Name != "updated-command" {
		t.Errorf("Expected updated command name 'updated-command', got '%s'", commands[0].Name)
	}
}

func TestFileModificationTimeTracking(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "customcommands_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create scanner
	scanner := NewScanner(tempDir)

	// Test with empty cache
	commands, errors, modTimes := scanner.ScanCommandsWithCache(make(map[string]time.Time))
	if len(commands) != 0 {
		t.Errorf("Expected 0 commands in empty directory, got %d", len(commands))
	}
	if len(errors) != 0 {
		t.Errorf("Expected 0 errors in empty directory, got %d", len(errors))
	}
	if len(modTimes) != 0 {
		t.Errorf("Expected 0 modification times in empty directory, got %d", len(modTimes))
	}

	// Create a test command file
	commandContent := `---
name: test-command
description: A test command
trigger: ["test"]
---

This is a test command.
`
	commandFile := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(commandFile, []byte(commandContent), 0644); err != nil {
		t.Fatalf("Failed to write test command file: %v", err)
	}

	// Scan with empty cache - should load the file
	commands, errors, modTimes = scanner.ScanCommandsWithCache(make(map[string]time.Time))
	if len(commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(commands))
	}
	if len(modTimes) != 1 {
		t.Errorf("Expected 1 modification time, got %d", len(modTimes))
	}

	// Scan again with the same modification times - should not reload
	commands2, errors2, modTimes2 := scanner.ScanCommandsWithCache(modTimes)
	if len(commands2) != 0 {
		t.Errorf("Expected 0 commands when file hasn't changed, got %d", len(commands2))
	}
	if len(errors2) != 0 {
		t.Errorf("Expected 0 errors when file hasn't changed, got %d", len(errors2))
	}
	if len(modTimes2) != 1 {
		t.Errorf("Expected 1 modification time even when file hasn't changed, got %d", len(modTimes2))
	}
}
