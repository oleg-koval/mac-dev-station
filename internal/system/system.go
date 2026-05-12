package system

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// MacOSVersion returns the current macOS version as major.minor
func MacOSVersion(ctx context.Context) (float64, error) {
	output, err := RunCmd(ctx, "sw_vers", "-productVersion")
	if err != nil {
		return 0, err
	}

	// Parse output like "14.6.1" and return 14.6
	parts := strings.Split(strings.TrimSpace(output), ".")
	if len(parts) < 2 {
		return 0, fmt.Errorf("unexpected macOS version format: %s", output)
	}

	major, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse macOS major version: %w", err)
	}

	minor, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse macOS minor version: %w", err)
	}

	return major + (minor / 10.0), nil
}

// Arch returns the system architecture (arm64 or amd64)
func Arch() string {
	return runtime.GOARCH
}

// IsARM64 returns true if running on Apple Silicon
func IsARM64() bool {
	return runtime.GOARCH == "arm64"
}

// IsIntel returns true if running on Intel Mac
func IsIntel() bool {
	return runtime.GOARCH == "amd64"
}
