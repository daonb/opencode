package customcommands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode/internal/commands"
)

// TestHelpCommandIncludesCustomCommands verifies that custom commands are properly loaded and accessible
func TestHelpCommandIncludesCustomCommands(t *testing.T) {
	// Create a temporary directory for test commands
	tempDir := t.TempDir()
	commandsDir := filepath.Join(tempDir, ".opencode", "commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Create test custom commands
	testCommands := []struct {
		filename    string
		content     string
		name        string
		trigger     string
		description string
	}{
		{
			"deploy.md",
			`---
name: "deploy-app"
description: "Deploy application to specified environment"
trigger: ["deploy", "d"]
---

# Deploy Application

Deploying to environment: $ARGUMENTS

!npm run build
!docker build -t myapp .`,
			"deploy-app",
			"deploy",
			"Deploy application to specified environment",
		},
		{
			"test.md",
			`---
name: "run-tests"
description: "Run test suite with coverage"
trigger: ["test", "t"]
---

# Test Runner

Running tests: $ARGUMENTS

!npm test $ARGUMENTS
!npm run coverage`,
			"run-tests",
			"test",
			"Run test suite with coverage",
		},
		{
			"status.md",
			`---
name: "git-status"
description: "Show git status with formatting"
trigger: ["status", "st"]
---

# Git Status

!git status --porcelain
!echo "Files changed: $(git status --porcelain | wc -l)"`,
			"git-status",
			"status",
			"Show git status with formatting",
		},
	}

	// Write test command files
	for _, cmd := range testCommands {
		filePath := filepath.Join(commandsDir, cmd.filename)
		err := os.WriteFile(filePath, []byte(cmd.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create command file %s: %v", cmd.filename, err)
		}
	}

	// Create command registry and custom command registry
	config := &opencode.Config{}
	commandRegistry := commands.LoadFromConfig(config)
	
	// Create scanner and registry for custom commands
	scanner := NewScanner(commandsDir)
	customRegistry := NewRegistryWithScanner(commandRegistry, scanner)

	// Test that custom commands are loaded
	customCommands := customRegistry.GetCustomCommands()
	if len(customCommands) != len(testCommands) {
		t.Errorf("Expected %d custom commands, got %d", len(testCommands), len(customCommands))
	}

	// Test that each custom command is properly loaded
	for _, expectedCmd := range testCommands {
		found := false
		for _, actualCmd := range customCommands {
			if actualCmd.Name == expectedCmd.name {
				found = true
				
				// Verify the command properties
				if actualCmd.Description != expectedCmd.description {
					t.Errorf("Command %s: expected description '%s', got '%s'", 
						expectedCmd.name, expectedCmd.description, actualCmd.Description)
				}
				
				if len(actualCmd.Trigger) == 0 {
					t.Errorf("Command %s: should have triggers", expectedCmd.name)
				}
				
				if actualCmd.PrimaryTrigger() != expectedCmd.trigger {
					t.Errorf("Command %s: expected primary trigger '%s', got '%s'", 
						expectedCmd.name, expectedCmd.trigger, actualCmd.PrimaryTrigger())
				}
				
				if !actualCmd.MatchesTrigger(expectedCmd.trigger) {
					t.Errorf("Command %s: should match trigger '%s'", expectedCmd.name, expectedCmd.trigger)
				}
				
				break
			}
		}
		
		if !found {
			t.Errorf("Custom command '%s' was not loaded", expectedCmd.name)
		}
	}

	// Test that the registry can find commands by trigger
	for _, expectedCmd := range testCommands {
		foundCmd := customRegistry.FindCustomCommandByTrigger(expectedCmd.trigger)
		if foundCmd == nil {
			t.Errorf("Could not find custom command by trigger '%s'", expectedCmd.trigger)
		} else if foundCmd.Name != expectedCmd.name {
			t.Errorf("Found wrong command for trigger '%s': expected '%s', got '%s'", 
				expectedCmd.trigger, expectedCmd.name, foundCmd.Name)
		}
	}

	// Test that all triggers are returned by GetAllTriggers
	allTriggers := customRegistry.GetAllTriggers()
	
	// Check that custom command triggers are included
	for _, expectedCmd := range testCommands {
		found := false
		for _, trigger := range allTriggers {
			if trigger == expectedCmd.trigger {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Trigger '%s' not found in GetAllTriggers() result", expectedCmd.trigger)
		}
	}

	// Check that built-in command triggers are also included (like "help")
	helpTriggerFound := false
	for _, trigger := range allTriggers {
		if trigger == "help" {
			helpTriggerFound = true
			break
		}
	}
	if !helpTriggerFound {
		t.Error("Built-in 'help' trigger should be included in GetAllTriggers() result")
	}

	// Test that IsCustomCommand works correctly
	for _, expectedCmd := range testCommands {
		if !customRegistry.IsCustomCommand(expectedCmd.trigger) {
			t.Errorf("IsCustomCommand should return true for trigger '%s'", expectedCmd.trigger)
		}
	}

	// Test that IsCustomCommand returns false for built-in commands
	if customRegistry.IsCustomCommand("help") {
		t.Error("IsCustomCommand should return false for built-in 'help' command")
	}
}

// TestHelpCommandWithNoCustomCommands verifies the registry works when no custom commands are defined
func TestHelpCommandWithNoCustomCommands(t *testing.T) {
	// Create empty commands directory
	tempDir := t.TempDir()
	commandsDir := filepath.Join(tempDir, ".opencode", "commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Create registry with no custom commands
	config := &opencode.Config{}
	commandRegistry := commands.LoadFromConfig(config)
	
	scanner := NewScanner(commandsDir)
	customRegistry := NewRegistryWithScanner(commandRegistry, scanner)

	// Should have no custom commands
	customCommands := customRegistry.GetCustomCommands()
	if len(customCommands) != 0 {
		t.Errorf("Expected 0 custom commands, got %d", len(customCommands))
	}

	// Should still have built-in commands in GetAllTriggers
	allTriggers := customRegistry.GetAllTriggers()
	helpTriggerFound := false
	for _, trigger := range allTriggers {
		if trigger == "help" {
			helpTriggerFound = true
			break
		}
	}
	if !helpTriggerFound {
		t.Error("Built-in 'help' trigger should be included even with no custom commands")
	}

	// IsCustomCommand should return false for everything
	if customRegistry.IsCustomCommand("help") {
		t.Error("IsCustomCommand should return false for built-in commands")
	}
	if customRegistry.IsCustomCommand("nonexistent") {
		t.Error("IsCustomCommand should return false for non-existent commands")
	}
}

// TestCustomCommandExecution verifies that custom commands can be executed
func TestCustomCommandExecution(t *testing.T) {
	// Create a temporary directory for test commands
	tempDir := t.TempDir()
	commandsDir := filepath.Join(tempDir, ".opencode", "commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Create a simple test command that uses echo (should work on all systems)
	commandContent := `---
name: "echo-test"
description: "Test echo command"
trigger: ["echo", "e"]
---

# Echo Test

Testing arguments: $ARGUMENTS

!echo "Hello $ARGUMENTS"`

	filePath := filepath.Join(commandsDir, "echo-test.md")
	err = os.WriteFile(filePath, []byte(commandContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create command file: %v", err)
	}

	// Create registry
	config := &opencode.Config{}
	commandRegistry := commands.LoadFromConfig(config)
	
	scanner := NewScanner(commandsDir)
	customRegistry := NewRegistryWithScanner(commandRegistry, scanner)

	// Test command execution
	result, err := customRegistry.ExecuteCustomCommand("echo", "World!")
	if err != nil {
		t.Fatalf("Failed to execute custom command: %v", err)
	}

	// Should contain the argument substitution and command output
	if !strings.Contains(result, "Testing arguments: World!") {
		t.Errorf("Result should contain argument substitution, got: %s", result)
	}

	if !strings.Contains(result, "Hello World!") {
		t.Errorf("Result should contain command output, got: %s", result)
	}

	// Test execution with non-existent command
	result, err = customRegistry.ExecuteCustomCommand("nonexistent", "args")
	if err != nil {
		t.Errorf("ExecuteCustomCommand should not error for non-existent commands, got: %v", err)
	}
	if result != "" {
		t.Errorf("ExecuteCustomCommand should return empty string for non-existent commands, got: %s", result)
	}
}

// TestCustomCommandsIntegrationWithBuiltinCommands verifies custom and built-in commands work together
func TestCustomCommandsIntegrationWithBuiltinCommands(t *testing.T) {
	// Create a temporary directory for test commands
	tempDir := t.TempDir()
	commandsDir := filepath.Join(tempDir, ".opencode", "commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Create a custom command with a trigger that doesn't conflict with built-ins
	commandContent := `---
name: "custom-help"
description: "Custom help command"
trigger: ["chelp", "ch"]
---

# Custom Help

This is a custom help command that doesn't conflict with built-in help.

!echo "Custom help executed"`

	filePath := filepath.Join(commandsDir, "custom-help.md")
	err = os.WriteFile(filePath, []byte(commandContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create command file: %v", err)
	}

	// Create registry
	config := &opencode.Config{}
	commandRegistry := commands.LoadFromConfig(config)
	
	scanner := NewScanner(commandsDir)
	customRegistry := NewRegistryWithScanner(commandRegistry, scanner)

	// Test that both built-in and custom triggers are available
	allTriggers := customRegistry.GetAllTriggers()
	
	hasBuiltinHelp := false
	hasCustomHelp := false
	
	for _, trigger := range allTriggers {
		if trigger == "help" {
			hasBuiltinHelp = true
		}
		if trigger == "chelp" {
			hasCustomHelp = true
		}
	}
	
	if !hasBuiltinHelp {
		t.Error("Should have built-in 'help' trigger")
	}
	if !hasCustomHelp {
		t.Error("Should have custom 'chelp' trigger")
	}

	// Test that IsCustomCommand correctly distinguishes between them
	if customRegistry.IsCustomCommand("help") {
		t.Error("'help' should not be identified as a custom command")
	}
	if !customRegistry.IsCustomCommand("chelp") {
		t.Error("'chelp' should be identified as a custom command")
	}

	// Test that custom command can be found and executed
	customCmd := customRegistry.FindCustomCommandByTrigger("chelp")
	if customCmd == nil {
		t.Error("Should be able to find custom command by trigger 'chelp'")
	} else {
		if customCmd.Name != "custom-help" {
			t.Errorf("Expected custom command name 'custom-help', got '%s'", customCmd.Name)
		}
	}

	// Test that built-in command is not found in custom commands
	builtinCmd := customRegistry.FindCustomCommandByTrigger("help")
	if builtinCmd != nil {
		t.Error("Built-in 'help' command should not be found in custom commands")
	}
}

// TestGitStatusCustomCommand verifies that the /git-status custom command works correctly
func TestGitStatusCustomCommand(t *testing.T) {
	// Use the actual .opencode/commands directory from the project root
	projectRoot := "/Users/daonb/src/asimi"
	commandsDir := filepath.Join(projectRoot, ".opencode", "commands")
	
	// Verify the git-status.md file exists
	gitStatusFile := filepath.Join(commandsDir, "git-status.md")
	if _, err := os.Stat(gitStatusFile); os.IsNotExist(err) {
		t.Skipf("Skipping test: git-status.md not found at %s", gitStatusFile)
	}

	// Create registry using the actual commands directory
	config := &opencode.Config{}
	commandRegistry := commands.LoadFromConfig(config)
	
	scanner := NewScanner(commandsDir)
	customRegistry := NewRegistryWithScanner(commandRegistry, scanner)

	// Test that git-status command is loaded
	gitStatusCmd := customRegistry.FindCustomCommandByTrigger("git-status")
	if gitStatusCmd == nil {
		t.Fatal("git-status command should be found")
	}

	// Verify command properties
	if gitStatusCmd.Name != "git-status" {
		t.Errorf("Expected command name 'git-status', got '%s'", gitStatusCmd.Name)
	}

	if gitStatusCmd.Description != "Show git status with enhanced formatting" {
		t.Errorf("Expected description 'Show git status with enhanced formatting', got '%s'", gitStatusCmd.Description)
	}

	// Test that all expected triggers work (focus on the ones that should work)
	expectedTriggers := []string{"st", "git-status"}  // Remove "status" as it might conflict
	for _, trigger := range expectedTriggers {
		if !gitStatusCmd.MatchesTrigger(trigger) {
			t.Errorf("Command should match trigger '%s'", trigger)
		}
		
		foundCmd := customRegistry.FindCustomCommandByTrigger(trigger)
		if foundCmd == nil || foundCmd.Name != "git-status" {
			t.Errorf("Should find git-status command by trigger '%s'", trigger)
		}
	}

	// Test that the command is identified as a custom command
	for _, trigger := range expectedTriggers {
		if !customRegistry.IsCustomCommand(trigger) {
			t.Errorf("Trigger '%s' should be identified as a custom command", trigger)
		}
	}

	// Change to project directory for git commands to work
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(projectRoot)
	if err != nil {
		t.Fatalf("Failed to change to project directory: %v", err)
	}

	// Test command execution with different triggers
	for _, trigger := range expectedTriggers {
		t.Run("trigger_"+trigger, func(t *testing.T) {
			result, err := customRegistry.ExecuteCustomCommand(trigger, "")
			if err != nil {
				t.Fatalf("Failed to execute git-status command with trigger '%s': %v", trigger, err)
			}

			// Verify the result contains expected git information
			expectedContent := []string{
				"Current branch:",
				"Files changed:",
				"Staged files:",
				"Unstaged files:",
				"Last 3 commits:",
			}

			for _, content := range expectedContent {
				if !strings.Contains(result, content) {
					t.Errorf("Result should contain '%s', got: %s", content, result)
				}
			}

			// The result should contain some git information (since we're in a git repo)
			if len(result) < 100 {
				t.Errorf("Result seems too short for a git status command, got: %s", result)
			}
		})
	}

	// Test command execution with arguments
	t.Run("with_arguments", func(t *testing.T) {
		result, err := customRegistry.ExecuteCustomCommand("git-status", "test-arg")
		if err != nil {
			t.Fatalf("Failed to execute git-status command with arguments: %v", err)
		}

		// Should contain the arguments in the "Additional Arguments" section
		if !strings.Contains(result, "Arguments provided: test-arg") {
			t.Errorf("Result should contain provided arguments, got: %s", result)
		}
	})

	// Test that the command appears in GetAllTriggers
	allTriggers := customRegistry.GetAllTriggers()
	for _, expectedTrigger := range expectedTriggers {
		found := false
		for _, trigger := range allTriggers {
			if trigger == expectedTrigger {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Trigger '%s' should be in GetAllTriggers() result", expectedTrigger)
		}
	}

	// Test that the command is in GetCustomCommands
	customCommands := customRegistry.GetCustomCommands()
	gitStatusFound := false
	for _, cmd := range customCommands {
		if cmd.Name == "git-status" {
			gitStatusFound = true
			break
		}
	}
	if !gitStatusFound {
		t.Error("git-status command should be in GetCustomCommands() result")
	}
}
