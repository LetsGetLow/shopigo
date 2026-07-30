package shared

import (
	"errors"
	"fmt"
)

// ConnectionFailed is used when a connection to a service fails, e.g., database connection.
type ConnectionFailed struct {
	ConnectionName string
	Err            error
}

func (e ConnectionFailed) Error() string {
	return fmt.Sprintf("failed to connect to %s: %s", e.ConnectionName, e.Err.Error())
}

// OSOperation is used for OS related errors, e.g., getting current working directory.
type OSOperation struct {
	Err error
}

func (e OSOperation) Error() string {
	return fmt.Sprintf("os operation error: %s", e.Err.Error())
}

// FileSystem is used for all file system related errors, e.g., reading/writing files.
type FileSystem struct {
	Err error
}

func (e FileSystem) Error() string {
	return fmt.Sprintf("file system error: %s", e.Err.Error())
}

// ErrNotFound is a sentinel error used when a required resource is not found.
var ErrNotFound = errors.New("not found")

// NewNotFound wraps ErrNotFound with additional context while preserving errors.Is checks.
func NewNotFound(message string) error {
	if message == "" {
		return ErrNotFound
	}
	return fmt.Errorf("%w: %s", ErrNotFound, message)
}

var ErrSqlExecutionFailed = errors.New("sql execution failed")

func NewSqlExecutionFailedError(err error) error {
	return fmt.Errorf("%w: %v", ErrSqlExecutionFailed, err)
}
