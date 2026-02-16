package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"splendor/backend/internal/game"
	"splendor/backend/internal/lobby"
	"splendor/backend/internal/ws"
)

type App struct {
	store    *lobby.Store
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

func New() *App {
	app := &App{
		store: lobby.NewStore(),
		hub:   ws.NewHub(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	app.startTimeoutLoop()
	return app
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", a.handleHealth)
	mux.HandleFunc("/api/rooms", a.handleRooms)
	mux.HandleFunc("/api/rooms/", a.handleRoomByID)
	mux.HandleFunc("/ws", a.handleWS)
	return withCORS(mux)
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

type createRoomRequest struct {
	HostName    string `json:"hostName"`
	TurnSeconds int    `json:"turnSeconds,omitempty"`
}

type createRoomResponse struct {
	Room   *lobby.Room  `json:"room"`
	Player lobby.Player `json:"player"`
}

func (a *App) handleRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}
	if strings.TrimSpace(req.HostName) == "" {
		writeError(w, http.StatusBadRequest, "invalid_host_name", "hostName is required")
		return
	}

	room, err := a.store.CreateRoom(req.HostName, req.TurnSeconds)
	if err != nil {
		writeLobbyError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, createRoomResponse{Room: room, Player: room.Players[0]})
}

type joinRoomRequest struct {
	PlayerName string `json:"playerName"`
}

type joinRoomResponse struct {
	Room   *lobby.Room  `json:"room"`
	Player lobby.Player `json:"player"`
}

type startGameRequest struct {
	PlayerID string `json:"playerId"`
}

type actionRequest struct {
	PlayerID string      `json:"playerId"`
	Action   game.Action `json:"action"`
}

func (a *App) handleRoomByID(w http.ResponseWriter, r *http.Request) {
	roomID, resource := parseRoomPath(r.URL.Path)
	if roomID == "" {
		writeError(w, http.StatusNotFound, "room_not_found", "room not found")
		return
	}

	switch {
	case resource == "" && r.Method == http.MethodGet:
		a.handleGetRoom(w, roomID)
	case resource == "join" && r.Method == http.MethodPost:
		a.handleJoinRoom(w, r, roomID)
	case resource == "start" && r.Method == http.MethodPost:
		a.handleStartGame(w, r, roomID)
	case resource == "state" && r.Method == http.MethodGet:
		a.handleGameState(w, roomID)
	case resource == "actions" && r.Method == http.MethodPost:
		a.handleAction(w, r, roomID)
	default:
		writeError(w, http.StatusNotFound, "route_not_found", "route not found")
	}
}

func parseRoomPath(path string) (roomID string, resource string) {
	trimmed := strings.TrimPrefix(path, "/api/rooms/")
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "", ""
	}

	roomID = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		resource = strings.ToLower(parts[1])
	}
	return roomID, resource
}

func (a *App) handleGetRoom(w http.ResponseWriter, roomID string) {
	room, err := a.store.GetRoom(roomID)
	if err != nil {
		writeLobbyError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, room)
}

func (a *App) handleJoinRoom(w http.ResponseWriter, r *http.Request, roomID string) {
	var req joinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}
	if strings.TrimSpace(req.PlayerName) == "" {
		writeError(w, http.StatusBadRequest, "invalid_player_name", "playerName is required")
		return
	}

	room, player, err := a.store.JoinRoom(roomID, req.PlayerName)
	if err != nil {
		writeLobbyError(w, err)
		return
	}

	a.broadcastRoomSnapshotRefs(room, "player_joined")
	writeJSON(w, http.StatusOK, joinRoomResponse{Room: room, Player: player})
}

func (a *App) handleStartGame(w http.ResponseWriter, r *http.Request, roomID string) {
	var req startGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}
	if strings.TrimSpace(req.PlayerID) == "" {
		writeError(w, http.StatusBadRequest, "invalid_player_id", "playerId is required")
		return
	}

	room, err := a.store.StartGame(roomID, req.PlayerID)
	if err != nil {
		writeLobbyError(w, err)
		return
	}

	a.broadcastRoomSnapshotRefs(room, "game_started")
	writeJSON(w, http.StatusOK, room)
}

func (a *App) handleGameState(w http.ResponseWriter, roomID string) {
	room, err := a.store.GetRoom(roomID)
	if err != nil {
		writeLobbyError(w, err)
		return
	}
	if room.Game == nil {
		writeError(w, http.StatusConflict, "game_not_started", "game not started")
		return
	}
	writeJSON(w, http.StatusOK, room.Game)
}

func (a *App) handleAction(w http.ResponseWriter, r *http.Request, roomID string) {
	var req actionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}
	if strings.TrimSpace(req.PlayerID) == "" {
		writeError(w, http.StatusBadRequest, "invalid_player_id", "playerId is required")
		return
	}

	room, err := a.store.ApplyAction(roomID, req.PlayerID, req.Action)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	a.broadcastRoomSnapshotRefs(room, "action_applied")
	writeJSON(w, http.StatusOK, room)
}

type wsClientMessage struct {
	Type   string      `json:"type"`
	Action game.Action `json:"action"`
}

func (a *App) handleWS(w http.ResponseWriter, r *http.Request) {
	roomID := strings.TrimSpace(r.URL.Query().Get("roomId"))
	playerID := strings.TrimSpace(r.URL.Query().Get("playerId"))
	if roomID == "" || playerID == "" {
		writeError(w, http.StatusBadRequest, "invalid_query", "roomId and playerId are required")
		return
	}

	room, err := a.store.GetRoom(roomID)
	if err != nil {
		writeLobbyError(w, err)
		return
	}

	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	a.hub.Add(roomID, conn)
	defer a.hub.Remove(roomID, conn)

	_ = a.store.SetConnected(roomID, playerID, true)
	defer func() {
		_ = a.store.SetConnected(roomID, playerID, false)
		if latestRoom, e := a.store.GetRoom(roomID); e == nil {
			a.broadcastRoomSnapshotRefs(latestRoom, "player_disconnected")
		}
	}()

	if room.Game != nil {
		_ = conn.WriteJSON(map[string]any{
			"type": "room_snapshot",
			"reason": "connected",
			"room": room,
		})
	}

	if latestRoom, e := a.store.GetRoom(roomID); e == nil {
		a.broadcastRoomSnapshotRefs(latestRoom, "player_connected")
	}

	for {
		var msg wsClientMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch strings.ToLower(strings.TrimSpace(msg.Type)) {
		case "action":
			updatedRoom, err := a.store.ApplyAction(roomID, playerID, msg.Action)
			if err != nil {
				_ = conn.WriteJSON(map[string]any{
					"type": "action_error",
					"error": err.Error(),
				})
				continue
			}
			a.broadcastRoomSnapshotRefs(updatedRoom, "action_applied")
		case "ping":
			_ = conn.WriteJSON(map[string]any{"type": "pong"})
		default:
			_ = conn.WriteJSON(map[string]any{
				"type": "action_error",
				"error": "unsupported message type",
			})
		}
	}
}

func (a *App) broadcastRoomSnapshot(roomID string, room *lobby.Room, reason string) {
	a.hub.Broadcast(roomID, map[string]any{
		"type":   "room_snapshot",
		"reason": reason,
		"room":   room,
	})
}

func (a *App) broadcastRoomSnapshotRefs(room *lobby.Room, reason string) {
	a.broadcastRoomSnapshot(room.ID, room, reason)
	if room.Code != "" && room.Code != room.ID {
		a.broadcastRoomSnapshot(room.Code, room, reason)
	}
}

func (a *App) startTimeoutLoop() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for now := range ticker.C {
			updates := a.store.ProcessTimeouts(now)
			for _, update := range updates {
				a.broadcastRoomSnapshotRefs(update.Room, "turn_timeout")
			}
		}
	}()
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, apiError{Code: code, Message: message})
}

func writeLobbyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, lobby.ErrRoomNotFound):
		writeError(w, http.StatusNotFound, "room_not_found", err.Error())
	case errors.Is(err, lobby.ErrRoomFull):
		writeError(w, http.StatusConflict, "room_full", err.Error())
	case errors.Is(err, lobby.ErrPlayerDuplicate):
		writeError(w, http.StatusConflict, "player_duplicate", err.Error())
	case errors.Is(err, lobby.ErrPlayerNotFound):
		writeError(w, http.StatusNotFound, "player_not_found", err.Error())
	case errors.Is(err, lobby.ErrInvalidTurnSeconds):
		writeError(w, http.StatusBadRequest, "invalid_turn_seconds", err.Error())
	case errors.Is(err, lobby.ErrOnlyHostCanStart):
		writeError(w, http.StatusForbidden, "only_host_can_start", err.Error())
	case errors.Is(err, lobby.ErrInvalidStartState), errors.Is(err, lobby.ErrGameAlreadyStarted), errors.Is(err, lobby.ErrGameNotStarted):
		writeError(w, http.StatusConflict, "invalid_room_state", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
	}
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, lobby.ErrRoomNotFound),
		errors.Is(err, lobby.ErrPlayerNotFound),
		errors.Is(err, lobby.ErrGameNotStarted):
		writeLobbyError(w, err)
	case errors.Is(err, game.ErrNotPlayerTurn),
		errors.Is(err, game.ErrUnknownAction),
		errors.Is(err, game.ErrInvalidAction),
		errors.Is(err, game.ErrGameFinished):
		writeError(w, http.StatusBadRequest, "invalid_action", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
