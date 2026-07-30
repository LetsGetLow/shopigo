package catalog

import (
	"errors"
	"fmt"
)

var ErrCategorySaveFailed = errors.New("category save failed")

func NewCategorySaveFailedError(err error) error {
	return fmt.Errorf("%w: %v", ErrCategorySaveFailed, err)
}

var ErrCategoryNotFound = errors.New("category not found")

func NewCategoryNotFoundError(err error) error {
	return fmt.Errorf("%w: %v", ErrCategoryNotFound, err)
}

var ErrCategoryDeleteFailed = errors.New("category delete failed")

func NewCategoryDeleteFailedError(err error) error {
	return fmt.Errorf("%w: %v", ErrCategoryDeleteFailed, err)
}
