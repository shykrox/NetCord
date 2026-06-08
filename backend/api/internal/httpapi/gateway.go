package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"netcord/backend/api/internal/gateway"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	gatewayReadLimit = 4096
	gatewayWriteWait = 10 * time.Second
	gatewayPongWait  = 70 * time.Second
)

var gatewayUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		parsed, err := url.Parse(origin)
		return err == nil && strings.EqualFold(parsed.Host, r.Host)
	},
}

type incomingGatewayEvent struct {
	Type string `json:"type"`
	Data struct {
		ChannelID string `json:"channel_id"`
	} `json:"data"`
}

func (s *Server) gatewayWS(w http.ResponseWriter, r *http.Request) {
	userID, ok := s.gatewayUserID(w, r)
	if !ok {
		return
	}

	if s.serverService == nil || s.gatewayHub == nil {
		writeError(w, http.StatusInternalServerError, "gateway_unavailable", "gateway is unavailable", nil)
		return
	}

	servers, err := s.serverService.ListServers(r.Context(), userID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	serverIDs := make([]uuid.UUID, 0, len(servers.Servers))
	for _, server := range servers.Servers {
		serverIDs = append(serverIDs, server.ID)
	}

	conn, err := gatewayUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	username := ""
	if s.authService != nil {
		if user, err := s.authService.GetMe(r.Context(), userID); err == nil {
			username = user.Username
		}
	}

	if s.presenceService != nil {
		if presence, err := s.presenceService.SetOnline(r.Context(), userID); err == nil {
			s.gatewayHub.BroadcastPresence(serverIDs, presence)
		}
	}

	client := s.gatewayHub.RegisterUser(userID, username, serverIDs)
	_ = client.Enqueue(gateway.HelloEvent(userID, client.ServerIDs()))

	errCh := make(chan error, 2)
	go func() {
		errCh <- s.readGatewayPump(r.Context(), conn, client)
	}()
	go func() {
		errCh <- writeGatewayPump(conn, client)
	}()

	<-errCh
	s.gatewayHub.Unregister(client)
	if s.presenceService != nil {
		if presence, err := s.presenceService.SetOffline(context.Background(), userID); err == nil {
			s.gatewayHub.BroadcastPresence(serverIDs, presence)
		}
	}
	_ = conn.Close()
}

func (s *Server) gatewayUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	if s.tokens == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required", nil)
		return uuid.Nil, false
	}

	tokenString := bearerToken(r)
	if tokenString == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required", nil)
		return uuid.Nil, false
	}

	userID, err := s.tokens.Validate(tokenString)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required", nil)
		return uuid.Nil, false
	}

	return userID, true
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	}

	return strings.TrimSpace(r.URL.Query().Get("token"))
}

func (s *Server) readGatewayPump(ctx context.Context, conn *websocket.Conn, client *gateway.Client) error {
	conn.SetReadLimit(gatewayReadLimit)
	_ = conn.SetReadDeadline(time.Now().Add(gatewayPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(gatewayPongWait))
	})

	for {
		var event incomingGatewayEvent
		if err := conn.ReadJSON(&event); err != nil {
			return err
		}

		switch event.Type {
		case gateway.EventHeartbeat:
			client.Enqueue(gateway.HeartbeatAckEvent())
		case gateway.EventTypingStart, gateway.EventTypingStop:
			s.handleTypingGatewayEvent(ctx, client, event)
		default:
			client.Enqueue(gateway.ErrorEvent("unknown_event", "unsupported gateway event"))
		}
	}
}

func (s *Server) handleTypingGatewayEvent(ctx context.Context, client *gateway.Client, event incomingGatewayEvent) {
	if s.serverService == nil || s.gatewayHub == nil {
		client.Enqueue(gateway.ErrorEvent("typing_unavailable", "typing is unavailable"))
		return
	}

	channelID, err := uuid.Parse(event.Data.ChannelID)
	if err != nil {
		client.Enqueue(gateway.ErrorEvent("invalid_channel", "channel_id must be a valid uuid"))
		return
	}

	channel, err := s.serverService.GetChannel(ctx, client.UserID, channelID)
	if err != nil {
		client.Enqueue(gateway.ErrorEvent("not_found", "channel not found"))
		return
	}

	if event.Type == gateway.EventTypingStart {
		s.gatewayHub.Broadcast(channel.ServerID, gateway.TypingStartEvent(channel.ServerID, channel.ID, client.UserID, client.Username))
		return
	}
	s.gatewayHub.Broadcast(channel.ServerID, gateway.TypingStopEvent(channel.ID, client.UserID))
}

func writeGatewayPump(conn *websocket.Conn, client *gateway.Client) error {
	pingTicker := time.NewTicker(gateway.HeartbeatInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-client.Done():
			return writeGatewayClose(conn)
		case event := <-client.Send():
			if err := writeGatewayJSON(conn, event); err != nil {
				return err
			}
		case <-pingTicker.C:
			if err := conn.SetWriteDeadline(time.Now().Add(gatewayWriteWait)); err != nil {
				return err
			}
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(gatewayWriteWait)); err != nil {
				return err
			}
		}
	}
}

func writeGatewayJSON(conn *websocket.Conn, event gateway.Event) error {
	if err := conn.SetWriteDeadline(time.Now().Add(gatewayWriteWait)); err != nil {
		return err
	}
	return conn.WriteJSON(event)
}

func writeGatewayClose(conn *websocket.Conn) error {
	err := conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(gatewayWriteWait),
	)
	if errors.Is(err, websocket.ErrCloseSent) {
		return nil
	}
	return err
}
