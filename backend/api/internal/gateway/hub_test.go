package gateway

import (
	"testing"

	"netcord/backend/api/internal/models"

	"github.com/google/uuid"
)

func TestHubBroadcastsOnlyToSubscribedServerMembers(t *testing.T) {
	hub := NewHub()
	serverID := uuid.New()
	otherServerID := uuid.New()
	member := hub.Register(uuid.New(), []uuid.UUID{serverID})
	otherMember := hub.Register(uuid.New(), []uuid.UUID{otherServerID})

	message := models.PublicMessage{
		ID:       uuid.New(),
		ServerID: serverID,
		Content:  "hello",
	}

	delivered := hub.BroadcastMessageCreated(message)
	if delivered != 1 {
		t.Fatalf("expected one delivery, got %d", delivered)
	}

	event := <-member.Send()
	if event.Type != EventMessageCreated {
		t.Fatalf("expected %s, got %s", EventMessageCreated, event.Type)
	}

	select {
	case event := <-otherMember.Send():
		t.Fatalf("unexpected event for non-member subscription: %+v", event)
	default:
	}
}

func TestHubUnregisterDisconnectsClient(t *testing.T) {
	hub := NewHub()
	client := hub.Register(uuid.New(), []uuid.UUID{uuid.New()})

	hub.Unregister(client)

	if hub.ClientCount() != 0 {
		t.Fatalf("expected no clients, got %d", hub.ClientCount())
	}
	if client.Enqueue(Event{Type: EventHeartbeatAck}) {
		t.Fatal("expected enqueue to fail after unregister")
	}
}
