package customcommands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCustomCommandParsing demonstrates how custom commands would be parsed
func TestCustomCommandParsing(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a sample command file
	commandContent := `---
name: "git-status"
description: "Show git status with custom formatting"
trigger: ["status", "st"]
---

# Git Status Command

This command shows the current git status.

!git status --porcelain
!echo "Files changed: $(git status --porcelain | wc -l)"

You can also add arguments: $ARGUMENTS
`

	commandFile := filepath.Join(tempDir, "git-status.md")
	err := os.WriteFile(commandFile, []byte(commandContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test command file: %v", err)
	}

	// Test loading the command
	cmd, err := LoadCustomCommand(commandFile)
	if err != nil {
		t.Fatalf("Failed to load custom command: %v", err)
	}

	// Verify the command was parsed correctly
	if cmd.Name != "git-status" {
		t.Errorf("Expected name 'git-status', got '%s'", cmd.Name)
	}

	if cmd.Description != "Show git status with custom formatting" {
		t.Errorf("Expected description 'Show git status with custom formatting', got '%s'", cmd.Description)
	}

	if len(cmd.Trigger) != 2 {
		t.Errorf("Expected 2 triggers, got %d", len(cmd.Trigger))
	}

	if cmd.Trigger[0] != "status" || cmd.Trigger[1] != "st" {
		t.Errorf("Expected triggers ['status', 'st'], got %v", cmd.Trigger)
	}

	if !cmd.HasTrigger() {
		t.Error("Expected command to have triggers")
	}

	if cmd.PrimaryTrigger() != "status" {
		t.Errorf("Expected primary trigger 'status', got '%s'", cmd.PrimaryTrigger())
	}

	if !cmd.MatchesTrigger("status") {
		t.Error("Expected command to match trigger 'status'")
	}

	if !cmd.MatchesTrigger("st") {
		t.Error("Expected command to match trigger 'st'")
	}

	if cmd.MatchesTrigger("invalid") {
		t.Error("Expected command not to match trigger 'invalid'")
	}
}

// TestScanner demonstrates how the scanner would work
func TestScanner(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()
	commandsDir := filepath.Join(tempDir, "commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Create multiple command files
	commands := []struct {
		filename string
		content  string
	}{
		{
			"git-status.md",
			`---
name: "git-status"
description: "Show git status"
trigger: ["status"]
---
!git status`,
		},
		{
			"build.md",
			`---
name: "build-project"
description: "Build the project"
trigger: ["build", "b"]
---
!make build`,
		},
		{
			"invalid.md",
			`This file has no frontmatter and should be ignored`,
		},
		{
			"readme.txt",
			`This is not a markdown file and should be ignored`,
		},
	}

	for _, cmd := range commands {
		filePath := filepath.Join(commandsDir, cmd.filename)
		err := os.WriteFile(filePath, []byte(cmd.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create command file %s: %v", cmd.filename, err)
		}
	}

	// Test the scanner
	scanner := NewScanner(commandsDir)
	loadedCommands, errors := scanner.ScanCommands()

	// Should have errors for invalid files but continue processing
	if len(errors) == 0 {
		t.Errorf("Expected errors for invalid files, but got none")
	}

	// Check that we got an error for the invalid.md file
	foundInvalidError := false
	for _, loadErr := range errors {
		errMsg := loadErr.Error.Error()
		if strings.Contains(errMsg, "invalid.md") && strings.Contains(errMsg, "does not start with YAML frontmatter") {
			foundInvalidError = true
			break
		}
	}
	if !foundInvalidError {
		t.Errorf("Expected error for invalid.md file, got errors: %v", errors)
	}

	// Should have loaded 2 valid commands (git-status and build-project)
	if len(loadedCommands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(loadedCommands))
	}

	// Verify the commands were loaded correctly
	foundGitStatus := false
	foundBuild := false

	for _, cmd := range loadedCommands {
		switch cmd.Name {
		case "git-status":
			foundGitStatus = true
			if len(cmd.Trigger) != 1 || cmd.Trigger[0] != "status" {
				t.Errorf("Git status command has incorrect triggers: %v", cmd.Trigger)
			}
		case "build-project":
			foundBuild = true
			if len(cmd.Trigger) != 2 || cmd.Trigger[0] != "build" || cmd.Trigger[1] != "b" {
				t.Errorf("Build command has incorrect triggers: %v", cmd.Trigger)
			}
		default:
			t.Errorf("Unexpected command loaded: %s", cmd.Name)
		}
	}

	if !foundGitStatus {
		t.Error("Git status command was not loaded")
	}

	if !foundBuild {
		t.Error("Build command was not loaded")
	}
}

// TestCommandExecution demonstrates how command execution would work
func TestCommandExecution(t *testing.T) {
	// Create a simple command that doesn't require external tools
	cmd := CustomCommand{
		Name:        "echo-test",
		Description: "Test echo command",
		Trigger:     []string{"echo"},
		Content:     "!echo Hello $ARGUMENTS",
		FilePath:    "/tmp/test.md",
	}

	// Test execution with arguments
	result, err := cmd.Execute("World!")
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	expected := "Hello World!\n"
	if result != expected {
		t.Errorf("Expected output '%s', got '%s'", expected, result)
	}
}

// TestArgumentSubstitution demonstrates argument substitution
func TestArgumentSubstitution(t *testing.T) {
	cmd := CustomCommand{
		Name:        "arg-test",
		Description: "Test argument substitution",
		Trigger:     []string{"argtest"},
		Content:     "Arguments provided: $ARGUMENTS\nEnd of arguments",
		FilePath:    "/tmp/test.md",
	}

	result, err := cmd.Execute("test args here")
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	expected := "Arguments provided: test args here\nEnd of arguments\n"
	if result != expected {
		t.Errorf("Expected output '%s', got '%s'", expected, result)
	}
}

// TestModelAndPromptFields tests the model and prompt functionality
func TestModelAndPromptFields(t *testing.T) {
	tempDir := t.TempDir()

	// Test case 1: Command with no model/prompt
	basicCommandContent := `---
name: "simple-echo"
description: "Simple echo without AI"
trigger: ["echo"]
---

# Simple Echo

!echo "Hello World"
`

	basicCommandFile := filepath.Join(tempDir, "simple-echo.md")
	err := os.WriteFile(basicCommandFile, []byte(basicCommandContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create basic command file: %v", err)
	}

	basicCmd, err := LoadCustomCommand(basicCommandFile)
	if err != nil {
		t.Fatalf("Failed to load basic command: %v", err)
	}

	if basicCmd.Name != "simple-echo" {
		t.Errorf("Expected command name 'simple-echo', got '%s'", basicCmd.Name)
	}

	if len(basicCmd.Trigger) != 1 || basicCmd.Trigger[0] != "echo" {
		t.Errorf("Expected trigger ['echo'], got %v", basicCmd.Trigger)
	}

	// Execute the command to verify it works
	result, err := basicCmd.Execute("")
	if err != nil {
		t.Fatalf("Failed to execute simple command: %v", err)
	}

	if !strings.Contains(result, "Hello World") {
		t.Errorf("Expected output to contain 'Hello World', got '%s'", result)
	}
}
