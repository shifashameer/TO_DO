package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"todo-app/internal/auth"

	"todo-app/internal/models"
)

func TestRequireAuthRedirectsWhenNoSession(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	userStore := &fakeUserStore{
		user: models.User{
			ID: "user-1",
		},
	}

	sessions := auth.NewSessions()
	middleware := RequireAuth(userStore, sessions, nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusSeeOther,
			rr.Code,
		)
	}

	if location := rr.Header().Get("Location"); location != "/login" {
		t.Errorf(
			"expected redirect to /login, got %q",
			location,
		)
	}
}

func TestRequireAuthAllowsValidSession(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	userStore := &fakeUserStore{
		user: models.User{
			ID: "user-1",
		},
	}

	sessions := auth.NewSessions()
	middleware := RequireAuth(userStore, sessions, nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	req.AddCookie(&http.Cookie{
		Name: "session",
		Value: func() string {
			token, err := sessions.Create("user-1")
			if err != nil {
				t.Fatal(err)
			}
			return token
		}(),
	})

	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusOK,
			rr.Code,
		)
	}
}

func TestRequireAuthRejectsInvalidSession(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	userStore := &fakeUserStore{
		user: models.User{
			ID: "user-1",
		},
	}

	sessions := auth.NewSessions()
	middleware := RequireAuth(userStore, sessions, nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "invalid-user",
	})

	rr := httptest.NewRecorder()

	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf(
			"expected status code %d, got %d",
			http.StatusSeeOther,
			rr.Code,
		)
	}

	if location := rr.Header().Get("Location"); location != "/login" {
		t.Errorf(
			"expected redirect to /login, got %q",
			location,
		)
	}
}
