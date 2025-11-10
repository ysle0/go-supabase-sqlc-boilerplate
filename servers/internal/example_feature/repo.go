package example_feature

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Repository handles data persistence for items
// In a real implementation, this would interact with a database (PostgreSQL, etc.)
type Repository struct {
	mu    sync.RWMutex
	items map[int64]*Item
	nextID int64
}

// NewRepository creates a new item repository
func NewRepository() *Repository {
	return &Repository{
		items: make(map[int64]*Item),
		nextID: 1,
	}
}

// Create creates a new item
func (r *Repository) Create(ctx context.Context, req CreateItemRequest) (*Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	item := &Item{
		ID:          r.nextID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Quantity:    req.Quantity,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	r.items[item.ID] = item
	r.nextID++

	return item, nil
}

// GetByID retrieves an item by ID
func (r *Repository) GetByID(ctx context.Context, id int64) (*Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, exists := r.items[id]
	if !exists {
		return nil, fmt.Errorf("item not found")
	}

	// Return a copy to prevent external modifications
	itemCopy := *item
	return &itemCopy, nil
}

// List retrieves all items with pagination
func (r *Repository) List(ctx context.Context, page, pageSize int) ([]Item, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalCount := int64(len(r.items))

	// Calculate pagination
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	items := make([]Item, 0, pageSize)
	i := 0
	for _, item := range r.items {
		if i >= offset && len(items) < pageSize {
			items = append(items, *item)
		}
		i++
	}

	return items, totalCount, nil
}

// Update updates an existing item
func (r *Repository) Update(ctx context.Context, id int64, req UpdateItemRequest) (*Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.items[id]
	if !exists {
		return nil, fmt.Errorf("item not found")
	}

	// Update only provided fields
	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.Quantity != nil {
		item.Quantity = *req.Quantity
	}
	item.UpdatedAt = time.Now()

	itemCopy := *item
	return &itemCopy, nil
}

// Delete deletes an item by ID
func (r *Repository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[id]; !exists {
		return fmt.Errorf("item not found")
	}

	delete(r.items, id)
	return nil
}
