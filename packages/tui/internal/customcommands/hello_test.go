package customcommands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sst/opencode/internal/commands"
)

// TestHelloCommandReturnsOutput tests that a simple hello command returns expected output
func TestHelloCommandReturnsOutput(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create commands directory and hello.md file
	commandsDir := filepath.Join(tempDir, ".opencode", "commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	// Create a simple hello command with markdown content and headers
	helloContent := `---
name: "hello"
description: "Say hello with markdown headers"
trigger: ["hello", "hi"]
---

# Hello Command

This is a greeting command.

## Output

!echo "Hello World"

## End

That's all!
`

	helloFile := filepath.Join(commandsDir, "hello.md")
	err = os.WriteFile(helloFile, []byte(helloContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create hello.md: %v", err)
	}

	// Create a registry to test command loading and execution
	commandRegistry := commands.CommandRegistry{} // Empty registry for testing
	scanner := NewScanner(commandsDir)
	registry := NewRegistryWithScanner(commandRegistry, scanner)

	// Load the commands
	registry.RefreshCommands()

	// Verify the command was loaded
	cmd := registry.FindCustomCommandByTrigger("hello")
	if cmd == nil {
		t.Fatal("Expected hello command to be loaded, but it was not found")
	}

	// Execute the command with no arguments
	result, err := registry.ExecuteCustomCommand("hello", "")
	if err != nil {
		t.Fatalf("Failed to execute hello command: %v", err)
	}

	// Verify the command returns output
	if result == "" {
		t.Errorf("Expected hello command to return output, but got empty string")
	}

	t.Logf("Command output: %q", result)

	// The output should contain the bash command output
	if !strings.Contains(result, "Hello World") {
		t.Errorf("Expected hello command output to contain 'Hello World', but got: %q", result)
	}

	// The output should contain actual content but NOT markdown headers (which are comments)
	expectedContent := []string{
		"This is a greeting command.",
		"That's all!",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected hello command output to contain %q, but it was missing from: %q", expected, result)
		}
	}

	// Verify that markdown headers are correctly skipped as comments
	skippedHeaders := []string{
		"# Hello Command",
		"## Output",
		"## End",
	}

	for _, header := range skippedHeaders {
		if strings.Contains(result, header) {
			t.Errorf("Expected markdown header %q to be skipped as comment, but it was included in output: %q", header, result)
		}
	}
}

// TestHelloCommandWithArguments tests that the hello command handles arguments
func TestHelloCommandWithArguments(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create commands directory
	commandsDir := filepath.Join(tempDir, ".opencode", "commands")
	err := os.MkdirAll(commandsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create commands directory: %v", err)
	}

	helloContent := `---
name: "hello"
description: "Say hello with arguments"
trigger: ["hello"]
---

# Hello Command

Arguments: $ARGUMENTS

!echo "Hello $ARGUMENTS"
`

	helloFile := filepath.Join(commandsDir, "hello.md")
	err = os.WriteFile(helloFile, []byte(helloContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create hello.md: %v", err)
	}

	// Create registry and test
	commandRegistry := commands.CommandRegistry{}
	scanner := NewScanner(commandsDir)
	registry := NewRegistryWithScanner(commandRegistry, scanner)

	registry.RefreshCommands()

	// Test with arguments
	result, err := registry.ExecuteCustomCommand("hello", "World!")
	if err != nil {
		t.Fatalf("Failed to execute hello command with arguments: %v", err)
	}

	if result == "" {
		t.Errorf("Expected hello command with arguments to return output, but got empty string")
	}

	// Should contain the bash command output with substituted arguments
	if !strings.Contains(result, "Hello World!") {
		t.Errorf("Expected hello command output to contain 'Hello World!', but got: %q", result)
	}

	// Should contain the argument substitution in regular text
	if !strings.Contains(result, "Arguments: World!") {
		t.Errorf("Expected hello command output to contain 'Arguments: World!', but got: %q", result)
	}

	// Verify that markdown headers are correctly skipped as comments
	if strings.Contains(result, "# Hello Command") {
		t.Errorf("Expected markdown header '# Hello Command' to be skipped as comment, but it was included in output: %q", result)
	}
}
