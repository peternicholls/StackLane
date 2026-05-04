package commands

// ExitCoder is implemented by command errors that should control the process
// exit code without being printed as ordinary failures.
type ExitCoder interface {
	error
	ExitCode() int
	Silent() bool
}
