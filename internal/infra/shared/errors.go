package shared

import (
	"fmt"
)

// ConnectionFailedError is used when a connection to a service fails, e.g., database connection.
type ConnectionFailedError struct {
	ConnectionName string
	Message        string
}

func (e ConnectionFailedError) Error() string {
	return fmt.Sprintf("failed to connect to %s:\n%s", e.ConnectionName, e.Message)
}

// SystemError is used for OS related Errors e.g. getting current working directory
type SystemError struct {
	Message string
}

func (e SystemError) Error() string {
	return fmt.Sprintf("system error: %s", e.Message)
}

type FileSystemError struct {
	Message string
}

func (e FileSystemError) Error() string {
	return fmt.Sprintf("file system error: %s", e.Message)
}

type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("migration error: %s", e.Message)
}
