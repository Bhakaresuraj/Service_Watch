package models

import (
	"testing"
)

func TestUserCreation(t *testing.T) {

	user := User{
		ID:    "123",
		Name:  "Suraj",
		Email: "suraj@example.com",
	}

	if user.Name != "Suraj" {
		t.Errorf("Expected name Suraj, got %s", user.Name)
	}

	if user.Email != "suraj@example.com" {
		t.Errorf("Expected email suraj@example.com, got %s", user.Email)
	}
}
