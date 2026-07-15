package domain

import "github.com/google/uuid"

type VariantID uuid.UUID
type Variant struct {
	ID         VariantID
	ProductID  ProductID
	SKU        string
	Attributes *AttributeMap
}

func NewVariant(productID ProductID, sku string) *Variant {
	return &Variant{
		ID:         VariantID(uuid.New()),
		ProductID:  productID,
		SKU:        sku,
		Attributes: NewAttributeMap(),
	}
}
