package onboarding

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

// CheckDockerBinary checks whether the Docker CLI binary exists at path.
// Pass an empty string to auto-detect via PATH.
func CheckDockerBinary(path string) StepResult {
	label := "Docker CLI"
	if path == "" {
		var err error
		path, err = exec.LookPath("docker")
		if err != nil {
			rem := remediationPtr("Install Docker Desktop from https://docs.docker.com/get-docker/")
			return StepResult{
				ID:          "docker.binary",
				Label:       label,
				Status:      StatusNeedsAction,
				Message:     "docker binary not found in PATH",
				Remediation: rem,
			}
		}
	}
	if _, err := os.Stat(path); err != nil {
		rem := remediationPtr("Install Docker Desktop from https://docs.docker.com/get-docker/")
		return StepResult{
			ID:          "docker.binary",
			Label:       label,
			Status:      StatusNeedsAction,
			Message:     fmt.Sprintf("docker binary not found at %s", path),
			Remediation: rem,
		}
	}
	return StepResult{
		ID:      "docker.binary",
		Label:   label,
		Status:  StatusReady,
		Message: fmt.Sprintf("docker found at %s", path),
	}
}

// CheckDockerDaemon checks whether the Docker daemon is reachable by running
// "docker info". The binary path is auto-detected via PATH.
func CheckDockerDaemon() StepResult {
	label := "Docker daemon"
	cmd := exec.Command("docker", "info", "--format", "{{.ServerVersion}}")
	out, err := cmd.Output()
	if err != nil {
		rem := remediationPtr("Start Docker Desktop or run: sudo systemctl start docker")
		return StepResult{
			ID:          "docker.daemon",
			Label:       label,
			Status:      StatusNeedsAction,
			Message:     "Docker daemon is not reachable",
			Remediation: rem,
		}
	}
	return StepResult{
		ID:      "docker.daemon",
		Label:   label,
		Status:  StatusReady,
		Message: fmt.Sprintf("Docker daemon running (server %s)", strings.TrimSpace(string(out))),
	}
}

// CheckStateDir checks whether the StackLane state directory exists and is
// accessible.
func CheckStateDir(stateDir string) StepResult {
	label := "State directory"
	info, err := os.Stat(stateDir)
	if err != nil {
		if os.IsNotExist(err) {
			rem := remediationPtr(fmt.Sprintf("mkdir -p %q", stateDir))
			return StepResult{
				ID:          "state.dir",
				Label:       label,
				Status:      StatusNeedsAction,
				Message:     fmt.Sprintf("state directory %q does not exist", stateDir),
				Remediation: rem,
			}
		}
		rem := remediationPtr(fmt.Sprintf("Check permissions on %q", stateDir))
		return StepResult{
			ID:          "state.dir",
			Label:       label,
			Status:      StatusError,
			Message:     fmt.Sprintf("cannot access state directory: %v", err),
			Remediation: rem,
		}
	}
	if !info.IsDir() {
		rem := remediationPtr(fmt.Sprintf("Remove %q and run 'stacklane setup' again", stateDir))
		return StepResult{
			ID:          "state.dir",
			Label:       label,
			Status:      StatusError,
			Message:     fmt.Sprintf("%q exists but is not a directory", stateDir),
			Remediation: rem,
		}
	}
	return StepResult{
		ID:      "state.dir",
		Label:   label,
		Status:  StatusReady,
		Message: fmt.Sprintf("state directory %q exists", stateDir),
	}
}

// CheckPort checks whether a TCP port on 127.0.0.1 is free.
// stepID should be "port.80" or "port.443".
func CheckPort(stepID string, port int) StepResult {
	label := fmt.Sprintf("Port %d", port)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		rem := remediationPtr(fmt.Sprintf("Find and stop the process using port %d: lsof -i :%d", port, port))
		return StepResult{
			ID:          stepID,
			Label:       label,
			Status:      StatusNeedsAction,
			Message:     fmt.Sprintf("port %d is already in use", port),
			Remediation: rem,
		}
	}
	ln.Close()
	return StepResult{
		ID:      stepID,
		Label:   label,
		Status:  StatusReady,
		Message: fmt.Sprintf("port %d is available", port),
	}
}

// CheckDNS probes whether the local DNS resolver for suffix is configured and
// reachable. On platforms other than darwin and linux this returns an
// unsupported-os step with Code="unsupported-os".
func CheckDNS(suffix string) StepResult {
	return checkDNS(suffix)
}

// CheckMkcert checks whether mkcert is installed and a local CA is available.
// It is macOS/Linux only — returns unsupported-os on other platforms.
func CheckMkcert() StepResult {
	return checkMkcert()
}
