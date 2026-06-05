// Package guidance owns bare-stage next-action planning. It collects cheap
// context, classifies the user's situation, and returns a renderer-neutral plan.
package guidance

import (
	"github.com/peternicholls/stageserve/core/onboarding"
	"github.com/peternicholls/stageserve/core/state"
)

// Situation is the planner's stable classification of the current context.
type Situation string

const (
	SituationMachineNotReady    Situation = "machine_not_ready"
	SituationProjectMissingConf Situation = "project_missing_config"
	SituationProjectReadyToRun  Situation = "project_ready_to_run"
	SituationProjectRunning     Situation = "project_running"
	SituationProjectDown        Situation = "project_down"
	SituationDriftDetected      Situation = "drift_detected"
	SituationNotProject         Situation = "not_project"
	SituationUnknownError       Situation = "unknown_error"
)

// TUICapability records the terminal facts that decide whether a TUI may run.
type TUICapability struct {
	StdinTTY      bool
	StdoutTTY     bool
	StderrTTY     bool
	NotUIFlag     bool
	CLIFlag       bool
	NoTUIShellEnv bool
	NoColor       bool
	Term          string
	Reason        string
}

// AllowsTUI reports whether the current invocation can use the interactive UI.
func (c TUICapability) AllowsTUI() bool {
	return c.StdinTTY && c.StdoutTTY && !c.NotUIFlag && !c.CLIFlag && !c.NoTUIShellEnv
}

// MachineReadinessSummary is intentionally compact so context collection can
// stay cheap and tests can inject expensive check results only when needed.
type MachineReadinessSummary struct {
	Checked      bool
	Blocked      bool
	Status       string
	NextFixLabel string
	NextCommand  string
	WorkItems    []WorkItem
}

// RuntimeSummary is the cheap runtime view used by the planner.
type RuntimeSummary struct {
	Checked  bool
	Running  bool
	Status   string
	Services []RuntimeServiceSummary
}

// RuntimeServiceSummary is the normalized service metadata the guided flow may
// use for user-goal actions. It intentionally avoids raw container dumps.
type RuntimeServiceSummary struct {
	ServiceName        string
	ContainerName      string
	Status             string
	EligibleForLogs    bool
	EligibleForRestart bool
}

// ServiceSelectionDecision records deterministic routing for service-scoped
// actions such as logs and restart.
type ServiceSelectionDecision struct {
	Action          string
	Mode            string
	SelectedService RuntimeServiceSummary
	Services        []RuntimeServiceSummary
	Reason          string
}

// GuidedContext is a no-mutation snapshot of the current StageServe situation.
type GuidedContext struct {
	CWD               string
	ProjectRoot       string
	ProjectEnvPath    string
	ProjectEnvExists  bool
	ProjectEnvValid   bool
	StackHome         string
	StackID           string
	StateDir          string
	Hostname          string
	LocalURL          string
	WebFolder         string
	SiteName          string
	SiteSuffix        string
	ProjectEnvPreview *onboarding.ProjectEnvPreview
	MachineReadiness  MachineReadinessSummary
	ProjectState      *state.Record
	Runtime           RuntimeSummary
	Capability        TUICapability
	NotProject        bool
	Warnings          []string
	Err               error
}

// NextActionPlan is the renderer-neutral plan consumed by TUI and text output.
type NextActionPlan struct {
	Situation       Situation
	Title           string
	Summary         string
	StatusHeader    string
	DecisionItems   []GuidedAction
	WorkItems       []WorkItem
	FooterActions   []GuidedAction
	Warnings        []string
	DirectCommands  []string
	VisibleDefaults []VisibleDefault
}

// ActionResult is the shell-visible result of running one guided action.
type ActionResult struct {
	Plan    NextActionPlan
	Message string
	Utility *UtilitySurface
}

// GuidedAction is one user-selectable or footer operation.
type GuidedAction struct {
	ID                   string
	Kind                 string
	Label                string
	Description          string
	InternalName         string
	MutatesState         bool
	RequiresConfirmation bool
	DirectCommand        string
	ExpectedResult       string
	Inputs               map[string]string
}

// WorkItem is a tool-owned setup, blocker, progress, or recovery row.
type WorkItem struct {
	Label         string
	Status        string
	Description   string
	DirectCommand string
}

// VisibleDefault is a value StageServe will use if the user accepts the plan.
type VisibleDefault struct {
	Label string
	Value string
	Note  string
}

// UtilitySurface is a temporary read-only panel such as status or logs.
type UtilitySurface struct {
	Title           string
	Body            string
	Footer          string
	DismissOnAnyKey bool
}
