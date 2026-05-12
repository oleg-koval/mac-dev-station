# mac-dev-station

> One-command dev environment setup for macOS. Built for Apple Silicon, macOS 14+.

Replicates a complete productivity stack on a fresh Mac:
**Karabiner Hyper key · AeroSpace tiling WM · Hammerspoon display auto-detect · kitty · Raycast · zsh + starship + modern CLI · PARA folders · backup repo wiring**

## Install via Homebrew

```bash
brew tap oleg-koval/tap
brew install mac-dev-station
mac-dev-station
```

## Install without Homebrew

```bash
curl -fsSL https://raw.githubusercontent.com/oleg-koval/mac-dev-station/main/install.sh | bash
```

## What it does

The script runs **13 phases**, each idempotent:

| # | Phase | Outcome |
|---|-------|---------|
| 0 | Pre-flight | macOS version, brew, xcode-select, gh auth check |
| 1 | Foundations | Install/update Homebrew, xcode-select |
| 2 | Brewfile | Install all CLI tools and GUI apps |
| 3 | Folders | Create `~/Work` (PARA), `~/oss`, `~/code` |
| 4 | Karabiner | Hyper key config + app launchers |
| 5 | AeroSpace | Tiling WM with auto-routing to workspaces |
| 6 | Hammerspoon | Display dock/undock auto-detection |
| 7 | kitty | Terminal config + project switcher (Cmd+P) |
| 8 | Shell | zsh + starship + aliases + 9am backup cron |
| 9 | Raycast | Open for first-time setup (Cmd+Space) |
| 10 | OSS starters | Clone `oleg-koval/starters` |
| 11 | Permissions | Walks you through Accessibility, Driver Extensions |
| 12 | Verification | Final smoke test of every component |
| 13 | Cheatsheet | Prints all hotkeys |

## What gets installed

**CLI tools** (via brew):
- `git`, `gh`, `fnm`, `uv`, `starship`, `zoxide`, `atuin`, `fzf`, `ripgrep`, `fd`, `bat`, `eza`, `glow`, `btop`, `jq`, `yq`, `delta`, `tmux`, `neovim`, `displayplacer`

**GUI apps** (via cask):
- `kitty`, `cursor`, `raycast`, `karabiner-elements`, `hammerspoon`, `aerospace`, `orbstack`, `1password-cli`

**Fonts**:
- Geist Mono, JetBrains Mono Nerd Font

## Hotkey map (after install)

### Hyper Key (Caps Lock)

| Key | Action |
|-----|--------|
| `Caps tap` | Escape |
| `Caps+T` | Workspace 3 → kitty |
| `Caps+B` | Workspace 1 → Chrome |
| `Caps+C` | Workspace 2 → Cursor |
| `Caps+S` | Workspace 4 → Slack |
| `Caps+L` | Workspace 4 → Linear |
| `Caps+F` | Workspace 5 → Figma |
| `Caps+N` | Workspace 6 → Notion |
| `Caps+H/J/K/;` | Arrow keys |

### AeroSpace (Alt prefix)

| Key | Action |
|-----|--------|
| `Alt+1-9` | Switch workspace |
| `Alt+Shift+1-9` | Move window to workspace |
| `Alt+H/J/K/L` | Focus window (vim) |
| `Alt+Shift+H/J/K/L` | Move window |
| `Alt+F` | Fullscreen |
| `Alt+/` | Tiling layout |
| `Alt+,` | Accordion layout |

### Hammerspoon

| Key | Action |
|-----|--------|
| `Cmd+Ctrl+Alt+D` | Force apply layout |
| (auto) | Detects monitor connect/disconnect with 9s debounce |

### kitty

| Key | Action |
|-----|--------|
| `Cmd+P` | Project switcher (fzf) |
| `Cmd+T` | New tab |
| `Cmd+D` | Vertical split |
| `Cmd+Shift+D` | Horizontal split |

## Shell aliases (after install)

```bash
dev <name>      # Full env: Cursor + kitty + GitHub in correct workspaces
work <name>     # kitty only, workspace 3
proj            # fzf project picker → cd
go              # fzf picker → full dev env
newproj <name>  # Scaffold new (project|area|scratch|oss)
sync_projects   # Refresh kitty picker
dock / undock   # Manual display layout
dclean          # Docker compose nuke + rebuild
```

## Manual steps after install

The script will pause and walk you through:

1. **Driver Extensions** for Karabiner (reboot required after first grant)
2. **Accessibility** for AeroSpace, Hammerspoon, Raycast
3. **Cmd+Space rebind** from Spotlight to Raycast
4. **Run `displayplacer list`** while docked, paste output into `~/oss/scripts/layout-docked.sh`
5. **Fill `~/.zsh/secrets.zsh`** with your API keys
6. **Sign in** to: gh, 1Password CLI, Cursor, Raycast (optional), Atuin (optional)

## Safety

- **Idempotent** - safe to re-run
- **Backups** - any existing config moved to `~/.dotfiles-backup-YYYYMMDD/`
- **Confirms before overwriting** - asks first if a config exists
- **No magic** - prints every command before running

## Uninstall

```bash
brew uninstall mac-dev-station
brew untap oleg-koval/tap
```

This removes only the bootstrap CLI - your installed apps and configs stay intact.

## Roadmap

- [ ] Optional flags: `--skip-aerospace`, `--skip-hammerspoon`, etc
- [ ] Dotfiles diff: detect drift between installed config and current repo version
- [ ] Update mode: `mac-dev-station update` pulls latest configs and re-applies

## License

MIT - see [LICENSE](./LICENSE)
