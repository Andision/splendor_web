package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type apiErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type roomPlayer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type gameState struct {
	Status          string `json:"status"`
	CurrentPlayerID string `json:"currentPlayerId"`
	Turn            int    `json:"turn"`
}

type roomDTO struct {
	ID      string       `json:"id"`
	HostID  string       `json:"hostId"`
	Status  string       `json:"status"`
	Players []roomPlayer `json:"players"`
	Game    *gameState   `json:"game,omitempty"`
}

type createRoomResp struct {
	Room   roomDTO     `json:"room"`
	Player roomPlayer  `json:"player"`
}

type joinRoomResp struct {
	Room   roomDTO    `json:"room"`
	Player roomPlayer `json:"player"`
}

type wsMessage struct {
	Type   string                 `json:"type"`
	Reason string                 `json:"reason,omitempty"`
	Room   *roomDTO               `json:"room,omitempty"`
	Error  string                 `json:"error,omitempty"`
	Extra  map[string]interface{} `json:"-"`
}

func TestHTTPRoomGameFlow(t *testing.T) {
	a := New()
	ts := httptest.NewServer(a.Routes())
	defer ts.Close()

	create := postJSON(t, ts.URL+"/api/rooms", map[string]any{"hostName": "Alice"}, http.StatusCreated)
	var createData createRoomResp
	decodeJSON(t, create, &createData)

	if createData.Room.ID == "" {
		t.Fatal("expected room id")
	}
	if createData.Player.ID == "" {
		t.Fatal("expected host player id")
	}

	join := postJSON(t, ts.URL+"/api/rooms/"+createData.Room.ID+"/join", map[string]any{"playerName": "Bob"}, http.StatusOK)
	var joinData joinRoomResp
	decodeJSON(t, join, &joinData)

	if len(joinData.Room.Players) != 2 {
		t.Fatalf("expected 2 players after join, got %d", len(joinData.Room.Players))
	}

	start := postJSON(t, ts.URL+"/api/rooms/"+createData.Room.ID+"/start", map[string]any{"playerId": createData.Player.ID}, http.StatusOK)
	var started roomDTO
	decodeJSON(t, start, &started)
	if started.Game == nil {
		t.Fatal("expected game state after start")
	}
	if started.Game.CurrentPlayerID != createData.Player.ID {
		t.Fatalf("expected first turn to host %s, got %s", createData.Player.ID, started.Game.CurrentPlayerID)
	}

	action := postJSON(t, ts.URL+"/api/rooms/"+createData.Room.ID+"/actions", map[string]any{
		"playerId": createData.Player.ID,
		"action": map[string]any{
			"type": "take_tokens",
			"payload": map[string]any{
				"colors": []string{"white", "blue", "green"},
			},
		},
	}, http.StatusOK)
	var actionRoom roomDTO
	decodeJSON(t, action, &actionRoom)
	if actionRoom.Game == nil {
		t.Fatal("expected game state after action")
	}
	if actionRoom.Game.Turn < 2 {
		t.Fatalf("expected turn increment, got %d", actionRoom.Game.Turn)
	}
	if actionRoom.Game.CurrentPlayerID == createData.Player.ID {
		t.Fatalf("expected turn moved to next player, still %s", actionRoom.Game.CurrentPlayerID)
	}
}

func TestHTTPActionNotPlayerTurn(t *testing.T) {
	a := New()
	ts := httptest.NewServer(a.Routes())
	defer ts.Close()

	create := postJSON(t, ts.URL+"/api/rooms", map[string]any{"hostName": "Alice"}, http.StatusCreated)
	var createData createRoomResp
	decodeJSON(t, create, &createData)

	join := postJSON(t, ts.URL+"/api/rooms/"+createData.Room.ID+"/join", map[string]any{"playerName": "Bob"}, http.StatusOK)
	var joinData joinRoomResp
	decodeJSON(t, join, &joinData)

	_ = postJSON(t, ts.URL+"/api/rooms/"+createData.Room.ID+"/start", map[string]any{"playerId": createData.Player.ID}, http.StatusOK)

	resp := postJSON(t, ts.URL+"/api/rooms/"+createData.Room.ID+"/actions", map[string]any{
		"playerId": joinData.Player.ID,
		"action": map[string]any{"type": "pass"},
	}, http.StatusBadRequest)
	var errBody apiErr
	decodeJSON(t, resp, &errBody)
	if errBody.Code != "invalid_action" {
		t.Fatalf("expected invalid_action, got %s", errBody.Code)
	}
}

func TestWebSocketPingAndInvalidAction(t *testing.T) {
	a := New()
	ts := httptest.NewServer(a.Routes())
	defer ts.Close()

	create := postJSON(t, ts.URL+"/api/rooms", map[string]any{"hostName": "Alice"}, http.StatusCreated)
	var createData createRoomResp
	decodeJSON(t, create, &createData)

	_ = postJSON(t, ts.URL+"/api/rooms/"+createData.Room.ID+"/join", map[string]any{"playerName": "Bob"}, http.StatusOK)
	_ = postJSON(t, ts.URL+"/api/rooms/"+createData.Room.ID+"/start", map[string]any{"playerId": createData.Player.ID}, http.StatusOK)

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws?roomId=" + createData.Room.ID + "&playerId=" + createData.Player.ID
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial failed: %v", err)
	}
	defer conn.Close()

	if _, err := readUntilType(t, conn, "room_snapshot"); err != nil {
		t.Fatalf("expected initial room snapshot: %v", err)
	}

	if err := conn.WriteJSON(map[string]any{"type": "ping"}); err != nil {
		t.Fatalf("write ping failed: %v", err)
	}
	if _, err := readUntilType(t, conn, "pong"); err != nil {
		t.Fatalf("expected pong: %v", err)
	}

	if err := conn.WriteJSON(map[string]any{
		"type": "action",
		"action": map[string]any{"type": "not_supported"},
	}); err != nil {
		t.Fatalf("write invalid action failed: %v", err)
	}

	msg, err := readUntilType(t, conn, "action_error")
	if err != nil {
		t.Fatalf("expected action_error: %v", err)
	}
	if !strings.Contains(msg.Error, "unknown action") {
		t.Fatalf("expected unknown action error, got %q", msg.Error)
	}
}

func TestCORSPreflight(t *testing.T) {
	a := New()
	ts := httptest.NewServer(a.Routes())
	defer ts.Close()

	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/api/rooms", nil)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("options request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("unexpected allow origin: %s", got)
	}
}

func readUntilType(t *testing.T, conn *websocket.Conn, want string) (wsMessage, error) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(400 * time.Millisecond))
		var raw map[string]any
		if err := conn.ReadJSON(&raw); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return wsMessage{}, err
			}
			continue
		}
		typ, _ := raw["type"].(string)
		if typ != want {
			continue
		}

		blob, _ := json.Marshal(raw)
		var msg wsMessage
		if err := json.Unmarshal(blob, &msg); err != nil {
			return wsMessage{}, err
		}
		return msg, nil
	}
	return wsMessage{}, &timeoutErr{want: want}
}

type timeoutErr struct {
	want string
}

func (e *timeoutErr) Error() string {
	return "timeout waiting for message type: " + e.want
}

func postJSON(t *testing.T, url string, payload any, wantStatus int) *http.Response {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	if resp.StatusCode != wantStatus {
		defer resp.Body.Close()
		var data map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&data)
		t.Fatalf("unexpected status: got %d want %d body=%v", resp.StatusCode, wantStatus, data)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, out any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		t.Fatalf("decode json failed: %v", err)
	}
}
