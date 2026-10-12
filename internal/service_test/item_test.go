package service_test

import (
	"errors"
	"myproject/internal/model"
	"myproject/internal/repository"
	"myproject/internal/service"
	"sync"
	"testing"
)

type FakeItemRepo struct {
	items map[string]model.Item
}

func NewFakeItemRepo() *FakeItemRepo {
	return &FakeItemRepo{items: make(map[string]model.Item)}
}

func (r *FakeItemRepo) GetAll() ([]model.Item, error) {
	result := make([]model.Item, 0, len(r.items))

	for _, it := range r.items {
		result = append(result, it)
	}

	return result, nil
}

func (r *FakeItemRepo) GetByID(id string) (model.Item, error) {
	item, ok := r.items[id]
	if !ok {
		return model.Item{}, model.ErrNotFound
	}

	return item, nil
}

func (r *FakeItemRepo) Create(item model.Item) error {
	if _, ok := r.items[item.ID]; ok {
		return model.ErrAlreadyExists
	}
	r.items[item.ID] = item

	return nil
}

func (r *FakeItemRepo) Update(item model.Item) error {
	if _, ok := r.items[item.ID]; !ok {
		return model.ErrNotFound
	}

	r.items[item.ID] = item

	return nil
}

func (r *FakeItemRepo) Delete(id string) error {
	if _, ok := r.items[id]; !ok {
		return model.ErrNotFound
	}
	delete(r.items, id)

	return nil
}

func TestItemService_Create(t *testing.T) {
	tests := []struct {
		name  string
		item  model.Item
		check func(t *testing.T, item model.Item, err error)
	}{
		{
			name: "empty title",
			item: model.Item{Price: 15},
			check: func(t *testing.T, item model.Item, err error) {
				var ve *model.ValidationError
				if !errors.As(err, &ve) {
					t.Fatalf("want ValidationError, got %v", err)
				}
				if ve.Field != "title" {
					t.Errorf("field = %q, want title", ve.Field)
				}
			},
		},

		{
			name: "negative price",
			item: model.Item{Title: "Хлеб", Price: -15},
			check: func(t *testing.T, item model.Item, err error) {
				var ve *model.ValidationError
				if !errors.As(err, &ve) {
					t.Fatalf("want ValidationError, got %v", err)
				}

				if ve.Field != "price" {
					t.Errorf("field = %q, want price", ve.Field)
				}
			},
		},

		{
			name: "success",
			item: model.Item{Title: "Хлеб", Price: 15},
			check: func(t *testing.T, item model.Item, err error) {
				if err != nil {
					t.Fatal(err)
				}

				if item.ID == "" {
					t.Error("want ID to be assigned")
				}

				if item.Title != "Хлеб" {
					t.Errorf("want title: %s, get: %s", "Хлеб", item.Title)
				}

				if item.Price != 15 {
					t.Errorf("want price: %d, get: %d", 15, item.Price)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewItemService(NewFakeItemRepo())
			created, err := svc.Create(tt.item)
			tt.check(t, created, err)
		})
	}
}

func TestGet_NoFound(t *testing.T) {
	svc := service.NewItemService(NewFakeItemRepo())

	_, err := svc.GetByID("999")
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatal("want ErrNotFound, get:", err)
	}
}

func TestDelete_NoFound(t *testing.T) {
	svc := service.NewItemService(NewFakeItemRepo())
	err := svc.Delete("67")

	if !errors.Is(err, model.ErrNotFound) {
		t.Fatal("want ErrNotFound, get:", err)
	}
}

func TestItemService_Update(t *testing.T) {
	svc := service.NewItemService(NewFakeItemRepo())

	created, err := svc.Create(model.Item{Title: "Хлеб", Price: 15})
	if err != nil {
		t.Fatalf("setup: не удалось создать item: %v", err)
	}

	updated, err := svc.Update(created.ID, model.Item{Title: "Батон", Price: 25})
	if err != nil {
		t.Fatalf("не ожидал ошибку: %v", err)
	}

	if updated.ID != created.ID {
		t.Errorf("id = %q, want %q", updated.ID, created.ID)
	}

	if updated.Title != "Батон" {
		t.Errorf("title = %q, want %q", updated.Title, "Батон")
	}

	if updated.Price != 25 {
		t.Errorf("price = %d, want %d", updated.Price, 25)
	}
}

// тест, который одновременно создаёт 100 элементов через настоящий репозиторий:
func TestItemService_ConcurrentCreate(t *testing.T) {
	svc := service.NewItemService(repository.NewInMemoryItemRepo[model.Item]())

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.Create(model.Item{Title: "Хлеб", Price: 10}); err != nil {
				t.Errorf("create: %v", err)
			}
		}()
	}
	wg.Wait()

	items, _ := svc.GetAll()
	if len(items) != 100 {
		t.Fatalf("items = %d, want 100", len(items))
	}
}
