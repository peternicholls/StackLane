package guidance

import (
	"fmt"
	"io"
)

// RenderText writes the plain text fallback for a next-action plan.
func RenderText(w io.Writer, plan NextActionPlan) error {
	if _, err := fmt.Fprintf(w, "StageServe\n%s\n\n%s\n", severityLabel(plan.Situation), plan.StatusHeader); err != nil {
		return err
	}
	if plan.Summary != "" {
		if _, err := fmt.Fprintf(w, "%s\n", plan.Summary); err != nil {
			return err
		}
	}
	if len(plan.VisibleDefaults) > 0 {
		if _, err := fmt.Fprintln(w, "\nKey facts:"); err != nil {
			return err
		}
		for _, item := range plan.VisibleDefaults {
			if item.Note != "" {
				if _, err := fmt.Fprintf(w, "  %-13s %s (%s)\n", item.Label+":", item.Value, item.Note); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(w, "  %-13s %s\n", item.Label+":", item.Value); err != nil {
				return err
			}
		}
	}
	if len(plan.DecisionItems) > 0 {
		if _, err := fmt.Fprintln(w, "\nWhat you can do:"); err != nil {
			return err
		}
		for i, action := range plan.DecisionItems {
			if action.ID == "more" {
				continue
			}
			if _, err := fmt.Fprintf(w, "  %d. %s\n     %s\n", i+1, action.Label, action.Description); err != nil {
				return err
			}
			if action.RequiresConfirmation {
				if _, err := fmt.Fprintf(w, "     Confirmation: %s\n", confirmationSummary(action, plan)); err != nil {
					return err
				}
			}
		}
	}
	if len(plan.WorkItems) > 0 {
		if _, err := fmt.Fprintln(w, "\nNext step:"); err != nil {
			return err
		}
		for _, item := range plan.WorkItems {
			if item.DirectCommand != "" {
				if _, err := fmt.Fprintf(w, "  %s: %s\n", item.Label, item.DirectCommand); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(w, "  %s\n", item.Label); err != nil {
				return err
			}
		}
	}
	if len(plan.Warnings) > 0 {
		if _, err := fmt.Fprintln(w, "\nNeeds attention:"); err != nil {
			return err
		}
		for _, warning := range plan.Warnings {
			if _, err := fmt.Fprintf(w, "  %s\n", warning); err != nil {
				return err
			}
		}
	}
	if len(plan.DirectCommands) > 0 {
		if _, err := fmt.Fprintln(w, "\nMore:"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "  Direct commands:"); err != nil {
			return err
		}
		for _, command := range plan.DirectCommands {
			if _, err := fmt.Fprintf(w, "    %s\n", command); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, "  Plain text output: stage --cli"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "  Advanced and troubleshooting: stage doctor"); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}
