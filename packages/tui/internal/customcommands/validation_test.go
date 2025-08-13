package customcommands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateCommandName tests command name validation
func TestValidateCommandName(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		expectError bool
	}{
		{"valid name", "git-status", false},
		{"valid with underscore", "build_project", false},
		{"valid alphanumeric", "test123", false},
		{"empty name", "", true},
		{"too long", "this-is-a-very-long-command-name-that-exceeds-fifty-characters", true},
		{"invalid characters", "git@status", true},
		{"starts with number", "123test", true},
		{"starts with hyphen", "-test", true},
		{"starts with underscore", "_test", true},
		{"contains spaces", "git status", true},
		{"contains special chars", "git!status", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommandName(tt.commandName)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for command name '%s', but got none", tt.commandName)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for command name '%s', but got: %v", tt.commandName, err)
			}
		})
	}
}

// TestValidateTriggers tests trigger validation
func TestValidateTriggers(t *testing.T) {
	tests := []struct {
		name        string
		triggers    []string
		expectError bool
	}{
		{"valid triggers", []string{"status", "st"}, false},
		{"single valid trigger", []string{"build"}, false},
		{"empty triggers", []string{}, true},
		{"empty trigger", []string{""}, true},
		{"trigger too long", []string{"this-trigger-is-too-long"}, true},
		{"invalid trigger chars", []string{"git@status"}, true},
		{"trigger starts with number", []string{"123test"}, true},
		{"mixed valid and invalid", []string{"valid", "123invalid"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTriggers(tt.triggers)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for triggers %v, but got none", tt.triggers)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for triggers %v, but got: %v", tt.triggers, err)
			}
		})
	}
}

// TestLoadCustomCommandWithValidation tests that LoadCustomCommand validates names and triggers
func TestLoadCustomCommandWithValidation(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			"valid command",
			`---
name: "git-status"
description: "Show git status"
trigger: ["status", "st"]
---
!git status`,
			false,
			"",
		},
		{
			"invalid command name",
			`---
name: "123invalid"
description: "Invalid name"
trigger: ["test"]
---
content`,
			true,
			"invalid command name",
		},
		{
			"invalid trigger",
			`---
name: "valid-name"
description: "Valid name but invalid trigger"
trigger: ["123invalid"]
---
content`,
			true,
			"invalid triggers",
		},
		{
			"command name too long",
			`---
name: "this-is-a-very-long-command-name-that-exceeds-fifty-characters"
description: "Too long name"
trigger: ["test"]
---
content`,
			true,
			"invalid command name",
		},
		{
			"trigger too long",
			`---
name: "valid-name"
description: "Valid name"
trigger: ["this-trigger-is-too-long"]
---
content`,
			true,
			"invalid triggers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandFile := filepath.Join(tempDir, tt.name+".md")
			err := os.WriteFile(commandFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test command file: %v", err)
			}

			_, err = LoadCustomCommand(commandFile)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for test '%s', but got none", tt.name)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for test '%s', but got: %v", tt.name, err)
			}
			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', but got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
