package auth

import (
	"strings"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if CheckPassword(password, hash) != true {
		t.Fatal("correct password should be accepted")
	}

	if CheckPassword("wrongpassword", hash) != false {
		t.Fatal("incorrect password should be rejected")
	}
}

func TestPasswordHashSafety(t *testing.T) {
	password := "a sufficiently long password"
	a, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	b, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if a == password || a == b {
		t.Fatal("password must be salted and hashed")
	}
	if CheckPassword(password, "not-a-hash") {
		t.Fatal("invalid hash accepted")
	}
	if _, err := HashPassword(strings.Repeat("x", 73)); err == nil {
		t.Fatal("bcrypt maximum length not enforced")
	}
}
