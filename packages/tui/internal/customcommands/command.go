package customcommands

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// CommandLoadError represents an error that occurred while loading a custom command
type CommandLoadError struct {
	FilePath string
	Error    error
}

// CustomCommand represents a custom slash command loaded from markdown files
type CustomCommand struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Trigger     []string `yaml:"trigger"`
	Content     string   `yaml:"-"` // The markdown content after frontmatter
	FilePath    string   `yaml:"-"` // Path to the source file
}

// CommandMetadata represents the YAML frontmatter in command files
type CommandMetadata struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Trigger     []string `yaml:"trigger"`
}

// HasTrigger returns true if the command has at least one trigger
func (c CustomCommand) HasTrigger() bool {
	return len(c.Trigger) > 0
}

// PrimaryTrigger returns the first trigger or empty string if none
func (c CustomCommand) PrimaryTrigger() string {
	if len(c.Trigger) > 0 {
		return c.Trigger[0]
	}
	return ""
}

// MatchesTrigger returns true if the given trigger matches any of the command's triggers
func (c CustomCommand) MatchesTrigger(trigger string) bool {
	for _, t := range c.Trigger {
		if t == trigger {
			return true
		}
	}
	return false
}

// resolvePath resolves a file path with support for home directory expansion and relative paths
func (c CustomCommand) resolvePath(filePath string) string {
	// Handle empty path
	if filePath == "" {
		return filePath
	}

	// Handle home directory expansion
	if strings.HasPrefix(filePath, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			if filePath == "~" {
				filePath = home
			} else if strings.HasPrefix(filePath, "~/") {
				filePath = filepath.Join(home, filePath[2:])
			}
		}
	}

	// Handle environment variable expansion
	filePath = os.ExpandEnv(filePath)

	// Resolve relative paths relative to the command file's directory
	if !filepath.IsAbs(filePath) {
		commandDir := filepath.Dir(c.FilePath)
		filePath = filepath.Join(commandDir, filePath)
	}

	return filePath
}

// Execute processes the command content and executes it
func (c CustomCommand) Execute(args string) (string, error) {
	content := c.Content

	// Replace $ARGUMENTS placeholder with sanitized arguments
	sanitizedArgs := sanitizeArguments(args)
	content = strings.ReplaceAll(content, "$ARGUMENTS", sanitizedArgs)

	// Expand environment variables in the content
	content = os.ExpandEnv(content)

	// Process the content line by line
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle bash execution with ! prefix
		if strings.HasPrefix(line, "!") {
			bashCmd := strings.TrimPrefix(line, "!")
			bashCmd = strings.TrimSpace(bashCmd)

			output, err := executeBashCommand(bashCmd)
			if err != nil {
				return "", fmt.Errorf("failed to execute bash command '%s': %w", bashCmd, err)
			}
			result.WriteString(output)
			continue
		}

		// Handle file references with @ prefix
		if strings.HasPrefix(line, "@") {
			filePath := strings.TrimPrefix(line, "@")
			filePath = strings.TrimSpace(filePath)

			// Resolve relative paths relative to the command file's directory
			if !filepath.IsAbs(filePath) {
				commandDir := filepath.Dir(c.FilePath)
				filePath = filepath.Join(commandDir, filePath)
			}

			content, err := readFileContent(filePath)
			if err != nil {
				return "", fmt.Errorf("failed to read file '%s': %w", filePath, err)
			}
			result.WriteString(content)
			result.WriteString("\n")
			continue
		}

		// Regular content line
		result.WriteString(line)
		result.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading command content: %w", err)
	}

	return result.String(), nil
}

// validateBashCommand validates a bash command for security
func validateBashCommand(command string) error {
	// Trim whitespace
	command = strings.TrimSpace(command)

	// Check for empty command
	if command == "" {
		return fmt.Errorf("empty command not allowed")
	}

	commandLower := strings.ToLower(command)

	// Check for dangerous command patterns but allow git command substitution
	if strings.Contains(commandLower, "rm -rf /") ||
		strings.Contains(commandLower, "rm -rf ~") ||
		strings.Contains(commandLower, "rm -rf *") {
		return fmt.Errorf("potentially dangerous rm command detected")
	}

	// Block dangerous patterns - but allow legitimate git command substitution
	dangerousPatterns := []string{
		"rm -rf", "sudo ", " su ", "sudo\t", "\tsu\t", // More precise sudo/su matching
		"chmod +x", "wget ", "curl ", " nc ", " netcat ", // More precise tool matching
		">/dev/sd", ">/dev/hd", "</dev/sd", "</dev/hd", // Block device access but allow /dev/null
		">/proc/", "</proc/", // Block /proc access
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(commandLower, pattern) {
			return fmt.Errorf("potentially dangerous command pattern detected: %s", pattern)
		}
	}

	return nil
}

// sanitizeArguments sanitizes user-provided arguments
func sanitizeArguments(args string) string {
	// Remove potentially dangerous characters
	args = strings.ReplaceAll(args, ";", "")
	args = strings.ReplaceAll(args, "&", "")
	args = strings.ReplaceAll(args, "|", "")
	args = strings.ReplaceAll(args, "`", "")
	args = strings.ReplaceAll(args, "$", "")
	args = strings.ReplaceAll(args, ">", "")
	args = strings.ReplaceAll(args, "<", "")

	// Trim whitespace
	return strings.TrimSpace(args)
}

// executeBashCommand executes a bash command and returns its output
func executeBashCommand(command string) (string, error) {
	// Validate the bash command before execution
	if err := validateBashCommand(command); err != nil {
		return "", fmt.Errorf("command validation failed: %w", err)
	}

	cmd := exec.Command("bash", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return "", fmt.Errorf("command failed: %s", stderr.String())
		}
		return "", err
	}

	return stdout.String(), nil
}

// readFileContent reads the content of a file
func readFileContent(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// validateCommandName validates a command name follows naming conventions
func validateCommandName(name string) error {
	if name == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	// Check length
	if len(name) > 50 {
		return fmt.Errorf("command name too long (max 50 characters): %s", name)
	}

	// Check for valid characters (alphanumeric, hyphens, underscores)
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("command name contains invalid characters (only alphanumeric, hyphens, underscores allowed): %s", name)
	}

	// Check doesn't start with number or special character
	if !regexp.MustCompile(`^[a-zA-Z]`).MatchString(name) {
		return fmt.Errorf("command name must start with a letter: %s", name)
	}

	return nil
}

// validateTriggers validates command triggers
func validateTriggers(triggers []string) error {
	if len(triggers) == 0 {
		return fmt.Errorf("at least one trigger is required")
	}

	for _, trigger := range triggers {
		if trigger == "" {
			return fmt.Errorf("trigger cannot be empty")
		}

		if len(trigger) > 20 {
			return fmt.Errorf("trigger too long (max 20 characters): %s", trigger)
		}

		// Same naming rules as command names
		if err := validateCommandName(trigger); err != nil {
			return fmt.Errorf("invalid trigger '%s': %w", trigger, err)
		}
	}

	return nil
}

// parseMarkdownWithFrontmatter parses a markdown file with YAML frontmatter
func parseMarkdownWithFrontmatter(content []byte) (*CommandMetadata, string, error) {
	contentStr := string(content)

	// Check if the file starts with YAML frontmatter
	if !strings.HasPrefix(contentStr, "---\n") {
		return nil, "", fmt.Errorf("file does not start with YAML frontmatter")
	}

	// Find the end of the frontmatter
	lines := strings.Split(contentStr, "\n")
	var frontmatterEnd int
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			frontmatterEnd = i
			break
		}
	}

	if frontmatterEnd == 0 {
		return nil, "", fmt.Errorf("could not find end of YAML frontmatter")
	}

	// Extract frontmatter and content
	frontmatterLines := lines[1:frontmatterEnd]
	contentLines := lines[frontmatterEnd+1:]

	frontmatterStr := strings.Join(frontmatterLines, "\n")
	markdownContent := strings.Join(contentLines, "\n")

	// Parse YAML frontmatter
	var metadata CommandMetadata
	if err := yaml.Unmarshal([]byte(frontmatterStr), &metadata); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	return &metadata, markdownContent, nil
}

// LoadCustomCommand loads a custom command from a markdown file
func LoadCustomCommand(filePath string) (*CustomCommand, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	metadata, markdownContent, err := parseMarkdownWithFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}

	// Validate required fields
	if metadata.Name == "" {
		return nil, fmt.Errorf("command name is required in %s", filePath)
	}

	if len(metadata.Trigger) == 0 {
		return nil, fmt.Errorf("at least one trigger is required in %s", filePath)
	}

	// Add validation after parsing metadata but before creating command
	if err := validateCommandName(metadata.Name); err != nil {
		return nil, fmt.Errorf("invalid command name in %s: %w", filePath, err)
	}

	if err := validateTriggers(metadata.Trigger); err != nil {
		return nil, fmt.Errorf("invalid triggers in %s: %w", filePath, err)
	}

	command := &CustomCommand{
		Name:        metadata.Name,
		Description: metadata.Description,
		Trigger:     metadata.Trigger,
		Content:     markdownContent,
		FilePath:    filePath,
	}

	return command, nil
}
