package httptransport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appcatalog "shopigo/internal/app/catalog"
)

type fakeRegistrar struct{}

func (fakeRegistrar) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /categories", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]appcatalog.CategoryResponse{{ID: "root", Name: "Root"}})
	})
}

func TestNewRouterPrefixesCategoryRoutesWithAPIV1(t *testing.T) {
	router := NewRouter(fakeRegistrar{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got []appcatalog.CategoryResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 1 || got[0].ID != "root" {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestNewRouterRedirectsAPIV1Root(t *testing.T) {
	router := NewRouter(fakeRegistrar{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Location"), "/api/v1/") {
		t.Fatalf("expected redirect location to include /api/v1/, got %q", rec.Header().Get("Location"))
	}
}
