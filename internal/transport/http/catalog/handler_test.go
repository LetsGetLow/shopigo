package cataloghttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appcatalog "shopigo/internal/app/catalog"
	domainshared "shopigo/internal/domain/shared"

	"github.com/google/uuid"
)

type fakeCategoryService struct {
	createFn func(context.Context, appcatalog.CreateCategoryRequest, domainshared.ActorID) (appcatalog.CreateCategoryResponse, error)
	updateFn func(context.Context, appcatalog.UpdateCategoryRequest, domainshared.ActorID) (appcatalog.UpdateCategoryResponse, error)
	deleteFn func(context.Context, appcatalog.DeleteCategoryRequest, domainshared.ActorID) (appcatalog.DeleteCategoryResponse, error)
	getFn    func(context.Context, string) (appcatalog.CategoryResponse, error)
	listFn   func(context.Context) ([]appcatalog.CategoryResponse, error)
}

func (f *fakeCategoryService) CreateCategory(ctx context.Context, req appcatalog.CreateCategoryRequest, user domainshared.ActorID) (appcatalog.CreateCategoryResponse, error) {
	return f.createFn(ctx, req, user)
}

func (f *fakeCategoryService) UpdateCategory(ctx context.Context, req appcatalog.UpdateCategoryRequest, user domainshared.ActorID) (appcatalog.UpdateCategoryResponse, error) {
	return f.updateFn(ctx, req, user)
}

func (f *fakeCategoryService) DeleteCategory(ctx context.Context, req appcatalog.DeleteCategoryRequest, user domainshared.ActorID) (appcatalog.DeleteCategoryResponse, error) {
	return f.deleteFn(ctx, req, user)
}

func (f *fakeCategoryService) GetCategory(ctx context.Context, id string) (appcatalog.CategoryResponse, error) {
	return f.getFn(ctx, id)
}

func (f *fakeCategoryService) ListCategories(ctx context.Context) ([]appcatalog.CategoryResponse, error) {
	return f.listFn(ctx)
}

func TestCategoryHandler_CreateCategory(t *testing.T) {
	parentID := uuid.NewString()
	svc := &fakeCategoryService{
		createFn: func(ctx context.Context, req appcatalog.CreateCategoryRequest, user domainshared.ActorID) (appcatalog.CreateCategoryResponse, error) {
			if req.Name != "Electronics" {
				t.Fatalf("unexpected name %q", req.Name)
			}
			if req.ParentID == nil || *req.ParentID != parentID {
				t.Fatalf("expected parent id %q, got %v", parentID, req.ParentID)
			}
			return appcatalog.CreateCategoryResponse{
				Category: appcatalog.CategoryResponse{ID: "cat-1", Name: req.Name, Description: req.Description, ParentID: req.ParentID},
			}, nil
		},
	}

	handler := NewCategoryHandler(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := strings.NewReader(`{"name":"Electronics","description":"Gadgets","parent_id":"` + parentID + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/categories", body)
	req.Header.Set(actorHeader, uuid.NewString())
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp appcatalog.CreateCategoryResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Category.ID != "cat-1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Category.ParentID == nil || *resp.Category.ParentID != parentID {
		t.Fatalf("expected parent id %q, got %v", parentID, resp.Category.ParentID)
	}
}

func TestCategoryHandler_GetUpdateDeleteCategory(t *testing.T) {
	parentID := uuid.NewString()
	getID := uuid.NewString()
	svc := &fakeCategoryService{
		updateFn: func(ctx context.Context, req appcatalog.UpdateCategoryRequest, user domainshared.ActorID) (appcatalog.UpdateCategoryResponse, error) {
			return appcatalog.UpdateCategoryResponse{
				Category: appcatalog.CategoryResponse{ID: req.ID, Name: req.Name, Description: req.Description, ParentID: req.ParentID},
			}, nil
		},
		deleteFn: func(ctx context.Context, req appcatalog.DeleteCategoryRequest, user domainshared.ActorID) (appcatalog.DeleteCategoryResponse, error) {
			return appcatalog.DeleteCategoryResponse{ID: req.ID}, nil
		},
		getFn: func(ctx context.Context, id string) (appcatalog.CategoryResponse, error) {
			return appcatalog.CategoryResponse{ID: id, Name: "Books", Description: "Reading"}, nil
		},
		listFn: func(ctx context.Context) ([]appcatalog.CategoryResponse, error) {
			return []appcatalog.CategoryResponse{{ID: "1", Name: "Root"}}, nil
		},
	}

	handler := NewCategoryHandler(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	getReq := httptest.NewRequest(http.MethodGet, "/categories/"+getID, nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected get status 200, got %d", getRec.Code)
	}
	var got appcatalog.CategoryResponse
	if err := json.NewDecoder(getRec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}
	if got.ID != getID {
		t.Fatalf("expected get id %q, got %q", getID, got.ID)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/categories", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRec.Code)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/categories/abc", strings.NewReader(`{"name":"Updated","description":"Desc","parent_id":"`+parentID+`"}`))
	updateReq.Header.Set(actorHeader, uuid.NewString())
	updateRec := httptest.NewRecorder()
	mux.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d", updateRec.Code)
	}
	var updated appcatalog.UpdateCategoryResponse
	if err := json.NewDecoder(updateRec.Body).Decode(&updated); err != nil {
		t.Fatalf("failed to decode update response: %v", err)
	}
	if updated.Category.ParentID == nil || *updated.Category.ParentID != parentID {
		t.Fatalf("expected updated parent id %q, got %v", parentID, updated.Category.ParentID)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/categories/abc", nil)
	deleteReq.Header.Set(actorHeader, uuid.NewString())
	deleteRec := httptest.NewRecorder()
	mux.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusOK {
		t.Fatalf("expected delete status 200, got %d", deleteRec.Code)
	}
}

func TestCategoryHandler_ReturnsErrors(t *testing.T) {
	id := uuid.NewString()
	svc := &fakeCategoryService{
		getFn: func(ctx context.Context, id string) (appcatalog.CategoryResponse, error) {
			return appcatalog.CategoryResponse{}, errors.New("boom")
		},
	}

	handler := NewCategoryHandler(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/categories/"+id, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for generic error, got %d", rec.Code)
	}
}

func TestCategoryHandler_GetCategoryRejectsInvalidID(t *testing.T) {
	svc := &fakeCategoryService{
		getFn: func(ctx context.Context, id string) (appcatalog.CategoryResponse, error) {
			t.Fatal("service should not be called for invalid id")
			return appcatalog.CategoryResponse{}, nil
		},
	}

	handler := NewCategoryHandler(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/categories/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for invalid category id, got %d", rec.Code)
	}
}
