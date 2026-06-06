package service

import "testing"

func TestValidateRegisterInputAcceptsValidInput(t *testing.T) {
	fields := validateRegisterInput(RegisterInput{
		Username: "shykrox",
		Email:    "user@example.com",
		Password: "strong-password",
	})
	if len(fields) != 0 {
		t.Fatalf("expected no validation errors, got %v", fields)
	}
}

func TestValidateRegisterInputRejectsInvalidInput(t *testing.T) {
	fields := validateRegisterInput(RegisterInput{
		Username: "no spaces",
		Email:    "not-an-email",
		Password: "short",
	})

	for _, field := range []string{"username", "email", "password"} {
		if fields[field] == "" {
			t.Fatalf("expected validation error for %s, got %v", field, fields)
		}
	}
}
