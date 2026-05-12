package system

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// RunCmd executes a command and returns stdout and stderr as strings
func RunCmd(ctx context.Context, cmd string, args ...string) (string, error) {
	logCommand(cmd, args)

	c := exec.CommandContext(ctx, cmd, args...)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	if err := c.Run(); err != nil {
		return "", fmt.Errorf("command failed: %s %v: %w\nstderr: %s", cmd, args, err, stderr.String())
	}

	return stdout.String(), nil
}

// RunCmdStream executes a command and streams output to io.Writer in real-time
func RunCmdStream(ctx context.Context, out io.Writer, cmd string, args ...string) error {
	logCommand(cmd, args)

	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdout = out
	c.Stderr = out
	c.Stdin = os.Stdin

	return c.Run()
}

// logCommand logs a command execution attempt
func logCommand(cmd string, args []string) {
	fullCmd := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	fmt.Fprintf(os.Stderr, "[system] %s\n", fullCmd)
}
