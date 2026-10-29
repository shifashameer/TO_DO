package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"todo-app/internal/auth"
	"todo-app/internal/models"
)

type fakeUserStore struct {
	user models.User
}

func (f *fakeUserStore) GetByUsername(username string) (models.User, error) {
	if username == f.user.Username {
		return f.user, nil
	}

	return models.User{}, nil
}
func (f *fakeUserStore) GetByID(id string) (models.User, error) {
	if id == f.user.ID {
		return f.user, nil
	}

	return models.User{}, errors.New("user not found")
}
func TestLoginCorrectCredentials(t *testing.T) {
	password := "testpassword123"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	userStore := &fakeUserStore{
		user: models.User{
			ID:           "user-1",
			Username:     "testuser",
			PasswordHash: hash,
		},
	}

	authHandler := NewAuthHandler(userStore)

	form := strings.NewReader("username=testuser&password=testpassword123")

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		form,
	)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	authHandler.Login(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status code %d, got %d", http.StatusSeeOther, rr.Code)
	}

	cookies := rr.Result().Cookies()

	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}
}
func TestLoginIncorrectCredentials(t *testing.T) {
	password := "testpassword123"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	userStore := &fakeUserStore{
		user: models.User{
			ID:           "user-1",
			Username:     "testuser",
			PasswordHash: hash,
		},
	}

	authHandler := NewAuthHandler(userStore)

	form := strings.NewReader("username=testuser&password=wrongpassword")

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		form,
	)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	authHandler.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	cookies := rr.Result().Cookies()

	if len(cookies) != 0 {
		t.Fatal("incorrect credentials should not create a session cookie")
	}
}
func TestLogoutClearsSession(t *testing.T) {
	authHandler := NewAuthHandler(&fakeUserStore{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/logout",
		nil,
	)

	rr := httptest.NewRecorder()

	authHandler.Logout(rr, req)

	cookies := rr.Result().Cookies()

	if len(cookies) == 0 {
		t.Fatal("expected session cookie to be cleared")
	}

	cookie := cookies[0]

	if cookie.Name != "session" {
		t.Errorf("expected cookie name %q, got %q", "session", cookie.Name)
	}

	if cookie.MaxAge != -1 {
		t.Errorf("expected MaxAge -1, got %d", cookie.MaxAge)
	}

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

func TestLoginSecureCookieAndMalformedForm(t *testing.T) {
	hash, err := auth.HashPassword("testpassword123")
	if err != nil {
		t.Fatal(err)
	}
	h := NewAuthHandler(&fakeUserStore{user: models.User{ID: "user-1", Username: "testuser", PasswordHash: hash}})
	h.SecureCookies = true
	for _, tc := range []struct {
		body   string
		status int
	}{{"username=testuser&password=testpassword123", 303}, {"username=testuser&password=%zz", 400}, {"username=unknown&password=testpassword123", 401}} {
		req := httptest.NewRequest("POST", "/login", strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		h.Login(rr, req)
		if rr.Code != tc.status {
			t.Fatalf("got %d want %d", rr.Code, tc.status)
		}
		cookies := rr.Result().Cookies()
		if tc.status == 303 {
			if len(cookies) != 1 || !cookies[0].Secure {
				t.Fatal("missing Secure cookie")
			}
		} else if len(cookies) != 0 {
			t.Fatal("failed login set cookie")
		}
	}
}
