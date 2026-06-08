package service

import (
	"context"
	"errors"
	"strings"

	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"

	"github.com/google/uuid"
)

const defaultMessageLimit = 50
const maxMessageLimit = 100

type CreateServerInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateChannelInput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type CreateMessageInput struct {
	Content     string      `json:"content"`
	Attachments []uuid.UUID `json:"attachments"`
}

type ListMessagesInput struct {
	Before uuid.UUID
	After  uuid.UUID
	Limit  int
}

type UpdateMessageInput struct {
	Content string `json:"content"`
}

type SearchMessagesInput struct {
	Query string
	Limit int
}

type ServersResponse struct {
	Servers []models.PublicServer `json:"servers"`
}

type ChannelsResponse struct {
	Channels []models.PublicChannel `json:"channels"`
}

type MessagesResponse struct {
	Messages []models.PublicMessage `json:"messages"`
}

type ServerService struct {
	servers repository.ServerRepository
}

func NewServerService(servers repository.ServerRepository) *ServerService {
	return &ServerService{servers: servers}
}

func (s *ServerService) CreateServer(ctx context.Context, userID uuid.UUID, input CreateServerInput) (models.PublicServer, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	if fields := validateCreateServerInput(input); len(fields) > 0 {
		return models.PublicServer{}, &ValidationError{Fields: fields}
	}

	var description *string
	if input.Description != "" {
		description = &input.Description
	}

	server := models.Server{
		ID:          uuid.New(),
		OwnerID:     userID,
		Name:        input.Name,
		Description: description,
	}

	created, err := s.servers.CreateServer(ctx, server, userID)
	if err != nil {
		return models.PublicServer{}, mapRepositoryError(err)
	}

	return created.Public(), nil
}

func (s *ServerService) ListServers(ctx context.Context, userID uuid.UUID) (ServersResponse, error) {
	servers, err := s.servers.ListServersForUser(ctx, userID)
	if err != nil {
		return ServersResponse{}, mapRepositoryError(err)
	}

	return ServersResponse{Servers: publicServers(servers)}, nil
}

func (s *ServerService) GetServer(ctx context.Context, userID, serverID uuid.UUID) (models.PublicServer, error) {
	server, err := s.servers.GetServerForUser(ctx, serverID, userID)
	if err != nil {
		return models.PublicServer{}, mapRepositoryError(err)
	}

	return server.Public(), nil
}

func (s *ServerService) CreateChannel(ctx context.Context, userID, serverID uuid.UUID, input CreateChannelInput) (models.PublicChannel, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Type = strings.TrimSpace(input.Type)
	if input.Type == "" {
		input.Type = models.ChannelTypeText
	}
	if fields := validateCreateChannelInput(input); len(fields) > 0 {
		return models.PublicChannel{}, &ValidationError{Fields: fields}
	}

	if _, err := s.servers.GetServerForUser(ctx, serverID, userID); err != nil {
		return models.PublicChannel{}, mapRepositoryError(err)
	}

	channel := models.Channel{
		ID:       uuid.New(),
		ServerID: serverID,
		Name:     input.Name,
		Type:     input.Type,
		Position: 0,
	}

	created, err := s.servers.CreateChannel(ctx, channel)
	if err != nil {
		if errors.Is(err, repository.ErrChannelConflict) {
			return models.PublicChannel{}, &ValidationError{
				Fields: map[string]string{
					"name": "channel name already exists in this server",
				},
			}
		}
		return models.PublicChannel{}, mapRepositoryError(err)
	}

	return created.Public(), nil
}

func (s *ServerService) ListChannels(ctx context.Context, userID, serverID uuid.UUID) (ChannelsResponse, error) {
	if _, err := s.servers.GetServerForUser(ctx, serverID, userID); err != nil {
		return ChannelsResponse{}, mapRepositoryError(err)
	}

	channels, err := s.servers.ListChannelsForUser(ctx, serverID, userID)
	if err != nil {
		return ChannelsResponse{}, mapRepositoryError(err)
	}

	return ChannelsResponse{Channels: publicChannels(channels)}, nil
}

func (s *ServerService) ListMessages(ctx context.Context, userID, channelID uuid.UUID, input ListMessagesInput) (MessagesResponse, error) {
	normalizeMessageListInput(&input)
	if fields := validateListMessagesInput(input); len(fields) > 0 {
		return MessagesResponse{}, &ValidationError{Fields: fields}
	}

	channel, err := s.servers.GetChannelForUser(ctx, channelID, userID)
	if err != nil {
		return MessagesResponse{}, mapRepositoryError(err)
	}
	if channel.Type != models.ChannelTypeText {
		return MessagesResponse{}, ErrNotFound
	}

	messages, err := s.servers.ListMessagesForChannelUser(ctx, channelID, userID, repository.MessageListOptions{
		Before: input.Before,
		After:  input.After,
		Limit:  input.Limit,
	})
	if err != nil {
		return MessagesResponse{}, mapRepositoryError(err)
	}

	return MessagesResponse{Messages: publicMessages(messages)}, nil
}

func (s *ServerService) GetChannel(ctx context.Context, userID, channelID uuid.UUID) (models.PublicChannel, error) {
	channel, err := s.servers.GetChannelForUser(ctx, channelID, userID)
	if err != nil {
		return models.PublicChannel{}, mapRepositoryError(err)
	}
	return channel.Public(), nil
}

func (s *ServerService) CreateMessage(ctx context.Context, userID, channelID uuid.UUID, input CreateMessageInput) (models.PublicMessage, error) {
	input.Content = strings.TrimSpace(input.Content)
	if fields := validateCreateMessageInput(input); len(fields) > 0 {
		return models.PublicMessage{}, &ValidationError{Fields: fields}
	}

	channel, err := s.servers.GetChannelForUser(ctx, channelID, userID)
	if err != nil {
		return models.PublicMessage{}, mapRepositoryError(err)
	}
	if channel.Type != models.ChannelTypeText {
		return models.PublicMessage{}, ErrNotFound
	}

	message := models.Message{
		ID:        uuid.New(),
		ServerID:  channel.ServerID,
		ChannelID: channel.ID,
		AuthorID:  userID,
		Content:   input.Content,
	}

	created, err := s.servers.CreateMessage(ctx, message, input.Attachments)
	if err != nil {
		return models.PublicMessage{}, mapRepositoryError(err)
	}

	return created.Public(), nil
}

func (s *ServerService) UpdateMessage(ctx context.Context, userID, messageID uuid.UUID, input UpdateMessageInput) (models.PublicMessage, error) {
	input.Content = strings.TrimSpace(input.Content)
	if fields := validateUpdateMessageInput(input); len(fields) > 0 {
		return models.PublicMessage{}, &ValidationError{Fields: fields}
	}

	message, err := s.servers.UpdateMessageForUser(ctx, messageID, userID, input.Content)
	if err != nil {
		return models.PublicMessage{}, mapRepositoryError(err)
	}

	return message.Public(), nil
}

func (s *ServerService) DeleteMessage(ctx context.Context, userID, messageID uuid.UUID) (models.PublicMessage, error) {
	message, err := s.servers.DeleteMessageForUser(ctx, messageID, userID)
	if err != nil {
		return models.PublicMessage{}, mapRepositoryError(err)
	}

	return message.Public(), nil
}

func (s *ServerService) SearchMessages(ctx context.Context, userID, channelID uuid.UUID, input SearchMessagesInput) (MessagesResponse, error) {
	input.Query = strings.TrimSpace(input.Query)
	if input.Limit <= 0 {
		input.Limit = defaultMessageLimit
	}
	if input.Limit > maxMessageLimit {
		input.Limit = maxMessageLimit
	}
	if fields := validateSearchMessagesInput(input); len(fields) > 0 {
		return MessagesResponse{}, &ValidationError{Fields: fields}
	}

	channel, err := s.servers.GetChannelForUser(ctx, channelID, userID)
	if err != nil {
		return MessagesResponse{}, mapRepositoryError(err)
	}
	if channel.Type != models.ChannelTypeText {
		return MessagesResponse{}, ErrNotFound
	}

	messages, err := s.servers.SearchMessagesForChannelUser(ctx, channelID, userID, input.Query, input.Limit)
	if err != nil {
		return MessagesResponse{}, mapRepositoryError(err)
	}

	return MessagesResponse{Messages: publicMessages(messages)}, nil
}

func mapRepositoryError(err error) error {
	switch {
	case errors.Is(err, repository.ErrServerNotFound),
		errors.Is(err, repository.ErrChannelNotFound),
		errors.Is(err, repository.ErrAttachmentNotFound),
		errors.Is(err, repository.ErrMessageNotFound):
		return ErrNotFound
	case errors.Is(err, repository.ErrForbidden):
		return ErrForbidden
	default:
		return err
	}
}

func normalizeMessageListInput(input *ListMessagesInput) {
	if input.Limit <= 0 {
		input.Limit = defaultMessageLimit
	}
	if input.Limit > maxMessageLimit {
		input.Limit = maxMessageLimit
	}
}

func publicServers(servers []models.Server) []models.PublicServer {
	public := make([]models.PublicServer, 0, len(servers))
	for _, server := range servers {
		public = append(public, server.Public())
	}
	return public
}

func publicChannels(channels []models.Channel) []models.PublicChannel {
	public := make([]models.PublicChannel, 0, len(channels))
	for _, channel := range channels {
		public = append(public, channel.Public())
	}
	return public
}

func publicMessages(messages []models.Message) []models.PublicMessage {
	public := make([]models.PublicMessage, 0, len(messages))
	for _, message := range messages {
		public = append(public, message.Public())
	}
	return public
}
