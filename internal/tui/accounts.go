package tui

import (
	"fmt"
	"strings"

	"github.com/bcmister/qs/internal/config"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// AccountsModel is the account management TUI
type AccountsModel struct {
	cfg        *config.Config
	accounts   []config.Account
	cursor     int
	width      int
	height     int
	editing    bool
	deleting   bool
	adding     bool
	inputs     []textinput.Model
	inputIdx   int
	message    string
}

// NewAccounts creates a new account management model
func NewAccounts(cfg *config.Config) AccountsModel {
	// Copy accounts so we can edit them
	accounts := make([]config.Account, len(cfg.Accounts))
	for i, a := range cfg.Accounts {
		accounts[i] = config.Account{
			ID:      a.ID,
			Label:   a.Label,
			Command: a.Command,
			Args:    append([]string{}, a.Args...),
			Icon:    a.Icon,
			Enabled: a.Enabled,
		}
	}

	return AccountsModel{
		cfg:      cfg,
		accounts: accounts,
	}
}

func (m AccountsModel) Init() tea.Cmd {
	return nil
}

func (m AccountsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Handle sub-forms
		if m.editing || m.adding {
			return m.updateForm(msg)
		}
		if m.deleting {
			return m.updateDelete(msg)
		}

		switch {
		case key.Matches(msg, DefaultKeyMap.Escape):
			// Save and quit
			m.cfg.Accounts = m.accounts
			_ = config.Save(m.cfg, "")
			return m, tea.Quit

		case key.Matches(msg, DefaultKeyMap.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, DefaultKeyMap.Down):
			if m.cursor < len(m.accounts)-1 {
				m.cursor++
			}

		case key.Matches(msg, DefaultKeyMap.Space):
			// Toggle enabled — but don't allow disabling last enabled account
			if m.accounts[m.cursor].Enabled {
				enabledCount := len(config.EnabledAccounts(m.accounts))
				if enabledCount <= 1 {
					m.message = "Cannot disable last enabled account"
					return m, nil
				}
			}
			m.accounts[m.cursor].Enabled = !m.accounts[m.cursor].Enabled
			m.message = ""

		case msg.String() == "a":
			m.adding = true
			m.inputs = makeAccountFormInputs("", "", "", "")
			m.inputIdx = 0
			m.inputs[0].Focus()
			m.message = ""
			return m, textinput.Blink

		case msg.String() == "e":
			a := m.accounts[m.cursor]
			m.editing = true
			m.inputs = makeAccountFormInputs(a.Label, a.Command, strings.Join(a.Args, " "), a.Icon)
			m.inputIdx = 0
			m.inputs[0].Focus()
			m.message = ""
			return m, textinput.Blink

		case key.Matches(msg, DefaultKeyMap.Delete):
			if len(m.accounts) <= 1 {
				m.message = "Cannot delete last account"
				return m, nil
			}
			if m.accounts[m.cursor].Enabled {
				enabledCount := len(config.EnabledAccounts(m.accounts))
				if enabledCount <= 1 {
					m.message = "Cannot delete last enabled account"
					return m, nil
				}
			}
			m.deleting = true
			m.message = ""
		}
	}
	return m, nil
}

func (m AccountsModel) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, DefaultKeyMap.Escape):
		m.editing = false
		m.adding = false
		return m, nil

	case key.Matches(msg, DefaultKeyMap.Tab):
		m.inputs[m.inputIdx].Blur()
		m.inputIdx = (m.inputIdx + 1) % len(m.inputs)
		m.inputs[m.inputIdx].Focus()
		return m, textinput.Blink

	case key.Matches(msg, DefaultKeyMap.Enter):
		if m.inputIdx < len(m.inputs)-1 {
			m.inputs[m.inputIdx].Blur()
			m.inputIdx++
			m.inputs[m.inputIdx].Focus()
			return m, textinput.Blink
		}

		// Submit
		name := m.inputs[0].Value()
		command := m.inputs[1].Value()
		args := m.inputs[2].Value()
		icon := m.inputs[3].Value()

		if name == "" || command == "" {
			m.message = "Name and command are required"
			return m, nil
		}

		if icon == "" {
			icon = "⬜"
		}

		var argList []string
		if args != "" {
			argList = strings.Fields(args)
		}

		if m.editing {
			m.accounts[m.cursor].Label = name
			m.accounts[m.cursor].Command = command
			m.accounts[m.cursor].Args = argList
			m.accounts[m.cursor].Icon = icon
			m.editing = false
		} else {
			id := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
			m.accounts = append(m.accounts, config.Account{
				ID:      id,
				Label:   name,
				Command: command,
				Args:    argList,
				Icon:    icon,
				Enabled: true,
			})
			m.cursor = len(m.accounts) - 1
			m.adding = false
		}
		m.message = ""
		return m, nil

	default:
		var cmd tea.Cmd
		m.inputs[m.inputIdx], cmd = m.inputs[m.inputIdx].Update(msg)
		return m, cmd
	}
}

func (m AccountsModel) updateDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.accounts = append(m.accounts[:m.cursor], m.accounts[m.cursor+1:]...)
		if m.cursor >= len(m.accounts) {
			m.cursor = len(m.accounts) - 1
		}
		m.deleting = false
		m.message = ""
	case "n", "N", "esc":
		m.deleting = false
		m.message = ""
	}
	return m, nil
}

func (m AccountsModel) View() string {
	var s strings.Builder

	s.WriteString(RenderLogo("accounts"))
	s.WriteString(RenderSep())
	s.WriteString("\n")

	if m.editing || m.adding {
		title := "Edit Account"
		if m.adding {
			title = "Add Account"
		}
		s.WriteString("  " + TitleStyle.Render(title) + "\n\n")

		labels := []string{"Name", "Command", "Args", "Icon"}
		for i, input := range m.inputs {
			active := i == m.inputIdx
			label := DimStyle.Render(fmt.Sprintf("  %-8s", labels[i]))
			if active {
				label = TitleStyle.Render(fmt.Sprintf("  %-8s", labels[i]))
			}
			s.WriteString(fmt.Sprintf("%s  %s\n", label, input.View()))
		}

		if m.message != "" {
			s.WriteString("\n  " + ErrorStyle.Render(m.message) + "\n")
		}

		s.WriteString("\n  " + DimStyle.Render("Tab next  Enter submit  Esc cancel") + "\n")
		return s.String()
	}

	if m.deleting {
		a := m.accounts[m.cursor]
		s.WriteString(fmt.Sprintf("  %s Delete %s %s? %s\n\n",
			WarningStyle.Render("⚠"),
			a.Icon, SubtitleStyle.Render(a.Label),
			DimStyle.Render("(y/n)")))
		return s.String()
	}

	// Main account list
	for i, a := range m.accounts {
		selected := i == m.cursor

		enabledMark := DimStyle.Render("[ ]")
		if a.Enabled {
			enabledMark = SuccessStyle.Render("[✓]")
		}

		prefix := "  "
		if selected {
			prefix = TitleStyle.Render("▸ ")
		}

		label := a.Icon + " " + a.Label
		if selected {
			label = SubtitleStyle.Render(a.Icon + " " + a.Label)
		} else if !a.Enabled {
			label = DimStyle.Render(a.Icon + " " + a.Label)
		} else {
			label = WhiteStyle.Render(a.Icon + " " + a.Label)
		}

		cmdStr := DimStyle.Render(truncate(a.FullCommand(), 40))

		s.WriteString(fmt.Sprintf("  %s %s %s  %s\n",
			prefix, enabledMark, label, cmdStr))
	}

	if m.message != "" {
		s.WriteString("\n  " + WarningStyle.Render(m.message) + "\n")
	}

	s.WriteString("\n  " + DimStyle.Render("Space toggle  a add  e edit  d delete  Esc save & quit") + "\n")
	return s.String()
}

func makeAccountFormInputs(name, command, args, icon string) []textinput.Model {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Tool Name"
	inputs[0].CharLimit = 32
	inputs[0].Width = 30
	inputs[0].SetValue(name)

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "command-name"
	inputs[1].CharLimit = 64
	inputs[1].Width = 30
	inputs[1].SetValue(command)

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "--flag1 --flag2"
	inputs[2].CharLimit = 128
	inputs[2].Width = 30
	inputs[2].SetValue(args)

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "⬜"
	inputs[3].CharLimit = 4
	inputs[3].Width = 10
	inputs[3].SetValue(icon)

	return inputs
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
