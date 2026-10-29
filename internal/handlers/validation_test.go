package handlers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"todo-app/internal/auth"
	"todo-app/internal/models"
	"todo-app/internal/store"
)

func todoRouter(s store.Store) http.Handler {
	h := NewTodoHandler(s)
	r := chi.NewRouter()
	r.Get("/", h.Index)
	r.Get("/api/todos", h.GetTodos)
	r.Post("/todos", h.CreateTodo)
	r.Get("/todos/{id}/edit", h.EditTodo)
	r.Post("/todos/{id}", h.UpdateTodo)
	r.Patch("/api/todos/{id}/toggle", h.ToggleComplete)
	r.Delete("/api/todos/{id}", h.DeleteTodo)
	return r
}
func TestTodoValidation(t *testing.T) {
	s := store.NewMemoryStore()
	s.Create(models.Todo{Title: "Existing"})
	r := todoRouter(s)
	for _, path := range []string{"/todos", "/todos/1"} {
		for _, body := range []string{"title=+++", "title=valid&description=%zz", "title="} {
			req := httptest.NewRequest("POST", path, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
			if rr.Code != 400 {
				t.Errorf("%s %q got %d", path, body, rr.Code)
			}
		}
	}
	for _, path := range []string{"/api/todos?status=nope", "/api/todos?sort=bad", "/api/todos?order=bad", "/?status=nope"} {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, httptest.NewRequest("GET", path, nil))
		if rr.Code != 400 {
			t.Errorf("%s got %d", path, rr.Code)
		}
	}
}
func TestGetTodosJSON(t *testing.T) {
	s := store.NewMemoryStore()
	s.Create(models.Todo{Title: "Zebra"})
	s.Create(models.Todo{Title: "Apple"})
	s.Create(models.Todo{Title: "Done", Completed: true})
	rr := httptest.NewRecorder()
	todoRouter(s).ServeHTTP(rr, httptest.NewRequest("GET", "/api/todos?status=active&sort=title&order=asc", nil))
	var got []models.Todo
	if rr.Code != 200 || json.Unmarshal(rr.Body.Bytes(), &got) != nil || len(got) != 2 || got[0].Title != "Apple" || got[1].Title != "Zebra" {
		t.Fatalf("bad response: %d %s", rr.Code, rr.Body)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatal("wrong content type")
	}
}
func TestMissingTodos(t *testing.T) {
	r := todoRouter(store.NewMemoryStore())
	for _, tc := range []struct{ method, path string }{{"GET", "/todos/missing/edit"}, {"POST", "/todos/missing"}, {"PATCH", "/api/todos/missing/toggle"}, {"DELETE", "/api/todos/missing"}} {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, httptest.NewRequest(tc.method, tc.path, nil))
		if rr.Code != 404 {
			t.Errorf("%s got %d", tc.path, rr.Code)
		}
	}
}
func TestRejectForgedUserIDAndAPIUnauthorized(t *testing.T) {
	users := &fakeUserStore{user: models.User{ID: "user-1"}}
	h := RequireAuth(users, auth.NewSessions(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	for _, token := range []string{"", "user-1", "garbage"} {
		req := httptest.NewRequest("GET", "/api/todos", nil)
		if token != "" {
			req.AddCookie(&http.Cookie{Name: "session", Value: token})
		}
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != 401 {
			t.Errorf("cookie %q got %d", token, rr.Code)
		}
	}
}

type brokenStore struct{ store.Store }

func (brokenStore) Create(models.Todo) error { return errors.New("secret database details") }
func (brokenStore) Get(string) (models.Todo, error) {
	return models.Todo{}, errors.New("secret database details")
}
func (brokenStore) List(store.ListOptions) ([]models.Todo, error) {
	return nil, errors.New("secret database details")
}
func (brokenStore) Update(models.Todo) error { return errors.New("secret database details") }
func (brokenStore) Delete(string) error      { return errors.New("secret database details") }
func TestStorageErrorsAreInternal(t *testing.T) {
	r := todoRouter(brokenStore{})
	for _, tc := range []struct{ method, path string }{{"GET", "/"}, {"GET", "/api/todos"}, {"POST", "/todos"}, {"GET", "/todos/1/edit"}, {"POST", "/todos/1"}, {"PATCH", "/api/todos/1/toggle"}, {"DELETE", "/api/todos/1"}} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader("title=valid"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != 500 || strings.Contains(rr.Body.String(), "secret") {
			t.Errorf("%s leaked or wrong status: %d %s", tc.path, rr.Code, rr.Body)
		}
	}
}
func TestInvalidContentTypeAndLargeBody(t *testing.T) {
	for _, tc := range []struct {
		body, content string
		status        int
	}{{`{"title":`, "application/json", 415}, {"title=" + strings.Repeat("x", 70<<10), "application/x-www-form-urlencoded", 400}} {
		req := httptest.NewRequest("POST", "/todos", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", tc.content)
		rr := httptest.NewRecorder()
		todoRouter(store.NewMemoryStore()).ServeHTTP(rr, req)
		if rr.Code != tc.status {
			t.Fatalf("got %d", rr.Code)
		}
	}
}

func TestFormContentTypeWithCharset(t *testing.T) {
	req := httptest.NewRequest("POST", "/todos", strings.NewReader("title=Hello"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	rr := httptest.NewRecorder()
	todoRouter(store.NewMemoryStore()).ServeHTTP(rr, req)
	if rr.Code != 303 {
		t.Fatalf("valid form rejected: %d", rr.Code)
	}
}

type updateFailureStore struct{ brokenStore }

func (updateFailureStore) Get(string) (models.Todo, error) {
	return models.Todo{ID: "1", Title: "Existing"}, nil
}
func TestUpdateFailure(t *testing.T) {
	r := todoRouter(updateFailureStore{})
	for _, tc := range []struct{ method, path string }{{"POST", "/todos/1"}, {"PATCH", "/api/todos/1/toggle"}} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader("title=Valid"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != 500 || strings.Contains(rr.Body.String(), "secret") {
			t.Fatal("update failure not safely handled")
		}
	}
}
