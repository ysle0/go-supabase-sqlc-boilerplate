package example_feature

import (
	"context"
	"fmt"
	"log/slog"
)

// Service contains business logic for items
type Service struct {
	repo   *Repository
	logger *slog.Logger
}

// NewService creates a new item service
func NewService(repo *Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateItem creates a new item with validation
func (s *Service) CreateItem(ctx context.Context, req CreateItemRequest) (*Item, error) {
	s.logger.Info("creating item", "name", req.Name)

	// Business logic validation
	if req.Price < 0 {
		return nil, fmt.Errorf("price cannot be negative")
	}
	if req.Quantity < 0 {
		return nil, fmt.Errorf("quantity cannot be negative")
	}

	item, err := s.repo.Create(ctx, req)
	if err != nil {
		s.logger.Error("failed to create item", "error", err)
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	s.logger.Info("item created successfully", "item_id", item.ID)
	return item, nil
}

// GetItem retrieves an item by ID
func (s *Service) GetItem(ctx context.Context, id int64) (*Item, error) {
	s.logger.Debug("getting item", "item_id", id)

	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get item", "item_id", id, "error", err)
		return nil, fmt.Errorf("item not found")
	}

	return item, nil
}

// ListItems retrieves all items with pagination
func (s *Service) ListItems(ctx context.Context, page, pageSize int) (*ListItemsResponse, error) {
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
func (s *Service) UpdateItem(ctx context.Context, id int64, req UpdateItemRequest) (*Item, error) {
	s.logger.Info("updating item", "item_id", id)

	// Business logic validation
	if req.Price != nil && *req.Price < 0 {
		return nil, fmt.Errorf("price cannot be negative")
	}
	if req.Quantity != nil && *req.Quantity < 0 {
		return nil, fmt.Errorf("quantity cannot be negative")
	}

	item, err := s.repo.Update(ctx, id, req)
	if err != nil {
		s.logger.Error("failed to update item", "item_id", id, "error", err)
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	s.logger.Info("item updated successfully", "item_id", item.ID)
	return item, nil
}

// DeleteItem deletes an item by ID
func (s *Service) DeleteItem(ctx context.Context, id int64) error {
	s.logger.Info("deleting item", "item_id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete item", "item_id", id, "error", err)
		return fmt.Errorf("failed to delete item: %w", err)
	}

	s.logger.Info("item deleted successfully", "item_id", id)
	return nil
}
