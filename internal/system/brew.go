// Package system provides macOS-specific helpers: command execution, app detection, and Homebrew utilities.
package system

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// BrewInstalled checks if Homebrew is installed
func BrewInstalled(ctx context.Context) bool {
	_, err := RunCmd(ctx, "brew", "--version")
	return err == nil
}

// BrewUpdate runs brew update
func BrewUpdate(ctx context.Context, out io.Writer) error {
	return RunCmdStream(ctx, out, "brew", "update")
}

// BrewUpgrade runs brew upgrade
func BrewUpgrade(ctx context.Context, out io.Writer) error {
	return RunCmdStream(ctx, out, "brew", "upgrade")
}

// BrewCleanup runs brew cleanup
func BrewCleanup(ctx context.Context, out io.Writer) error {
	return RunCmdStream(ctx, out, "brew", "cleanup")
}

// BrewBundle runs brew bundle with a Brewfile
func BrewBundle(ctx context.Context, out io.Writer, brewfilePath string) error {
	return RunCmdStream(ctx, out, "brew", "bundle", "--file="+brewfilePath)
}

// BrewList returns list of installed packages
func BrewList(ctx context.Context) ([]string, error) {
	output, err := RunCmd(ctx, "brew", "list")
	if err != nil {
		return nil, err
	}

	var packages []string
	for _, line := range strings.Split(output, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			packages = append(packages, trimmed)
		}
	}
	return packages, nil
}

// AppInstalled checks if an application is installed in /Applications
func AppInstalled(ctx context.Context, appName string) bool {
	out, err := RunCmd(ctx, "ls", "-d", fmt.Sprintf("/Applications/%s.app", appName))
	return err == nil && out != ""
}
