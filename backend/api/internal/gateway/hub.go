package gateway

import (
	"sync"

	"netcord/backend/api/internal/models"

	"github.com/google/uuid"
)

type Hub struct {
	mu       sync.RWMutex
	clients  map[*Client]struct{}
	byServer map[uuid.UUID]map[*Client]struct{}
	byUser   map[uuid.UUID]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:  make(map[*Client]struct{}),
		byServer: make(map[uuid.UUID]map[*Client]struct{}),
		byUser:   make(map[uuid.UUID]map[*Client]struct{}),
	}
}

func (h *Hub) Register(userID uuid.UUID, serverIDs []uuid.UUID) *Client {
	return h.RegisterUser(userID, "", serverIDs)
}

func (h *Hub) RegisterUser(userID uuid.UUID, username string, serverIDs []uuid.UUID) *Client {
	client := newClient(userID, username, serverIDs)

	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = struct{}{}
	if h.byUser[userID] == nil {
		h.byUser[userID] = make(map[*Client]struct{})
	}
	h.byUser[userID][client] = struct{}{}
	for serverID := range client.serverIDs {
		if h.byServer[serverID] == nil {
			h.byServer[serverID] = make(map[*Client]struct{})
		}
		h.byServer[serverID][client] = struct{}{}
	}

	return client
}

func (h *Hub) Unregister(client *Client) {
	if client == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; !ok {
		return
	}

	delete(h.clients, client)
	userClients := h.byUser[client.UserID]
	delete(userClients, client)
	if len(userClients) == 0 {
		delete(h.byUser, client.UserID)
	}
	for serverID := range client.serverIDs {
		clients := h.byServer[serverID]
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.byServer, serverID)
		}
	}
	client.close()
}

func (h *Hub) BroadcastMessageCreated(message models.PublicMessage) int {
	return h.Broadcast(message.ServerID, MessageCreatedEvent(message))
}

func (h *Hub) BroadcastMessageUpdated(message models.PublicMessage) int {
	return h.Broadcast(message.ServerID, MessageUpdatedEvent(message))
}

func (h *Hub) BroadcastMessageDeleted(messageID, serverID, channelID uuid.UUID) int {
	return h.Broadcast(serverID, MessageDeletedEvent(messageID, serverID, channelID))
}

func (h *Hub) BroadcastPresence(serverIDs []uuid.UUID, presence models.PublicPresence) int {
	delivered := 0
	for _, serverID := range serverIDs {
		delivered += h.Broadcast(serverID, PresenceUpdateEvent(presence))
	}
	return delivered
}

func (h *Hub) BroadcastUser(userID uuid.UUID, event Event) int {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.byUser[userID]))
	for client := range h.byUser[userID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	delivered := 0
	for _, client := range clients {
		if client.Enqueue(event) {
			delivered++
		}
	}
	return delivered
}

func (h *Hub) BroadcastUsers(userIDs []uuid.UUID, event Event) int {
	delivered := 0
	seen := make(map[uuid.UUID]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		delivered += h.BroadcastUser(userID, event)
	}
	return delivered
}

func (h *Hub) Broadcast(serverID uuid.UUID, event Event) int {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.byServer[serverID]))
	for client := range h.byServer[serverID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	delivered := 0
	for _, client := range clients {
		if client.Enqueue(event) {
			delivered++
		}
	}
	return delivered
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
