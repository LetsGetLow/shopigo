package catalog

import (
	"context"
	"fmt"

	domain "shopigo/internal/domain/catalog"
	domainshared "shopigo/internal/domain/shared"

	"github.com/google/uuid"
)

// CategoryService coordinates category use cases.
type CategoryService struct {
	repo domain.CategoryRepository
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(repo domain.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func categoryResponseFromDomain(category *domain.Category) CategoryResponse {
	resp := CategoryResponse{
		ID:          uuid.UUID(category.ID).String(),
		Name:        category.Name,
		Description: category.Description,
	}

	if category.ParentID != nil {
		parentID := uuid.UUID(*category.ParentID).String()
		resp.ParentID = &parentID
	}

	return resp
}

func categoryIDFromString(value string) (domain.CategoryID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return domain.CategoryID(uuid.Nil), fmt.Errorf("invalid category id %q: %w", value, err)
	}
	return domain.CategoryID(id), nil
}

func parentIDFromString(value string) (domain.ParentCategoryID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return domain.ParentCategoryID(uuid.Nil), fmt.Errorf("invalid parent category id %q: %w", value, err)
	}
	return domain.ParentCategoryID(id), nil
}

func applyParentID(category *domain.Category, parentID *string) error {
	if parentID == nil {
		return nil
	}

	if *parentID == "" {
		category.MakeRoot()
		return nil
	}

	resolvedParentID, err := parentIDFromString(*parentID)
	if err != nil {
		return err
	}

	category.MoveToParent(resolvedParentID)
	return nil
}

// CreateCategory creates and saves a new root category.
func (s *CategoryService) CreateCategory(ctx context.Context, req CreateCategoryRequest, user domainshared.ActorID) (CreateCategoryResponse, error) {
	category := domain.NewCategory(req.Name, req.Description)
	if err := applyParentID(category, req.ParentID); err != nil {
		return CreateCategoryResponse{}, err
	}

	if err := s.repo.Save(ctx, *category, user); err != nil {
		return CreateCategoryResponse{}, err
	}

	return CreateCategoryResponse{Category: categoryResponseFromDomain(category)}, nil
}

// UpdateCategory updates the name and description of an existing category.
func (s *CategoryService) UpdateCategory(ctx context.Context, req UpdateCategoryRequest, user domainshared.ActorID) (UpdateCategoryResponse, error) {
	id, err := categoryIDFromString(req.ID)
	if err != nil {
		return UpdateCategoryResponse{}, err
	}

	category, err := s.repo.Get(ctx, id)
	if err != nil {
		return UpdateCategoryResponse{}, err
	}

	category.Name = req.Name
	category.Description = req.Description
	if err := applyParentID(category, req.ParentID); err != nil {
		return UpdateCategoryResponse{}, err
	}

	if err := s.repo.Save(ctx, *category, user); err != nil {
		return UpdateCategoryResponse{}, err
	}

	return UpdateCategoryResponse{Category: categoryResponseFromDomain(category)}, nil
}

// MoveCategoryToParent moves an existing category under a new parent.
func (s *CategoryService) MoveCategoryToParent(ctx context.Context, id string, parentID string, user domainshared.ActorID) (CategoryResponse, error) {
	categoryID, err := categoryIDFromString(id)
	if err != nil {
		return CategoryResponse{}, err
	}
	resolvedParentID, err := parentIDFromString(parentID)
	if err != nil {
		return CategoryResponse{}, err
	}

	category, err := s.repo.Get(ctx, categoryID)
	if err != nil {
		return CategoryResponse{}, err
	}

	category.MoveToParent(resolvedParentID)

	if err := s.repo.Save(ctx, *category, user); err != nil {
		return CategoryResponse{}, err
	}

	return categoryResponseFromDomain(category), nil
}

// MakeCategoryRoot removes the parent relationship from an existing category.
func (s *CategoryService) MakeCategoryRoot(ctx context.Context, id string, user domainshared.ActorID) (CategoryResponse, error) {
	categoryID, err := categoryIDFromString(id)
	if err != nil {
		return CategoryResponse{}, err
	}

	category, err := s.repo.Get(ctx, categoryID)
	if err != nil {
		return CategoryResponse{}, err
	}

	category.MakeRoot()

	if err := s.repo.Save(ctx, *category, user); err != nil {
		return CategoryResponse{}, err
	}

	return categoryResponseFromDomain(category), nil
}

// GetCategory retrieves a category by ID.
func (s *CategoryService) GetCategory(ctx context.Context, id string) (CategoryResponse, error) {
	categoryID, err := categoryIDFromString(id)
	if err != nil {
		return CategoryResponse{}, err
	}

	category, err := s.repo.Get(ctx, categoryID)
	if err != nil {
		return CategoryResponse{}, err
	}

	return categoryResponseFromDomain(category), nil
}

// ListCategories returns all categories.
func (s *CategoryService) ListCategories(ctx context.Context) ([]CategoryResponse, error) {
	categories, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]CategoryResponse, 0, len(categories))
	for i := range categories {
		category := categories[i]
		out = append(out, categoryResponseFromDomain(&category))
	}

	return out, nil
}

// ListCategoriesByParent returns categories by parent relationship.
func (s *CategoryService) ListCategoriesByParent(ctx context.Context, id *string) ([]CategoryResponse, error) {
	var parentID *domain.ParentCategoryID
	if id != nil {
		resolvedParentID, err := parentIDFromString(*id)
		if err != nil {
			return nil, err
		}
		parentID = &resolvedParentID
	}

	categories, err := s.repo.ListByParent(ctx, parentID)
	if err != nil {
		return nil, err
	}

	out := make([]CategoryResponse, 0, len(categories))
	for i := range categories {
		category := categories[i]
		out = append(out, categoryResponseFromDomain(&category))
	}

	return out, nil
}

// DeleteCategory soft-deletes a category.
func (s *CategoryService) DeleteCategory(ctx context.Context, req DeleteCategoryRequest, user domainshared.ActorID) (DeleteCategoryResponse, error) {
	id, err := categoryIDFromString(req.ID)
	if err != nil {
		return DeleteCategoryResponse{}, err
	}

	if err := s.repo.Delete(ctx, id, user); err != nil {
		return DeleteCategoryResponse{}, err
	}

	return DeleteCategoryResponse{ID: req.ID}, nil
}
