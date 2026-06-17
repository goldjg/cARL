// Package cmdutil defines shared types for cARL CLI commands.
package cmdutil

import "context"

// Command is the interface all cARL CLI commands must implement.
// Adding a new command only requires implementing this interface and
// registering it in cmd/carl/main.go.
type Command interface {
	// Name returns the command name as it appears on the command line.
	Name() string
	// Synopsis returns a short one-line description shown in usage output.
	Synopsis() string
	// Run executes the command. args contains any arguments after the
	// command name. Non-nil errors are printed to stderr by the dispatcher.
	Run(ctx context.Context, args []string) error
}
