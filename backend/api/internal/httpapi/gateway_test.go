package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"netcord/backend/api/internal/auth"
	"netcord/backend/api/internal/gateway"
	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/repository"
	"netcord/backend/api/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func TestGatewayWebSocketHelloAndHeartbeat(t *testing.T) {
	userID := uuid.New()
	serverID := uuid.New()
	repo := newHTTPFakeServerRepository()
	repo.servers[serverID] = models.Server{ID: serverID, OwnerID: userID, Name: "NetCord"}
	repo.members[serverID] = map[uuid.UUID]string{userID: models.ServerRoleOwner}

	tokenManager := newTestTokenManager(t)
	token := newTestToken(t, tokenManager, userID)
	router := NewRouter(nil, service.NewServerService(repo), nil, tokenManager, gateway.NewHub())
	server := httptest.NewServer(router)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(wsURL(server.URL, token), nil)
	if err != nil {
		t.Fatalf("dial gateway: %v", err)
	}
	defer conn.Close()

	var hello gateway.Event
	if err := conn.ReadJSON(&hello); err != nil {
		t.Fatalf("read hello: %v", err)
	}
	if hello.Type != gateway.EventHello {
		t.Fatalf("expected hello event, got %s", hello.Type)
	}

	if err := conn.WriteJSON(gateway.Event{Type: gateway.EventHeartbeat}); err != nil {
		t.Fatalf("write heartbeat: %v", err)
	}

	var ack gateway.Event
	if err := conn.ReadJSON(&ack); err != nil {
		t.Fatalf("read heartbeat ack: %v", err)
	}
	if ack.Type != gateway.EventHeartbeatAck {
		t.Fatalf("expected heartbeat ack, got %s", ack.Type)
	}
}

func TestCreateMessageBroadcastsGatewayEvent(t *testing.T) {
	userID := uuid.New()
	serverID := uuid.New()
	channelID := uuid.New()
	repo := newHTTPFakeServerRepository()
	repo.servers[serverID] = models.Server{ID: serverID, OwnerID: userID, Name: "NetCord"}
	repo.members[serverID] = map[uuid.UUID]string{userID: models.ServerRoleOwner}
	repo.channels[channelID] = models.Channel{
		ID:       channelID,
		ServerID: serverID,
		Name:     "general",
		Type:     models.ChannelTypeText,
	}

	tokenManager := newTestTokenManager(t)
	token := newTestToken(t, tokenManager, userID)
	hub := gateway.NewHub()
	client := hub.Register(userID, []uuid.UUID{serverID})
	router := NewRouter(nil, service.NewServerService(repo), nil, tokenManager, hub)

	request := httptest.NewRequest(
		http.MethodPost,
		"/channels/"+channelID.String()+"/messages",
		bytes.NewBufferString(`{"content":"hello realtime"}`),
	)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	event := <-client.Send()
	if event.Type != gateway.EventMessageCreated {
		t.Fatalf("expected message.created, got %s", event.Type)
	}

	message, ok := event.Data.(models.PublicMessage)
	if !ok {
		t.Fatalf("expected public message data, got %T", event.Data)
	}
	if message.Content != "hello realtime" || message.ChannelID != channelID || message.ServerID != serverID {
		t.Fatalf("unexpected message payload: %+v", message)
	}
}

func newTestTokenManager(t *testing.T) *auth.TokenManager {
	t.Helper()
	tokenManager, err := auth.NewTokenManager("test-secret", "netcord-test", time.Hour)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}
	return tokenManager
}

func newTestToken(t *testing.T, tokenManager *auth.TokenManager, userID uuid.UUID) string {
	t.Helper()
	token, err := tokenManager.Generate(userID)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func wsURL(httpURL, token string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http") + "/gateway/ws?token=" + url.QueryEscape(token)
}

type httpFakeServerRepository struct {
	servers  map[uuid.UUID]models.Server
	members  map[uuid.UUID]map[uuid.UUID]string
	channels map[uuid.UUID]models.Channel
	messages []models.Message
}

func newHTTPFakeServerRepository() *httpFakeServerRepository {
	return &httpFakeServerRepository{
		servers:  make(map[uuid.UUID]models.Server),
		members:  make(map[uuid.UUID]map[uuid.UUID]string),
		channels: make(map[uuid.UUID]models.Channel),
		messages: make([]models.Message, 0),
	}
}

func (r *httpFakeServerRepository) CreateServer(ctx context.Context, server models.Server, ownerID uuid.UUID) (models.Server, error) {
	server.OwnerID = ownerID
	server.CreatedAt = time.Now().UTC()
	server.UpdatedAt = server.CreatedAt
	r.servers[server.ID] = server
	r.members[server.ID] = map[uuid.UUID]string{ownerID: models.ServerRoleOwner}
	return server, nil
}

func (r *httpFakeServerRepository) ListServersForUser(ctx context.Context, userID uuid.UUID) ([]models.Server, error) {
	servers := make([]models.Server, 0)
	for serverID, members := range r.members {
		if _, ok := members[userID]; ok {
			servers = append(servers, r.servers[serverID])
		}
	}
	return servers, nil
}

func (r *httpFakeServerRepository) GetServerForUser(ctx context.Context, serverID, userID uuid.UUID) (models.Server, error) {
	if _, ok := r.members[serverID][userID]; !ok {
		return models.Server{}, repository.ErrServerNotFound
	}
	server, ok := r.servers[serverID]
	if !ok {
		return models.Server{}, repository.ErrServerNotFound
	}
	return server, nil
}

func (r *httpFakeServerRepository) CreateChannel(ctx context.Context, channel models.Channel) (models.Channel, error) {
	channel.CreatedAt = time.Now().UTC()
	channel.UpdatedAt = channel.CreatedAt
	r.channels[channel.ID] = channel
	return channel, nil
}

func (r *httpFakeServerRepository) ListChannelsForUser(ctx context.Context, serverID, userID uuid.UUID) ([]models.Channel, error) {
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

func (r *httpFakeServerRepository) GetChannelForUser(ctx context.Context, channelID, userID uuid.UUID) (models.Channel, error) {
	channel, ok := r.channels[channelID]
	if !ok {
		return models.Channel{}, repository.ErrChannelNotFound
	}
	if _, ok := r.members[channel.ServerID][userID]; !ok {
		return models.Channel{}, repository.ErrChannelNotFound
	}
	return channel, nil
}

func (r *httpFakeServerRepository) CreateMessage(ctx context.Context, message models.Message, attachmentIDs []uuid.UUID) (models.Message, error) {
	message.CreatedAt = time.Now().UTC()
	message.UpdatedAt = message.CreatedAt
	r.messages = append(r.messages, message)
	return message, nil
}

func (r *httpFakeServerRepository) ListMessagesForChannelUser(ctx context.Context, channelID, userID uuid.UUID, limit int) ([]models.Message, error) {
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
