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

func NewConnectionFailedError(connectionName string, err error) error {
	return &ConnectionFailed{
		ConnectionName: connectionName,
		Err:            err,
	}
}

func (e ConnectionFailed) Error() string {
	return fmt.Sprintf("failed to connect to %s: %s", e.ConnectionName, e.Err.Error())
}

// ErrOSOperation is a sentinel error used for OS related errors, e.g., getting current working directory.
var ErrOSOperation = errors.New("os operation error")

// NewOSOperationError wraps ErrOSOperation with the underlying error while preserving errors.Is checks.
func NewOSOperationError(err error) error {
	return fmt.Errorf("%w: %v", ErrOSOperation, err)
}

// ErrFileSystem is a sentinel error used for all file system related errors, e.g., reading/writing files.
var ErrFileSystem = errors.New("file system error")

// NewFileSystemError wraps ErrFileSystem with the underlying error while preserving errors.Is checks.
func NewFileSystemError(err error) error {
	return fmt.Errorf("%w: %v", ErrFileSystem, err)
}

// ErrNotFound is a sentinel error used when a required resource is not found.
var ErrNotFound = errors.New("not found")

// NewNotFoundError wraps ErrNotFound with additional context while preserving errors.Is checks.
func NewNotFoundError(message string) error {
	if message == "" {
		return ErrNotFound
	}
	return fmt.Errorf("%w: %s", ErrNotFound, message)
}

var ErrSqlExecutionFailed = errors.New("sql execution failed")

func NewSqlExecutionFailedError(err error) error {
	return fmt.Errorf("%w: %v", ErrSqlExecutionFailed, err)
}
