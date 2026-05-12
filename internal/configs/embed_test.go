package configs

import "testing"

func TestEmbeddedConfigs(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"Brewfile", BrewfileContent},
		{"karabiner.json", KarabinerContent},
		{"aerospace.toml", AerospaceContent},
		{"hammerspoon/init.lua", HammerspoonInitContent},
		{"hammerspoon/display-watcher.lua", HammerspoonDisplayWatcherContent},
		{"kitty/kitty.conf", KittyConfContent},
		{"kitty/one-dark.conf", KittyColorSchemeContent},
		{"kitty/projects.py", KittyProjectsContent},
		{"shell/zshrc", ZshrcContent},
		{"shell/secrets.zsh.template", SecretsTemplateContent},
		{"shell/backup-zsh.sh", BackupScriptContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.data) == 0 {
				t.Errorf("embedded config %s is empty", tt.name)
			}
		})
	}
}

func TestGetConfigFile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"Brewfile", true},
		{"karabiner.json", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetConfigFile(tt.name)
			hasContent := len(got) > 0
			if hasContent != tt.want {
				t.Errorf("GetConfigFile(%q) returned %v, want %v", tt.name, len(got), tt.want)
			}
		})
	}
}
