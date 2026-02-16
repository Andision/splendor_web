package lobby

import (
	"testing"

	"splendor/backend/internal/game"
)

func TestCreateJoinStartAndAction(t *testing.T) {
	store := NewStore()

	room := store.CreateRoom("host")
	if room.ID == "" {
		t.Fatal("expected room id")
	}

	joinedRoom, player, err := store.JoinRoom(room.ID, "friend")
	if err != nil {
		t.Fatalf("join room failed: %v", err)
	}
	if player.ID == "" {
		t.Fatal("expected player id")
	}
	if len(joinedRoom.Players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(joinedRoom.Players))
	}

	startedRoom, err := store.StartGame(room.ID, room.HostID)
	if err != nil {
		t.Fatalf("start game failed: %v", err)
	}
	if startedRoom.Game == nil {
		t.Fatal("expected game state")
	}
	if startedRoom.Status != RoomPlaying {
		t.Fatalf("expected room status %s, got %s", RoomPlaying, startedRoom.Status)
	}

	action := game.Action{Type: "take_tokens", Payload: game.ActionInput{Colors: []string{"white", "blue", "green"}}}
	actionedRoom, err := store.ApplyAction(room.ID, room.HostID, action)
	if err != nil {
		t.Fatalf("apply action failed: %v", err)
	}
	if actionedRoom.Game == nil {
		t.Fatal("expected game state after action")
	}
}

func TestOnlyHostCanStart(t *testing.T) {
	store := NewStore()
	room := store.CreateRoom("host")
	_, player, err := store.JoinRoom(room.ID, "friend")
	if err != nil {
		t.Fatalf("join failed: %v", err)
	}

	_, err = store.StartGame(room.ID, player.ID)
	if err == nil {
		t.Fatal("expected only host can start error")
	}
	if err != ErrOnlyHostCanStart {
		t.Fatalf("expected ErrOnlyHostCanStart, got %v", err)
	}
}
