// stage dns-setup: bootstrap the local DNS so .test hostnames resolve.
// Returns a typed message on platforms where the
// flow is unsupported (FR-012).
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/peternicholls/stageserve/platform/dns"
)

func NewDNSSetup(flags *SharedFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "dns-setup",
		Short: "Configure local DNS for .test hostnames",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(flags)
			if err != nil {
				return err
			}
			settings := dns.Settings{
				Provider: cfg.LocalDNS.Provider,
				IP:       cfg.LocalDNS.IP,
				Port:     cfg.LocalDNS.Port,
				Suffix:   cfg.LocalDNS.Suffix,
				StateDir: cfg.StateDir,
			}
			provider := dns.NewProvider()
			if status := provider.Status(settings); status.Code == dns.CodeReady {
				fmt.Fprintln(os.Stdout, "local DNS already ready: "+status.Message)
				return nil
			}
			if err := provider.Bootstrap(settings); err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "local DNS ready")
			return nil
		},
	}
}
