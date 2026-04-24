// Tiny filesystem helpers used by the orchestrator. Kept separate so the
// orchestrator file stays focused on flow control.
package lifecycle

import "os"

func mkdirAll(p string) error { return os.MkdirAll(p, 0o755) }
func writeFile(p string, b []byte) error {
	return os.WriteFile(p, b, 0o644)
}
