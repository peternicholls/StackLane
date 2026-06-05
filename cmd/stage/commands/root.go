// Cobra root + shared flag plumbing. Every subcommand attaches to the root
// returned by NewRoot.
package commands

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/guidance"
	"github.com/peternicholls/stageserve/core/lifecycle"
	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/compose"
	"github.com/peternicholls/stageserve/infra/docker"
	"github.com/peternicholls/stageserve/infra/gateway"
	"github.com/peternicholls/stageserve/platform/ports"
	stls "github.com/peternicholls/stageserve/platform/tls"
)

// SharedFlags is bound to the root command and inherited by every subcommand.
type SharedFlags struct {
	ProjectDir      string
	SiteName        string
	SiteHostname    string
	SiteSuffix      string
	DocRoot         string
	PHPVersion      string
	MySQLDatabase   string
	MySQLUser       string
	MySQLPort       string
	PMAPort         string
	HostPort        string
	WaitTimeoutSecs int
	DryRun          bool
	Profile         []string
	All             bool
	StackHome       string
}

type rootInteractionFlags struct {
	NotUI bool
	CLI   bool
}

// NewRoot returns the configured root command.
func NewRoot(version string) *cobra.Command {
	flags := &SharedFlags{}
	interaction := &rootInteractionFlags{}
	root := &cobra.Command{
		Use:           "stage",
		Short:         "Run local projects with StageServe",
		Long:          "StageServe runs local project environments and keeps their browser URLs routed consistently. Run stage with no subcommand for guided next steps, or use a direct command when you already know what you want to do.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRootGuidance(cmd, flags, interaction)
		},
	}
	root.Flags().BoolVar(&interaction.NotUI, "notui", false, "Use plain text output")
	root.Flags().BoolVar(&interaction.CLI, "cli", false, "Use plain text output")
	pf := root.PersistentFlags()
	pf.StringVar(&flags.ProjectDir, "project-dir", "", "Project directory (defaults to cwd)")
	pf.StringVar(&flags.SiteName, "site-name", "", "Project name (defaults to project dir basename)")
	pf.StringVar(&flags.SiteHostname, "site-hostname", "", "Hostname (defaults to <slug>.<suffix>)")
	pf.StringVar(&flags.SiteSuffix, "site-suffix", "", "Site suffix (default: test)")
	pf.StringVar(&flags.DocRoot, "docroot", "", "Document root inside the project")
	pf.StringVar(&flags.PHPVersion, "php-version", "", "PHP version (e.g. 8.5)")
	pf.StringVar(&flags.MySQLDatabase, "mysql-database", "", "MySQL database name")
	pf.StringVar(&flags.MySQLUser, "mysql-user", "", "MySQL user")
	pf.StringVar(&flags.MySQLPort, "mysql-port", "", "MySQL host port")
	pf.StringVar(&flags.PMAPort, "pma-port", "", "phpMyAdmin host port")
	pf.StringVar(&flags.HostPort, "host-port", "", "HTTP host port (advanced)")
	pf.IntVar(&flags.WaitTimeoutSecs, "wait-timeout", 0, "Healthcheck wait timeout in seconds (default 120)")
	pf.BoolVar(&flags.DryRun, "dry-run", false, "Print planned actions without executing")
	pf.StringSliceVar(&flags.Profile, "profile", nil, "Compose profile to enable (currently: debug)")
	pf.BoolVar(&flags.All, "all", false, "Apply to every recorded project")
	pf.StringVar(&flags.StackHome, "stack-home", "", "Path to the stageserve install (default: auto)")

	root.AddCommand(NewUp(flags))
	root.AddCommand(NewDown(flags))
	root.AddCommand(NewDetach(flags))
	root.AddCommand(NewAttach(flags))
	root.AddCommand(NewStatus(flags))
	root.AddCommand(NewLogs(flags))
	root.AddCommand(NewDNSSetup(flags))
	root.AddCommand(NewSetup(flags))
	root.AddCommand(NewDoctor(flags))
	root.AddCommand(NewInit(flags))
	root.AddCommand(NewGuidancePlan(flags))
	root.AddCommand(NewVersion(version))
	return root
}

func runRootGuidance(cmd *cobra.Command, flags *SharedFlags, interaction *rootInteractionFlags) error {
	capability := guidance.DetectCapability(os.Stdin, os.Stdout, os.Stderr, interaction.NotUI, interaction.CLI)
	cfg, err := loadConfig(flags)
	if err != nil {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			cwd = "."
		}
		plan := guidance.Plan(guidance.ContextFromError(cwd, capability, err))
		return guidance.RenderText(cmd.OutOrStdout(), plan)
	}
	context := collectRootGuidedContext(cmd.Context(), cfg, capability)
	plan := guidance.Plan(context)
	if capability.AllowsTUI() {
		return runGuidedTUI(cmd.Context(), cfg, plan, capability, cmd.OutOrStdout())
	}
	return guidance.RenderText(cmd.OutOrStdout(), plan)
}

// loadConfig produces a ProjectConfig honouring the precedence chain.
func loadConfig(flags *SharedFlags) (config.ProjectConfig, error) {
	loader := config.NewLoader()
	loader.StackHomeOverride = flags.StackHome
	cli := config.CLIFlags{
		ProjectDir:      flags.ProjectDir,
		SiteName:        flags.SiteName,
		SiteHostname:    flags.SiteHostname,
		SiteSuffix:      flags.SiteSuffix,
		DocRoot:         flags.DocRoot,
		PHPVersion:      flags.PHPVersion,
		MySQLDatabase:   flags.MySQLDatabase,
		MySQLUser:       flags.MySQLUser,
		MySQLPort:       flags.MySQLPort,
		PMAPort:         flags.PMAPort,
		HostPort:        flags.HostPort,
		WaitTimeoutSecs: flags.WaitTimeoutSecs,
	}
	return loader.Load(flags.ProjectDir, cli)
}

// buildOrchestrator wires the production set of dependencies.
func buildOrchestrator(cfg config.ProjectConfig) (*lifecycle.Orchestrator, error) {
	store, err := state.NewStore(cfg.StateDir)
	if err != nil {
		return nil, err
	}
	return lifecycle.New(lifecycle.Deps{
		Docker:  docker.NewSDKClient(),
		Compose: compose.NewCLI(),
		Gateway: gateway.NewManager(cfg.SharedGateway.ConfigFile),
		State:   store,
		Ports:   ports.NewAllocator(cfg.StateDir),
		TLS:     stls.NewMkcert(),
	}), nil
}
