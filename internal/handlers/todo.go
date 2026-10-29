package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"todo-app/internal/models"
	"todo-app/internal/store"

	"github.com/go-chi/chi/v5"
)

type TodoHandler struct {
	store store.Store
}

func NewTodoHandler(s store.Store) *TodoHandler {
	return &TodoHandler{
		store: s,
	}
}
func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	todos, err := h.store.List(store.ListOptions{
		Status: status,
		Sort:   sort,
		Order:  order,
	})

	if err != nil {
		storeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}
func (h *TodoHandler) Index(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	todos, err := h.store.List(store.ListOptions{
		Status: status,
		Sort:   sort,
		Order:  order,
	})

	if err != nil {
		storeError(w, err)
		return
	}

	render(w, "index.html", todos)
}
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	title := strings.TrimSpace(r.PostForm.Get("title"))
	description := r.PostForm.Get("description")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	todo := models.Todo{

		Title:       title,
		Description: description,
		Completed:   false,
	}

	err := h.store.Create(todo)
	if err != nil {
		storeError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)

}
func (h *TodoHandler) ToggleComplete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	todo, err := h.store.Get(id)
	if err != nil {
		storeError(w, err)
		return
	}
	todo.Completed = !todo.Completed
	err = h.store.Update(todo)
	if err != nil {
		storeError(w, err)
		return
	}

}
func (h *TodoHandler) EditTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	todo, err := h.store.Get(id)
	if err != nil {
		storeError(w, err)
		return
	}

	render(w, "edit.html", todo)
}
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	todo, err := h.store.Get(id)
	if err != nil {
		storeError(w, err)
		return
	}
	if !parseForm(w, r) {
		return
	}

	title := strings.TrimSpace(r.PostForm.Get("title"))
	description := r.PostForm.Get("description")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	todo.Title = title
	todo.Description = description
	err = h.store.Update(todo)
	if err != nil {
		storeError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	err := h.store.Delete(id)
	if err != nil {
		storeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func storeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		http.Error(w, "Todo not found", http.StatusNotFound)
	case errors.Is(err, store.ErrInvalidOptions):
		http.Error(w, "Invalid filter or sort options", http.StatusBadRequest)
	default:
		http.Error(w, "Storage operation failed", http.StatusInternalServerError)
	}
}
