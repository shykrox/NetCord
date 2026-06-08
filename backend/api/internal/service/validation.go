package service

import (
	"net/mail"
	"strings"
	"unicode"

	"netcord/backend/api/internal/models"

	"github.com/google/uuid"
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

func validateUpdateMeInput(input UpdateMeInput) map[string]string {
	fields := make(map[string]string)
	if len(input.DisplayName) > 80 {
		fields["display_name"] = "display name must be 80 characters or fewer"
	}
	switch input.Status {
	case models.PresenceOnline, models.PresenceIdle, models.PresenceDND, models.PresenceOffline:
	default:
		fields["status"] = "status must be online, idle, dnd, or offline"
	}
	return fields
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

func validateUpdateServerInput(input UpdateServerInput) map[string]string {
	return validateCreateServerInput(CreateServerInput{
		Name:        input.Name,
		Description: input.Description,
	})
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
	if input.Type != models.ChannelTypeText && input.Type != models.ChannelTypeVoice {
		fields["type"] = "channel type must be text or voice"
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

func validateUpdateChannelInput(input UpdateChannelInput) map[string]string {
	return validateCreateChannelInput(CreateChannelInput{
		Name: input.Name,
		Type: input.Type,
	})
}

func validateCreateMessageInput(input CreateMessageInput) map[string]string {
	fields := make(map[string]string)
	if input.Content == "" && len(input.Attachments) == 0 {
		fields["content"] = "message content or attachments are required"
	} else if len(input.Content) > 4000 {
		fields["content"] = "message content must be 4000 characters or fewer"
	}
	if len(input.Attachments) > 10 {
		fields["attachments"] = "a message can include at most 10 attachments"
	}

	seen := make(map[string]struct{}, len(input.Attachments))
	for _, attachmentID := range input.Attachments {
		if attachmentID == uuid.Nil {
			fields["attachments"] = "attachment ids must be valid"
			break
		}
		key := attachmentID.String()
		if _, ok := seen[key]; ok {
			fields["attachments"] = "attachment ids must be unique"
			break
		}
		seen[key] = struct{}{}
	}
	return fields
}

func validateListMessagesInput(input ListMessagesInput) map[string]string {
	fields := make(map[string]string)
	if input.Before != uuid.Nil && input.After != uuid.Nil {
		fields["cursor"] = "before and after cannot be used together"
	}
	if input.Limit < 1 || input.Limit > maxMessageLimit {
		fields["limit"] = "limit must be between 1 and 100"
	}
	return fields
}

func validateUpdateMessageInput(input UpdateMessageInput) map[string]string {
	fields := make(map[string]string)
	if input.Content == "" {
		fields["content"] = "message content is required"
	} else if len(input.Content) > 4000 {
		fields["content"] = "message content must be 4000 characters or fewer"
	}
	return fields
}

func validateSearchMessagesInput(input SearchMessagesInput) map[string]string {
	fields := make(map[string]string)
	if input.Query == "" {
		fields["q"] = "search query is required"
	} else if len(input.Query) > 200 {
		fields["q"] = "search query must be 200 characters or fewer"
	}
	if input.Limit < 1 || input.Limit > maxMessageLimit {
		fields["limit"] = "limit must be between 1 and 100"
	}
	return fields
}

func validatePresenceInput(input UpdatePresenceInput) map[string]string {
	fields := make(map[string]string)
	switch input.Status {
	case models.PresenceOnline, models.PresenceIdle, models.PresenceDND, models.PresenceOffline:
	default:
		fields["status"] = "status must be online, idle, dnd, or offline"
	}
	if len(input.CustomStatus) > 120 {
		fields["custom_status"] = "custom status must be 120 characters or fewer"
	}
	return fields
}

func validateCreateFriendRequestInput(requesterID uuid.UUID, input CreateFriendRequestInput) map[string]string {
	fields := make(map[string]string)
	if input.RecipientID == uuid.Nil {
		fields["recipient_id"] = "recipient_id is required"
	} else if input.RecipientID == requesterID {
		fields["recipient_id"] = "cannot send a friend request to yourself"
	}
	return fields
}

func validateCreateDMInput(memberIDs []uuid.UUID, input CreateDMInput) map[string]string {
	fields := make(map[string]string)
	if len(memberIDs) < 2 {
		fields["member_ids"] = "at least one other member is required"
	} else if len(memberIDs) > 10 {
		fields["member_ids"] = "group DMs can include at most 10 members"
	}
	if len(input.Name) > 80 {
		fields["name"] = "name must be 80 characters or fewer"
	}
	return fields
}

func validateCreateDMMessageInput(input CreateDMMessageInput) map[string]string {
	fields := make(map[string]string)
	if input.Content == "" {
		fields["content"] = "message content is required"
	} else if len(input.Content) > 4000 {
		fields["content"] = "message content must be 4000 characters or fewer"
	}
	return fields
}

func validateCreateRoleInput(input CreateRoleInput) map[string]string {
	role := models.Role{
		Name:        input.Name,
		Color:       input.Color,
		Position:    input.Position,
		Permissions: input.Permissions,
	}
	return validateRole(role)
}

func validateRole(role models.Role) map[string]string {
	fields := make(map[string]string)
	if role.Name == "" {
		fields["name"] = "role name is required"
	} else if len(role.Name) > 64 {
		fields["name"] = "role name must be 64 characters or fewer"
	}
	if role.Color != nil && *role.Color != "" {
		color := *role.Color
		if len(color) != 7 || !strings.HasPrefix(color, "#") {
			fields["color"] = "color must be a hex value like #37d7a7"
		}
	}
	if role.Position < 0 {
		fields["position"] = "position must be zero or greater"
	}
	if role.Permissions < 0 || role.Permissions&^models.AllPermissions != 0 {
		fields["permissions"] = "permissions contains unsupported bits"
	}
	return fields
}

func validateCreateInviteInput(input CreateInviteInput) map[string]string {
	fields := make(map[string]string)
	if input.MaxUses != nil && *input.MaxUses < 1 {
		fields["max_uses"] = "max_uses must be greater than zero"
	}
	if input.ExpiresAt != nil && input.ExpiresAt.IsZero() {
		fields["expires_at"] = "expires_at must be a valid timestamp"
	}
	return fields
}

func validateCreateAIJobInput(command string, input CreateAIJobInput) map[string]string {
	fields := make(map[string]string)
	if command != models.AICommandAsk && command != models.AICommandDraw {
		fields["command"] = "command must be ask or draw"
	}
	if input.ChannelID == uuid.Nil {
		fields["channel_id"] = "channel_id is required"
	}
	if input.Prompt == "" {
		fields["prompt"] = "prompt is required"
	} else if len(input.Prompt) > 2000 {
		fields["prompt"] = "prompt must be 2000 characters or fewer"
	}
	return fields
}
