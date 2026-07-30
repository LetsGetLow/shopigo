package httptransport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appcatalog "shopigo/internal/app/catalog"
	domainshared "shopigo/internal/domain/shared"
	cataloghttp "shopigo/internal/transport/http/catalog"
)

type routerCategoryService struct{}

func (routerCategoryService) CreateCategory(ctx context.Context, req appcatalog.CreateCategoryRequest, user domainshared.ActorID) (appcatalog.CreateCategoryResponse, error) {
	return appcatalog.CreateCategoryResponse{
		Category: appcatalog.CategoryResponse{ID: "cat-1", Name: req.Name, Description: req.Description},
	}, nil
}

func (routerCategoryService) UpdateCategory(ctx context.Context, req appcatalog.UpdateCategoryRequest, user domainshared.ActorID) (appcatalog.UpdateCategoryResponse, error) {
	return appcatalog.UpdateCategoryResponse{
		Category: appcatalog.CategoryResponse{ID: req.ID, Name: req.Name, Description: req.Description},
	}, nil
}

func (routerCategoryService) DeleteCategory(ctx context.Context, req appcatalog.DeleteCategoryRequest, user domainshared.ActorID) (appcatalog.DeleteCategoryResponse, error) {
	return appcatalog.DeleteCategoryResponse{ID: req.ID}, nil
}

func (routerCategoryService) GetCategory(ctx context.Context, id string) (appcatalog.CategoryResponse, error) {
	return appcatalog.CategoryResponse{ID: id, Name: "Books", Description: "Reading"}, nil
}

func (routerCategoryService) ListCategories(ctx context.Context) ([]appcatalog.CategoryResponse, error) {
	return []appcatalog.CategoryResponse{{ID: "root", Name: "Root"}}, nil
}

func TestNewRouterPrefixesCategoryRoutesWithAPIV1(t *testing.T) {
	router := NewCategoryRouter(cataloghttp.NewCategoryHandler(routerCategoryService{}))

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
	router := NewCategoryRouter(cataloghttp.NewCategoryHandler(routerCategoryService{}))

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
