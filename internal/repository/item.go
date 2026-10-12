package repository

import (
	"fmt"
	"myproject/internal/model"
	"sync"
)

type Identifiable interface {
	GetID() string
}

type ItemRepo[V Identifiable] interface {
	GetAll() ([]V, error)
	GetByID(id string) (V, error)
	Create(item V) error
	Update(item V) error
	Delete(id string) error
}

type InMemoryItemRepo[V Identifiable] struct {
	mu    sync.RWMutex
	items map[string]V
}

func NewInMemoryItemRepo[V Identifiable]() *InMemoryItemRepo[V] {
	return &InMemoryItemRepo[V]{
		items: make(map[string]V),
	}
}

func (r *InMemoryItemRepo[V]) GetAll() ([]V, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]V, 0, len(r.items))
	for _, it := range r.items {
		result = append(result, it)
	}

	return result, nil
}

func (r *InMemoryItemRepo[V]) GetByID(id string) (V, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[id]
	if !ok {
		var zero V
		return zero, fmt.Errorf("get item %q: %w", id, model.ErrNotFound)
	}

	return item, nil
}

func (r *InMemoryItemRepo[V]) Create(item V) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[item.GetID()]; ok {
		return fmt.Errorf("create item %q: %w", item.GetID(), model.ErrAlreadyExists)
	}
	r.items[item.GetID()] = item
	return nil
}

func (r *InMemoryItemRepo[V]) Update(item V) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[item.GetID()]; !ok {
		return fmt.Errorf("update item %q: %w", item.GetID(), model.ErrNotFound)
	}
	r.items[item.GetID()] = item
	return nil
}

func (r *InMemoryItemRepo[V]) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return fmt.Errorf("delete item %q: %w", id, model.ErrNotFound)
	}
	delete(r.items, id)

	return nil
}
