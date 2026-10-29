package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"todo-app/internal/auth"
	"todo-app/internal/models"
	"todo-app/internal/store"
)

type testUsers struct{ user models.User }

func (u testUsers) GetByUsername(name string) (models.User, error) {
	if name == u.user.Username {
		return u.user, nil
	}
	return models.User{}, store.ErrNotFound
}
func (u testUsers) GetByID(id string) (models.User, error) {
	if id == u.user.ID {
		return u.user, nil
	}
	return models.User{}, store.ErrNotFound
}
func TestRouterWorkflow(t *testing.T) {
	hash, err := auth.HashPassword("test-password-123")
	if err != nil {
		t.Fatal(err)
	}
	r := newRouter(store.NewMemoryStore(), testUsers{models.User{ID: "user", Username: "alice", PasswordHash: hash}}, false)
	request := func(method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		if cookie != nil {
			req.AddCookie(cookie)
		}
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		return rr
	}
	for _, path := range []string{"/healthz", "/login", "/static/style.css"} {
		rr := request("GET", path, "", nil)
		if rr.Code != 200 {
			t.Fatalf("%s: %d", path, rr.Code)
		}
		if path == "/healthz" && rr.Body.String() != "ok" {
			t.Fatal("incorrect health body")
		}
	}
	for _, tc := range []struct {
		method, path string
		status       int
	}{{"GET", "/", 303}, {"GET", "/todos/1/edit", 303}, {"POST", "/todos", 303}, {"POST", "/todos/1", 303}, {"GET", "/api/todos", 401}, {"PATCH", "/api/todos/1/toggle", 401}, {"DELETE", "/api/todos/1", 401}} {
		if rr := request(tc.method, tc.path, "", nil); rr.Code != tc.status {
			t.Errorf("%s %s: %d", tc.method, tc.path, rr.Code)
		}
	}
	bad := request("POST", "/login", "username=alice&password=wrong", nil)
	if bad.Code != 401 || len(bad.Result().Cookies()) != 0 {
		t.Fatal("bad login accepted")
	}
	login := request("POST", "/login", "username=alice&password=test-password-123", &http.Cookie{Name: "session", Value: "user"})
	if login.Code != 303 || len(login.Result().Cookies()) != 1 {
		t.Fatalf("login: %d", login.Code)
	}
	cookie := login.Result().Cookies()[0]
	if cookie.Value == "user" || len(cookie.Value) < 40 || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.MaxAge <= 0 || cookie.Expires.IsZero() {
		t.Fatal("unsafe session cookie")
	}
	form := url.Values{"title": {"<b>Read Go</b>"}, "description": {"Learning templates"}}.Encode()
	created := request("POST", "/todos", form, cookie)
	if created.Code != 303 || created.Header().Get("Location") != "/" {
		t.Fatal("create failed")
	}
	page := request("GET", "/", "", cookie)
	if page.Code != 200 || !strings.Contains(page.Body.String(), "&lt;b&gt;Read Go&lt;/b&gt;") || !strings.Contains(page.Body.String(), "Learning templates") || !strings.Contains(page.Body.String(), "/static/style.css") {
		t.Fatal("list rendering/escaping failed")
	}
	list := request("GET", "/api/todos", "", cookie)
	var items []models.Todo
	if err := json.Unmarshal(list.Body.Bytes(), &items); err != nil || len(items) != 1 {
		t.Fatal("invalid JSON list")
	}
	id := items[0].ID
	edit := request("GET", "/todos/"+id+"/edit", "", cookie)
	if edit.Code != 200 || !strings.Contains(edit.Body.String(), "Learning templates") {
		t.Fatal("edit page failed")
	}
	updated := request("POST", "/todos/"+id, "title=Changed&description=Updated", cookie)
	if updated.Code != 303 {
		t.Fatal("update failed")
	}
	for _, want := range []bool{true, false} {
		toggle := request("PATCH", "/api/todos/"+id+"/toggle", "", cookie)
		if toggle.Code != 200 {
			t.Fatal("toggle failed")
		}
		list = request("GET", "/api/todos", "", cookie)
		if err := json.Unmarshal(list.Body.Bytes(), &items); err != nil || len(items) != 1 || items[0].Completed != want {
			t.Fatal("toggle state wrong")
		}
	}
	if rr := request("DELETE", "/api/todos/"+id, "", cookie); rr.Code != 204 {
		t.Fatal("delete failed")
	}
	list = request("GET", "/api/todos", "", cookie)
	if strings.TrimSpace(list.Body.String()) != "[]" {
		t.Fatal("empty list must be array")
	}
	// A second login rotates and revokes the previous token.
	rotated := request("POST", "/login", "username=alice&password=test-password-123", cookie)
	next := rotated.Result().Cookies()[0]
	if next.Value == cookie.Value {
		t.Fatal("session not rotated")
	}
	if request("GET", "/api/todos", "", cookie).Code != 401 {
		t.Fatal("old token still valid")
	}
	logout := request("POST", "/logout", "", next)
	if logout.Code != 303 || logout.Result().Cookies()[0].MaxAge != -1 {
		t.Fatal("logout failed")
	}
	if request("GET", "/api/todos", "", next).Code != 401 {
		t.Fatal("logout did not revoke session")
	}
	if request("GET", "/logout", "", nil).Code != 405 {
		t.Fatal("GET logout must not mutate session")
	}
	req := httptest.NewRequest("POST", "/login", strings.NewReader("username=alice&password=test-password-123"))
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != 403 {
		t.Fatal("cross-origin login accepted")
	}
}
