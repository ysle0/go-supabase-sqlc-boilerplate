package repository

import (
	"context"
	"fmt"
	"sync"
	"time"
)

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

// ItemRepository handles data persistence for items
// In a real implementation, this would interact with a database (PostgreSQL, etc.)
type ItemRepository struct {
	mu     sync.RWMutex
	items  map[int64]*Item
	nextID int64
}

// NewItemRepository creates a new item repository
func NewItemRepository() *ItemRepository {
	return &ItemRepository{
		items:  make(map[int64]*Item),
		nextID: 1,
	}
}

// Create creates a new item
func (r *ItemRepository) Create(ctx context.Context, name, description string, price float64, quantity int) (*Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	item := &Item{
		ID:          r.nextID,
		Name:        name,
		Description: description,
		Price:       price,
		Quantity:    quantity,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	r.items[item.ID] = item
	r.nextID++

	return item, nil
}

// GetByID retrieves an item by ID
func (r *ItemRepository) GetByID(ctx context.Context, id int64) (*Item, error) {
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
func (r *ItemRepository) List(ctx context.Context, page, pageSize int) ([]Item, int64, error) {
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
func (r *ItemRepository) Update(ctx context.Context, id int64, name, description *string, price *float64, quantity *int) (*Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, exists := r.items[id]
	if !exists {
		return nil, fmt.Errorf("item not found")
	}

	// Update only provided fields
	if name != nil {
		item.Name = *name
	}
	if description != nil {
		item.Description = *description
	}
	if price != nil {
		item.Price = *price
	}
	if quantity != nil {
		item.Quantity = *quantity
	}
	item.UpdatedAt = time.Now()

	itemCopy := *item
	return &itemCopy, nil
}

// Delete deletes an item by ID
func (r *ItemRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.items[id]; !exists {
		return fmt.Errorf("item not found")
	}

	delete(r.items, id)
	return nil
}
