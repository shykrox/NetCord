package service

import (
	"context"
	"errors"
	"strings"
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

func TestCreateMessageAttachesUploadedFiles(t *testing.T) {
	repo := newFakeServerRepository()
	service := NewServerService(repo)
	userID := uuid.New()
	serverID := uuid.New()
	channelID := uuid.New()
	attachmentID := uuid.New()
	repo.servers[serverID] = models.Server{ID: serverID, OwnerID: userID, Name: "NetCord"}
	repo.members[serverID] = map[uuid.UUID]string{userID: models.ServerRoleOwner}
	repo.channels[channelID] = models.Channel{
		ID:       channelID,
		ServerID: serverID,
		Name:     "general",
		Type:     models.ChannelTypeText,
	}
	repo.attachments[attachmentID] = models.MessageAttachment{
		ID:               attachmentID,
		UploaderID:       userID,
		Bucket:           "attachments",
		ObjectKey:        "attachments/random/file",
		OriginalFilename: "hello.txt",
		ContentType:      "text/plain; charset=utf-8",
		SizeBytes:        5,
	}

	message, err := service.CreateMessage(context.Background(), userID, channelID, CreateMessageInput{
		Attachments: []uuid.UUID{attachmentID},
	})
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	if len(message.Attachments) != 1 {
		t.Fatalf("expected one attachment, got %d", len(message.Attachments))
	}
	if message.Attachments[0].ID != attachmentID {
		t.Fatalf("expected attachment %s, got %s", attachmentID, message.Attachments[0].ID)
	}
}

func TestUpdateMessageRequiresAuthor(t *testing.T) {
	repo := newFakeServerRepository()
	service := NewServerService(repo)
	authorID := uuid.New()
	otherID := uuid.New()
	serverID := uuid.New()
	channelID := uuid.New()
	messageID := uuid.New()
	repo.servers[serverID] = models.Server{ID: serverID, OwnerID: authorID, Name: "NetCord"}
	repo.members[serverID] = map[uuid.UUID]string{
		authorID: models.ServerRoleOwner,
		otherID:  models.ServerRoleMember,
	}
	repo.channels[channelID] = models.Channel{ID: channelID, ServerID: serverID, Name: "general", Type: models.ChannelTypeText}
	repo.messages = append(repo.messages, models.Message{
		ID:        messageID,
		ServerID:  serverID,
		ChannelID: channelID,
		AuthorID:  authorID,
		Content:   "original",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	_, err := service.UpdateMessage(context.Background(), otherID, messageID, UpdateMessageInput{Content: "nope"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateMessageEditsOwnContent(t *testing.T) {
	repo := newFakeServerRepository()
	service := NewServerService(repo)
	userID := uuid.New()
	serverID := uuid.New()
	channelID := uuid.New()
	messageID := uuid.New()
	repo.servers[serverID] = models.Server{ID: serverID, OwnerID: userID, Name: "NetCord"}
	repo.members[serverID] = map[uuid.UUID]string{userID: models.ServerRoleOwner}
	repo.channels[channelID] = models.Channel{ID: channelID, ServerID: serverID, Name: "general", Type: models.ChannelTypeText}
	repo.messages = append(repo.messages, models.Message{
		ID:        messageID,
		ServerID:  serverID,
		ChannelID: channelID,
		AuthorID:  userID,
		Content:   "original",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	message, err := service.UpdateMessage(context.Background(), userID, messageID, UpdateMessageInput{Content: " edited "})
	if err != nil {
		t.Fatalf("update message: %v", err)
	}
	if message.Content != "edited" {
		t.Fatalf("expected edited content, got %q", message.Content)
	}
	if message.EditedAt == nil {
		t.Fatalf("expected edited_at to be set")
	}
}

type fakeServerRepository struct {
	servers     map[uuid.UUID]models.Server
	members     map[uuid.UUID]map[uuid.UUID]string
	channels    map[uuid.UUID]models.Channel
	attachments map[uuid.UUID]models.MessageAttachment
	messages    []models.Message
}

func newFakeServerRepository() *fakeServerRepository {
	return &fakeServerRepository{
		servers:     make(map[uuid.UUID]models.Server),
		members:     make(map[uuid.UUID]map[uuid.UUID]string),
		channels:    make(map[uuid.UUID]models.Channel),
		attachments: make(map[uuid.UUID]models.MessageAttachment),
		messages:    make([]models.Message, 0),
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

func (r *fakeServerRepository) CreateMessage(ctx context.Context, message models.Message, attachmentIDs []uuid.UUID) (models.Message, error) {
	now := time.Now().UTC()
	message.CreatedAt = now
	message.UpdatedAt = now
	for _, attachmentID := range attachmentIDs {
		attachment, ok := r.attachments[attachmentID]
		if !ok || attachment.UploaderID != message.AuthorID || attachment.MessageID != nil {
			return models.Message{}, repository.ErrAttachmentNotFound
		}
		messageID := message.ID
		serverID := message.ServerID
		channelID := message.ChannelID
		attachment.MessageID = &messageID
		attachment.ServerID = &serverID
		attachment.ChannelID = &channelID
		r.attachments[attachmentID] = attachment
		message.Attachments = append(message.Attachments, attachment)
	}
	r.messages = append(r.messages, message)
	return message, nil
}

func (r *fakeServerRepository) ListMessagesForChannelUser(ctx context.Context, channelID, userID uuid.UUID, options repository.MessageListOptions) ([]models.Message, error) {
	channel, err := r.GetChannelForUser(ctx, channelID, userID)
	if err != nil {
		return nil, err
	}

	messages := make([]models.Message, 0)
	for _, message := range r.messages {
		if message.ChannelID == channel.ID && message.DeletedAt == nil {
			messages = append(messages, message)
		}
	}
	return messages, nil
}

func (r *fakeServerRepository) SearchMessagesForChannelUser(ctx context.Context, channelID, userID uuid.UUID, query string, limit int) ([]models.Message, error) {
	channel, err := r.GetChannelForUser(ctx, channelID, userID)
	if err != nil {
		return nil, err
	}

	messages := make([]models.Message, 0)
	for _, message := range r.messages {
		if message.ChannelID == channel.ID && message.DeletedAt == nil && strings.Contains(message.Content, query) {
			messages = append(messages, message)
		}
	}
	return messages, nil
}

func (r *fakeServerRepository) UpdateMessageForUser(ctx context.Context, messageID, userID uuid.UUID, content string) (models.Message, error) {
	for i, message := range r.messages {
		if message.ID != messageID || message.DeletedAt != nil {
			continue
		}
		if _, ok := r.members[message.ServerID][userID]; !ok {
			return models.Message{}, repository.ErrMessageNotFound
		}
		if message.AuthorID != userID {
			return models.Message{}, repository.ErrForbidden
		}
		now := time.Now().UTC()
		r.messages[i].Content = content
		r.messages[i].UpdatedAt = now
		r.messages[i].EditedAt = &now
		return r.messages[i], nil
	}
	return models.Message{}, repository.ErrMessageNotFound
}

func (r *fakeServerRepository) DeleteMessageForUser(ctx context.Context, messageID, userID uuid.UUID) (models.Message, error) {
	for i, message := range r.messages {
		if message.ID != messageID || message.DeletedAt != nil {
			continue
		}
		if _, ok := r.members[message.ServerID][userID]; !ok {
			return models.Message{}, repository.ErrMessageNotFound
		}
		if message.AuthorID != userID {
			return models.Message{}, repository.ErrForbidden
		}
		now := time.Now().UTC()
		r.messages[i].UpdatedAt = now
		r.messages[i].DeletedAt = &now
		return r.messages[i], nil
	}
	return models.Message{}, repository.ErrMessageNotFound
}
