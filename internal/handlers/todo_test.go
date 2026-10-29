package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"todo-app/internal/models"
	"todo-app/internal/store"

	"github.com/go-chi/chi/v5"
)

func TestIndex(t *testing.T) {
	todoStore := store.NewMemoryStore()
	todoStore.Create(models.Todo{
		ID:          "1",
		Title:       "Learn Go",
		Description: "Testing",
		Completed:   false,
	})
	todoHandler := NewTodoHandler(todoStore)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	todoHandler.Index(rr, req)
	t.Log(rr.Body.String())
	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Learn Go") {
		t.Errorf("response body does not contain todo title")
	}
}
func TestCreateTodo(t *testing.T) {
	todoStore := store.NewMemoryStore()
	todoHandler := NewTodoHandler(todoStore)
	form := strings.NewReader("title=Learn+Go&description=Learn+HTTP+forms")
	req := httptest.NewRequest(
		http.MethodPost,
		"/todos",
		form,
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	todoHandler.CreateTodo(rr, req)
	todo, err := todoStore.Get("1")
	if err != nil {
		t.Fatalf("todo was not saved: %v", err)
	}
	if todo.Title != "Learn Go" {
		t.Errorf("expected title %q, got %q", "Learn Go", todo.Title)
	}
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected redirect status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}
func TestToggleComplete(t *testing.T) {
	todoStore := store.NewMemoryStore()

	todoStore.Create(models.Todo{
		ID:          "1",
		Title:       "Learn Go",
		Description: "Testing",
		Completed:   false,
	})

	todoHandler := NewTodoHandler(todoStore)

	req := httptest.NewRequest(http.MethodPatch, "/api/todos/1/toggle", nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	rr := httptest.NewRecorder()

	todoHandler.ToggleComplete(rr, req)
	todo, err := todoStore.Get("1")
	if err != nil {
		t.Fatalf("todo was not found: %v", err)
	}

	if todo.Completed != true {
		t.Errorf("expected todo to be completed")
	}
}
func TestEditTodo(t *testing.T) {
	todoStore := store.NewMemoryStore()

	todoStore.Create(models.Todo{
		ID:          "1",
		Title:       "Learn Go",
		Description: "Testing",
		Completed:   false,
	})

	todoHandler := NewTodoHandler(todoStore)

	req := httptest.NewRequest(http.MethodGet, "/todos/1/edit", nil)

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	rr := httptest.NewRecorder()

	todoHandler.EditTodo(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Learn Go") {
		t.Errorf("response does not contain todo title")
	}
}
func TestUpdateTodo(t *testing.T) {
	todoStore := store.NewMemoryStore()

	todoStore.Create(models.Todo{
		ID:          "1",
		Title:       "Learn Go",
		Description: "Testing",
		Completed:   false,
	})

	todoHandler := NewTodoHandler(todoStore)

	form := strings.NewReader("title=Learn+Advanced+Go&description=Updated+description")

	req := httptest.NewRequest(
		http.MethodPost,
		"/todos/1",
		form,
	)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))

	rr := httptest.NewRecorder()

	todoHandler.UpdateTodo(rr, req)

	todo, err := todoStore.Get("1")
	if err != nil {
		t.Fatalf("todo was not found: %v", err)
	}

	if todo.Title != "Learn Advanced Go" {
		t.Errorf("expected title %q, got %q", "Learn Advanced Go", todo.Title)
	}

	if todo.Description != "Updated description" {
		t.Errorf("expected description %q, got %q", "Updated description", todo.Description)
	}

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected redirect status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}
func TestDeleteTodo(t *testing.T) {
	todoStore := store.NewMemoryStore() //fake memory store

	todoStore.Create(models.Todo{ //fake content
		ID:          "1",
		Title:       "Learn Go",
		Description: "Testing",
		Completed:   false,
	})

	todoHandler := NewTodoHandler(todoStore) //fake new handler

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/todos/1",
		nil,
	)

	routeCtx := chi.NewRouteContext() //fake router
	routeCtx.URLParams.Add("id", "1")
	req = req.WithContext(
		context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx), //connection req info to fake router so that it can connect to handler
	)

	rr := httptest.NewRecorder() //response of handler

	todoHandler.DeleteTodo(rr, req)

	_, err := todoStore.Get("1")
	if err == nil {
		t.Errorf("expected todo to be deleted")
	}

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status code %d, got %d", http.StatusNoContent, rr.Code)
	}
}
