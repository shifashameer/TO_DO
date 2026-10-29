package store

import (
	"sync"
	"testing"
	"time"
	"todo-app/internal/models"
)

// Both implementations must satisfy this contract in a fresh, isolated store.
func storeContract(t *testing.T, s Store) {
	t.Helper()
	if _, err := s.Get("missing"); err == nil {
		t.Error("Get missing must fail")
	}
	if err := s.Update(models.Todo{ID: "missing", Title: "absent"}); err == nil {
		t.Error("Update missing must fail")
	}
	if err := s.Delete("missing"); err == nil {
		t.Error("Delete missing must fail")
	}
	for i, title := range []string{"Zebra", "Apple", "Middle"} {
		if err := s.Create(models.Todo{Title: title, Description: "description", Completed: i == 1, CreatedAt: time.Unix(int64(i+1), 0).UTC()}); err != nil {
			t.Fatal(err)
		}
	}
	for _, status := range []string{"all", "active", "completed"} {
		for _, field := range []string{"title", "created_at"} {
			for _, order := range []string{"asc", "desc"} {
				got, err := s.List(ListOptions{Status: status, Sort: field, Order: order})
				if err != nil {
					t.Fatal(err)
				}
				want := 3
				if status == "active" {
					want = 2
				}
				if status == "completed" {
					want = 1
				}
				if len(got) != want {
					t.Fatalf("%s: got %d want %d", status, len(got), want)
				}
				for i, item := range got {
					if item.ID == "" || item.CreatedAt.IsZero() {
						t.Fatal("missing generated fields")
					}
					if status == "active" && item.Completed || status == "completed" && !item.Completed {
						t.Fatal("wrong status")
					}
					if i > 0 {
						less := got[i-1].Title < item.Title
						if field == "created_at" {
							less = got[i-1].CreatedAt.Before(item.CreatedAt)
						}
						if less != (order == "asc") {
							t.Fatalf("wrong order: %s %s", field, order)
						}
					}
				}
			}
		}
	}
	for _, opts := range []ListOptions{{Status: "bad"}, {Sort: "title; DROP TABLE todos"}, {Order: "bad"}} {
		if _, err := s.List(opts); err == nil {
			t.Error("invalid options accepted")
		}
	}
	items, err := s.List(ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("got %d items", len(items))
	}
	item := items[0]
	item.Title = "Updated"
	item.Description = "changed"
	item.Completed = !item.Completed
	if err := s.Update(item); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(item.ID)
	if err != nil || got.Title != item.Title || got.Description != item.Description || got.Completed != item.Completed {
		t.Fatalf("update not saved: %+v %v", got, err)
	}
	got.Completed = false
	if err := s.Update(got); err != nil {
		t.Fatal(err)
	}
	got, err = s.Get(item.ID)
	if err != nil || got.Completed {
		t.Fatal("false update not saved")
	}
	if err := s.Delete(item.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(item.ID); err == nil {
		t.Fatal("delete not saved")
	}
}
func TestMemoryStoreContract(t *testing.T) { storeContract(t, NewMemoryStore()) }

func TestMemoryStoreConcurrentAccess(t *testing.T) {
	s := NewMemoryStore()
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Create(models.Todo{Title: "Concurrent"}); err != nil {
				t.Error(err)
			}
			items, err := s.List(ListOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			for _, item := range items {
				s.Get(item.ID)
				s.Update(item)
			}
		}()
	}
	wg.Wait()
	items, err := s.List(ListOptions{})
	if err != nil || len(items) != 30 {
		t.Fatalf("lost writes: %d %v", len(items), err)
	}
	for _, item := range items {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if err := s.Delete(id); err != nil {
				t.Error(err)
			}
		}(item.ID)
	}
	wg.Wait()
}
