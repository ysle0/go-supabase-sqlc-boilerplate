package example_item

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yourusername/servers/internal/repository"
)

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
	Items      []repository.Item `json:"items"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ItemService contains business logic for items
type ItemService struct {
	repo   *repository.ItemRepository
	logger *slog.Logger
}

// NewItemService creates a new item service
func NewItemService(repo *repository.ItemRepository, logger *slog.Logger) *ItemService {
	return &ItemService{
		repo:   repo,
		logger: logger,
	}
}

// CreateItem creates a new item with validation
func (s *ItemService) CreateItem(ctx context.Context, req CreateItemRequest) (*repository.Item, error) {
	s.logger.Info("creating item", "name", req.Name)

	// Business logic validation
	if req.Price < 0 {
		return nil, fmt.Errorf("price cannot be negative")
	}
	if req.Quantity < 0 {
		return nil, fmt.Errorf("quantity cannot be negative")
	}

	item, err := s.repo.Create(ctx, req.Name, req.Description, req.Price, req.Quantity)
	if err != nil {
		s.logger.Error("failed to create item", "error", err)
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	s.logger.Info("item created successfully", "item_id", item.ID)
	return item, nil
}

// GetItem retrieves an item by ID
func (s *ItemService) GetItem(ctx context.Context, id int64) (*repository.Item, error) {
	s.logger.Debug("getting item", "item_id", id)

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get item", "item_id", id, "error", err)
		return nil, fmt.Errorf("item not found")
	}

	return item, nil
}

// ListItems retrieves all items with pagination
func (s *ItemService) ListItems(ctx context.Context, page, pageSize int) (*ListItemsResponse, error) {
	s.logger.Debug("listing items", "page", page, "page_size", pageSize)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	items, totalCount, err := s.repo.List(ctx, page, pageSize)
	if err != nil {
		s.logger.Error("failed to list items", "error", err)
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	return &ListItemsResponse{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// UpdateItem updates an existing item
func (s *ItemService) UpdateItem(ctx context.Context, id int64, req UpdateItemRequest) (*repository.Item, error) {
	s.logger.Info("updating item", "item_id", id)

	// Business logic validation
	if req.Price != nil && *req.Price < 0 {
		return nil, fmt.Errorf("price cannot be negative")
	}
	if req.Quantity != nil && *req.Quantity < 0 {
		return nil, fmt.Errorf("quantity cannot be negative")
	}

	item, err := s.repo.Update(ctx, id, req.Name, req.Description, req.Price, req.Quantity)
	if err != nil {
		s.logger.Error("failed to update item", "item_id", id, "error", err)
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	s.logger.Info("item updated successfully", "item_id", item.ID)
	return item, nil
}

// DeleteItem deletes an item by ID
func (s *ItemService) DeleteItem(ctx context.Context, id int64) error {
	s.logger.Info("deleting item", "item_id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete item", "item_id", id, "error", err)
		return fmt.Errorf("failed to delete item: %w", err)
	}

	s.logger.Info("item deleted successfully", "item_id", id)
	return nil
}
