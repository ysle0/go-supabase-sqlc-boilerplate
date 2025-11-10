package example_feature

import "time"

// Item represents a simple item entity
type Item struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateItemRequest represents a request to create an item
type CreateItemRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description string  `json:"description" validate:"max=500"`
	Price       float64 `json:"price" validate:"required,gte=0"`
	Quantity    int     `json:"quantity" validate:"required,gte=0"`
}

// UpdateItemRequest represents a request to update an item
type UpdateItemRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	Price       *float64 `json:"price,omitempty" validate:"omitempty,gte=0"`
	Quantity    *int     `json:"quantity,omitempty" validate:"omitempty,gte=0"`
}

// ListItemsResponse represents a paginated list of items
type ListItemsResponse struct {
	Items      []Item `json:"items"`
	TotalCount int64  `json:"total_count"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
