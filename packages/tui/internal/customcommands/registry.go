package customcommands

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/sst/opencode/internal/commands"
)

// CommandCache holds cached command data with modification time tracking
type CommandCache struct {
	commands     []CustomCommand
	errors       []CommandLoadError
	lastScan     time.Time
	fileModTimes map[string]time.Time
	mutex        sync.RWMutex
}

// Registry manages custom commands and integrates them with the existing command system
type Registry struct {
	scanner         *Scanner
	commandRegistry commands.CommandRegistry
	cache           *CommandCache
	initialized     bool
	mutex           sync.RWMutex
}

// NewRegistry creates a new custom command registry
// NewRegistryWithProjectRoot creates a new custom command registry using the specified project root
func NewRegistry(commandRegistry commands.CommandRegistry, projectRoot string) *Registry {
	scanner := NewScannerWithProjectRoot(projectRoot)
	return &Registry{
		scanner:         scanner,
		commandRegistry: commandRegistry,
		cache: &CommandCache{
			fileModTimes: make(map[string]time.Time),
		},
		initialized: false,
	}
}

// NewRegistryWithScanner creates a new custom command registry with a specific scanner
func NewRegistryWithScanner(commandRegistry commands.CommandRegistry, scanner *Scanner) *Registry {
	return &Registry{
		scanner:         scanner,
		commandRegistry: commandRegistry,
		cache: &CommandCache{
			fileModTimes: make(map[string]time.Time),
		},
		initialized: false,
	}
}

// ensureLoaded ensures commands are loaded (lazy loading)
func (r *Registry) ensureLoaded() {
	r.mutex.RLock()
	if r.initialized && !r.needsRefresh() {
		r.mutex.RUnlock()
		return
	}
	r.mutex.RUnlock()

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Double-check after acquiring write lock
	if r.initialized && !r.needsRefresh() {
		return
	}

	r.loadCustomCommands()
	r.initialized = true
}

// needsRefresh checks if cache needs refreshing based on file modification times
func (r *Registry) needsRefresh() bool {
	if r.cache == nil {
		return true
	}

	// Check if any command directories have been modified
	for _, dir := range r.scanner.GetDirectories() {
		if info, err := os.Stat(dir); err == nil {
			if info.ModTime().After(r.cache.lastScan) {
				return true
			}
		}
	}

	return false
}

// validateAgainstBuiltinCommands checks for conflicts with built-in commands
func (r *Registry) validateAgainstBuiltinCommands(cmd *CustomCommand) error {
	// Check if any trigger conflicts with built-in commands
	for _, trigger := range cmd.Trigger {
		for _, builtinCmd := range r.commandRegistry.Sorted() {
			if builtinCmd.HasTrigger() {
				for _, builtinTrigger := range builtinCmd.Trigger {
					if trigger == builtinTrigger {
						return fmt.Errorf("trigger '%s' conflicts with built-in command", trigger)
					}
				}
			}
		}
	}
	return nil
}

// loadCustomCommands loads custom commands from the scanner
func (r *Registry) loadCustomCommands() []CommandLoadError {
	commands, errors, newModTimes := r.scanner.ScanCommandsWithCache(r.cache.fileModTimes)

	// Validate commands against built-ins and filter out invalid ones
	var validCommands []CustomCommand
	for _, cmd := range commands {
		if err := r.validateAgainstBuiltinCommands(&cmd); err != nil {
			errors = append(errors, CommandLoadError{
				FilePath: cmd.FilePath,
				Error:    err,
			})
			continue
		}
		validCommands = append(validCommands, cmd)
	}

	// Update cache
	r.cache.mutex.Lock()
	r.cache.commands = validCommands
	r.cache.errors = errors
	r.cache.lastScan = time.Now()
	r.cache.fileModTimes = newModTimes
	r.cache.mutex.Unlock()

	if len(errors) > 0 {
		slog.Warn("Some custom commands failed to load", "errorCount", len(errors))
		for _, err := range errors {
			slog.Warn("Failed to load command", "file", err.FilePath, "error", err.Error)
		}
	} else {
		slog.Info("Loaded custom commands", "count", len(validCommands))
	}

	return errors
}

// RefreshCommands reloads custom commands from disk
func (r *Registry) RefreshCommands() {
	r.mutex.Lock()
	r.initialized = false
	r.mutex.Unlock()
	r.ensureLoaded()
}

// GetCustomCommands returns all loaded custom commands
func (r *Registry) GetCustomCommands() []CustomCommand {
	r.ensureLoaded()
	r.cache.mutex.RLock()
	defer r.cache.mutex.RUnlock()
	return append([]CustomCommand(nil), r.cache.commands...) // Return a copy
}

// FindCustomCommandByTrigger finds a custom command by its trigger
func (r *Registry) FindCustomCommandByTrigger(trigger string) *CustomCommand {
	r.ensureLoaded()
	r.cache.mutex.RLock()
	defer r.cache.mutex.RUnlock()

	for _, cmd := range r.cache.commands {
		if cmd.MatchesTrigger(trigger) {
			return &cmd
		}
	}
	return nil
}

// FindCustomCommandByName finds a custom command by its name
func (r *Registry) FindCustomCommandByName(name string) *CustomCommand {
	r.ensureLoaded()
	r.cache.mutex.RLock()
	defer r.cache.mutex.RUnlock()

	for _, cmd := range r.cache.commands {
		if cmd.Name == name {
			return &cmd
		}
	}
	return nil
}

// GetAllTriggers returns all triggers from both built-in and custom commands
func (r *Registry) GetAllTriggers() []string {
	r.ensureLoaded()
	r.cache.mutex.RLock()
	defer r.cache.mutex.RUnlock()

	var triggers []string

	// Add built-in command triggers
	for _, cmd := range r.commandRegistry.Sorted() {
		if cmd.HasTrigger() {
			triggers = append(triggers, cmd.Trigger...)
		}
	}

	// Add custom command triggers
	for _, cmd := range r.cache.commands {
		if cmd.HasTrigger() {
			triggers = append(triggers, cmd.Trigger...)
		}
	}

	return triggers
}

// IsCustomCommand checks if a trigger corresponds to a custom command
func (r *Registry) IsCustomCommand(trigger string) bool {
	return r.FindCustomCommandByTrigger(trigger) != nil
}

// ExecuteCustomCommand executes a custom command with the given arguments
func (r *Registry) ExecuteCustomCommand(trigger, args string) (string, error) {
	cmd := r.FindCustomCommandByTrigger(trigger)
	if cmd == nil {
		return "", nil // Return empty string for non-existent commands, not error
	}

	result, err := cmd.Execute(args)
	if err != nil {
		return "", fmt.Errorf("failed to execute custom command '%s': %w", cmd.Name, err)
	}

	return result, nil
}

// ExecuteCustomCommandWithContext executes a custom command and returns execution context
type CustomCommandResult struct {
	Output      string
	CommandName string
}

func (r *Registry) ExecuteCustomCommandWithContext(trigger, args string) (*CustomCommandResult, error) {
	cmd := r.FindCustomCommandByTrigger(trigger)
	if cmd == nil {
		return nil, nil // Return nil for non-existent commands, not error
	}

	result, err := cmd.Execute(args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute custom command '%s': %w", cmd.Name, err)
	}

	return &CustomCommandResult{
		Output:      result,
		CommandName: cmd.Name,
	}, nil
}

// GetSortedCustomCommands returns custom commands sorted by name
func (r *Registry) GetSortedCustomCommands() []CustomCommand {
	r.ensureLoaded()
	r.cache.mutex.RLock()
	defer r.cache.mutex.RUnlock()

	commands := append([]CustomCommand(nil), r.cache.commands...)
	slices.SortFunc(commands, func(a, b CustomCommand) int {
		return strings.Compare(a.Name, b.Name)
	})
	return commands
}

// GetCommandsWithTrigger returns only custom commands that have triggers
func (r *Registry) GetCommandsWithTrigger() []CustomCommand {
	r.ensureLoaded()
	r.cache.mutex.RLock()
	defer r.cache.mutex.RUnlock()

	var commands []CustomCommand
	for _, cmd := range r.cache.commands {
		if cmd.HasTrigger() {
			commands = append(commands, cmd)
		}
	}
	return commands
}

// AddDirectory adds a directory to scan for custom commands
func (r *Registry) AddDirectory(dir string) {
	r.scanner.AddDirectory(dir)
	r.RefreshCommands() // Reload commands after adding directory
}

// GetScannedDirectories returns the directories being scanned for commands
func (r *Registry) GetScannedDirectories() []string {
	return r.scanner.GetDirectories()
}
