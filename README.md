# qs

**Launch AI coding agents across all your monitors with a single command.**

```
qs
```

Pick a project, pick a tool, start coding. Terminals auto-arrange across your monitors.

---

## What It Does

You have multiple monitors. You want to vibe code on several projects at once. `qs` handles the tedious part:

1. Spawns terminal windows across all your monitors
2. Each window gets a project picker with fuzzy search
3. Pick an AI coding tool (Claude Code, Codex, Gemini, and more)
4. Windows auto-position in your configured layout

```
┌─────────────────────┬─────────────────────┬───────────────────┐
│   Monitor 1 (2x2)   │   Monitor 2 (split) │  Laptop (full)    │
│ ┌────────┬────────┐ │ ┌────────┬────────┐ │ ┌───────────────┐ │
│ │ Claude │ Codex  │ │ │ Claude │ Gemini │ │ │  Claude Code   │ │
│ ├────────┼────────┤ │ │  Code  │  CLI   │ │ │               │ │
│ │ Claude │ Claude │ │ │        │        │ │ │               │ │
│ │  Code  │  Code  │ │ │        │        │ │ │               │ │
│ └────────┴────────┘ │ └────────┴────────┘ │ └───────────────┘ │
└─────────────────────┴─────────────────────┴───────────────────┘
```

---

## Install

Requires **Go 1.21+** and **Windows Terminal**.

```powershell
git clone https://github.com/blakemister/quickstart.git
cd quickstart
.\install.ps1
```

This builds `qs.exe`, installs it to `~/.qs/bin/`, and adds it to your PATH.

Or build manually:

```bash
go build -o qs.exe .
```

---

## Usage

```bash
qs                # Launch project picker
qs setup          # Run the setup wizard
qs accounts       # Manage AI tool accounts
qs monitors       # List detected monitors
qs version        # Print version
```

### First Run

On first launch, `qs` will prompt you to either run the full setup wizard or quickly set your projects directory. The setup wizard walks through:

1. **Projects folder** - where your project directories live
2. **Monitor layout** - how many windows per monitor and their arrangement
3. **AI tool accounts** - which tools to enable

### Project Picker

The main TUI shows your project folders with fuzzy search. Type to filter, arrow keys to navigate, Enter to select. You can also create new project folders inline.

### Account Selection

After picking a project, choose which AI coding tool to launch. If only one tool is enabled, it launches automatically.

---

## Supported Tools

| Tool | Command | Default |
|------|---------|---------|
| Claude Code | `claude --dangerously-skip-permissions` | Enabled |
| OpenAI Codex | `codex --dangerously-bypass-approvals-and-sandbox` | Enabled |
| Gemini CLI | `gemini --yolo` | Enabled |
| OpenCode (z.ai) | `opencode` | Enabled |
| Cursor Agent | `agent` | Enabled |
| Aider | `aider --yes-always` | Disabled |
| Continue Dev | `continue` | Disabled |

Add custom tools through the setup wizard or `qs accounts`.

---

## Configuration

Config file: `~/.qs/config.yaml`

```yaml
version: 4
projectsRoot: "C:/Users/you/dev"
defaultAccount: claude
lastAccount: claude
accounts:
  - id: claude
    label: Claude Code
    command: claude
    args: ["--dangerously-skip-permissions"]
    icon: "\U0001F7E0"
    enabled: true
  - id: codex
    label: OpenAI Codex
    command: codex
    args: ["--dangerously-bypass-approvals-and-sandbox"]
    icon: "\U0001F7E2"
    enabled: true
monitors:
  - layout: grid
    windows:
      - tool: claude
      - tool: claude
      - tool: codex
      - tool: claude
  - layout: vertical
    windows:
      - tool: claude
      - tool: claude
  - layout: full
    windows:
      - tool: claude
```

### Layouts

| Layout | Description |
|--------|-------------|
| `full` | Single fullscreen window |
| `vertical` | Side-by-side columns |
| `horizontal` | Stacked rows |
| `grid` | 2x2, 3x3, etc. based on window count |

---

## How It Works

1. **Monitor detection** - Uses Win32 `EnumDisplayMonitors` API to find all connected displays
2. **Window spawning** - Launches Windows Terminal (`wt.exe`) instances with unique titles
3. **Window positioning** - Uses Win32 `SetWindowPos` to arrange windows according to layout config
4. **Project picker** - Each terminal runs the Bubble Tea TUI for project/tool selection
5. **Tool launch** - Hands off to the selected AI coding tool via `tea.ExecProcess`

---

## Project Structure

```
qs/
├── main.go                    # Entry point
├── go.mod
├── install.ps1                # Build + install script
├── internal/
│   ├── cmd/                   # CLI commands (Cobra)
│   │   ├── root.go            # Main command + first-run flow
│   │   ├── setup.go           # Setup wizard command
│   │   ├── accounts.go        # Account management command
│   │   ├── monitors.go        # Monitor listing command
│   │   └── version.go         # Version command
│   ├── config/                # Configuration (YAML, migration)
│   │   ├── config.go          # Load/save/migrate config
│   │   └── accounts.go        # Account definitions
│   ├── launcher/              # Window management (Win32 API)
│   │   └── launcher.go        # Terminal launch + positioning
│   ├── monitor/               # Monitor detection (Win32 API)
│   │   └── monitor.go         # EnumDisplayMonitors wrapper
│   └── tui/                   # Terminal UI (Bubble Tea)
│       ├── picker.go          # Project + account picker
│       ├── setup.go           # Setup wizard
│       ├── first_run.go       # First-run flow
│       ├── accounts.go        # Account management UI
│       ├── keys.go            # Key bindings
│       └── styles.go          # Colors + styles
└── README.md
```

---

## Requirements

- **Windows 10/11**
- **Windows Terminal** (default on Windows 11, or install from Microsoft Store)
- **Go 1.21+** (to build from source)
- At least one AI coding tool installed (`claude`, `codex`, `gemini`, etc.)

---

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- Win32 API - Monitor detection + window positioning

---

## License

MIT
