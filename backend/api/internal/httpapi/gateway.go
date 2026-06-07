package httpapi

import (
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

	client := s.gatewayHub.Register(userID, serverIDs)
	_ = client.Enqueue(gateway.HelloEvent(userID, client.ServerIDs()))

	errCh := make(chan error, 2)
	go func() {
		errCh <- readGatewayPump(conn, client)
	}()
	go func() {
		errCh <- writeGatewayPump(conn, client)
	}()

	<-errCh
	s.gatewayHub.Unregister(client)
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

func readGatewayPump(conn *websocket.Conn, client *gateway.Client) error {
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
		default:
			client.Enqueue(gateway.ErrorEvent("unknown_event", "unsupported gateway event"))
		}
	}
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
