package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mac-dev-station",
		Short:   "One-command macOS dev environment setup",
		Long:    "mac-dev-station: Replicate a complete productivity stack on a fresh Mac with Karabiner, AeroSpace, Hammerspoon, kitty, Raycast, zsh, and more.",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup(cmd, args)
		},
	}

	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newCheatsheetCmd())

	return cmd
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run verification only (phase 12)",
		Long:  "Verify that all components are installed and configured correctly.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("==> doctor: verification only")
			return nil
		},
	}
}

func newCheatsheetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cheatsheet",
		Short: "Print hotkey map",
		Long:  "Display all keyboard shortcuts and aliases.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("==> Cheatsheet")
			fmt.Println("Hyper Key (Caps Lock)")
			fmt.Println("  Caps tap      → Escape")
			fmt.Println("  Caps+T        → Workspace 3 (kitty)")
			fmt.Println("  Caps+B        → Workspace 1 (Chrome)")
			fmt.Println("  Caps+C        → Workspace 2 (Cursor)")
			fmt.Println("  Caps+S        → Workspace 4 (Slack)")
			fmt.Println("  Caps+L        → Workspace 4 (Linear)")
			fmt.Println("  Caps+F        → Workspace 5 (Figma)")
			fmt.Println("  Caps+N        → Workspace 6 (Notion)")
			fmt.Println("  Caps+H/J/K/;  → Arrow keys")
			return nil
		},
	}
}

func runSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("mac-dev-station: setup phases")
	fmt.Println("(implementation in progress)")
	return nil
}

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
