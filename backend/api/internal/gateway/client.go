package gateway

import (
	"sync"

	"github.com/google/uuid"
)

type Client struct {
	UserID    uuid.UUID
	Username  string
	serverIDs map[uuid.UUID]struct{}
	send      chan Event
	done      chan struct{}
	mu        sync.RWMutex
	closed    bool
	closeOnce sync.Once
}

func newClient(userID uuid.UUID, username string, serverIDs []uuid.UUID) *Client {
	client := &Client{
		UserID:    userID,
		Username:  username,
		serverIDs: make(map[uuid.UUID]struct{}, len(serverIDs)),
		send:      make(chan Event, clientSendBufferSize),
		done:      make(chan struct{}),
	}

	for _, serverID := range serverIDs {
		if serverID != uuid.Nil {
			client.serverIDs[serverID] = struct{}{}
		}
	}

	return client
}

func (c *Client) Enqueue(event Event) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return false
	}

	select {
	case c.send <- event:
		return true
	default:
		return false
	}
}

func (c *Client) Send() <-chan Event {
	return c.send
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) ServerIDs() []uuid.UUID {
	serverIDs := make([]uuid.UUID, 0, len(c.serverIDs))
	for serverID := range c.serverIDs {
		serverIDs = append(serverIDs, serverID)
	}
	return serverIDs
}

func (c *Client) close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		c.closed = true
		close(c.done)
	})
}
