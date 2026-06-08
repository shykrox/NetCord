package service

import (
	"context"
	"strings"

	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"

	"github.com/google/uuid"
)

type UpdatePresenceInput struct {
	Status       string `json:"status"`
	CustomStatus string `json:"custom_status"`
}

type PresenceService struct {
	presence repository.PresenceRepository
}

func NewPresenceService(presence repository.PresenceRepository) *PresenceService {
	return &PresenceService{presence: presence}
}

func (s *PresenceService) SetOnline(ctx context.Context, userID uuid.UUID) (models.PublicPresence, error) {
	return s.UpdatePresence(ctx, userID, UpdatePresenceInput{Status: models.PresenceOnline})
}

func (s *PresenceService) SetOffline(ctx context.Context, userID uuid.UUID) (models.PublicPresence, error) {
	return s.UpdatePresence(ctx, userID, UpdatePresenceInput{Status: models.PresenceOffline})
}

func (s *PresenceService) UpdatePresence(ctx context.Context, userID uuid.UUID, input UpdatePresenceInput) (models.PublicPresence, error) {
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	input.CustomStatus = strings.TrimSpace(input.CustomStatus)
	if input.Status == "" {
		input.Status = models.PresenceOnline
	}

	if fields := validatePresenceInput(input); len(fields) > 0 {
		return models.PublicPresence{}, &ValidationError{Fields: fields}
	}

	var customStatus *string
	if input.CustomStatus != "" {
		customStatus = &input.CustomStatus
	}

	presence, err := s.presence.UpsertPresence(ctx, models.UserPresence{
		UserID:       userID,
		Status:       input.Status,
		CustomStatus: customStatus,
	})
	if err != nil {
		return models.PublicPresence{}, mapRepositoryError(err)
	}

	return presence.Public(), nil
}
