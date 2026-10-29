package store

import (
	"sort"
	"strconv"
	"sync"
	"time"

	"todo-app/internal/models"
)

type MemoryStore struct {
	todos  map[string]models.Todo
	mu     sync.Mutex
	nextID int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		todos:  make(map[string]models.Todo),
		nextID: 1,
	}
}

func (m *MemoryStore) Create(todo models.Todo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	todo.ID = strconv.Itoa(m.nextID)
	m.nextID++

	// Set CreatedAt if it was not provided.
	if todo.CreatedAt.IsZero() {
		todo.CreatedAt = time.Now()
	}

	m.todos[todo.ID] = todo

	return nil
}

func (m *MemoryStore) Get(id string) (models.Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	todo, exists := m.todos[id]

	if !exists {
		return models.Todo{}, ErrNotFound
	}

	return todo, nil
}

func (m *MemoryStore) List(options ListOptions) ([]models.Todo, error) {
	var err error
	options, err = options.Normalize()
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	todos := make([]models.Todo, 0)

	// Filter by status
	for _, todo := range m.todos {
		switch options.Status {
		case "completed":
			if !todo.Completed {
				continue
			}

		case "active":
			if todo.Completed {
				continue
			}
		}

		todos = append(todos, todo)
	}

	// Sort
	order := options.Order

	if order != "desc" {
		order = "asc"
	}

	switch options.Sort {

	case "title":
		sort.Slice(todos, func(i, j int) bool {
			if order == "desc" {
				return todos[i].Title > todos[j].Title
			}

			return todos[i].Title < todos[j].Title
		})

	case "created_at":
		sort.Slice(todos, func(i, j int) bool {
			if order == "desc" {
				return todos[i].CreatedAt.After(todos[j].CreatedAt)
			}

			return todos[i].CreatedAt.Before(todos[j].CreatedAt)
		})
	}

	return todos, nil
}

func (m *MemoryStore) Update(todo models.Todo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.todos[todo.ID]

	if !exists {
		return ErrNotFound
	}

	m.todos[todo.ID] = todo

	return nil
}

func (m *MemoryStore) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, exists := m.todos[id]

	if !exists {
		return ErrNotFound
	}

	delete(m.todos, id)

	return nil
}

var _ Store = (*MemoryStore)(nil)
