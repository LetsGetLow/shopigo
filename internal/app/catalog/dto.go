package catalog

// CreateCategoryRequest represents the data needed to create a category.
type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id,omitempty"`
}

// UpdateCategoryRequest represents the data needed to update a category.
type UpdateCategoryRequest struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id,omitempty"`
}

// DeleteCategoryRequest represents the data needed to delete a category.
type DeleteCategoryRequest struct {
	ID string `json:"id"`
}

// CategoryResponse represents a category returned by the application layer.
type CategoryResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id,omitempty"`
}

// CreateCategoryResponse is returned after creating a category.
type CreateCategoryResponse struct {
	Category CategoryResponse `json:"category"`
}

// UpdateCategoryResponse is returned after updating a category.
type UpdateCategoryResponse struct {
	Category CategoryResponse `json:"category"`
}

// DeleteCategoryResponse is returned after deleting a category.
type DeleteCategoryResponse struct {
	ID string `json:"id"`
}
