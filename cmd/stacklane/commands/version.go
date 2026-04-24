// stacklane version: print the build-time version stamped via -ldflags.
package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewVersion(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the stacklane version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}
