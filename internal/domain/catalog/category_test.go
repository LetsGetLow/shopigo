package catalog

import (
	"testing"

	"github.com/google/uuid"
)

func TestParentCategoryIDFromNullUUID(t *testing.T) {
	t.Run("returns nil for invalid uuid", func(t *testing.T) {
		if got := ParentCategoryIDFromNullUUID(uuid.NullUUID{}); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})

	t.Run("returns parent id for valid uuid", func(t *testing.T) {
		parentUUID := uuid.New()
		got := ParentCategoryIDFromNullUUID(uuid.NullUUID{UUID: parentUUID, Valid: true})
		if got == nil {
			t.Fatal("expected non-nil parent id")
		}
		if uuid.UUID(*got) != parentUUID {
			t.Fatalf("expected %v, got %v", parentUUID, *got)
		}
	})
}

func TestNewCategoryAsRoot(t *testing.T) {
	name := "Electronics"
	description := "Electronic products"

	category := NewCategory(name, description)

	if category.Name != name {
		t.Errorf("expected name %s, got %s", name, category.Name)
	}
	if category.Description != description {
		t.Errorf("expected description %s, got %s", description, category.Description)
	}
	if category.ParentID != nil {
		t.Error("expected nil parent for root category")
	}
	if !category.IsRoot() {
		t.Error("expected root category")
	}
}

func TestMoveCategory(t *testing.T) {
	category := NewCategory("Phones", "Phone products")
	newParentID := ParentCategoryID(uuid.New())

	category.MoveToParent(newParentID)

	if category.ParentID == nil {
		t.Error("expected non-nil parent after move")
	}
	if *category.ParentID != newParentID {
		t.Errorf("expected parent %v, got %v", newParentID, *category.ParentID)
	}
}

func TestMoveCategoryIntoParent(t *testing.T) {
	category := NewCategory("Accessories", "Accessories")
	oldParentID := ParentCategoryID(uuid.New())
	category.MoveToParent(oldParentID)

	newParentID := ParentCategoryID(uuid.New())
	category.MoveToParent(newParentID)

	if *category.ParentID != newParentID {
		t.Errorf("expected new parent %v, got %v", newParentID, *category.ParentID)
	}
}

func TestMakeRootCategory(t *testing.T) {
	category := NewCategory("Tablets", "Tablet products")
	parentID := ParentCategoryID(uuid.New())
	category.MoveToParent(parentID)

	if category.IsRoot() {
		t.Error("expected non-root category")
	}

	category.MakeRoot()

	if !category.IsRoot() {
		t.Error("expected nil parent after MakeRootCategory")
	}
}

func TestCategoryIDUniqueness(t *testing.T) {
	category1 := NewCategory("Category1", "Desc1")
	category2 := NewCategory("Category2", "Desc2")

	if category1.ID == category2.ID {
		t.Error("expected unique IDs for different categories")
	}
}
