package configs

import _ "embed"

//go:embed Brewfile
var BrewfileContent []byte

//go:embed karabiner.json
var KarabinerContent []byte

//go:embed aerospace.toml
var AerospaceContent []byte

//go:embed hammerspoon/init.lua
var HammerspoonInitContent []byte

//go:embed hammerspoon/display-watcher.lua
var HammerspoonDisplayWatcherContent []byte

//go:embed kitty/kitty.conf
var KittyConfContent []byte

//go:embed kitty/one-dark.conf
var KittyColorSchemeContent []byte

//go:embed kitty/projects.py
var KittyProjectsContent []byte

//go:embed shell/zshrc
var ZshrcContent []byte

//go:embed shell/secrets.zsh.template
var SecretsTemplateContent []byte

//go:embed shell/backup-zsh.sh
var BackupScriptContent []byte

// GetConfigFile returns the embedded config file content by name
func GetConfigFile(name string) []byte {
	switch name {
	case "Brewfile":
		return BrewfileContent
	case "karabiner.json":
		return KarabinerContent
	case "aerospace.toml":
		return AerospaceContent
	case "hammerspoon/init.lua":
		return HammerspoonInitContent
	case "hammerspoon/display-watcher.lua":
		return HammerspoonDisplayWatcherContent
	case "kitty/kitty.conf":
		return KittyConfContent
	case "kitty/one-dark.conf":
		return KittyColorSchemeContent
	case "kitty/projects.py":
		return KittyProjectsContent
	case "shell/zshrc":
		return ZshrcContent
	case "shell/secrets.zsh.template":
		return SecretsTemplateContent
	case "shell/backup-zsh.sh":
		return BackupScriptContent
	}
	return nil
}
