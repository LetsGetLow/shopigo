package catalog

import (
	"context"
	shared "shopigo/internal/domain/shared"

	"github.com/google/uuid"
)

type CategoryID uuid.UUID
type ParentCategoryID uuid.UUID

type Category struct {
	ID          CategoryID
	Name        string
	Description string
	ParentID    *ParentCategoryID
	shared.Audit
}

type CategoryRepository interface {
	Save(ctx context.Context, category *Category) error
	Get(ctx context.Context, id CategoryID) (*Category, error)
	List(ctx context.Context) ([]Category, error)
	ListByParent(ctx context.Context, id ParentCategoryID) ([]Category, error)
	Delete(ctx context.Context, id CategoryID) error
	Close() error
}

func NewCategory(name string, description string) *Category {
	return &Category{
		ID:          CategoryID(uuid.New()),
		Name:        name,
		Description: description,
		ParentID:    nil, // default root category
	}
}

func (c *Category) MoveInto(newParentID ParentCategoryID) {
	c.ParentID = &newParentID
}

func (c *Category) MakeRoot() {
	c.ParentID = nil
}

func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}
