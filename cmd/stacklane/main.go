// stacklane CLI entrypoint. Wires cobra subcommands to the lifecycle
// orchestrator.
package main

import (
	"fmt"
	"os"

	"github.com/peternicholls/stacklane/cmd/stacklane/commands"
)

// version is overridden at build time via -ldflags.
var version = "dev"

func main() {
	root := commands.NewRoot(version)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
