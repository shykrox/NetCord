package gateway

import (
	"time"

	"netcord/backend/api/internal/models"

	"github.com/google/uuid"
)

const (
	EventHello           = "hello"
	EventHeartbeat       = "heartbeat"
	EventHeartbeatAck    = "heartbeat_ack"
	EventMessageCreated  = "message.created"
	EventError           = "error"
	HeartbeatInterval    = 30 * time.Second
	clientSendBufferSize = 64
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

type HelloData struct {
	UserID              uuid.UUID   `json:"user_id"`
	ServerIDs           []uuid.UUID `json:"server_ids"`
	HeartbeatIntervalMS int64       `json:"heartbeat_interval_ms"`
}

type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func HelloEvent(userID uuid.UUID, serverIDs []uuid.UUID) Event {
	return Event{
		Type: EventHello,
		Data: HelloData{
			UserID:              userID,
			ServerIDs:           serverIDs,
			HeartbeatIntervalMS: HeartbeatInterval.Milliseconds(),
		},
	}
}

func HeartbeatAckEvent() Event {
	return Event{Type: EventHeartbeatAck}
}

func ErrorEvent(code, message string) Event {
	return Event{
		Type: EventError,
		Data: ErrorData{
			Code:    code,
			Message: message,
		},
	}
}

func MessageCreatedEvent(message models.PublicMessage) Event {
	return Event{
		Type: EventMessageCreated,
		Data: message,
	}
}
