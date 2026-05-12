// Package phases defines the ordered setup phases for mac-dev-station.
package phases

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oleg-koval/mac-dev-station/internal/configs"
	"github.com/oleg-koval/mac-dev-station/internal/system"
)

var homeDir = os.ExpandEnv("$HOME")

type Status int

const (
	StatusUnknown Status = iota
	StatusSatisfied
	StatusPartial
	StatusMissing
)

type Phase interface {
	Name() string
	Description() string
	Check(ctx context.Context) (Status, error)
	Apply(ctx context.Context) error
}

// backupConfig backs up an existing config file to ~/.dotfiles-backup-YYYYMMDD/
func backupConfig(srcPath string) error {
	if _, err := os.Stat(srcPath); err != nil {
		return nil // File doesn't exist, no backup needed
	}

	backupDir := filepath.Join(homeDir, ".dotfiles-backup-"+time.Now().Format("20060102"))
	if err := os.MkdirAll(backupDir, 0o755); err != nil { //nolint:gosec // 0o755 is correct for $HOME subdirs
		return fmt.Errorf("failed to create backup dir: %w", err)
	}

	content, err := os.ReadFile(srcPath) //nolint:gosec // path is derived from $HOME
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %w", err)
	}

	relPath, _ := filepath.Rel(homeDir, srcPath)
	backupPath := filepath.Join(backupDir, relPath)
	if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil { //nolint:gosec // 0o755 for $HOME subdirs
		return fmt.Errorf("failed to create backup subdirs: %w", err)
	}

	if err := os.WriteFile(backupPath, content, 0o644); err != nil { //nolint:gosec // backup files match source perms
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// fileSHA256 computes the SHA256 hash of a file's contents
func fileSHA256(path string) (string, error) {
	content, err := os.ReadFile(path) //nolint:gosec // path is an internal config path derived from $HOME
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(content)), nil
}

var Registry = []Phase{
	&PreflightPhase{},
	&FoundationsPhase{},
	&BrewfilePhase{},
	&FoldersPhase{},
	&KarabinerPhase{},
	&AerospacePhase{},
	&HammerspoonPhase{},
	&KittyPhase{},
	&ShellPhase{},
	&RaycastPhase{},
	&StartersPhase{},
	&PermissionsWalkPhase{},
	&VerifyPhase{},
	&CheatsheetPhase{},
}

// PreflightPhase checks system requirements
type PreflightPhase struct{}

func (p *PreflightPhase) Name() string {
	return "Preflight"
}

func (p *PreflightPhase) Description() string {
	return "Check macOS version, brew, xcode-select, gh auth"
}

func (p *PreflightPhase) Check(ctx context.Context) (Status, error) {
	// Check macOS version
	version, err := system.MacOSVersion(ctx)
	if err != nil {
		return StatusMissing, err
	}
	if version < 14.0 {
		return StatusMissing, fmt.Errorf("macOS 14+ required, found %.1f", version)
	}

	// Check Homebrew
	if !system.BrewInstalled(ctx) {
		return StatusMissing, nil
	}

	// Check xcode-select
	_, err = system.RunCmd(ctx, "xcode-select", "-p")
	if err != nil {
		return StatusMissing, nil
	}

	// Check gh auth
	_, err = system.RunCmd(ctx, "gh", "auth", "status")
	if err != nil {
		return StatusPartial, nil // gh not authenticated but brew/xcode exist
	}

	return StatusSatisfied, nil
}

func (p *PreflightPhase) Apply(ctx context.Context) error {
	return nil
}

// FoundationsPhase installs Homebrew and xcode-select
type FoundationsPhase struct{}

func (p *FoundationsPhase) Name() string {
	return "Foundations"
}

func (p *FoundationsPhase) Description() string {
	return "Install/update Homebrew, xcode-select"
}

func (p *FoundationsPhase) Check(ctx context.Context) (Status, error) {
	if !system.BrewInstalled(ctx) {
		return StatusMissing, nil
	}
	_, err := system.RunCmd(ctx, "xcode-select", "-p")
	if err != nil {
		return StatusMissing, nil
	}
	return StatusSatisfied, nil
}

func (p *FoundationsPhase) Apply(ctx context.Context) error {
	if !system.BrewInstalled(ctx) {
		return fmt.Errorf("brew not installed, install manually or run earlier phase")
	}
	_, err := system.RunCmd(context.Background(), "xcode-select", "--install")
	if err != nil && err.Error() != "exit status 1" {
		return fmt.Errorf("xcode-select install failed: %w", err)
	}
	if err := system.BrewUpdate(context.Background(), os.Stdout); err != nil {
		return fmt.Errorf("brew update failed: %w", err)
	}
	if err := system.BrewUpgrade(context.Background(), os.Stdout); err != nil {
		return fmt.Errorf("brew upgrade failed: %w", err)
	}
	if err := system.BrewCleanup(context.Background(), os.Stdout); err != nil {
		return fmt.Errorf("brew cleanup failed: %w", err)
	}
	return nil
}

// BrewfilePhase installs all brew packages
type BrewfilePhase struct{}

func (p *BrewfilePhase) Name() string {
	return "Brewfile"
}

func (p *BrewfilePhase) Description() string {
	return "Install all CLI tools and GUI apps"
}

func (p *BrewfilePhase) Check(ctx context.Context) (Status, error) {
	// Simple check: verify a few key tools are installed
	checks := []string{"brew", "git", "gh", "starship"}
	allPresent := true

	for _, tool := range checks {
		_, err := system.RunCmd(ctx, "which", tool)
		if err != nil {
			allPresent = false
			break
		}
	}

	// Check for at least one GUI app
	appsPresent := system.AppInstalled(ctx, "kitty") || system.AppInstalled(ctx, "Cursor")

	if allPresent && appsPresent {
		return StatusSatisfied, nil
	}
	if allPresent {
		return StatusPartial, nil
	}
	return StatusMissing, nil
}

func (p *BrewfilePhase) Apply(ctx context.Context) error {
	brewfilePath := filepath.Join(homeDir, "Brewfile")

	// Write embedded Brewfile
	if err := os.WriteFile(brewfilePath, configs.BrewfileContent, 0o644); err != nil { //nolint:gosec // Brewfile must be world-readable
		return fmt.Errorf("failed to write Brewfile: %w", err)
	}

	// Run brew bundle
	return system.BrewBundle(ctx, os.Stdout, brewfilePath)
}

// FoldersPhase creates PARA folders
type FoldersPhase struct{}

func (p *FoldersPhase) Name() string {
	return "Folders"
}

func (p *FoldersPhase) Description() string {
	return "Create ~/Work (PARA), ~/oss, ~/code"
}

func (p *FoldersPhase) Check(ctx context.Context) (Status, error) {
	folders := []string{
		filepath.Join(homeDir, "Work"),
		filepath.Join(homeDir, "oss"),
		filepath.Join(homeDir, "code"),
	}

	allExist := true
	for _, folder := range folders {
		if _, err := os.Stat(folder); err != nil {
			allExist = false
			break
		}
	}

	if allExist {
		return StatusSatisfied, nil
	}
	return StatusMissing, nil
}

func (p *FoldersPhase) Apply(ctx context.Context) error {
	folders := []string{
		filepath.Join(homeDir, "Work"),
		filepath.Join(homeDir, "oss"),
		filepath.Join(homeDir, "code"),
	}

	for _, folder := range folders {
		if err := os.MkdirAll(folder, 0o755); err != nil { //nolint:gosec // 0o755 for user workspace dirs
			return err
		}
	}
	return nil
}

// KarabinerPhase configures Karabiner (Hyper key)
type KarabinerPhase struct{}

func (p *KarabinerPhase) Name() string {
	return "Karabiner"
}

func (p *KarabinerPhase) Description() string {
	return "Hyper key config + app launchers"
}

func (p *KarabinerPhase) Check(ctx context.Context) (Status, error) {
	configPath := filepath.Join(homeDir, ".config/karabiner/karabiner.json")

	// Check if Karabiner is installed
	if !system.AppInstalled(ctx, "Karabiner-Elements") {
		return StatusMissing, nil
	}

	// Check if config exists and matches embedded version
	if _, err := os.Stat(configPath); err != nil {
		return StatusMissing, nil
	}

	embeddedHash := fmt.Sprintf("%x", sha256.Sum256(configs.KarabinerContent))
	fileHash, err := fileSHA256(configPath)
	if err == nil && embeddedHash == fileHash {
		return StatusSatisfied, nil
	}

	return StatusPartial, nil
}

func (p *KarabinerPhase) Apply(ctx context.Context) error {
	configDir := filepath.Join(homeDir, ".config/karabiner")
	configPath := filepath.Join(configDir, "karabiner.json")

	// Backup existing config
	if err := backupConfig(configPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil { //nolint:gosec // 0o755 for $HOME/.config subdirs
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	if err := os.WriteFile(configPath, configs.KarabinerContent, 0o644); err != nil { //nolint:gosec // app config files are world-readable
		return fmt.Errorf("failed to write Karabiner config: %w", err)
	}

	return nil
}

// AerospacePhase configures AeroSpace (tiling WM)
type AerospacePhase struct{}

func (p *AerospacePhase) Name() string {
	return "AeroSpace"
}

func (p *AerospacePhase) Description() string {
	return "Tiling WM with auto-routing to workspaces"
}

func (p *AerospacePhase) Check(ctx context.Context) (Status, error) {
	configPath := filepath.Join(homeDir, ".config/aerospace/aerospace.toml")
	if !system.AppInstalled(ctx, "AeroSpace") {
		return StatusMissing, nil
	}
	if _, err := os.Stat(configPath); err != nil {
		return StatusMissing, nil
	}
	embeddedHash := fmt.Sprintf("%x", sha256.Sum256(configs.AerospaceContent))
	fileHash, err := fileSHA256(configPath)
	if err == nil && embeddedHash == fileHash {
		return StatusSatisfied, nil
	}
	return StatusPartial, nil
}

func (p *AerospacePhase) Apply(ctx context.Context) error {
	configDir := filepath.Join(homeDir, ".config/aerospace")
	configPath := filepath.Join(configDir, "aerospace.toml")
	if err := backupConfig(configPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil { //nolint:gosec // 0o755 for $HOME/.config subdirs
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.WriteFile(configPath, configs.AerospaceContent, 0o644); err != nil { //nolint:gosec // app config files are world-readable
		return fmt.Errorf("failed to write AeroSpace config: %w", err)
	}
	return nil
}

// HammerspoonPhase configures Hammerspoon (display auto-detect)
type HammerspoonPhase struct{}

func (p *HammerspoonPhase) Name() string {
	return "Hammerspoon"
}

func (p *HammerspoonPhase) Description() string {
	return "Display dock/undock auto-detection"
}

func (p *HammerspoonPhase) Check(ctx context.Context) (Status, error) {
	initPath := filepath.Join(homeDir, ".hammerspoon/init.lua")
	if !system.AppInstalled(ctx, "Hammerspoon") {
		return StatusMissing, nil
	}
	if _, err := os.Stat(initPath); err != nil {
		return StatusMissing, nil
	}
	embeddedHash := fmt.Sprintf("%x", sha256.Sum256(configs.HammerspoonInitContent))
	fileHash, err := fileSHA256(initPath)
	if err == nil && embeddedHash == fileHash {
		return StatusSatisfied, nil
	}
	return StatusPartial, nil
}

func (p *HammerspoonPhase) Apply(ctx context.Context) error {
	hsDir := filepath.Join(homeDir, ".hammerspoon")
	initPath := filepath.Join(hsDir, "init.lua")
	watcherPath := filepath.Join(hsDir, "display-watcher.lua")

	if err := backupConfig(initPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	if err := os.MkdirAll(hsDir, 0o755); err != nil { //nolint:gosec // 0o755 for $HOME/.hammerspoon
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.WriteFile(initPath, configs.HammerspoonInitContent, 0o644); err != nil { //nolint:gosec // Lua config files are world-readable
		return fmt.Errorf("failed to write init.lua: %w", err)
	}
	if err := os.WriteFile(watcherPath, configs.HammerspoonDisplayWatcherContent, 0o644); err != nil { //nolint:gosec // Lua config files are world-readable
		return fmt.Errorf("failed to write display-watcher.lua: %w", err)
	}
	return nil
}

// KittyPhase configures kitty (terminal)
type KittyPhase struct{}

func (p *KittyPhase) Name() string {
	return "kitty"
}

func (p *KittyPhase) Description() string {
	return "Terminal config + project switcher (Cmd+P)"
}

func (p *KittyPhase) Check(ctx context.Context) (Status, error) {
	confPath := filepath.Join(homeDir, ".config/kitty/kitty.conf")
	if !system.AppInstalled(ctx, "kitty") {
		return StatusMissing, nil
	}
	if _, err := os.Stat(confPath); err != nil {
		return StatusMissing, nil
	}
	embeddedHash := fmt.Sprintf("%x", sha256.Sum256(configs.KittyConfContent))
	fileHash, err := fileSHA256(confPath)
	if err == nil && embeddedHash == fileHash {
		return StatusSatisfied, nil
	}
	return StatusPartial, nil
}

func (p *KittyPhase) Apply(ctx context.Context) error {
	kittyDir := filepath.Join(homeDir, ".config/kitty")
	confPath := filepath.Join(kittyDir, "kitty.conf")
	themePath := filepath.Join(kittyDir, "one-dark.conf")
	projPath := filepath.Join(kittyDir, "projects.py")

	if err := backupConfig(confPath); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	if err := os.MkdirAll(kittyDir, 0o755); err != nil { //nolint:gosec // 0o755 for $HOME/.config/kitty
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	if err := os.WriteFile(confPath, configs.KittyConfContent, 0o644); err != nil { //nolint:gosec // kitty config is world-readable
		return fmt.Errorf("failed to write kitty.conf: %w", err)
	}
	if err := os.WriteFile(themePath, configs.KittyColorSchemeContent, 0o644); err != nil { //nolint:gosec // kitty config is world-readable
		return fmt.Errorf("failed to write color scheme: %w", err)
	}
	if err := os.WriteFile(projPath, configs.KittyProjectsContent, 0o755); err != nil { //nolint:gosec // Python script must be executable
		return fmt.Errorf("failed to write projects.py: %w", err)
	}
	return nil
}

// ShellPhase configures shell (zsh + starship + aliases + cron)
type ShellPhase struct{}

func (p *ShellPhase) Name() string {
	return "Shell"
}

func (p *ShellPhase) Description() string {
	return "zsh + starship + aliases + 9am backup cron"
}

func (p *ShellPhase) Check(ctx context.Context) (Status, error) {
	zshPath := filepath.Join(homeDir, ".zshrc")
	if _, err := os.Stat(zshPath); err == nil {
		return StatusPartial, nil
	}
	return StatusMissing, nil
}

func (p *ShellPhase) Apply(ctx context.Context) error {
	zshDir := filepath.Join(homeDir, ".zsh")
	if err := os.MkdirAll(zshDir, 0o755); err != nil { //nolint:gosec // 0o755 for $HOME/.zsh
		return fmt.Errorf("failed to create zsh dir: %w", err)
	}

	zshPath := filepath.Join(homeDir, ".zshrc")
	if err := os.WriteFile(zshPath, configs.ZshrcContent, 0o644); err != nil { //nolint:gosec // .zshrc must be world-readable
		return fmt.Errorf("failed to write zshrc: %w", err)
	}

	// Write secrets template only if it doesn't exist — protect user's API keys.
	secretsPath := filepath.Join(zshDir, "secrets.zsh")
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		if err := os.WriteFile(secretsPath, configs.SecretsTemplateContent, 0o600); err != nil {
			return fmt.Errorf("failed to write secrets template: %w", err)
		}
	}

	scriptDir := filepath.Join(homeDir, "code/oss/scripts")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil { //nolint:gosec // 0o755 for scripts dir in $HOME/code
		return fmt.Errorf("failed to create scripts dir: %w", err)
	}
	scriptPath := filepath.Join(scriptDir, "backup-zsh.sh")
	if err := os.WriteFile(scriptPath, configs.BackupScriptContent, 0o755); err != nil { //nolint:gosec // shell script must be executable
		return fmt.Errorf("failed to write backup script: %w", err)
	}

	return nil
}

// RaycastPhase opens Raycast for first-time setup
type RaycastPhase struct{}

func (p *RaycastPhase) Name() string {
	return "Raycast"
}

func (p *RaycastPhase) Description() string {
	return "Open for first-time setup (Cmd+Space)"
}

func (p *RaycastPhase) Check(ctx context.Context) (Status, error) {
	if system.AppInstalled(ctx, "Raycast") {
		return StatusSatisfied, nil
	}
	return StatusMissing, nil
}

func (p *RaycastPhase) Apply(ctx context.Context) error {
	return system.RunCmdStream(ctx, os.Stdout, "open", "-a", "Raycast")
}

// StartersPhase clones oleg-koval/starters
type StartersPhase struct{}

func (p *StartersPhase) Name() string {
	return "OSS starters"
}

func (p *StartersPhase) Description() string {
	return "Clone oleg-koval/starters"
}

func (p *StartersPhase) Check(ctx context.Context) (Status, error) {
	startersPath := filepath.Join(homeDir, "code/starters")
	if _, err := os.Stat(startersPath); err == nil {
		return StatusSatisfied, nil
	}
	return StatusMissing, nil
}

func (p *StartersPhase) Apply(ctx context.Context) error {
	ossDir := filepath.Join(homeDir, "code")
	if err := os.MkdirAll(ossDir, 0o755); err != nil { //nolint:gosec // 0o755 for $HOME/code
		return fmt.Errorf("failed to create code dir: %w", err)
	}
	startersPath := filepath.Join(ossDir, "starters")
	if _, err := os.Stat(startersPath); err == nil {
		return nil // Already cloned
	}
	return system.RunCmdStream(ctx, os.Stdout, "git", "-C", ossDir, "clone", "git@github.com:oleg-koval/starters.git")
}

// PermissionsWalkPhase guides through Accessibility + Driver Extensions
type PermissionsWalkPhase struct{}

func (p *PermissionsWalkPhase) Name() string {
	return "Permissions"
}

func (p *PermissionsWalkPhase) Description() string {
	return "Walks through Accessibility, Driver Extensions"
}

func (p *PermissionsWalkPhase) Check(ctx context.Context) (Status, error) {
	return StatusPartial, nil
}

func (p *PermissionsWalkPhase) Apply(ctx context.Context) error {
	fmt.Println("Please grant accessibility permissions to these apps:")
	fmt.Println("  - Karabiner-Elements → System Settings → Privacy → Accessibility")
	fmt.Println("  - AeroSpace → System Settings → Privacy → Accessibility")
	fmt.Println("  - Hammerspoon → System Settings → Privacy → Accessibility")
	fmt.Println("  - Raycast → System Settings → Privacy → Accessibility")
	fmt.Println("\nYou may need to restart your Mac for changes to take effect.")
	return nil
}

// VerifyPhase runs smoke tests
type VerifyPhase struct{}

func (p *VerifyPhase) Name() string {
	return "Verification"
}

func (p *VerifyPhase) Description() string {
	return "Final smoke test of every component"
}

func (p *VerifyPhase) Check(ctx context.Context) (Status, error) {
	checks := []struct {
		name string
		fn   func(context.Context) bool
	}{
		{"brew", system.BrewInstalled},
		{"kitty", func(c context.Context) bool { return system.AppInstalled(c, "kitty") }},
		{"karabiner", func(c context.Context) bool { return system.AppInstalled(c, "Karabiner-Elements") }},
		{"aerospace", func(c context.Context) bool { return system.AppInstalled(c, "AeroSpace") }},
	}

	allPass := true
	for _, check := range checks {
		if !check.fn(ctx) {
			fmt.Printf("  ✗ %s not found\n", check.name)
			allPass = false
		}
	}

	if allPass {
		return StatusSatisfied, nil
	}
	return StatusPartial, nil
}

func (p *VerifyPhase) Apply(ctx context.Context) error {
	return nil
}

// CheatsheetPhase prints hotkey map
type CheatsheetPhase struct{}

func (p *CheatsheetPhase) Name() string {
	return "Cheatsheet"
}

func (p *CheatsheetPhase) Description() string {
	return "Prints all hotkeys"
}

func (p *CheatsheetPhase) Check(ctx context.Context) (Status, error) {
	return StatusSatisfied, nil
}

func (p *CheatsheetPhase) Apply(ctx context.Context) error {
	fmt.Println("=== Hyper Key (Caps Lock) ===")
	fmt.Println("  Caps (tap)    → Escape")
	fmt.Println("  Caps+T        → Open kitty")
	fmt.Println("  Caps+B        → Open Chrome")
	fmt.Println("  Caps+C        → Open Cursor")
	fmt.Println("  Caps+S        → Open Slack")
	fmt.Println("  Caps+L        → Linear (browser)")
	fmt.Println("  Caps+F        → Figma (browser)")
	fmt.Println("  Caps+N        → Notion (browser)")
	fmt.Println("  Caps+H/J/K/;  → Arrow keys (vim)")
	fmt.Println("\n=== AeroSpace Tiling ===")
	fmt.Println("  Alt+H/J/K/L   → Focus left/down/up/right")
	fmt.Println("  Alt+Shift+H/J/K/L → Resize window")
	fmt.Println("  Alt+A         → Toggle accordion layout")
	fmt.Println("  Alt+T         → Toggle tiles layout")
	return nil
}
