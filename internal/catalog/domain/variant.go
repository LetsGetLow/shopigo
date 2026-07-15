package domain

import (
	"context"

	"github.com/google/uuid"
)

type VariantID uuid.UUID
type Variant struct {
	ID         VariantID
	ProductID  ProductID
	SKU        string
	Attributes *AttributeMap
}

type VariantRepository interface {
	Save(ctx context.Context, variant *Variant) error
	Get(ctx context.Context, id VariantID) (*Variant, error)
	ListByProduct(ctx context.Context, productID ProductID) ([]Variant, error)
	Delete(ctx context.Context, id VariantID) error
}

func NewVariant(productID ProductID, sku string) *Variant {
	return &Variant{
		ID:         VariantID(uuid.New()),
		ProductID:  productID,
		SKU:        sku,
		Attributes: NewAttributeMap(),
	}
}
