package customcommands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode/internal/commands"
)

// TestNewRegistry verifies that the registry can use a specific project root
func TestNewRegistry(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "TestProjectRoot")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .opencode/commands directory
	commandsDir := filepath.Join(tmpDir, ".opencode", "commands")
	err = os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create commands dir: %v", err)
	}

	// Create a test command file
	testCommandContent := `---
name: "test-project-root"
description: "Test command for project root functionality"
trigger: ["test-root", "tr"]
---

# Test Project Root Command

This command tests that the project root is used correctly.

!echo "Project root test command executed"
`

	testCommandFile := filepath.Join(commandsDir, "test-project-root.md")
	err = os.WriteFile(testCommandFile, []byte(testCommandContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test command file: %v", err)
	}

	// Create command registry
	config := &opencode.Config{}
	commandRegistry := commands.LoadFromConfig(config)

	// Test the new registry constructor with project root
	registry := NewRegistry(commandRegistry, tmpDir)

	// Verify the command was loaded
	customCommands := registry.GetCustomCommands()
	if len(customCommands) != 1 {
		t.Errorf("Expected 1 custom command, got %d", len(customCommands))
	}

	// Verify the specific command
	cmd := registry.FindCustomCommandByTrigger("test-root")
	if cmd == nil {
		t.Error("Expected to find 'test-root' command, but it was not found")
	} else {
		if cmd.Name != "test-project-root" {
			t.Errorf("Expected command name 'test-project-root', got '%s'", cmd.Name)
		}
		if len(cmd.Trigger) != 2 {
			t.Errorf("Expected 2 triggers, got %d", len(cmd.Trigger))
		}
	}

	// Test that the command can be executed
	result, err := registry.ExecuteCustomCommand("test-root", "")
	if err != nil {
		t.Errorf("Failed to execute custom command: %v", err)
	}
	if result == "" {
		t.Error("Expected command output, got empty string")
	}
}

// TestProjectRootVsCurrentDirectory verifies that project root takes precedence over current directory
func TestProjectRootVsCurrentDirectory(t *testing.T) {
	// Create two temporary directories
	projectRootDir, err := os.MkdirTemp("", "TestProjectRoot")
	if err != nil {
		t.Fatalf("Failed to create project root dir: %v", err)
	}
	defer os.RemoveAll(projectRootDir)

	currentDir, err := os.MkdirTemp("", "TestCurrentDir")
	if err != nil {
		t.Fatalf("Failed to create current dir: %v", err)
	}
	defer os.RemoveAll(currentDir)

	// Create commands in project root
	projectCommandsDir := filepath.Join(projectRootDir, ".opencode", "commands")
	err = os.MkdirAll(projectCommandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create project commands dir: %v", err)
	}

	projectCommandContent := `---
name: "project-command"
description: "Command from project root"
trigger: ["project"]
---

# Project Command

!echo "From project root"
`

	err = os.WriteFile(filepath.Join(projectCommandsDir, "project.md"), []byte(projectCommandContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write project command file: %v", err)
	}

	// Create commands in current directory
	currentCommandsDir := filepath.Join(currentDir, ".opencode", "commands")
	err = os.MkdirAll(currentCommandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create current commands dir: %v", err)
	}

	currentCommandContent := `---
name: "current-command"
description: "Command from current directory"
trigger: ["current"]
---

# Current Command

!echo "From current directory"
`

	err = os.WriteFile(filepath.Join(currentCommandsDir, "current.md"), []byte(currentCommandContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write current command file: %v", err)
	}

	// Change to current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(currentDir)
	if err != nil {
		t.Fatalf("Failed to change to current directory: %v", err)
	}

	// Create command registry
	config := &opencode.Config{}
	commandRegistry := commands.LoadFromConfig(config)

	// Test with project root - should find project command, not current directory command
	registry := NewRegistry(commandRegistry, projectRootDir)

	customCommands := registry.GetCustomCommands()
	if len(customCommands) != 1 {
		t.Errorf("Expected 1 custom command from project root, got %d", len(customCommands))
	}

	// Should find project command
	projectCmd := registry.FindCustomCommandByTrigger("project")
	if projectCmd == nil {
		t.Error("Expected to find 'project' command from project root")
	}

	// Should NOT find current directory command
	currentCmd := registry.FindCustomCommandByTrigger("current")
	if currentCmd != nil {
		t.Error("Should not find 'current' command when using project root")
	}
}
