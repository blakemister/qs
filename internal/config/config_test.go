package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultAccounts(t *testing.T) {
	if len(DefaultAccounts) != 7 {
		t.Fatalf("expected 7 default accounts, got %d", len(DefaultAccounts))
	}

	// Check first account is Claude
	if DefaultAccounts[0].ID != "claude" {
		t.Errorf("expected first account ID 'claude', got %q", DefaultAccounts[0].ID)
	}
	if DefaultAccounts[0].Command != "claude" {
		t.Errorf("expected first account command 'claude', got %q", DefaultAccounts[0].Command)
	}
	if !DefaultAccounts[0].Enabled {
		t.Error("expected first account to be enabled")
	}

	// Check that Aider is disabled by default
	aider := AccountByID(DefaultAccounts, "aider")
	if aider == nil {
		t.Fatal("expected to find aider account")
	}
	if aider.Enabled {
		t.Error("expected aider to be disabled by default")
	}
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig("C:/dev")

	if cfg.Version != 4 {
		t.Fatalf("expected version 4, got %d", cfg.Version)
	}
	if cfg.ProjectsRoot != "C:/dev" {
		t.Fatalf("expected projects root C:/dev, got %q", cfg.ProjectsRoot)
	}
	if cfg.DefaultAccount == "" {
		t.Fatal("expected default account to be set")
	}
	if cfg.LastAccount == "" {
		t.Fatal("expected last account to be set")
	}
	if len(cfg.Accounts) != len(DefaultAccounts) {
		t.Fatalf("expected %d accounts, got %d", len(DefaultAccounts), len(cfg.Accounts))
	}
	if len(cfg.Monitors) != 1 {
		t.Fatalf("expected 1 monitor, got %d", len(cfg.Monitors))
	}
	if cfg.Monitors[0].WindowCount() != 1 {
		t.Fatalf("expected monitor to have 1 window, got %d", cfg.Monitors[0].WindowCount())
	}
}

func TestEnsureDefaults(t *testing.T) {
	cfg := &Config{}
	EnsureDefaults(cfg)

	if cfg.Version != 4 {
		t.Fatalf("expected version 4, got %d", cfg.Version)
	}
	if cfg.DefaultAccount == "" {
		t.Fatal("expected default account to be populated")
	}
	if cfg.LastAccount == "" {
		t.Fatal("expected last account to be populated")
	}
	if len(cfg.Accounts) == 0 {
		t.Fatal("expected accounts to be populated")
	}
	if len(cfg.Monitors) == 0 {
		t.Fatal("expected monitors to be populated")
	}
	if cfg.Monitors[0].WindowCount() == 0 {
		t.Fatal("expected at least one monitor window")
	}
}

func TestAccountByID(t *testing.T) {
	accounts := []Account{
		{ID: "claude", Label: "Claude Code"},
		{ID: "codex", Label: "OpenAI Codex"},
	}

	// Found
	a := AccountByID(accounts, "claude")
	if a == nil {
		t.Fatal("expected to find claude")
	}
	if a.Label != "Claude Code" {
		t.Errorf("expected label 'Claude Code', got %q", a.Label)
	}

	// Not found
	if AccountByID(accounts, "unknown") != nil {
		t.Error("expected nil for unknown account")
	}
}

func TestEnabledAccounts(t *testing.T) {
	accounts := []Account{
		{ID: "a", Enabled: true},
		{ID: "b", Enabled: false},
		{ID: "c", Enabled: true},
		{ID: "d", Enabled: false},
	}

	enabled := EnabledAccounts(accounts)
	if len(enabled) != 2 {
		t.Fatalf("expected 2 enabled accounts, got %d", len(enabled))
	}
	if enabled[0].ID != "a" || enabled[1].ID != "c" {
		t.Errorf("expected enabled accounts [a, c], got [%s, %s]", enabled[0].ID, enabled[1].ID)
	}
}

func TestMigrateV3toV4(t *testing.T) {
	v3yaml := []byte(`version: 3
projectsRoot: /home/test/.1dev
monitors:
  - layout: vertical
    windows:
      - tool: cc
      - tool: cx
  - layout: full
    windows:
      - tool: cc
`)

	cfg, err := migrateV3(v3yaml)
	if err != nil {
		t.Fatalf("migrateV3 failed: %v", err)
	}

	if cfg.Version != 4 {
		t.Errorf("expected version 4, got %d", cfg.Version)
	}

	if cfg.ProjectsRoot != "/home/test/.1dev" {
		t.Errorf("expected projectsRoot /home/test/.1dev, got %s", cfg.ProjectsRoot)
	}

	if cfg.DefaultAccount != "claude" {
		t.Errorf("expected defaultAccount 'claude', got %q", cfg.DefaultAccount)
	}

	if len(cfg.Monitors) != 2 {
		t.Fatalf("expected 2 monitors, got %d", len(cfg.Monitors))
	}

	// Check tool mapping: cc -> claude, cx -> codex
	if cfg.Monitors[0].Windows[0].Tool != "claude" {
		t.Errorf("expected window 0 tool 'claude', got %q", cfg.Monitors[0].Windows[0].Tool)
	}
	if cfg.Monitors[0].Windows[1].Tool != "codex" {
		t.Errorf("expected window 1 tool 'codex', got %q", cfg.Monitors[0].Windows[1].Tool)
	}

	// Check accounts are seeded
	if len(cfg.Accounts) != len(DefaultAccounts) {
		t.Errorf("expected %d accounts, got %d", len(DefaultAccounts), len(cfg.Accounts))
	}
}

func TestMigrateV2toV4(t *testing.T) {
	v2yaml := []byte(`version: 2
projectsRoot: /home/test/.1dev
monitors:
  - windows: 2
    layout: vertical
  - windows: 1
    layout: full
`)

	cfg, err := migrateV2(v2yaml)
	if err != nil {
		t.Fatalf("migrateV2 failed: %v", err)
	}

	if cfg.Version != 4 {
		t.Errorf("expected version 4, got %d", cfg.Version)
	}

	if len(cfg.Monitors) != 2 {
		t.Fatalf("expected 2 monitors, got %d", len(cfg.Monitors))
	}

	if cfg.Monitors[0].WindowCount() != 2 {
		t.Errorf("monitor 0: expected 2 windows, got %d", cfg.Monitors[0].WindowCount())
	}

	// V2 windows all map to "claude"
	for j := 0; j < 2; j++ {
		if cfg.Monitors[0].Windows[j].Tool != "claude" {
			t.Errorf("monitor 0, window %d: expected tool 'claude', got %q", j, cfg.Monitors[0].Windows[j].Tool)
		}
	}

	if cfg.Monitors[1].WindowCount() != 1 {
		t.Errorf("monitor 1: expected 1 window, got %d", cfg.Monitors[1].WindowCount())
	}
}

func TestLoadV4Direct(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	v4yaml := []byte(`version: 4
projectsRoot: /test
defaultAccount: claude
lastAccount: codex
accounts:
  - id: claude
    label: Claude Code
    command: claude
    args: ["--dangerously-skip-permissions"]
    icon: "C"
    enabled: true
  - id: codex
    label: OpenAI Codex
    command: codex
    args: ["--full-auto"]
    icon: "X"
    enabled: true
monitors:
  - layout: vertical
    windows:
      - tool: claude
      - tool: codex
`)

	if err := os.WriteFile(path, v4yaml, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Version != 4 {
		t.Errorf("expected version 4, got %d", cfg.Version)
	}
	if cfg.DefaultAccount != "claude" {
		t.Errorf("expected defaultAccount 'claude', got %q", cfg.DefaultAccount)
	}
	if cfg.LastAccount != "codex" {
		t.Errorf("expected lastAccount 'codex', got %q", cfg.LastAccount)
	}
	if len(cfg.Accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(cfg.Accounts))
	}
	if cfg.Monitors[0].ToolFor(0) != "claude" {
		t.Errorf("expected window 0 tool 'claude', got %q", cfg.Monitors[0].ToolFor(0))
	}
	if cfg.Monitors[0].ToolFor(1) != "codex" {
		t.Errorf("expected window 1 tool 'codex', got %q", cfg.Monitors[0].ToolFor(1))
	}
}

func TestSaveAndLoadV4(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		Version:        4,
		ProjectsRoot:   "/test/projects",
		DefaultAccount: "claude",
		LastAccount:    "codex",
		Accounts: []Account{
			{ID: "claude", Label: "Claude", Command: "claude", Args: []string{"--skip"}, Enabled: true},
			{ID: "codex", Label: "Codex", Command: "codex", Args: []string{"--auto"}, Enabled: true},
		},
		Monitors: []MonitorConfig{
			{
				Layout: "vertical",
				Windows: []WindowConfig{
					{Tool: "codex"},
					{Tool: "claude"},
				},
			},
		},
	}

	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != 4 {
		t.Errorf("expected version 4, got %d", loaded.Version)
	}
	if loaded.DefaultAccount != "claude" {
		t.Errorf("expected defaultAccount 'claude', got %q", loaded.DefaultAccount)
	}
	if loaded.LastAccount != "codex" {
		t.Errorf("expected lastAccount 'codex', got %q", loaded.LastAccount)
	}
	if loaded.Monitors[0].ToolFor(0) != "codex" {
		t.Errorf("expected window 0 tool 'codex', got %q", loaded.Monitors[0].ToolFor(0))
	}
	if loaded.Monitors[0].ToolFor(1) != "claude" {
		t.Errorf("expected window 1 tool 'claude', got %q", loaded.Monitors[0].ToolFor(1))
	}
}

func TestLoadFallbackToLegacy(t *testing.T) {
	// Create a temp directory structure simulating home
	dir := t.TempDir()
	legacyDir := filepath.Join(dir, ".cc")
	os.MkdirAll(legacyDir, 0755)

	legacyPath := filepath.Join(legacyDir, "config.yaml")
	v3yaml := []byte(`version: 3
projectsRoot: /legacy
monitors:
  - layout: full
    windows:
      - tool: cc
`)

	if err := os.WriteFile(legacyPath, v3yaml, 0644); err != nil {
		t.Fatalf("failed to write legacy config: %v", err)
	}

	// Load with explicit legacy path (since we can't mock home dir easily)
	cfg, err := Load(legacyPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Should have been migrated to v4
	if cfg.Version != 4 {
		t.Errorf("expected version 4 after migration, got %d", cfg.Version)
	}

	// cc -> claude mapping
	if cfg.Monitors[0].Windows[0].Tool != "claude" {
		t.Errorf("expected tool 'claude' after migration, got %q", cfg.Monitors[0].Windows[0].Tool)
	}
}

func TestWindowConfigToolMapping(t *testing.T) {
	// Test the v3 tool mapping function
	tests := []struct {
		input    string
		expected string
	}{
		{"cc", "claude"},
		{"cx", "codex"},
		{"claude", "claude"},
		{"codex", "codex"},
		{"", "claude"},
	}

	for _, tt := range tests {
		got := mapV3Tool(tt.input)
		if got != tt.expected {
			t.Errorf("mapV3Tool(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDefaultConfigPaths(t *testing.T) {
	// Just verify they don't panic and return non-empty strings
	path := DefaultConfigPath()
	if path == "" {
		t.Error("DefaultConfigPath() returned empty string")
	}

	legacy := LegacyConfigPath()
	if legacy == "" {
		t.Error("LegacyConfigPath() returned empty string")
	}

	root := DefaultProjectsRoot()
	if root == "" {
		t.Error("DefaultProjectsRoot() returned empty string")
	}

	// Verify paths contain expected directory names
	if !containsPath(path, ".qs") {
		t.Errorf("DefaultConfigPath() should contain .qs, got %q", path)
	}
	if !containsPath(legacy, ".cc") {
		t.Errorf("LegacyConfigPath() should contain .cc, got %q", legacy)
	}
}

func containsPath(path, segment string) bool {
	for _, p := range filepath.SplitList(path) {
		if p == segment {
			return true
		}
	}
	// Also check with filepath split
	dir := path
	for dir != "." && dir != "/" && dir != "" {
		base := filepath.Base(dir)
		if base == segment {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return false
}
