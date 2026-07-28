package shared

import "testing"

func TestLookupConfigValueWithFallback(t *testing.T) {
	t.Run("returns fallback when key is missing", func(t *testing.T) {
		const key = "SHOPIGO_CONFIG_TEST_MISSING"
		const fallback = "default-value"

		if got := LookupConfigValueWithFallback(key, fallback); got != fallback {
			t.Fatalf("expected fallback %q, got %q", fallback, got)
		}
	})

	t.Run("returns env value when key is set", func(t *testing.T) {
		const key = "SHOPIGO_CONFIG_TEST_PRESENT"
		const value = "configured-value"
		t.Setenv(key, value)

		if got := LookupConfigValueWithFallback(key, "default-value"); got != value {
			t.Fatalf("expected env value %q, got %q", value, got)
		}
	})

	t.Run("returns empty string when env value is empty", func(t *testing.T) {
		const key = "SHOPIGO_CONFIG_TEST_EMPTY"
		t.Setenv(key, "")

		if got := LookupConfigValueWithFallback(key, "default-value"); got != "" {
			t.Fatalf("expected empty string, got %q", got)
		}
	})
}

func TestLookupConfigValue(t *testing.T) {
	t.Run("returns false when key is missing", func(t *testing.T) {
		const key = "SHOPIGO_CONFIG_TEST_LOOKUP_MISSING"

		got, found := LookupConfigValue(key)
		if found {
			t.Fatal("expected key to be missing")
		}
		if got != "" {
			t.Fatalf("expected empty value, got %q", got)
		}
	})

	t.Run("returns value and true when key is set", func(t *testing.T) {
		const key = "SHOPIGO_CONFIG_TEST_LOOKUP_PRESENT"
		const value = "configured-value"
		t.Setenv(key, value)

		got, found := LookupConfigValue(key)
		if !found {
			t.Fatal("expected key to be found")
		}
		if got != value {
			t.Fatalf("expected value %q, got %q", value, got)
		}
	})

	t.Run("returns empty string and true when env value is empty", func(t *testing.T) {
		const key = "SHOPIGO_CONFIG_TEST_LOOKUP_EMPTY"
		t.Setenv(key, "")

		got, found := LookupConfigValue(key)
		if !found {
			t.Fatal("expected key to be found")
		}
		if got != "" {
			t.Fatalf("expected empty value, got %q", got)
		}
	})
}
