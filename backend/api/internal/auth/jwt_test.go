package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTokenManagerGenerateAndValidate(t *testing.T) {
	manager, err := NewTokenManager("test-secret", "netcord-test", time.Hour)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	userID := uuid.New()
	token, err := manager.Generate(userID)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	got, err := manager.Validate(token)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if got != userID {
		t.Fatalf("expected %s, got %s", userID, got)
	}
}

func TestTokenManagerRejectsWrongSecret(t *testing.T) {
	manager, err := NewTokenManager("test-secret", "netcord-test", time.Hour)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}
	otherManager, err := NewTokenManager("other-secret", "netcord-test", time.Hour)
	if err != nil {
		t.Fatalf("new other token manager: %v", err)
	}

	token, err := manager.Generate(uuid.New())
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	_, err = otherManager.Validate(token)
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
