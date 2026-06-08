package httpapi

import (
	"bufio"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"netcord/backend/api/internal/auth"
	"netcord/backend/api/internal/gateway"
	"netcord/backend/api/internal/middleware"
	"netcord/backend/api/internal/repository"
	"netcord/backend/api/internal/service"
)

type Server struct {
	authService       *service.AuthService
	serverService     *service.ServerService
	fileService       *service.FileService
	presenceService   *service.PresenceService
	socialService     *service.SocialService
	permissionService *service.PermissionService
	voiceService      *service.VoiceService
	aiService         *service.AIService
	tokens            *auth.TokenManager
	gatewayHub        *gateway.Hub
}

type RouterConfig struct {
	AuthService       *service.AuthService
	ServerService     *service.ServerService
	FileService       *service.FileService
	PresenceService   *service.PresenceService
	SocialService     *service.SocialService
	PermissionService *service.PermissionService
	VoiceService      *service.VoiceService
	AIService         *service.AIService
	Tokens            *auth.TokenManager
	GatewayHub        *gateway.Hub
}

func NewRouter(authService *service.AuthService, serverService *service.ServerService, fileService *service.FileService, presenceService *service.PresenceService, tokens *auth.TokenManager, gatewayHub *gateway.Hub) http.Handler {
	return NewRouterWithConfig(RouterConfig{
		AuthService:     authService,
		ServerService:   serverService,
		FileService:     fileService,
		PresenceService: presenceService,
		Tokens:          tokens,
		GatewayHub:      gatewayHub,
	})
}

func NewRouterWithConfig(config RouterConfig) http.Handler {
	server := &Server{
		authService:       config.AuthService,
		serverService:     config.ServerService,
		fileService:       config.FileService,
		presenceService:   config.PresenceService,
		socialService:     config.SocialService,
		permissionService: config.PermissionService,
		voiceService:      config.VoiceService,
		aiService:         config.AIService,
		tokens:            config.Tokens,
		gatewayHub:        config.GatewayHub,
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("POST /auth/register", server.register)
	mux.HandleFunc("POST /auth/login", server.login)
	mux.Handle("GET /users/me", protected(config.Tokens, server.me))
	mux.Handle("PATCH /users/me", protected(config.Tokens, server.updateMe))
	mux.Handle("POST /users/me/avatar", protected(config.Tokens, server.uploadAvatar))
	mux.Handle("POST /users/me/banner", protected(config.Tokens, server.uploadBanner))
	mux.Handle("GET /users/{user_id}", protected(config.Tokens, server.getUser))
	mux.Handle("PATCH /users/me/presence", protected(config.Tokens, server.updatePresence))
	mux.HandleFunc("GET /gateway/ws", server.gatewayWS)
	mux.Handle("POST /files/upload", protected(config.Tokens, server.uploadFile))
	mux.Handle("GET /files/{file_id}", protected(config.Tokens, server.getFile))

	mux.Handle("POST /servers", protected(config.Tokens, server.createServer))
	mux.Handle("GET /servers", protected(config.Tokens, server.listServers))
	mux.Handle("GET /servers/{server_id}", protected(config.Tokens, server.getServer))
	mux.Handle("PATCH /servers/{server_id}", protected(config.Tokens, server.updateServer))
	mux.Handle("DELETE /servers/{server_id}", protected(config.Tokens, server.deleteServer))
	mux.Handle("GET /servers/{server_id}/members", protected(config.Tokens, server.listServerMembers))
	mux.Handle("POST /servers/{server_id}/channels", protected(config.Tokens, server.createChannel))
	mux.Handle("GET /servers/{server_id}/channels", protected(config.Tokens, server.listChannels))
	mux.Handle("PATCH /channels/{channel_id}", protected(config.Tokens, server.updateChannel))
	mux.Handle("DELETE /channels/{channel_id}", protected(config.Tokens, server.deleteChannel))
	mux.Handle("GET /channels/{channel_id}/messages", protected(config.Tokens, server.listMessages))
	mux.Handle("GET /channels/{channel_id}/messages/search", protected(config.Tokens, server.searchMessages))
	mux.Handle("POST /channels/{channel_id}/messages", protected(config.Tokens, server.createMessage))
	mux.Handle("PATCH /messages/{message_id}", protected(config.Tokens, server.updateMessage))
	mux.Handle("DELETE /messages/{message_id}", protected(config.Tokens, server.deleteMessage))

	mux.Handle("POST /friends/requests", protected(config.Tokens, server.createFriendRequest))
	mux.Handle("GET /friends/requests", protected(config.Tokens, server.listFriendRequests))
	mux.Handle("POST /friends/requests/{request_id}/accept", protected(config.Tokens, server.acceptFriendRequest))
	mux.Handle("POST /friends/requests/{request_id}/decline", protected(config.Tokens, server.declineFriendRequest))
	mux.Handle("GET /friends", protected(config.Tokens, server.listFriends))
	mux.Handle("DELETE /friends/{user_id}", protected(config.Tokens, server.removeFriend))
	mux.Handle("POST /dm", protected(config.Tokens, server.createDM))
	mux.Handle("GET /dm", protected(config.Tokens, server.listDMs))
	mux.Handle("GET /dm/{conversation_id}/messages", protected(config.Tokens, server.listDMMessages))
	mux.Handle("POST /dm/{conversation_id}/messages", protected(config.Tokens, server.createDMMessage))

	mux.Handle("POST /servers/{server_id}/roles", protected(config.Tokens, server.createRole))
	mux.Handle("GET /servers/{server_id}/roles", protected(config.Tokens, server.listRoles))
	mux.Handle("PATCH /roles/{role_id}", protected(config.Tokens, server.updateRole))
	mux.Handle("DELETE /roles/{role_id}", protected(config.Tokens, server.deleteRole))
	mux.Handle("PUT /servers/{server_id}/members/{user_id}/roles/{role_id}", protected(config.Tokens, server.assignRole))
	mux.Handle("DELETE /servers/{server_id}/members/{user_id}/roles/{role_id}", protected(config.Tokens, server.removeRole))
	mux.Handle("POST /servers/{server_id}/invites", protected(config.Tokens, server.createInvite))
	mux.HandleFunc("GET /invites/{code}", server.getInvite)
	mux.Handle("POST /invites/{code}/join", protected(config.Tokens, server.joinInvite))

	mux.Handle("POST /voice/join", protected(config.Tokens, server.joinVoice))
	mux.Handle("POST /voice/leave", protected(config.Tokens, server.leaveVoice))

	mux.Handle("POST /ai/ask", protected(config.Tokens, server.createAskJob))
	mux.Handle("POST /ai/draw", protected(config.Tokens, server.createDrawJob))
	mux.Handle("GET /ai/jobs", protected(config.Tokens, server.listAIJobs))
	mux.Handle("GET /ai/jobs/{job_id}", protected(config.Tokens, server.getAIJob))

	return requestLogger(mux)
}

func protected(tokens *auth.TokenManager, handler http.HandlerFunc) http.Handler {
	return middleware.RequireAuth(tokens)(http.HandlerFunc(handler))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func (r *statusRecorder) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}

func (s *Server) writeServiceError(w http.ResponseWriter, err error) {
	var validation *service.ValidationError
	switch {
	case errors.As(err, &validation):
		writeError(w, http.StatusBadRequest, "validation_error", "request validation failed", validation.Fields)
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found", nil)
	case errors.Is(err, service.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "forbidden", nil)
	case errors.Is(err, service.ErrVoiceNotConfigured):
		writeError(w, http.StatusServiceUnavailable, "voice_not_configured", "LiveKit voice is not configured", nil)
	case errors.Is(err, service.ErrUnavailable):
		writeError(w, http.StatusServiceUnavailable, "service_unavailable", "service unavailable", nil)
	case errors.Is(err, repository.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "user_not_found", "user not found", nil)
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
