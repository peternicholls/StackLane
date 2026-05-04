// stacklane CLI entrypoint. Wires cobra subcommands to the lifecycle
// orchestrator.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/peternicholls/stacklane/cmd/stacklane/commands"
)

// version is overridden at build time via -ldflags.
var version = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := commands.NewRoot(version)
	if err := root.ExecuteContext(ctx); err != nil {
		var exitCoder commands.ExitCoder
		if errors.As(err, &exitCoder) {
			if !exitCoder.Silent() {
				fmt.Fprintln(os.Stderr, err)
			}
			os.Exit(exitCoder.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
