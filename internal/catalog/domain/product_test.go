package domain

import (
	"testing"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
)

func TestNewProduct(t *testing.T) {
	name := "Test Product"
	description := "A test product"
	price := *money.New(1000, money.EUR)

	product := NewProduct(name, description, price)

	if product.Name != name {
		t.Errorf("expected name %s, got %s", name, product.Name)
	}
	if product.Description != description {
		t.Errorf("expected description %s, got %s", description, product.Description)
	}
	if product.Price != price {
		t.Errorf("expected price %v, got %v", price, product.Price)
	}
	if product.ID == ProductID(uuid.Nil) {
		t.Error("expected non-nil product ID")
	}
	if len(product.CategoryIDs()) != 0 {
		t.Error("expected empty category IDs on new product")
	}
}

func TestAddToCategory(t *testing.T) {
	product := NewProduct("Test", "Test", *money.New(100, money.EUR))
	catID := CategoryID(uuid.New())

	product.AddToCategory(catID)
	categories := product.CategoryIDs()

	if len(categories) != 1 {
		t.Errorf("expected 1 category, got %d", len(categories))
	}
	if categories[0] != catID {
		t.Errorf("expected category %v, got %v", catID, categories[0])
	}
}

func TestAddToMultipleCategories(t *testing.T) {
	product := NewProduct("Test", "Test", *money.New(100, money.EUR))
	catID1 := CategoryID(uuid.New())
	catID2 := CategoryID(uuid.New())

	product.AddToCategory(catID1)
	product.AddToCategory(catID2)
	categories := product.CategoryIDs()

	if len(categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(categories))
	}
}

func TestAddDuplicateCategory(t *testing.T) {
	product := NewProduct("Test", "Test", *money.New(100, money.EUR))
	catID := CategoryID(uuid.New())

	product.AddToCategory(catID)
	product.AddToCategory(catID)
	categories := product.CategoryIDs()

	if len(categories) != 1 {
		t.Errorf("expected 1 category (duplicate ignored), got %d", len(categories))
	}
}

func TestRemoveFromCategory(t *testing.T) {
	product := NewProduct("Test", "Test", *money.New(100, money.EUR))
	catID := CategoryID(uuid.New())

	product.AddToCategory(catID)
	product.RemoveFromCategory(catID)
	categories := product.CategoryIDs()

	if len(categories) != 0 {
		t.Errorf("expected 0 categories after removal, got %d", len(categories))
	}
}

func TestRemoveFromCategoryNotPresent(t *testing.T) {
	product := NewProduct("Test", "Test", *money.New(100, money.EUR))
	catID := CategoryID(uuid.New())

	product.RemoveFromCategory(catID)
	categories := product.CategoryIDs()

	if len(categories) != 0 {
		t.Errorf("expected 0 categories, got %d", len(categories))
	}
}

func TestRemoveFromMultipleCategories(t *testing.T) {
	product := NewProduct("Test", "Test", *money.New(100, money.EUR))
	catID1 := CategoryID(uuid.New())
	catID2 := CategoryID(uuid.New())

	product.AddToCategory(catID1)
	product.AddToCategory(catID2)
	product.RemoveFromCategory(catID1)
	categories := product.CategoryIDs()

	if len(categories) != 1 {
		t.Errorf("expected 1 category after removal, got %d", len(categories))
	}
	if categories[0] != catID2 {
		t.Errorf("expected remaining category %v, got %v", catID2, categories[0])
	}
}

func TestCategoryIDsReturnsSnapshot(t *testing.T) {
	product := NewProduct("Test", "Test", *money.New(100, money.EUR))
	catID := CategoryID(uuid.New())

	product.AddToCategory(catID)
	categories1 := product.CategoryIDs()
	categories2 := product.CategoryIDs()

	if len(categories1) != len(categories2) {
		t.Error("expected consistent results from multiple CategoryIDs() calls")
	}
}
