package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/oleg-koval/mac-dev-station/internal/phases"
	"github.com/oleg-koval/mac-dev-station/internal/reporter"
)

var version = "dev"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mac-dev-station",
		Short:   "One-command macOS dev environment setup",
		Long:    "mac-dev-station: Replicate a complete productivity stack on a fresh Mac with Karabiner, AeroSpace, Hammerspoon, kitty, Raycast, zsh, and more.",
		Version: version,
		RunE: runSetup,
	}

	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newCheatsheetCmd())

	return cmd
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run verification only (phase 13)",
		Long:  "Verify that all components are installed and configured correctly.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			rep := reporter.New(false)

			verifyPhase := phases.Registry[12] // VerifyPhase is at index 12 (14th phase)
			rep.Phase(13, len(phases.Registry), verifyPhase.Name())
			rep.Step(verifyPhase.Description())

			status, err := verifyPhase.Check(ctx)
			if err != nil {
				rep.Error(err.Error())
				return err
			}

			switch status {
			case phases.StatusSatisfied:
				rep.OK("All components verified")
			case phases.StatusPartial:
				rep.Warn("Some components missing")
				if err := verifyPhase.Apply(ctx); err != nil {
					rep.Error(err.Error())
				}
			case phases.StatusMissing:
				rep.Error("Critical components missing")
			}

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
			ctx := context.Background()
			cheatsheetPhase := phases.Registry[13] // CheatsheetPhase is at index 13 (14th phase)
			return cheatsheetPhase.Apply(ctx)
		},
	}
}

func runSetup(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	rep := reporter.New(false)

	fmt.Println("==> mac-dev-station setup")

	for i, phase := range phases.Registry {
		rep.Phase(i+1, len(phases.Registry), phase.Name())
		rep.Step(phase.Description())

		status, err := phase.Check(ctx)
		if err != nil {
			rep.Error(fmt.Sprintf("%v", err))
			continue
		}

		switch status {
		case phases.StatusSatisfied:
			rep.OK("Already configured")
		case phases.StatusPartial:
			rep.Warn("Partial setup, continuing")
			if err := phase.Apply(ctx); err != nil {
				rep.Error(fmt.Sprintf("%v", err))
			} else {
				rep.OK("Applied")
			}
		case phases.StatusMissing:
			rep.Step("Applying...")
			if err := phase.Apply(ctx); err != nil {
				rep.Error(fmt.Sprintf("%v", err))
			} else {
				rep.OK("Applied")
			}
		case phases.StatusUnknown:
			rep.Step("Checking...")
			if err := phase.Apply(ctx); err != nil {
				rep.Error(fmt.Sprintf("%v", err))
			} else {
				rep.OK("Applied")
			}
		}
	}

	fmt.Println("\n==> Setup complete!")
	return nil
}

func main() {
	root := newRootCmd()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
