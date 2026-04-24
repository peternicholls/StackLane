// Package state defines the typed StateStore contract for per-project JSON
// state and registry projection.
package state

import "github.com/peternicholls/stacklane/core/config"

// AttachmentState records the intended lifecycle state for a project.
type AttachmentState string

const (
	StateAttached AttachmentState = "attached"
	StateDown     AttachmentState = "down"
)

// ContainerIdentity is the recorded identity of one runtime container, captured
// after `compose up` so subsequent status / logs invocations have something
// stable to compare against the live container set.
type ContainerIdentity struct {
	Name   string
	ID     string
	Status string
}

// RuntimeIdentity captures the recorded identities of the well-known services
// per project.
type RuntimeIdentity struct {
	Nginx       ContainerIdentity
	Apache      ContainerIdentity
	MariaDB     ContainerIdentity
	PhpMyAdmin  ContainerIdentity
	SummaryLine string
}

// Record is the persisted, per-project view that round-trips through the
// state store. It is the on-disk schema; format/version live alongside.
type Record struct {
	SchemaVersion   int                  `json:"schema_version"`
	Project         config.ProjectConfig `json:"project"`
	AttachmentState AttachmentState      `json:"attachment_state"`
	Runtime         RuntimeIdentity      `json:"runtime"`
}

// RegistryRow is the typed projection of a single project across the registry.
type RegistryRow struct {
	Slug             string
	AttachmentState  AttachmentState
	Name             string
	Dir              string
	Hostname         string
	DocRoot          string
	ComposeProject   string
	RuntimeNetwork   string
	DatabaseVolume   string
	PHPVersion       string
	MySQLDatabase    string
	MySQLPort        int
	PMAPort          int
	WebNetworkAlias  string
	ContainerSummary string
}

// StateStore is the contract for persisting and loading per-project state.
// Implementations MUST guarantee atomic writes (FR-008) and MUST tolerate
// missing slugs by returning a typed error.
type StateStore interface {
	Save(rec Record) error
	Load(slug string) (Record, error)
	Remove(slug string) error
	Registry() ([]RegistryRow, error)
	StateFileForSelector(selector string) (Record, string, error)
	StateDir() string
}
