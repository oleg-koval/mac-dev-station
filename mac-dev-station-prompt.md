# Build `mac-dev-station` - Go CLI for macOS dev environment setup

You are Claude Code. Build a complete Go CLI named `mac-dev-station` that automates setting up a macOS development environment to match a known-good reference. Use my Go starter as the template.

## Step 0: Bootstrap the project from my starter

```bash
cd ~/oss/mine
gh repo create oleg-koval/mac-dev-station \
  --public \
  --template oleg-koval/go-starter \
  --clone \
  --description "One-command macOS dev environment bootstrap (Karabiner, AeroSpace, Hammerspoon, kitty, Raycast, zsh)"
cd mac-dev-station
```

The starter already gives you:
- Standard layout (`cmd/`, `internal/`)
- `golangci-lint` config
- `Makefile` with `test`, `lint`, `cover`, `build`
- GitHub Actions CI matrix (Go 1.23, 1.24)
- Dependabot
- `AGENTS.md`, `CONTRIBUTING.md`, `LICENSE` (MIT), `.editorconfig`
- README skeleton with badges

**Rename the module** from `github.com/oleg-koval/go-starter` to `github.com/oleg-koval/mac-dev-station`:

```bash
find . -type f \( -name "*.go" -o -name "go.mod" -o -name "README.md" -o -name "Makefile" \) \
  -not -path "./.git/*" \
  -exec sed -i '' 's|github.com/oleg-koval/go-starter|github.com/oleg-koval/mac-dev-station|g' {} +
```

Rename `cmd/app/` to `cmd/mac-dev-station/`:

```bash
mv cmd/app cmd/mac-dev-station
sed -i '' 's|cmd/app|cmd/mac-dev-station|g' Makefile README.md
sed -i '' 's|BINARY := bin/app|BINARY := bin/mac-dev-station|' Makefile
```

Commit the rename before adding code:

```bash
go mod tidy
git add -A
git commit -m "chore: rename module from go-starter template"
git push
```

## Project structure to build

```
mac-dev-station/
├── cmd/mac-dev-station/
│   └── main.go                    # entry point, cobra CLI
├── internal/
│   ├── reporter/
│   │   └── reporter.go            # colored status output (ANSI, no library)
│   ├── system/
│   │   ├── system.go              # macOS version, arch, brew detection
│   │   ├── brew.go                # brew install/upgrade wrappers
│   │   └── commands.go            # safe shell-out with logging
│   ├── permissions/
│   │   └── permissions.go         # accessibility checks, System Settings deep links
│   ├── phases/
│   │   ├── phases.go              # Phase interface + Registry
│   │   ├── preflight.go
│   │   ├── foundations.go
│   │   ├── brewfile.go
│   │   ├── folders.go
│   │   ├── karabiner.go
│   │   ├── aerospace.go
│   │   ├── hammerspoon.go
│   │   ├── kitty.go
│   │   ├── shell.go
│   │   ├── raycast.go
│   │   ├── starters.go
│   │   ├── permissions_walk.go
│   │   ├── verify.go
│   │   └── cheatsheet.go
│   └── configs/
│       └── embed.go               # //go:embed all the config files
└── configs/                        # source-of-truth files (embedded)
    ├── Brewfile
    ├── karabiner.json
    ├── aerospace.toml
    ├── hammerspoon/
    │   ├── init.lua
    │   └── display-watcher.lua
    ├── kitty/
    │   ├── kitty.conf
    │   ├── one-dark.conf
    │   └── projects.py
    └── shell/
        ├── zshrc
        ├── secrets.zsh.template
        └── backup-zsh.sh
```

All files in `configs/` are embedded into the binary via `//go:embed` so the binary is self-contained.

## Library choices (keep minimal)

- **CLI framework:** `github.com/spf13/cobra` for subcommands + flags
- **Logger:** `log/slog` from stdlib
- **Colors:** ANSI escape codes directly, no library
- **TOML/JSON:** stdlib (we only embed and write, no parsing needed)

```bash
go get github.com/spf13/cobra@latest
```

That's the only dependency. No viper, no charm, no bubbletea, no color libs. Keep it boring.

## CLI surface

```
mac-dev-station                       # interactive run, all phases
mac-dev-station --dry-run             # show what would happen, change nothing
mac-dev-station --skip aerospace,hammerspoon   # skip phases
mac-dev-station --only karabiner      # only run named phase
mac-dev-station doctor                # run verification only (phase 12)
mac-dev-station cheatsheet            # just print the hotkey map
mac-dev-station --version
```

## Core types

### `internal/phases/phases.go`

```go
package phases

import "context"

// Phase is one logical step in the bootstrap process.
type Phase interface {
    Name() string                                // "preflight", "brewfile", etc.
    Description() string                         // one-line human description
    Check(ctx context.Context) (Status, error)   // is this phase already satisfied?
    Apply(ctx context.Context) error             // perform the work
}

type Status int

const (
    StatusUnknown   Status = iota
    StatusSatisfied        // nothing to do
    StatusPartial          // some work needed
    StatusMissing          // full apply needed
)

// Registry is the ordered list of phases run by main.
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
```

### `internal/reporter/reporter.go`

Output format (use ANSI codes, no library):

```
==> [3/13] Folders
    PARA work folders        ✓ exists
    OSS workspace            ✓ created
    Code workspace           ✓ created

==> [4/13] Karabiner
    karabiner.json           → backing up existing to ~/.dotfiles-backup-20260512/
                             → writing new config
    Karabiner-Elements       → restarting
    ⚠ Manual step: Grant Driver Extensions permission. See:
      System Settings → Privacy & Security → Driver Extensions
    Press ENTER when done...
```

Methods:
- `Phase(n, total int, name string)`
- `Step(msg string)`
- `OK(msg string)`
- `Skip(msg string)`
- `Warn(msg string)`
- `Error(msg string)`
- `ManualStep(msg, link string)` - opens link with `open` command and waits for ENTER
- `Confirm(msg string) bool` - y/N prompt, defaults to no

## Per-phase implementation notes

### Phase 0: Preflight

Check and report (don't fail unless macOS < 14):
- `sw_vers -productVersion` → reject if < 14
- `uname -m` → expect `arm64`, warn if `x86_64`
- `command -v brew`
- `xcode-select -p`
- `gh auth status`
- Existing configs at: `~/.config/karabiner/karabiner.json`, `~/.config/aerospace/aerospace.toml`, `~/.hammerspoon/init.lua`, `~/.zshrc`, `~/.zsh/aliases.zsh`

If any existing config found → print list, prompt: `These will be backed up to ~/.dotfiles-backup-YYYYMMDD/ before overwriting. Continue? [y/N]`

### Phase 1: Foundations

If brew missing: install via official script. Write `eval "$(/opt/homebrew/bin/brew shellenv)"` to `~/.zprofile`.

If brew present: `brew update && brew upgrade && brew cleanup`.

`xcode-select --install` if missing (opens a GUI dialog, just trigger and continue).

### Phase 2: Brewfile

Embed `configs/Brewfile`. Write to `~/Brewfile`. Run `brew bundle --file=~/Brewfile`. Stream stdout to reporter.

Verify each cask is in `/Applications/`:

```go
caskApps := map[string]string{
    "kitty":              "kitty.app",
    "cursor":             "Cursor.app",
    "raycast":            "Raycast.app",
    "karabiner-elements": "Karabiner-Elements.app",
    "hammerspoon":        "Hammerspoon.app",
    "aerospace":          "AeroSpace.app",
    "orbstack":           "OrbStack.app",
}
```

### Phase 3: Folders

Create with `os.MkdirAll`. Idempotent.

### Phase 4-7: Config phases (Karabiner, AeroSpace, Hammerspoon, kitty)

Same pattern for each:
1. Read embedded config
2. Compare to existing file at target path
3. If different, back up existing to `~/.dotfiles-backup-YYYYMMDD/` preserving directory structure
4. Write new config
5. Restart the relevant app via `osascript`

AeroSpace phase additionally runs:

```go
cmds := []string{
    "defaults write com.apple.dock expose-animation-duration -float 0",
    "defaults write com.apple.dock workspaces-auto-swoosh -bool false",
    "defaults write NSGlobalDomain NSWindowResizeTime -float 0.001",
    "defaults write com.apple.universalaccess reduceMotion -bool true",
    "killall Dock",
}
```

Hammerspoon phase also writes placeholder `~/oss/scripts/layout-docked.sh` and `~/oss/scripts/layout-mobile.sh` with `chmod +x`.

### Phase 8: Shell

Trickiest phase:
1. Clone `git@github.com:oleg-koval/zshrc-backups.git` to `~/code/zshrc-backups` (skip if exists)
2. Symlink `~/code/zshrc-backups/aliases.zsh` → `~/.zsh/aliases.zsh`
3. Write `~/.zsh/secrets.zsh` from template (chmod 600)
4. Write `~/.zshrc` from embedded template
5. Write `~/code/oss/scripts/backup-zsh.sh` (chmod +x)
6. Install crontab entry for 9am daily backup:

```go
// Read current crontab, filter old backup-zsh lines, append new line
existing, _ := exec.Command("crontab", "-l").Output()
lines := strings.Split(string(existing), "\n")
filtered := []string{}
for _, l := range lines {
    if !strings.Contains(l, "backup-zsh") && l != "" {
        filtered = append(filtered, l)
    }
}
filtered = append(filtered,
    "0 9 * * * /bin/bash $HOME/code/oss/scripts/backup-zsh.sh >> $HOME/.zsh/backup.log 2>&1")
// Pipe back to crontab -
```

### Phase 9: Raycast

`open -a Raycast`. Print message that user must accept Cmd+Space takeover and configure Quicklinks manually. List the Quicklink URLs.

### Phase 10: Starters

```bash
git -C ~/code clone git@github.com:oleg-koval/starters.git 2>/dev/null || true
```

Idempotent.

### Phase 11: Permissions walk

```go
type permPrompt struct {
    App     string
    Setting string
    Link    string  // x-apple.systempreferences URL
    Reboot  bool
}

var perms = []permPrompt{
    {"Karabiner-Elements", "Driver Extensions",
     "x-apple.systempreferences:com.apple.preference.security?Privacy_DriverExtension", true},
    {"Karabiner-Elements", "Input Monitoring",
     "x-apple.systempreferences:com.apple.preference.security?Privacy_ListenEvent", false},
    {"AeroSpace", "Accessibility",
     "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility", false},
    {"Hammerspoon", "Accessibility",
     "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility", false},
    {"Raycast", "Accessibility",
     "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility", false},
}
```

For each: open System Settings deep link with `exec.Command("open", p.Link)`, wait for ENTER, mark reboot-required ones at the end with a single combined message.

### Phase 12: Verify (also `mac-dev-station doctor`)

Pure read-only smoke test:

```
==> Verification
    Brew                     ✓ 4.5.2
    Apps installed (7/7)     ✓
    Configs in place (7/7)   ✓
    Aliases sourcable        ✓ 23 functions found
    Cron set                 ✓ 0 9 * * * backup-zsh.sh
    Processes running        ✓ AeroSpace, Hammerspoon, kitty
```

### Phase 13: Cheatsheet

Print formatted hotkey map (also accessible via `mac-dev-station cheatsheet`).

## Embedded files

Use `//go:embed` for everything in `configs/`:

```go
package configs

import _ "embed"

//go:embed Brewfile
var Brewfile []byte

//go:embed karabiner.json
var KarabinerJSON []byte

//go:embed aerospace.toml
var AerospaceTOML []byte

//go:embed hammerspoon/init.lua
var HammerspoonInit []byte

//go:embed hammerspoon/display-watcher.lua
var HammerspoonWatcher []byte

//go:embed kitty/kitty.conf
var KittyConf []byte

//go:embed kitty/one-dark.conf
var KittyTheme []byte

//go:embed kitty/projects.py
var KittyProjectsKitten []byte

//go:embed shell/zshrc
var Zshrc []byte

//go:embed shell/secrets.zsh.template
var SecretsTemplate []byte

//go:embed shell/backup-zsh.sh
var BackupScript []byte
```

The `configs/` directory becomes the source of truth - edit those files, rebuild, ship.

## Makefile additions

The starter already has `build`, `test`, `lint`. Add:

```makefile
.PHONY: install release

install: build
	cp $(BINARY) /opt/homebrew/bin/mac-dev-station

release:
	@test -n "$(VERSION)" || (echo "VERSION=vX.Y.Z required" && exit 1)
	git tag -a $(VERSION) -m "$(VERSION)"
	git push origin $(VERSION)
	@echo "→ GoReleaser workflow will publish via GitHub Actions"
```

## GoReleaser for Homebrew distribution

`.goreleaser.yaml`:

```yaml
version: 2
project_name: mac-dev-station

before:
  hooks:
    - go mod tidy

builds:
  - id: mac-dev-station
    main: ./cmd/mac-dev-station
    binary: mac-dev-station
    env: [CGO_ENABLED=0]
    goos: [darwin]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}

checksum:
  name_template: "checksums.txt"

brews:
  - repository:
      owner: oleg-koval
      name: homebrew-tap
    homepage: "https://github.com/oleg-koval/mac-dev-station"
    description: "One-command macOS dev environment bootstrap"
    license: "MIT"
    test: |
      system "#{bin}/mac-dev-station --version"
    install: |
      bin.install "mac-dev-station"
```

`.github/workflows/release.yml`:

```yaml
name: release
on:
  push:
    tags: ['v*']

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v6
        with:
          go-version: '1.24'
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.TAP_GITHUB_TOKEN }}
```

## One-time tap setup

```bash
gh repo create oleg-koval/homebrew-tap --public --description "Homebrew tap for my tools"
mkdir -p ~/oss/mine/homebrew-tap/Formula
cd ~/oss/mine/homebrew-tap
git init && git remote add origin git@github.com:oleg-koval/homebrew-tap.git
echo "# homebrew-tap" > README.md
git add . && git commit -m "init"
git push -u origin main
```

Create a Personal Access Token with `repo` scope, save as `TAP_GITHUB_TOKEN` in mac-dev-station repo secrets (`gh secret set TAP_GITHUB_TOKEN`).

## End-user install flow

```bash
brew tap oleg-koval/tap
brew install mac-dev-station
mac-dev-station
```

## Build order

Work through this sequence, **committing after each step**:

1. **Bootstrap from go-starter** (above). Confirm `make test` passes.
2. **Add cobra**, create `cmd/mac-dev-station/main.go` with `--version` working.
3. **Build `reporter` package**, write a tiny test program exercising all methods.
4. **Define `Phase` interface and `Registry`** (empty implementations).
5. **Implement `FoldersPhase` end-to-end** as the reference pattern.
6. **Implement `PreflightPhase`** - the gatekeeper, exercises confirm prompts.
7. **Implement `BrewfilePhase`** - first phase to use the embed pattern.
8. **Implement `KarabinerPhase`** - establishes the backup + write + restart pattern.
9. **Copy that pattern for AeroSpace, Hammerspoon, kitty.**
10. **Implement `ShellPhase`** - the trickiest one (git clone + symlinks + crontab).
11. **Implement `RaycastPhase`, `StartersPhase`** - thin phases.
12. **Implement `PermissionsWalkPhase`** - the manual-step pattern.
13. **Implement `VerifyPhase`** + `doctor` subcommand.
14. **Implement `CheatsheetPhase`** + `cheatsheet` subcommand.
15. **GoReleaser config + release workflow.**
16. **Test on a fresh user account** (`sudo dscl . -create /Users/test`).
17. **Tag v0.1.0**, release, install via brew, verify end-to-end.

## Operating principles

1. **Idempotent first** - every phase checks state before applying
2. **Backup before overwriting** - `~/.dotfiles-backup-YYYYMMDD/` with original directory structure preserved
3. **No silent failures** - every shell-out is logged, every error is wrapped with context
4. **Manual steps clearly marked** - they're not failures, they're intentional pauses
5. **Dry-run must work** - `--dry-run` shows the plan without touching anything
6. **`--skip` and `--only` as escape hatches** - never force-run a phase the user disabled

## What I'll handle manually after install

- Run `displayplacer list` while docked, paste into `~/oss/scripts/layout-docked.sh`
- Fill `~/.zsh/secrets.zsh` with API keys
- Sign in: gh, 1Password CLI, Cursor, Raycast (optional), Atuin (optional)
- Configure Raycast Quicklinks (GUI-only step)

## Start now

Report back after **step 1** with:
- Confirmation the repo is created at `github.com/oleg-koval/mac-dev-station`
- Module renamed (grep should find zero `go-starter` references)
- `make test` passes
- Initial commit pushed
