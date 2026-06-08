package service

import (
	"context"
	"strings"
	"time"

	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JoinVoiceInput struct {
	ChannelID uuid.UUID `json:"channel_id"`
}

type VoiceSessionResponse struct {
	URL       string    `json:"url"`
	Token     string    `json:"token"`
	Room      string    `json:"room"`
	ServerID  uuid.UUID `json:"server_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	UserID    uuid.UUID `json:"user_id"`
}

type VoiceService struct {
	servers     repository.ServerRepository
	permissions *PermissionService
	liveKitURL  string
	apiKey      string
	apiSecret   string
}

func NewVoiceService(servers repository.ServerRepository, permissions *PermissionService, liveKitURL, apiKey, apiSecret string) *VoiceService {
	return &VoiceService{
		servers:     servers,
		permissions: permissions,
		liveKitURL:  strings.TrimSpace(liveKitURL),
		apiKey:      strings.TrimSpace(apiKey),
		apiSecret:   strings.TrimSpace(apiSecret),
	}
}

func (s *VoiceService) Join(ctx context.Context, userID uuid.UUID, input JoinVoiceInput) (VoiceSessionResponse, error) {
	if input.ChannelID == uuid.Nil {
		return VoiceSessionResponse{}, &ValidationError{Fields: map[string]string{"channel_id": "channel_id is required"}}
	}

	channel, err := s.servers.GetChannelForUser(ctx, input.ChannelID, userID)
	if err != nil {
		return VoiceSessionResponse{}, mapRepositoryError(err)
	}
	if channel.Type != models.ChannelTypeVoice {
		return VoiceSessionResponse{}, ErrNotFound
	}
	if s.permissions != nil {
		if err := s.permissions.RequirePermission(ctx, channel.ServerID, userID, models.PermissionConnectVoice); err != nil {
			return VoiceSessionResponse{}, err
		}
	}
	if s.liveKitURL == "" || s.apiKey == "" || s.apiSecret == "" {
		return VoiceSessionResponse{}, ErrVoiceNotConfigured
	}

	room := voiceRoomName(channel.ServerID, channel.ID)
	token, err := s.liveKitToken(userID, room)
	if err != nil {
		return VoiceSessionResponse{}, err
	}
	return VoiceSessionResponse{
		URL:       s.liveKitURL,
		Token:     token,
		Room:      room,
		ServerID:  channel.ServerID,
		ChannelID: channel.ID,
		UserID:    userID,
	}, nil
}

func (s *VoiceService) Leave(ctx context.Context, userID uuid.UUID, input JoinVoiceInput) (VoiceSessionResponse, error) {
	channel, err := s.servers.GetChannelForUser(ctx, input.ChannelID, userID)
	if err != nil {
		return VoiceSessionResponse{}, mapRepositoryError(err)
	}
	if channel.Type != models.ChannelTypeVoice {
		return VoiceSessionResponse{}, ErrNotFound
	}
	return VoiceSessionResponse{
		Room:      voiceRoomName(channel.ServerID, channel.ID),
		ServerID:  channel.ServerID,
		ChannelID: channel.ID,
		UserID:    userID,
	}, nil
}

func (s *VoiceService) liveKitToken(userID uuid.UUID, room string) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss":  s.apiKey,
		"sub":  userID.String(),
		"name": userID.String(),
		"nbf":  now.Unix(),
		"exp":  now.Add(6 * time.Hour).Unix(),
		"video": map[string]interface{}{
			"room":         room,
			"roomJoin":     true,
			"canPublish":   true,
			"canSubscribe": true,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.apiSecret))
}

func voiceRoomName(serverID, channelID uuid.UUID) string {
	return "server_" + serverID.String() + "_channel_" + channelID.String()
}
