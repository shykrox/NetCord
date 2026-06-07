package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"

	"github.com/google/uuid"
)

func TestCreateServerMakesCreatorOwner(t *testing.T) {
	repo := newFakeServerRepository()
	service := NewServerService(repo)
	userID := uuid.New()

	server, err := service.CreateServer(context.Background(), userID, CreateServerInput{
		Name:        "NetCord",
		Description: "private workspace",
	})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	if server.OwnerID != userID {
		t.Fatalf("expected owner %s, got %s", userID, server.OwnerID)
	}
	if repo.members[server.ID][userID] != models.ServerRoleOwner {
		t.Fatalf("creator was not inserted as owner member")
	}
}

func TestCreateChannelRequiresServerMembership(t *testing.T) {
	repo := newFakeServerRepository()
	service := NewServerService(repo)
	serverID := uuid.New()
	repo.servers[serverID] = models.Server{ID: serverID, OwnerID: uuid.New(), Name: "Private"}

	_, err := service.CreateChannel(context.Background(), uuid.New(), serverID, CreateChannelInput{Name: "general"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for non-member, got %v", err)
	}
}

func TestCreateMessageStoresAuthorAndChannelServer(t *testing.T) {
	repo := newFakeServerRepository()
	service := NewServerService(repo)
	userID := uuid.New()
	serverID := uuid.New()
	channelID := uuid.New()
	repo.servers[serverID] = models.Server{ID: serverID, OwnerID: userID, Name: "NetCord"}
	repo.members[serverID] = map[uuid.UUID]string{userID: models.ServerRoleOwner}
	repo.channels[channelID] = models.Channel{
		ID:       channelID,
		ServerID: serverID,
		Name:     "general",
		Type:     models.ChannelTypeText,
	}

	message, err := service.CreateMessage(context.Background(), userID, channelID, CreateMessageInput{
		Content: " hello ",
	})
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	if message.AuthorID != userID || message.ChannelID != channelID || message.ServerID != serverID {
		t.Fatalf("message identifiers were not populated from authenticated user/channel: %+v", message)
	}
	if message.Content != "hello" {
		t.Fatalf("expected trimmed content, got %q", message.Content)
	}
}

type fakeServerRepository struct {
	servers  map[uuid.UUID]models.Server
	members  map[uuid.UUID]map[uuid.UUID]string
	channels map[uuid.UUID]models.Channel
	messages []models.Message
}

func newFakeServerRepository() *fakeServerRepository {
	return &fakeServerRepository{
		servers:  make(map[uuid.UUID]models.Server),
		members:  make(map[uuid.UUID]map[uuid.UUID]string),
		channels: make(map[uuid.UUID]models.Channel),
		messages: make([]models.Message, 0),
	}
}

func (r *fakeServerRepository) CreateServer(ctx context.Context, server models.Server, ownerID uuid.UUID) (models.Server, error) {
	now := time.Now().UTC()
	server.OwnerID = ownerID
	server.CreatedAt = now
	server.UpdatedAt = now
	r.servers[server.ID] = server
	r.members[server.ID] = map[uuid.UUID]string{ownerID: models.ServerRoleOwner}
	return server, nil
}

func (r *fakeServerRepository) ListServersForUser(ctx context.Context, userID uuid.UUID) ([]models.Server, error) {
	servers := make([]models.Server, 0)
	for serverID := range r.members {
		if _, ok := r.members[serverID][userID]; ok {
			servers = append(servers, r.servers[serverID])
		}
	}
	return servers, nil
}

func (r *fakeServerRepository) GetServerForUser(ctx context.Context, serverID, userID uuid.UUID) (models.Server, error) {
	if _, ok := r.members[serverID][userID]; !ok {
		return models.Server{}, repository.ErrServerNotFound
	}
	server, ok := r.servers[serverID]
	if !ok {
		return models.Server{}, repository.ErrServerNotFound
	}
	return server, nil
}

func (r *fakeServerRepository) CreateChannel(ctx context.Context, channel models.Channel) (models.Channel, error) {
	for _, existing := range r.channels {
		if existing.ServerID == channel.ServerID && existing.Name == channel.Name {
			return models.Channel{}, repository.ErrChannelConflict
		}
	}
	now := time.Now().UTC()
	channel.CreatedAt = now
	channel.UpdatedAt = now
	r.channels[channel.ID] = channel
	return channel, nil
}

func (r *fakeServerRepository) ListChannelsForUser(ctx context.Context, serverID, userID uuid.UUID) ([]models.Channel, error) {
	if _, ok := r.members[serverID][userID]; !ok {
		return nil, repository.ErrServerNotFound
	}
	channels := make([]models.Channel, 0)
	for _, channel := range r.channels {
		if channel.ServerID == serverID {
			channels = append(channels, channel)
		}
	}
	return channels, nil
}

func (r *fakeServerRepository) GetChannelForUser(ctx context.Context, channelID, userID uuid.UUID) (models.Channel, error) {
	channel, ok := r.channels[channelID]
	if !ok {
		return models.Channel{}, repository.ErrChannelNotFound
	}
	if _, ok := r.members[channel.ServerID][userID]; !ok {
		return models.Channel{}, repository.ErrChannelNotFound
	}
	return channel, nil
}

func (r *fakeServerRepository) CreateMessage(ctx context.Context, message models.Message) (models.Message, error) {
	now := time.Now().UTC()
	message.CreatedAt = now
	message.UpdatedAt = now
	r.messages = append(r.messages, message)
	return message, nil
}

func (r *fakeServerRepository) ListMessagesForChannelUser(ctx context.Context, channelID, userID uuid.UUID, limit int) ([]models.Message, error) {
	channel, err := r.GetChannelForUser(ctx, channelID, userID)
	if err != nil {
		return nil, err
	}

	messages := make([]models.Message, 0)
	for _, message := range r.messages {
		if message.ChannelID == channel.ID {
			messages = append(messages, message)
		}
	}
	return messages, nil
}
