package service

import (
	"fmt"
	"myproject/internal/model"
	"myproject/internal/repository"
	"strconv"
	"sync/atomic"
)

type ItemService struct {
	repo   repository.ItemRepo[model.Item]
	nextID atomic.Int64
}

func NewItemService(repo repository.ItemRepo[model.Item]) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) GetAll() ([]model.Item, error) {
	items, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("service get all %w", err)
	}

	return items, nil
}

func (s *ItemService) GetByID(id string) (model.Item, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return model.Item{}, fmt.Errorf("service get %q: %w", id, err)
	}
	return item, nil
}

func (s *ItemService) Create(item model.Item) (model.Item, error) {
	if item.Title == "" {
		return model.Item{}, &model.ValidationError{
			Field:   "title",
			Message: "должно быть",
		}
	}

	if item.Price < 0 {
		return model.Item{}, &model.ValidationError{
			Field:   "price",
			Message: "только неотрицательные числа",
		}
	}

	// Обратите внимание: значение берётся из результата Add, а не отдельным Load() потом.
	// Иначе между Add и Load другая горутина успела бы сделать свой Add,
	// и снова получились бы одинаковые ID.
	id := s.nextID.Add(1)
	item.ID = strconv.FormatInt(id, 10)

	if err := s.repo.Create(item); err != nil {
		return model.Item{}, fmt.Errorf("service create: %w", err)
	}

	return item, nil
}

func (s *ItemService) Update(id string, item model.Item) (model.Item, error) {
	if item.Title == "" {
		return model.Item{}, &model.ValidationError{
			Field:   "title",
			Message: "должно быть",
		}
	}

	if item.Price < 0 {
		return model.Item{}, &model.ValidationError{
			Field:   "price",
			Message: "только неотрицательные числа",
		}
	}

	item.ID = id

	if err := s.repo.Update(item); err != nil {
		return model.Item{}, fmt.Errorf("service update %q: %w", id, err)
	}

	return item, nil
}

func (s *ItemService) Delete(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("service delete %q: %w", id, err)
	}
	return nil
}
