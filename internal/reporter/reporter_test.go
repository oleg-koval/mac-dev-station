package reporter

import "testing"

func TestReporter(t *testing.T) {
	r := New(true)
	r.Phase(1, 13, "Preflight")
	r.Step("Checking macOS version...")
	r.OK("macOS 14.6")
	r.Skip("Skipped: xcode-select")
	r.Warn("Manual step required")
	r.Error("Failed to install brew")
}
