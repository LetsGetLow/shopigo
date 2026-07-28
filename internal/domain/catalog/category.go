package catalog

import (
	"context"
	shared "shopigo/internal/domain/shared"

	"github.com/google/uuid"
)

type CategoryID uuid.UUID
type ParentCategoryID uuid.UUID

func ParentCategoryIDFromNullUUID(id uuid.NullUUID) *ParentCategoryID {
	if !id.Valid {
		return nil
	}
	parentID := ParentCategoryID(id.UUID)
	return &parentID
}

// CategoryRepository defines persistence operations for categories.
type CategoryRepository interface {
	// Save creates or updates a category and records the acting user.
	Save(ctx context.Context, category Category, user shared.ActorID) error
	// Get returns a category by ID.
	Get(ctx context.Context, id CategoryID) (*Category, error)
	// List returns all categories.
	List(ctx context.Context) ([]Category, error)
	// ListByParent returns all categories with the given parent ID. If parent ID is nil, it returns all root categories.
	ListByParent(ctx context.Context, id *ParentCategoryID) ([]Category, error)
	// Delete marks a category as deleted and records the acting user.
	Delete(ctx context.Context, id CategoryID, user shared.ActorID) error
	// Close releases any repository resources.
	Close() error
}

type Category struct {
	ID          CategoryID
	Name        string
	Description string
	ParentID    *ParentCategoryID
	shared.Audit
}

func NewCategory(name string, description string) *Category {
	return &Category{
		ID:          CategoryID(uuid.New()),
		Name:        name,
		Description: description,
		ParentID:    nil, // default root category
	}
}

func (c *Category) MoveToParent(newParentID ParentCategoryID) {
	c.ParentID = &newParentID
}

func (c *Category) MakeRoot() {
	c.ParentID = nil
}

func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}
