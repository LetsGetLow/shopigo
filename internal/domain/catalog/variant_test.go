package catalog

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewVariant(t *testing.T) {
	productID := ProductID(uuid.New())
	sku := "SKU-123"

	variant := NewVariant(productID, sku)

	if variant.ID == VariantID(uuid.Nil) {
		t.Error("expected non-nil variant ID")
	}
	if variant.ProductID != productID {
		t.Errorf("expected product ID %v, got %v", productID, variant.ProductID)
	}
	if variant.SKU != sku {
		t.Errorf("expected SKU %s, got %s", sku, variant.SKU)
	}
	if variant.Attributes == nil {
		t.Fatal("expected initialized attributes")
	}
	if len(variant.Attributes.Attributes()) != 0 {
		t.Errorf("expected empty attributes on new variant, got %d", len(variant.Attributes.Attributes()))
	}
}
