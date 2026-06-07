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
}

func NewHub() *Hub {
	return &Hub{
		clients:  make(map[*Client]struct{}),
		byServer: make(map[uuid.UUID]map[*Client]struct{}),
	}
}

func (h *Hub) Register(userID uuid.UUID, serverIDs []uuid.UUID) *Client {
	client := newClient(userID, serverIDs)

	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = struct{}{}
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
