package commands

import (
	"context"
	"encoding/json"
	"os"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/guidance"
	"github.com/spf13/cobra"
)

type guidancePlanView struct {
	Situation       guidance.Situation        `json:"situation"`
	StatusHeader    string                    `json:"status_header"`
	Summary         string                    `json:"summary,omitempty"`
	DecisionItems   []guidance.GuidedAction   `json:"decision_items,omitempty"`
	WorkItems       []guidance.WorkItem       `json:"work_items,omitempty"`
	VisibleDefaults []guidance.VisibleDefault `json:"visible_defaults,omitempty"`
	DirectCommands  []string                  `json:"direct_commands,omitempty"`
	Warnings        []string                  `json:"warnings,omitempty"`
}

func NewGuidancePlan(shared *SharedFlags) *cobra.Command {
	var skipReadiness bool
	cmd := &cobra.Command{
		Use:    "guidance-plan",
		Hidden: true,
		Short:  "Inspect the current guided plan",
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			capability := guidance.DetectCapability(os.Stdin, os.Stdout, os.Stderr, true, true)
			cfg, err := loadConfig(shared)
			var plan guidance.NextActionPlan
			if err != nil {
				cwd, cwdErr := os.Getwd()
				if cwdErr != nil {
					cwd = "."
				}
				plan = guidance.Plan(guidance.ContextFromError(cwd, capability, err))
			} else {
				opts := guidance.CollectOptions{
					Capability:    capability,
					RuntimeStatus: guidedRuntimeStatus,
				}
				if !skipReadiness {
					opts.MachineReadiness = func(_ context.Context, cfg config.ProjectConfig) guidance.MachineReadinessSummary {
						return guidance.MachineReadinessFromResult(buildMachineReadinessResultForConfig(cfg))
					}
				}
				collected := guidance.Collect(cmd.Context(), cfg, opts)
				plan = guidance.Plan(collected)
			}
			view := guidancePlanView{
				Situation:       plan.Situation,
				StatusHeader:    plan.StatusHeader,
				Summary:         plan.Summary,
				DecisionItems:   plan.DecisionItems,
				WorkItems:       plan.WorkItems,
				VisibleDefaults: plan.VisibleDefaults,
				DirectCommands:  plan.DirectCommands,
				Warnings:        plan.Warnings,
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(view)
		},
	}
	cmd.Flags().BoolVar(&skipReadiness, "skip-readiness", false, "Skip setup checks in the planner inspection output")
	_ = cmd.Flags().MarkHidden("skip-readiness")
	return cmd
}
