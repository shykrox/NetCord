package gateway

import (
	"time"

	"netcord/backend/api/internal/models"

	"github.com/google/uuid"
)

const (
	EventHello            = "hello"
	EventHeartbeat        = "heartbeat"
	EventHeartbeatAck     = "heartbeat_ack"
	EventMessageCreated   = "message.created"
	EventMessageUpdated   = "message.updated"
	EventMessageDeleted   = "message.deleted"
	EventPresenceUpdate   = "presence.update"
	EventTypingStart      = "typing.start"
	EventTypingStop       = "typing.stop"
	EventFriendRequested  = "friend.requested"
	EventFriendAccepted   = "friend.accepted"
	EventDMMessageCreated = "dm.message.created"
	EventVoiceJoined      = "voice.joined"
	EventVoiceLeft        = "voice.left"
	EventVoiceState       = "voice.state"
	EventJobProgress      = "job.progress"
	EventJobCompleted     = "job.completed"
	EventError            = "error"
	HeartbeatInterval     = 30 * time.Second
	clientSendBufferSize  = 64
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

type TypingData struct {
	ServerID  uuid.UUID `json:"server_id,omitempty"`
	ChannelID uuid.UUID `json:"channel_id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	StartedAt time.Time `json:"started_at,omitempty"`
}

type MessageDeletedData struct {
	MessageID uuid.UUID `json:"message_id"`
	ServerID  uuid.UUID `json:"server_id"`
	ChannelID uuid.UUID `json:"channel_id"`
}

type VoiceStateData struct {
	ServerID  uuid.UUID `json:"server_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	UserID    uuid.UUID `json:"user_id"`
	Room      string    `json:"room"`
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

func MessageUpdatedEvent(message models.PublicMessage) Event {
	return Event{
		Type: EventMessageUpdated,
		Data: message,
	}
}

func MessageDeletedEvent(messageID, serverID, channelID uuid.UUID) Event {
	return Event{
		Type: EventMessageDeleted,
		Data: MessageDeletedData{
			MessageID: messageID,
			ServerID:  serverID,
			ChannelID: channelID,
		},
	}
}

func PresenceUpdateEvent(presence models.PublicPresence) Event {
	return Event{Type: EventPresenceUpdate, Data: presence}
}

func TypingStartEvent(serverID, channelID, userID uuid.UUID, username string) Event {
	return Event{
		Type: EventTypingStart,
		Data: TypingData{
			ServerID:  serverID,
			ChannelID: channelID,
			UserID:    userID,
			Username:  username,
			StartedAt: time.Now().UTC(),
		},
	}
}

func TypingStopEvent(channelID, userID uuid.UUID) Event {
	return Event{
		Type: EventTypingStop,
		Data: TypingData{
			ChannelID: channelID,
			UserID:    userID,
		},
	}
}

func FriendRequestedEvent(request models.PublicFriendRequest) Event {
	return Event{Type: EventFriendRequested, Data: request}
}

func FriendAcceptedEvent(request models.PublicFriendRequest) Event {
	return Event{Type: EventFriendAccepted, Data: request}
}

func DMMessageCreatedEvent(message models.PublicDMMessage) Event {
	return Event{Type: EventDMMessageCreated, Data: message}
}

func VoiceJoinedEvent(serverID, channelID, userID uuid.UUID, room string) Event {
	return Event{
		Type: EventVoiceJoined,
		Data: VoiceStateData{
			ServerID:  serverID,
			ChannelID: channelID,
			UserID:    userID,
			Room:      room,
		},
	}
}

func VoiceLeftEvent(serverID, channelID, userID uuid.UUID, room string) Event {
	return Event{
		Type: EventVoiceLeft,
		Data: VoiceStateData{
			ServerID:  serverID,
			ChannelID: channelID,
			UserID:    userID,
			Room:      room,
		},
	}
}

func JobProgressEvent(job models.PublicAIJob) Event {
	return Event{Type: EventJobProgress, Data: job}
}

func JobCompletedEvent(job models.PublicAIJob) Event {
	return Event{Type: EventJobCompleted, Data: job}
}
