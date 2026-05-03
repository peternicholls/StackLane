package onboarding

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONProjector writes a CommandResult as a stable JSON envelope to w.
type JSONProjector struct {
	W io.Writer
}

// Project renders the result as indented JSON.
func (p *JSONProjector) Project(r CommandResult) error {
	enc := json.NewEncoder(p.W)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("json projection: %w", err)
	}
	return nil
}
