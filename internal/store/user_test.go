package store

import (
	"testing"
	"todo-app/internal/models"
)

func TestGetByUsername(t *testing.T) {
	userStore := newTestStore(t)

	user := models.User{
		ID:           "test-user-1",
		Username:     "testuser",
		PasswordHash: "hashedpassword",
	}

	t.Cleanup(func() {
		userStore.db.Delete(&models.User{}, "id = ?", user.ID)
	})

	err := userStore.db.Create(&user).Error
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	foundUser, err := userStore.GetByUsername("testuser")
	if err != nil {
		t.Fatalf("failed to find user: %v", err)
	}

	if foundUser.Username != "testuser" {
		t.Errorf("expected username %q, got %q", "testuser", foundUser.Username)
	}
}

func TestCreateUser(t *testing.T) {
	userStore := newTestStore(t)

	user := models.User{
		ID:           "test-user-create",
		Username:     "newuser",
		PasswordHash: "hashedpassword",
	}

	t.Cleanup(func() {
		userStore.db.Delete(&models.User{}, "id = ?", user.ID)
	})

	err := userStore.CreateUser(user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	foundUser, err := userStore.GetByUsername("newuser")
	if err != nil {
		t.Fatalf("failed to find created user: %v", err)
	}

	if foundUser.Username != "newuser" {
		t.Errorf("expected username %q, got %q", "newuser", foundUser.Username)
	}
}
