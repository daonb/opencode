package completions

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/commands"
	"github.com/sst/opencode/internal/customcommands"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
)

type CommandCompletionProvider struct {
	app *app.App
}

func NewCommandCompletionProvider(app *app.App) CompletionProvider {
	return &CommandCompletionProvider{app: app}
}

func (c *CommandCompletionProvider) GetId() string {
	return "commands"
}

func (c *CommandCompletionProvider) GetEmptyMessage() string {
	return "no matching commands"
}

func (c *CommandCompletionProvider) getCommandCompletionItem(
	cmd commands.Command,
	space int,
) CompletionSuggestion {
	displayFunc := func(s styles.Style) string {
		t := theme.CurrentTheme()
		spacer := strings.Repeat(" ", space)
		display := "  /" + cmd.PrimaryTrigger() + s.
			Foreground(t.TextMuted()).
			Render(spacer+cmd.Description)
		return display
	}

	value := string(cmd.Name)
	return CompletionSuggestion{
		Display:    displayFunc,
		Value:      value,
		ProviderID: c.GetId(),
		RawData:    cmd,
	}
}

func (c *CommandCompletionProvider) getCustomCommandCompletionItem(
	cmd customcommands.CustomCommand,
	space int,
) CompletionSuggestion {
	displayFunc := func(s styles.Style) string {
		t := theme.CurrentTheme()
		spacer := strings.Repeat(" ", space)
		display := "  /" + cmd.PrimaryTrigger() + s.
			Foreground(t.TextMuted()).
			Render(spacer+cmd.Description)
		return display
	}

	value := cmd.Name
	return CompletionSuggestion{
		Display:    displayFunc,
		Value:      value,
		ProviderID: c.GetId(),
		RawData:    cmd,
	}
}

func (c *CommandCompletionProvider) GetChildEntries(
	query string,
) ([]CompletionSuggestion, error) {
	commands := c.app.Commands

	// Get custom commands from the app
	customCommands := c.app.CustomCommandRegistry.GetCommandsWithTrigger()

	// Calculate spacing for alignment - check both built-in and custom commands
	space := 1
	for _, cmd := range c.app.Commands {
		if cmd.HasTrigger() && lipgloss.Width(cmd.PrimaryTrigger()) > space {
			space = lipgloss.Width(cmd.PrimaryTrigger())
		}
	}
	for _, cmd := range customCommands {
		if cmd.HasTrigger() && lipgloss.Width(cmd.PrimaryTrigger()) > space {
			space = lipgloss.Width(cmd.PrimaryTrigger())
		}
	}
	space += 2

	sorted := commands.Sorted()
	if query == "" {
		// If no query, return all commands (built-in + custom)
		items := []CompletionSuggestion{}

		// Add built-in commands
		for _, cmd := range sorted {
			if !cmd.HasTrigger() {
				continue
			}
			itemSpace := space - lipgloss.Width(cmd.PrimaryTrigger())
			items = append(items, c.getCommandCompletionItem(cmd, itemSpace))
		}

		// Add custom commands
		for _, cmd := range customCommands {
			if !cmd.HasTrigger() {
				continue
			}
			itemSpace := space - lipgloss.Width(cmd.PrimaryTrigger())
			items = append(items, c.getCustomCommandCompletionItem(cmd, itemSpace))
		}

		return items, nil
	}

	var commandNames []string
	commandMap := make(map[string]CompletionSuggestion)

	// Add built-in commands to search
	for _, cmd := range sorted {
		if !cmd.HasTrigger() {
			continue
		}
		itemSpace := space - lipgloss.Width(cmd.PrimaryTrigger())
		for _, trigger := range cmd.Trigger {
			commandNames = append(commandNames, trigger)
			commandMap[trigger] = c.getCommandCompletionItem(cmd, itemSpace)
		}
	}

	// Add custom commands to search
	for _, cmd := range customCommands {
		if !cmd.HasTrigger() {
			continue
		}
		itemSpace := space - lipgloss.Width(cmd.PrimaryTrigger())
		for _, trigger := range cmd.Trigger {
			commandNames = append(commandNames, trigger)
			commandMap[trigger] = c.getCustomCommandCompletionItem(cmd, itemSpace)
		}
	}

	matches := fuzzy.RankFindFold(query, commandNames)
	sort.Sort(matches)

	// Convert matches to completion items, deduplicating by command name
	items := []CompletionSuggestion{}
	seen := make(map[string]bool)
	for _, match := range matches {
		if item, ok := commandMap[match.Target]; ok {
			// Use the command's value (name) as the deduplication key
			if !seen[item.Value] {
				seen[item.Value] = true
				items = append(items, item)
			}
		}
	}
	return items, nil
}
