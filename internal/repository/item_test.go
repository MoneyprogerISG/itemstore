package repository_test

import (
	"myproject/internal/model"
	"myproject/internal/repository"
	"strconv"
	"sync"
	"testing"
)

// отдельный тест на сам репозиторий, где читатели и писатели работают одновременно
func TestInMemoryRepo_Concurrent(t *testing.T) {
	repo := repository.NewInMemoryItemRepo[model.Item]()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		id := strconv.Itoa(i)
		wg.Add(2)
		go func() { // писатель
			defer wg.Done()
			_ = repo.Create(model.Item{ID: id, Title: "a"})
			_ = repo.Update(model.Item{ID: id, Title: "b"})
		}()
		go func() { // читатель
			defer wg.Done()
			_, _ = repo.GetAll()
			_, _ = repo.GetByID(id)
		}()
	}
	wg.Wait()

	items, _ := repo.GetAll()
	if len(items) != 100 {
		t.Fatalf("items = %d, want 100", len(items))
	}
}
