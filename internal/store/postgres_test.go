package store

import (
	"testing"
	"time"
	"todo-app/internal/models"
)

func TestPostgresStoreCreateAndGet(t *testing.T) {
	todoStore := newTestStore(t)

	todo := models.Todo{
		Title:       "Learn GORM",
		Description: "Connect Go to PostgreSQL",
		Completed:   false,
	}

	err := todoStore.Create(todo)
	if err != nil {
		t.Fatalf("failed to create todo: %v", err)
	}

	// Find the created todo so we can get its generated UUID.
	todos, err := todoStore.List(ListOptions{})
	if err != nil {
		t.Fatalf("failed to list todos: %v", err)
	}

	var createdTodo models.Todo
	for _, item := range todos {
		if item.Title == "Learn GORM" {
			createdTodo = item
			break
		}
	}

	if createdTodo.ID == "" {
		t.Fatal("could not find created todo")
	}

	got, err := todoStore.Get(createdTodo.ID)
	if err != nil {
		t.Fatalf("failed to get todo: %v", err)
	}

	if got.Title != todo.Title {
		t.Errorf("expected title %q, got %q", todo.Title, got.Title)
	}
}
func TestPostgresStoreList(t *testing.T) {
	todoStore := newTestStore(t)

	todo1 := models.Todo{
		Title:       "Todo 1",
		Description: "First todo",
		Completed:   false,
	}

	todo2 := models.Todo{
		Title:       "Todo 2",
		Description: "Second todo",
		Completed:   true,
	}

	if err := todoStore.Create(todo1); err != nil {
		t.Fatalf("failed to create todo1: %v", err)
	}

	if err := todoStore.Create(todo2); err != nil {
		t.Fatalf("failed to create todo2: %v", err)
	}

	todos, err := todoStore.List(ListOptions{})
	if err != nil {
		t.Fatalf("failed to list todos: %v", err)
	}

	if len(todos) < 2 {
		t.Errorf("expected at least 2 todos, got %d", len(todos))
	}
}
func TestPostgresStoreUpdate(t *testing.T) {
	todoStore := newTestStore(t)

	todo := models.Todo{
		Title:       "Original Title",
		Description: "Original Description",
		Completed:   false,
	}

	if err := todoStore.Create(todo); err != nil {
		t.Fatalf("failed to create todo: %v", err)
	}

	// Get the created todo so we have its generated ID.
	todos, err := todoStore.List(ListOptions{})
	if err != nil {
		t.Fatalf("failed to list todos: %v", err)
	}

	var createdTodo models.Todo
	for _, item := range todos {
		if item.Title == "Original Title" {
			createdTodo = item
			break
		}
	}

	if createdTodo.ID == "" {
		t.Fatal("could not find created todo")
	}

	createdTodo.Title = "Updated Title"
	createdTodo.Description = "Updated Description"
	createdTodo.Completed = true

	if err := todoStore.Update(createdTodo); err != nil {
		t.Fatalf("failed to update todo: %v", err)
	}

	updated, err := todoStore.Get(createdTodo.ID)
	if err != nil {
		t.Fatalf("failed to get updated todo: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("expected title %q, got %q", "Updated Title", updated.Title)
	}

	if updated.Description != "Updated Description" {
		t.Errorf("expected description %q, got %q", "Updated Description", updated.Description)
	}

	if !updated.Completed {
		t.Errorf("expected todo to be completed")
	}
}
func TestPostgresStoreDelete(t *testing.T) {
	todoStore := newTestStore(t)

	todo := models.Todo{
		Title:       "Todo to Delete",
		Description: "This should be deleted",
		Completed:   false,
	}

	if err := todoStore.Create(todo); err != nil {
		t.Fatalf("failed to create todo: %v", err)
	}

	// Find the created todo so we can get its generated ID.
	todos, err := todoStore.List(ListOptions{})
	if err != nil {
		t.Fatalf("failed to list todos: %v", err)
	}

	var createdTodo models.Todo
	for _, item := range todos {
		if item.Title == "Todo to Delete" {
			createdTodo = item
			break
		}
	}

	if createdTodo.ID == "" {
		t.Fatal("could not find created todo")
	}

	// Delete it.
	if err := todoStore.Delete(createdTodo.ID); err != nil {
		t.Fatalf("failed to delete todo: %v", err)
	}

	// Try to get it again.
	_, err = todoStore.Get(createdTodo.ID)
	if err == nil {
		t.Errorf("expected todo to be deleted")
	}
}
func TestPostgresStoreListCompleted(t *testing.T) {
	todoStore := newTestStore(t)

	completedTodo := models.Todo{
		Title:       "Completed Todo",
		Description: "This is completed",
		Completed:   true,
	}

	activeTodo := models.Todo{
		Title:       "Active Todo",
		Description: "This is still active",
		Completed:   false,
	}

	if err := todoStore.Create(completedTodo); err != nil {
		t.Fatalf("failed to create completed todo: %v", err)
	}

	if err := todoStore.Create(activeTodo); err != nil {
		t.Fatalf("failed to create active todo: %v", err)
	}

	todos, err := todoStore.List(ListOptions{
		Status: "completed",
	})

	if err != nil {
		t.Fatalf("failed to list completed todos: %v", err)
	}

	for _, todo := range todos {
		if !todo.Completed {
			t.Errorf("expected only completed todos, but got active todo %q", todo.Title)
		}
	}
}
func TestPostgresStoreListSortByTitleAsc(t *testing.T) {
	todoStore := newTestStore(t)
	todoStore.db.Exec("DELETE FROM todos")

	todo1 := models.Todo{
		Title:       "checkpoint7 Zebra",
		Description: "Last alphabetically",
	}

	todo2 := models.Todo{
		Title:       "checkpoint7 Apple",
		Description: "First alphabetically",
	}

	if err := todoStore.Create(todo1); err != nil {
		t.Fatalf("failed to create todo1: %v", err)
	}

	if err := todoStore.Create(todo2); err != nil {
		t.Fatalf("failed to create todo2: %v", err)
	}

	todos, err := todoStore.List(ListOptions{
		Sort:  "title",
		Order: "asc",
	})
	if err != nil {
		t.Fatalf("failed to list todos: %v", err)
	}
	for _, todo := range todos {
		t.Logf("Title: %q", todo.Title)
	}

	var appleIndex = -1
	var zebraIndex = -1

	for i, todo := range todos {
		if todo.Title == "checkpoint7 Apple" {
			appleIndex = i
		}

		if todo.Title == "checkpoint7 Zebra" {
			zebraIndex = i
		}
	}

	if appleIndex == -1 {
		t.Fatal("could not find checkpoint7 Apple")
	}

	if zebraIndex == -1 {
		t.Fatal("could not find checkpoint7 Zebra")
	}

	if appleIndex >= zebraIndex {
		t.Errorf("expected checkpoint7 Apple before checkpoint7 Zebra")
	}
}
func TestPostgresStoreListSortByTitleDesc(t *testing.T) {
	todoStore := newTestStore(t)
	todoStore.db.Exec("DELETE FROM todos")

	todo1 := models.Todo{
		Title: "checkpoint7 Zebra Desc",
	}

	todo2 := models.Todo{
		Title: "checkpoint7 Apple Desc",
	}

	if err := todoStore.Create(todo1); err != nil {
		t.Fatalf("failed to create todo1: %v", err)
	}

	if err := todoStore.Create(todo2); err != nil {
		t.Fatalf("failed to create todo2: %v", err)
	}

	todos, err := todoStore.List(ListOptions{
		Sort:  "title",
		Order: "desc",
	})
	if err != nil {
		t.Fatalf("failed to list todos: %v", err)
	}

	var appleIndex = -1
	var zebraIndex = -1

	for i, todo := range todos {
		if todo.Title == "checkpoint7 Zebra Desc" {
			zebraIndex = i
		}

		if todo.Title == "checkpoint7 Apple Desc" {
			appleIndex = i
		}
	}

	if appleIndex == -1 {
		t.Fatal("could not find checkpoint7 Apple Desc")
	}

	if zebraIndex == -1 {
		t.Fatal("could not find checkpoint7 Zebra Desc")
	}

	if zebraIndex >= appleIndex {
		t.Errorf("expected checkpoint7 Zebra Desc before checkpoint7 Apple Desc")
	}
}
func TestPostgresStoreListSortByCreatedAtAsc(t *testing.T) {
	todoStore := newTestStore(t)
	todoStore.db.Exec("DELETE FROM todos")

	todo1 := models.Todo{
		Title:     "checkpoint7 Older",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	todo2 := models.Todo{
		Title:     "checkpoint7 Newer",
		CreatedAt: time.Now(),
	}

	if err := todoStore.Create(todo1); err != nil {
		t.Fatalf("failed to create todo1: %v", err)
	}

	if err := todoStore.Create(todo2); err != nil {
		t.Fatalf("failed to create todo2: %v", err)
	}

	todos, err := todoStore.List(ListOptions{
		Sort:  "created_at",
		Order: "asc",
	})
	if err != nil {
		t.Fatalf("failed to list todos: %v", err)
	}

	var olderIndex = -1
	var newerIndex = -1

	for i, todo := range todos {
		if todo.Title == "checkpoint7 Older" {
			olderIndex = i
		}

		if todo.Title == "checkpoint7 Newer" {
			newerIndex = i
		}
	}

	if olderIndex == -1 {
		t.Fatal("could not find checkpoint7 Older")
	}

	if newerIndex == -1 {
		t.Fatal("could not find checkpoint7 Newer")
	}

	if olderIndex >= newerIndex {
		t.Errorf("expected checkpoint7 Older before checkpoint7 Newer")
	}
}
func TestPostgresStoreListSortByCreatedAtDesc(t *testing.T) {
	todoStore := newTestStore(t)
	todoStore.db.Exec("DELETE FROM todos")

	todo1 := models.Todo{
		Title:     "checkpoint7 Older Desc",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}

	todo2 := models.Todo{
		Title:     "checkpoint7 Newer Desc",
		CreatedAt: time.Now(),
	}

	if err := todoStore.Create(todo1); err != nil {
		t.Fatalf("failed to create todo1: %v", err)
	}

	if err := todoStore.Create(todo2); err != nil {
		t.Fatalf("failed to create todo2: %v", err)
	}

	todos, err := todoStore.List(ListOptions{
		Sort:  "created_at",
		Order: "desc",
	})
	if err != nil {
		t.Fatalf("failed to list todos: %v", err)
	}
	for _, todo := range todos {
		t.Logf("Title: %q, CreatedAt: %v", todo.Title, todo.CreatedAt)
	}

	var olderIndex = -1
	var newerIndex = -1

	for i, todo := range todos {
		if todo.Title == "checkpoint7 Older Desc" {
			olderIndex = i
		}

		if todo.Title == "checkpoint7 Newer Desc" {
			newerIndex = i
		}
	}

	if olderIndex == -1 {
		t.Fatal("could not find checkpoint7 Older Desc")
	}

	if newerIndex == -1 {
		t.Fatal("could not find checkpoint7 Newer Desc")
	}

	if newerIndex >= olderIndex {
		t.Errorf("expected checkpoint7 Newer Desc before checkpoint7 Older Desc")
	}
}
