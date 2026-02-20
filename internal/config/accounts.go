package config

import "strings"

// Account represents a configured AI tool account
type Account struct {
	ID      string   `yaml:"id"`
	Label   string   `yaml:"label"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	Icon    string   `yaml:"icon"`
	Enabled bool     `yaml:"enabled"`
}

// DefaultAccounts returns the built-in account definitions
var DefaultAccounts = []Account{
	{
		ID:      "claude",
		Label:   "Claude Code",
		Command: "claude",
		Args:    []string{"--dangerously-skip-permissions"},
		Icon:    "\U0001F7E0",
		Enabled: true,
	},
	{
		ID:      "codex",
		Label:   "OpenAI Codex",
		Command: "codex",
		Args:    []string{"--dangerously-bypass-approvals-and-sandbox"},
		Icon:    "\U0001F7E2",
		Enabled: true,
	},
	{
		ID:      "gemini",
		Label:   "Gemini CLI",
		Command: "gemini",
		Args:    []string{"--yolo"},
		Icon:    "\U0001F535",
		Enabled: true,
	},
	{
		ID:      "opencode",
		Label:   "OpenCode (z.ai)",
		Command: "opencode",
		Args:    []string{},
		Icon:    "\u26AB",
		Enabled: true,
	},
	{
		ID:      "cursor",
		Label:   "Cursor Agent",
		Command: "agent",
		Args:    []string{},
		Icon:    "\U0001F7E1",
		Enabled: true,
	},
	{
		ID:      "aider",
		Label:   "Aider",
		Command: "aider",
		Args:    []string{"--yes-always"},
		Icon:    "\U0001F7E3",
		Enabled: false,
	},
	{
		ID:      "continue",
		Label:   "Continue Dev",
		Command: "continue",
		Args:    []string{},
		Icon:    "\U0001F537",
		Enabled: false,
	},
}

// AccountByID finds an account by its ID, returns nil if not found
func AccountByID(accounts []Account, id string) *Account {
	for i := range accounts {
		if accounts[i].ID == id {
			return &accounts[i]
		}
	}
	return nil
}

// EnabledAccounts returns only accounts that are enabled
func EnabledAccounts(accounts []Account) []Account {
	var result []Account
	for _, a := range accounts {
		if a.Enabled {
			result = append(result, a)
		}
	}
	return result
}

// FullCommand returns the display string "command args..." for an account
func (a *Account) FullCommand() string {
	if len(a.Args) == 0 {
		return a.Command
	}
	return a.Command + " " + strings.Join(a.Args, " ")
}
