// Package lifecycle coordinates every cross-module call needed for
// up / down / attach / detach / status / logs.
//
// Errors flowing back to the operator MUST be StepError so the cobra layer
// can render them with the failing step name, the affected project, and a
// stated next action (FR-013).
package lifecycle
