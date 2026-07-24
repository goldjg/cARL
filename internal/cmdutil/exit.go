package cmdutil

// ExitError represents a command failure with an explicit process exit code.
// When SuppressPrefix is true, the dispatcher must print Message as-is.
type ExitError struct {
	Code           int
	Message        string
	SuppressPrefix bool
}

// Error implements the error interface.
func (e *ExitError) Error() string {
	return e.Message
}
