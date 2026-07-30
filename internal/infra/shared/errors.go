package shared

import (
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

// NotFound is an error type used when a required resource is not found, e.g., a file or DB row.
type NotFound struct {
	Message string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("not found: %s", e.Message)
}
