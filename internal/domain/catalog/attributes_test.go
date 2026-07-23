package catalog

import "testing"

func TestNewAttributeMapStartsEmpty(t *testing.T) {
	set := NewAttributeMap()

	if len(set.Attributes()) != 0 {
		t.Errorf("expected empty attributes, got %d", len(set.Attributes()))
	}
}

func TestAttributeMapSetAndGet(t *testing.T) {
	set := NewAttributeMap()
	set.Set("color", "red")

	value, exists := set.Get("color")
	if !exists {
		t.Fatal("expected attribute to exist")
	}
	if value != "red" {
		t.Errorf("expected red, got %s", value)
	}
}

func TestAttributeMapGetMissingReturnsFalse(t *testing.T) {
	set := NewAttributeMap()

	_, exists := set.Get("size")
	if exists {
		t.Fatal("expected missing attribute to return exists=false")
	}
}

func TestAttributeMapSetOverridesValue(t *testing.T) {
	set := NewAttributeMap()
	set.Set("color", "red")
	set.Set("color", "blue")

	value, exists := set.Get("color")
	if !exists {
		t.Fatal("expected attribute to exist")
	}
	if value != "blue" {
		t.Errorf("expected blue, got %s", value)
	}
}

func TestAttributeMapRemove(t *testing.T) {
	set := NewAttributeMap()
	set.Set("size", "L")
	set.Remove("size")

	_, exists := set.Get("size")
	if exists {
		t.Fatal("expected removed attribute to not exist")
	}
}

func TestAttributeMapAttributesReturnsCopy(t *testing.T) {
	set := NewAttributeMap()
	set.Set("material", "cotton")

	attrs := set.Attributes()
	attrs["material"] = "wool"
	attrs["new"] = "value"

	value, _ := set.Get("material")
	if value != "cotton" {
		t.Errorf("expected original value cotton, got %s", value)
	}
	if _, exists := set.Get("new"); exists {
		t.Fatal("expected mutation on returned map not to affect source")
	}
}
