package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bcmister/qs/internal/config"
	"github.com/bcmister/qs/internal/monitor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type setupStep int

const (
	stepWelcome setupStep = iota
	stepProjectsRoot
	stepMonitors
	stepAccounts
	stepConfirm
	stepDone
)

// SetupModel is the multi-step setup wizard
type SetupModel struct {
	existingCfg  *config.Config
	step         setupStep
	width        int
	height       int
	err          error

	// Step 1: Projects root
	rootInput    textinput.Model
	projectsRoot string

	// Step 2: Monitors
	monitors     []monitor.Monitor
	monitorIdx   int
	windowCounts []int

	// Step 3: Accounts
	accounts     []config.Account
	accountIdx   int

	// Step 3 sub-form: add custom account
	addingAccount  bool
	addInputs      []textinput.Model
	addInputIdx    int

	// Final config
	savedPath    string
}

// NewSetup creates a new setup wizard model
func NewSetup(existingCfg *config.Config) SetupModel {
	// Projects root input
	rootInput := textinput.New()
	rootInput.Placeholder = config.DefaultProjectsRoot()
	rootInput.CharLimit = 256
	rootInput.Width = 50

	defaultRoot := config.DefaultProjectsRoot()
	if existingCfg != nil && existingCfg.ProjectsRoot != "" {
		defaultRoot = existingCfg.ProjectsRoot
	}
	rootInput.SetValue(defaultRoot)

	// Copy default accounts
	accounts := make([]config.Account, len(config.DefaultAccounts))
	for i, a := range config.DefaultAccounts {
		accounts[i] = config.Account{
			ID:      a.ID,
			Label:   a.Label,
			Command: a.Command,
			Args:    append([]string{}, a.Args...),
			Icon:    a.Icon,
			Enabled: a.Enabled,
		}
	}

	// If existing config has accounts, use those instead
	if existingCfg != nil && len(existingCfg.Accounts) > 0 {
		accounts = make([]config.Account, len(existingCfg.Accounts))
		for i, a := range existingCfg.Accounts {
			accounts[i] = config.Account{
				ID:      a.ID,
				Label:   a.Label,
				Command: a.Command,
				Args:    append([]string{}, a.Args...),
				Icon:    a.Icon,
				Enabled: a.Enabled,
			}
		}
	}

	return SetupModel{
		existingCfg: existingCfg,
		step:        stepWelcome,
		rootInput:   rootInput,
		accounts:    accounts,
	}
}

func (m SetupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case monitorsDetectedMsg:
		m.monitors = msg.monitors
		m.windowCounts = make([]int, len(m.monitors))
		for i := range m.windowCounts {
			m.windowCounts[i] = 1
			// Use existing config if available
			if m.existingCfg != nil && i < len(m.existingCfg.Monitors) {
				count := m.existingCfg.Monitors[i].WindowCount()
				if count >= 1 {
					m.windowCounts[i] = count
				}
			}
		}
		return m, nil

	case tea.KeyMsg:
		// Handle add account sub-form
		if m.addingAccount {
			return m.updateAddAccount(msg)
		}

		switch m.step {
		case stepWelcome:
			return m.updateWelcome(msg)
		case stepProjectsRoot:
			return m.updateProjectsRoot(msg)
		case stepMonitors:
			return m.updateMonitors(msg)
		case stepAccounts:
			return m.updateAccounts(msg)
		case stepConfirm:
			return m.updateConfirm(msg)
		case stepDone:
			return m, tea.Quit
		}
	}

	// Pass to text input if active
	if m.step == stepProjectsRoot {
		var cmd tea.Cmd
		m.rootInput, cmd = m.rootInput.Update(msg)
		return m, cmd
	}

	if m.addingAccount {
		var cmd tea.Cmd
		m.addInputs[m.addInputIdx], cmd = m.addInputs[m.addInputIdx].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m SetupModel) updateWelcome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, DefaultKeyMap.Enter):
		m.step = stepProjectsRoot
		m.rootInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, DefaultKeyMap.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m SetupModel) updateProjectsRoot(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, DefaultKeyMap.Enter):
		m.projectsRoot = m.rootInput.Value()
		if m.projectsRoot == "" {
			m.projectsRoot = config.DefaultProjectsRoot()
		}

		// Create directory if it doesn't exist
		if _, err := os.Stat(m.projectsRoot); os.IsNotExist(err) {
			os.MkdirAll(m.projectsRoot, 0755)
		}

		// Detect monitors and move to that step
		m.step = stepMonitors
		return m, detectMonitors
	case key.Matches(msg, DefaultKeyMap.Escape):
		m.step = stepWelcome
		m.rootInput.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		m.rootInput, cmd = m.rootInput.Update(msg)
		return m, cmd
	}
}

func (m SetupModel) updateMonitors(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, DefaultKeyMap.Enter):
		m.step = stepAccounts
		return m, nil
	case key.Matches(msg, DefaultKeyMap.Escape):
		m.step = stepProjectsRoot
		m.rootInput.Focus()
		return m, textinput.Blink
	case msg.String() == "left", msg.String() == "h":
		if m.monitorIdx > 0 {
			m.monitorIdx--
		}
	case msg.String() == "right", msg.String() == "l":
		if m.monitorIdx < len(m.monitors)-1 {
			m.monitorIdx++
		}
	case key.Matches(msg, DefaultKeyMap.Up):
		if len(m.monitors) > 0 && m.windowCounts[m.monitorIdx] < 9 {
			m.windowCounts[m.monitorIdx]++
		}
	case key.Matches(msg, DefaultKeyMap.Down):
		if len(m.monitors) > 0 && m.windowCounts[m.monitorIdx] > 1 {
			m.windowCounts[m.monitorIdx]--
		}
	}
	return m, nil
}

func (m SetupModel) updateAccounts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, DefaultKeyMap.Enter):
		// Check at least one account enabled
		enabled := config.EnabledAccounts(m.accounts)
		if len(enabled) == 0 {
			return m, nil
		}
		m.step = stepConfirm
		return m, nil
	case key.Matches(msg, DefaultKeyMap.Escape):
		m.step = stepMonitors
		return m, nil
	case key.Matches(msg, DefaultKeyMap.Up):
		if m.accountIdx > 0 {
			m.accountIdx--
		}
	case key.Matches(msg, DefaultKeyMap.Down):
		if m.accountIdx < len(m.accounts)-1 {
			m.accountIdx++
		}
	case key.Matches(msg, DefaultKeyMap.Space):
		m.accounts[m.accountIdx].Enabled = !m.accounts[m.accountIdx].Enabled
	case msg.String() == "a":
		m.addingAccount = true
		m.addInputs = makeAddAccountInputs()
		m.addInputIdx = 0
		m.addInputs[0].Focus()
		return m, textinput.Blink
	}
	return m, nil
}

func (m SetupModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, DefaultKeyMap.Enter):
		// Build and save config
		cfg := m.buildConfig()
		path := config.DefaultConfigPath()
		if err := config.Save(cfg, path); err != nil {
			m.err = err
			return m, nil
		}
		m.savedPath = path
		m.step = stepDone
		return m, nil
	case key.Matches(msg, DefaultKeyMap.Escape):
		m.step = stepAccounts
		return m, nil
	}
	return m, nil
}

func (m SetupModel) updateAddAccount(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, DefaultKeyMap.Escape):
		m.addingAccount = false
		return m, nil
	case key.Matches(msg, DefaultKeyMap.Tab):
		m.addInputs[m.addInputIdx].Blur()
		m.addInputIdx = (m.addInputIdx + 1) % len(m.addInputs)
		m.addInputs[m.addInputIdx].Focus()
		return m, textinput.Blink
	case key.Matches(msg, DefaultKeyMap.Enter):
		if m.addInputIdx < len(m.addInputs)-1 {
			// Move to next field
			m.addInputs[m.addInputIdx].Blur()
			m.addInputIdx++
			m.addInputs[m.addInputIdx].Focus()
			return m, textinput.Blink
		}
		// Submit: create account
		name := m.addInputs[0].Value()
		command := m.addInputs[1].Value()
		args := m.addInputs[2].Value()
		icon := m.addInputs[3].Value()

		if name == "" || command == "" {
			return m, nil
		}

		id := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		if icon == "" {
			icon = "⬜"
		}

		var argList []string
		if args != "" {
			argList = strings.Fields(args)
		}

		m.accounts = append(m.accounts, config.Account{
			ID:      id,
			Label:   name,
			Command: command,
			Args:    argList,
			Icon:    icon,
			Enabled: true,
		})
		m.addingAccount = false
		m.accountIdx = len(m.accounts) - 1
		return m, nil
	default:
		var cmd tea.Cmd
		m.addInputs[m.addInputIdx], cmd = m.addInputs[m.addInputIdx].Update(msg)
		return m, cmd
	}
}

func (m SetupModel) View() string {
	var s strings.Builder

	s.WriteString(RenderLogo("setup"))

	switch m.step {
	case stepWelcome:
		s.WriteString(m.viewWelcome())
	case stepProjectsRoot:
		s.WriteString(m.viewProjectsRoot())
	case stepMonitors:
		s.WriteString(m.viewMonitors())
	case stepAccounts:
		if m.addingAccount {
			s.WriteString(m.viewAddAccount())
		} else {
			s.WriteString(m.viewAccounts())
		}
	case stepConfirm:
		s.WriteString(m.viewConfirm())
	case stepDone:
		s.WriteString(m.viewDone())
	}

	return s.String()
}

func (m SetupModel) viewWelcome() string {
	var s strings.Builder
	s.WriteString(RenderSep())
	s.WriteString("\n")
	s.WriteString("  " + SubtitleStyle.Render("Welcome to quickstart!") + "\n\n")
	s.WriteString("  " + DimStyle.Render("This wizard will configure your terminal launcher.") + "\n")
	s.WriteString("  " + DimStyle.Render("You'll set up:") + "\n\n")
	s.WriteString("    " + TitleStyle.Render("1.") + " " + WhiteStyle.Render("Projects folder") + "\n")
	s.WriteString("    " + TitleStyle.Render("2.") + " " + WhiteStyle.Render("Monitor layout") + "\n")
	s.WriteString("    " + TitleStyle.Render("3.") + " " + WhiteStyle.Render("AI tool accounts") + "\n")
	s.WriteString("\n")
	s.WriteString("  " + DimStyle.Render("Press Enter to begin, q to quit") + "\n")
	return s.String()
}

func (m SetupModel) viewProjectsRoot() string {
	var s strings.Builder
	s.WriteString(RenderSep())
	s.WriteString("\n")
	s.WriteString("  " + TitleStyle.Render("Step 1") + " " + SubtitleStyle.Render("Projects Root") + "\n\n")
	s.WriteString("  " + DimStyle.Render("Directory containing your project folders:") + "\n\n")
	s.WriteString("  " + m.rootInput.View() + "\n\n")

	// Check if directory exists
	if m.rootInput.Value() != "" {
		if _, err := os.Stat(m.rootInput.Value()); os.IsNotExist(err) {
			s.WriteString("  " + WarningStyle.Render("Directory does not exist — will be created") + "\n")
		} else {
			s.WriteString("  " + SuccessStyle.Render("✓ Directory exists") + "\n")
		}
	}

	s.WriteString("\n  " + DimStyle.Render("Enter to continue, Esc to go back") + "\n")
	return s.String()
}

func (m SetupModel) viewMonitors() string {
	var s strings.Builder
	s.WriteString(RenderSep())
	s.WriteString("\n")
	s.WriteString("  " + TitleStyle.Render("Step 2") + " " + SubtitleStyle.Render("Monitor Layout") + "\n\n")

	if len(m.monitors) == 0 {
		s.WriteString("  " + DimStyle.Render("Detecting monitors...") + "\n")
		return s.String()
	}

	s.WriteString(fmt.Sprintf("  %s %d monitors detected\n\n",
		SuccessStyle.Render("✓"),
		len(m.monitors)))

	for i, mon := range m.monitors {
		selected := i == m.monitorIdx

		label := fmt.Sprintf("Monitor %d", i+1)
		if mon.Primary {
			label += " (Primary)"
		}

		res := fmt.Sprintf("%d×%d", mon.Width, mon.Height)
		windows := m.windowCounts[i]
		layout := layoutForCount(windows)

		var panel string
		if selected {
			panel = fmt.Sprintf("  %s %s  %s  %s windows  %s\n",
				TitleStyle.Render("▸"),
				SubtitleStyle.Render(label),
				DimStyle.Render(res),
				TitleStyle.Render(strconv.Itoa(windows)),
				DimStyle.Render("("+layout+")"))
		} else {
			panel = fmt.Sprintf("    %s  %s  %d windows  %s\n",
				DimStyle.Render(label),
				DimStyle.Render(res),
				windows,
				DimStyle.Render("("+layout+")"))
		}
		s.WriteString(panel)

		// Draw ASCII layout preview for selected monitor
		if selected {
			s.WriteString(renderLayoutPreview(windows))
			s.WriteString("\n")
		}
	}

	s.WriteString("\n  " + DimStyle.Render("←→ select monitor  ↑↓ adjust windows  Enter to continue") + "\n")
	return s.String()
}

func (m SetupModel) viewAccounts() string {
	var s strings.Builder
	s.WriteString(RenderSep())
	s.WriteString("\n")
	s.WriteString("  " + TitleStyle.Render("Step 3") + " " + SubtitleStyle.Render("AI Tool Accounts") + "\n\n")
	s.WriteString("  " + DimStyle.Render("Toggle accounts with Space, add custom with 'a'") + "\n\n")

	for i, a := range m.accounts {
		selected := i == m.accountIdx

		// Check if command is on PATH
		onPath := false
		if _, err := exec.LookPath(a.Command); err == nil {
			onPath = true
		}

		// Build status indicators
		enabledMark := DimStyle.Render("[ ]")
		if a.Enabled {
			enabledMark = SuccessStyle.Render("[✓]")
		}

		pathMark := ErrorStyle.Render("✗")
		if onPath {
			pathMark = SuccessStyle.Render("✓")
		}

		prefix := "  "
		if selected {
			prefix = TitleStyle.Render("▸ ")
		}

		label := a.Icon + " " + a.Label
		if selected {
			label = SubtitleStyle.Render(a.Icon + " " + a.Label)
		} else {
			label = DimStyle.Render(a.Icon + " " + a.Label)
		}

		s.WriteString(fmt.Sprintf("  %s %s %s  %s  %s\n",
			prefix, enabledMark, label, pathMark,
			DimStyle.Render(a.Command)))
	}

	enabledCount := len(config.EnabledAccounts(m.accounts))
	s.WriteString(fmt.Sprintf("\n  %s %d accounts enabled\n",
		DimStyle.Render("·"),
		enabledCount))

	s.WriteString("\n  " + DimStyle.Render("Space toggle  a add  Enter to continue  Esc back") + "\n")
	return s.String()
}

func (m SetupModel) viewAddAccount() string {
	var s strings.Builder
	s.WriteString(RenderSep())
	s.WriteString("\n")
	s.WriteString("  " + TitleStyle.Render("Add Account") + "\n\n")

	labels := []string{"Name", "Command", "Args", "Icon"}
	for i, input := range m.addInputs {
		active := i == m.addInputIdx
		label := DimStyle.Render(labels[i] + ":")
		if active {
			label = TitleStyle.Render(labels[i] + ":")
		}
		s.WriteString(fmt.Sprintf("  %s  %s\n", label, input.View()))
	}

	s.WriteString("\n  " + DimStyle.Render("Tab next field  Enter submit  Esc cancel") + "\n")
	return s.String()
}

func (m SetupModel) viewConfirm() string {
	var s strings.Builder
	s.WriteString(RenderSep())
	s.WriteString("\n")
	s.WriteString("  " + TitleStyle.Render("Step 4") + " " + SubtitleStyle.Render("Confirm") + "\n\n")

	// Projects root
	s.WriteString(fmt.Sprintf("  %s  %s\n",
		DimStyle.Render("Projects:"),
		WhiteStyle.Render(m.projectsRoot)))

	// Monitors
	for i, count := range m.windowCounts {
		layout := layoutForCount(count)
		s.WriteString(fmt.Sprintf("  %s  %d windows (%s)\n",
			DimStyle.Render(fmt.Sprintf("Monitor %d:", i+1)),
			count, layout))
	}

	// Enabled accounts
	s.WriteString("\n  " + DimStyle.Render("Accounts:") + "\n")
	for _, a := range m.accounts {
		if a.Enabled {
			s.WriteString(fmt.Sprintf("    %s %s  %s\n",
				a.Icon,
				WhiteStyle.Render(a.Label),
				DimStyle.Render(a.FullCommand())))
		}
	}

	if m.err != nil {
		s.WriteString("\n  " + ErrorStyle.Render("Error: "+m.err.Error()) + "\n")
	}

	s.WriteString("\n  " + DimStyle.Render("Enter to save, Esc to go back") + "\n")
	return s.String()
}

func (m SetupModel) viewDone() string {
	var s strings.Builder
	s.WriteString(RenderSep())
	s.WriteString("\n")
	s.WriteString("  " + SuccessStyle.Render("✓") + " " + SubtitleStyle.Render("Configuration saved!") + "\n\n")
	s.WriteString(fmt.Sprintf("  %s %s\n",
		DimStyle.Render("▸"),
		DimStyle.Render(m.savedPath)))
	s.WriteString("\n  " + DimStyle.Render("Run") + " " + TitleStyle.Render("qs") + " " + DimStyle.Render("to launch.") + "\n\n")
	return s.String()
}

func (m SetupModel) buildConfig() *config.Config {
	monitors := make([]config.MonitorConfig, len(m.windowCounts))
	for i, count := range m.windowCounts {
		windows := make([]config.WindowConfig, count)
		for j := range windows {
			windows[j] = config.WindowConfig{Tool: "claude"}
		}
		monitors[i] = config.MonitorConfig{
			Layout:  layoutForCount(count),
			Windows: windows,
		}
	}

	return &config.Config{
		Version:        4,
		ProjectsRoot:   m.projectsRoot,
		DefaultAccount: "claude",
		LastAccount:    "claude",
		Accounts:       m.accounts,
		Monitors:       monitors,
	}
}

func layoutForCount(count int) string {
	switch {
	case count <= 1:
		return "full"
	case count == 2:
		return "vertical"
	default:
		return "grid"
	}
}

func renderLayoutPreview(count int) string {
	var s strings.Builder
	switch {
	case count <= 1:
		s.WriteString("      ┌──────────┐\n")
		s.WriteString("      │          │\n")
		s.WriteString("      │          │\n")
		s.WriteString("      └──────────┘")
	case count == 2:
		s.WriteString("      ┌─────┬─────┐\n")
		s.WriteString("      │     │     │\n")
		s.WriteString("      │     │     │\n")
		s.WriteString("      └─────┴─────┘")
	case count == 3:
		s.WriteString("      ┌─────┬─────┐\n")
		s.WriteString("      │     │     │\n")
		s.WriteString("      ├─────┤     │\n")
		s.WriteString("      │     │     │\n")
		s.WriteString("      └─────┴─────┘")
	default:
		s.WriteString("      ┌─────┬─────┐\n")
		s.WriteString("      │     │     │\n")
		s.WriteString("      ├─────┼─────┤\n")
		s.WriteString("      │     │     │\n")
		s.WriteString("      └─────┴─────┘")
	}
	return DimStyle.Render(s.String())
}

func makeAddAccountInputs() []textinput.Model {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "My Tool"
	inputs[0].CharLimit = 32
	inputs[0].Width = 30

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "tool-name"
	inputs[1].CharLimit = 64
	inputs[1].Width = 30

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "--flag1 --flag2"
	inputs[2].CharLimit = 128
	inputs[2].Width = 30

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "⬜"
	inputs[3].CharLimit = 4
	inputs[3].Width = 10

	return inputs
}

// monitorsDetectedMsg is sent when monitor detection completes
type monitorsDetectedMsg struct {
	monitors []monitor.Monitor
}

func detectMonitors() tea.Msg {
	monitors, err := monitor.Detect()
	if err != nil {
		return monitorsDetectedMsg{monitors: nil}
	}
	return monitorsDetectedMsg{monitors: monitors}
}
