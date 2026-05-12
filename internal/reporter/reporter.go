package reporter

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Status int

const (
	StatusUnknown Status = iota
	StatusSatisfied
	StatusPartial
	StatusMissing
)

type Reporter struct {
	dryRun bool
}

func New(dryRun bool) *Reporter {
	return &Reporter{dryRun: dryRun}
}

func (r *Reporter) Phase(n, total int, name string) {
	fmt.Printf("\n==> [%d/%d] %s\n", n, total, name)
}

func (r *Reporter) Step(msg string) {
	fmt.Printf("    %s\n", msg)
}

func (r *Reporter) OK(msg string) {
	fmt.Printf("    %s  %s\n", msg, greenCheck())
}

func (r *Reporter) Skip(msg string) {
	fmt.Printf("    %s  %s\n", msg, yellowSkip())
}

func (r *Reporter) Warn(msg string) {
	fmt.Printf("    %s ⚠️\n", msg)
}

func (r *Reporter) Error(msg string) {
	fmt.Printf("    %s  %s\n", msg, redX())
}

func (r *Reporter) ManualStep(msg, link string) {
	r.Warn("Manual step: " + msg)
	if link != "" {
		r.Step("See: " + link)
		if !r.dryRun {
			_ = exec.Command("open", link).Run()
		}
	}
	r.promptEnter()
}

func (r *Reporter) Confirm(msg string) bool {
	if r.dryRun {
		return true
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("    %s [y/N] ", msg)
	resp, _ := reader.ReadString('\n')
	return strings.ToLower(strings.TrimSpace(resp)) == "y"
}

func (r *Reporter) promptEnter() {
	if r.dryRun {
		return
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("    Press ENTER when done...")
	_, _ = reader.ReadString('\n')
}

func greenCheck() string { return "\033[32m✓\033[0m" }
func yellowSkip() string { return "\033[33m↓\033[0m" }
func redX() string       { return "\033[31m✗\033[0m" }
