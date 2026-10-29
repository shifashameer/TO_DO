package store

import (
	"errors"
	"todo-app/internal/models"
)

type ListOptions struct {
	Status string
	Sort   string
	Order  string
}

type Store interface {
	Create(todo models.Todo) error
	Get(id string) (models.Todo, error)
	List(options ListOptions) ([]models.Todo, error)
	Update(todo models.Todo) error
	Delete(id string) error
}
type UserStore interface {
	GetByUsername(username string) (models.User, error)
	GetByID(id string) (models.User, error)
}

var ErrNotFound = errors.New("record not found")
var ErrInvalidOptions = errors.New("invalid list options")

func (o ListOptions) Normalize() (ListOptions, error) {
	if o.Status == "" {
		o.Status = "all"
	}
	if o.Sort == "" {
		o.Sort = "created_at"
	}
	if o.Order == "" {
		o.Order = "asc"
	}
	if (o.Status != "all" && o.Status != "active" && o.Status != "completed") ||
		(o.Sort != "created_at" && o.Sort != "title") || (o.Order != "asc" && o.Order != "desc") {
		return o, ErrInvalidOptions
	}
	return o, nil
}
