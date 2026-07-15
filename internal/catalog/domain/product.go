package domain

import (
	"context"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
)

type ProductID uuid.UUID
type Product struct {
	ID          ProductID
	Name        string
	Description string
	Price       money.Money
	categoryIDs map[CategoryID]struct{}
	Attributes  *AttributeMap // attributes will be inherited to all variants
}

type ProductRepository interface {
	Save(ctx context.Context, product *Product) error
	Get(ctx context.Context, id ProductID) (*Product, error)
	List(ctx context.Context) ([]Product, error)
	ListByCategory(ctx context.Context, categoryID CategoryID) ([]Product, error)
	Delete(ctx context.Context, id ProductID) error
}

func NewProduct(name string, description string, price money.Money) *Product {
	return &Product{
		ID:          ProductID(uuid.New()),
		Name:        name,
		Description: description,
		Price:       price,
		categoryIDs: make(map[CategoryID]struct{}),
		Attributes:  NewAttributeMap(),
	}
}

func (p *Product) AddToCategory(id CategoryID) {
	p.categoryIDs[id] = struct{}{}
}

func (p *Product) RemoveFromCategory(id CategoryID) {
	delete(p.categoryIDs, id)
}

func (p *Product) CategoryIDs() []CategoryID {
	ids := make([]CategoryID, 0, len(p.categoryIDs))
	for id := range p.categoryIDs {
		ids = append(ids, id)
	}
	return ids
}
