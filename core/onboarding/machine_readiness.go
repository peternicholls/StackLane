package onboarding

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

var portListen = net.Listen
var portOwnerLookup = lookupBusyPortOwner

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

// CheckStateDir checks whether the StageServe state directory exists and is
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
		rem := remediationPtr(fmt.Sprintf("Remove %q and run 'stage setup' again", stateDir))
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

// CheckRequiredFile verifies that a required StageServe file exists.
func CheckRequiredFile(stepID, label, path string) StepResult {
	if strings.TrimSpace(path) == "" {
		return StepResult{
			ID:      stepID,
			Label:   label,
			Status:  StatusError,
			Message: "required file path is empty",
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			remediation := requiredFileRemediation(stepID, path)
			return StepResult{
				ID:          stepID,
				Label:       label,
				Status:      StatusNeedsAction,
				Message:     fmt.Sprintf("required file is missing at %s", path),
				Remediation: remediation,
			}
		}
		return StepResult{
			ID:      stepID,
			Label:   label,
			Status:  StatusError,
			Message: fmt.Sprintf("cannot access required file %s: %v", path, err),
		}
	}

	if info.IsDir() {
		return StepResult{
			ID:      stepID,
			Label:   label,
			Status:  StatusError,
			Message: fmt.Sprintf("expected a file at %s, but found a directory", path),
		}
	}

	return StepResult{
		ID:      stepID,
		Label:   label,
		Status:  StatusReady,
		Message: fmt.Sprintf("found %s", path),
	}
}

func requiredFileRemediation(stepID, path string) *string {
	switch stepID {
	case "stack.shared_file", "stack.project_file":
		return remediationPtr(fmt.Sprintf("Reinstall StageServe or restore the bundled runtime file at %s, then run stage doctor", path))
	default:
		return nil
	}
}

// CheckPort checks whether a TCP port on 127.0.0.1 is free.
// stepID should be "port.80" or "port.443".
func CheckPort(stepID string, port int) StepResult {
	label := fmt.Sprintf("Port %d", port)
	loopbackAddr := fmt.Sprintf("127.0.0.1:%d", port)
	ln, err := portListen("tcp", loopbackAddr)
	if err != nil {
		if isPermissionDenied(err) {
			wildcardAddr := fmt.Sprintf("0.0.0.0:%d", port)
			wildcardLn, wildcardErr := portListen("tcp", wildcardAddr)
			switch {
			case wildcardErr == nil:
				wildcardLn.Close()
				return StepResult{
					ID:      stepID,
					Label:   label,
					Status:  StatusReady,
					Message: fmt.Sprintf("port %d is available", port),
				}
			case isAddrInUse(wildcardErr):
				return busyPortStep(stepID, label, port)
			default:
				rem := remediationPtr(fmt.Sprintf("Check local networking permissions while probing port %d: %v", port, wildcardErr))
				return StepResult{
					ID:          stepID,
					Label:       label,
					Status:      StatusError,
					Message:     fmt.Sprintf("could not probe port %d availability", port),
					Remediation: rem,
				}
			}
		}
		if isAddrInUse(err) {
			return busyPortStep(stepID, label, port)
		}
		rem := remediationPtr(fmt.Sprintf("Check local networking permissions while probing port %d: %v", port, err))
		return StepResult{
			ID:          stepID,
			Label:       label,
			Status:      StatusError,
			Message:     fmt.Sprintf("could not probe port %d availability", port),
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

func busyPortStep(stepID, label string, port int) StepResult {
	message := fmt.Sprintf("port %d is already in use", port)
	rem := fmt.Sprintf("lsof -nP -iTCP:%d -sTCP:LISTEN", port)
	if owner := portOwnerLookup(port); owner != "" {
		if owner == "another process (owner hidden without sudo)" {
			message = fmt.Sprintf("port %d is already in use — owner requires sudo to identify", port)
			rem = fmt.Sprintf("sudo lsof -nP -iTCP:%d -sTCP:LISTEN", port)
		} else {
			message = fmt.Sprintf("port %d is already in use by %s", port, owner)
		}
	}
	return StepResult{
		ID:          stepID,
		Label:       label,
		Status:      StatusNeedsAction,
		Message:     message,
		Remediation: remediationPtr(rem),
	}
}

func lookupBusyPortOwner(port int) string {
	args := []string{"-nP", fmt.Sprintf("-iTCP:%d", port), "-sTCP:LISTEN"}
	if owner := parseBusyPortOwnerOutput(mustCombinedOutput(exec.Command("lsof", args...))); owner != "" {
		return owner
	}
	privilegedOut := mustCombinedOutput(exec.Command("sudo", append([]string{"-n", "lsof"}, args...)...))
	if owner := parseBusyPortOwnerOutput(privilegedOut); owner != "" {
		return owner
	}
	if strings.Contains(strings.ToLower(string(privilegedOut)), "password is required") {
		return "another process (owner hidden without sudo)"
	}
	return ""
}

func mustCombinedOutput(cmd *exec.Cmd) []byte {
	out, _ := cmd.CombinedOutput()
	return out
}

func parseBusyPortOwnerOutput(out []byte) string {
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "COMMAND ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if _, err := strconv.Atoi(fields[1]); err != nil {
			continue
		}
		return fmt.Sprintf("%s (pid %s)", fields[0], fields[1])
	}
	return ""
}

func isAddrInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE)
}

func isPermissionDenied(err error) bool {
	return errors.Is(err, syscall.EACCES) || errors.Is(err, syscall.EPERM)
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
