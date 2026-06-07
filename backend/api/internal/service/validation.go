package service

import (
	"net/mail"
	"strings"
	"unicode"
)

func validateRegisterInput(input RegisterInput) map[string]string {
	fields := make(map[string]string)

	if err := validateUsername(input.Username); err != "" {
		fields["username"] = err
	}
	if err := validateEmail(input.Email); err != "" {
		fields["email"] = err
	}
	if err := validatePassword(input.Password); err != "" {
		fields["password"] = err
	}

	return fields
}

func validateUsername(username string) string {
	if username == "" {
		return "username is required"
	}
	if len(username) < 3 || len(username) > 32 {
		return "username must be between 3 and 32 characters"
	}
	for _, r := range username {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			continue
		}
		return "username can contain only letters, numbers, underscores, and hyphens"
	}
	return ""
}

func validateEmail(email string) string {
	if email == "" {
		return "email is required"
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || strings.Contains(email, " ") {
		return "email must be valid"
	}
	return ""
}

func validatePassword(password string) string {
	if password == "" {
		return "password is required"
	}
	if len(password) < 12 {
		return "password must be at least 12 characters"
	}
	return ""
}

func validateCreateServerInput(input CreateServerInput) map[string]string {
	fields := make(map[string]string)
	if input.Name == "" {
		fields["name"] = "server name is required"
	} else if len(input.Name) > 80 {
		fields["name"] = "server name must be 80 characters or fewer"
	}

	if len(input.Description) > 500 {
		fields["description"] = "description must be 500 characters or fewer"
	}

	return fields
}

func validateCreateChannelInput(input CreateChannelInput) map[string]string {
	fields := make(map[string]string)
	if input.Name == "" {
		fields["name"] = "channel name is required"
		return fields
	}
	if len(input.Name) > 64 {
		fields["name"] = "channel name must be 64 characters or fewer"
		return fields
	}
	for _, r := range input.Name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			continue
		}
		fields["name"] = "channel name can contain only letters, numbers, underscores, and hyphens"
		return fields
	}
	return fields
}

func validateCreateMessageInput(input CreateMessageInput) map[string]string {
	fields := make(map[string]string)
	if input.Content == "" {
		fields["content"] = "message content is required"
	} else if len(input.Content) > 4000 {
		fields["content"] = "message content must be 4000 characters or fewer"
	}
	return fields
}
