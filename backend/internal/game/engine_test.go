package game

import "testing"

func TestNewGame(t *testing.T) {
	engine, err := New([]Seat{{ID: "p1", Name: "A"}, {ID: "p2", Name: "B"}})
	if err != nil {
		t.Fatalf("new game failed: %v", err)
	}

	s := engine.Snapshot()
	if s.Status != StatusPlaying {
		t.Fatalf("expected status %s, got %s", StatusPlaying, s.Status)
	}
	if s.CurrentPlayerID != "p1" {
		t.Fatalf("expected current player p1, got %s", s.CurrentPlayerID)
	}
	if len(s.Tier1) != 4 || len(s.Tier2) != 4 || len(s.Tier3) != 4 {
		t.Fatalf("expected 4 face-up cards per tier")
	}
	if len(s.Nobles) != 3 {
		t.Fatalf("expected 3 nobles for 2 players, got %d", len(s.Nobles))
	}
}

func TestTakeTokensAndTurnSwitch(t *testing.T) {
	engine, err := New([]Seat{{ID: "p1", Name: "A"}, {ID: "p2", Name: "B"}})
	if err != nil {
		t.Fatalf("new game failed: %v", err)
	}

	err = engine.Apply("p1", Action{Type: "take_tokens", Payload: ActionInput{Colors: []string{"white", "blue", "green"}}})
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	s := engine.Snapshot()
	if s.CurrentPlayerID != "p2" {
		t.Fatalf("expected current player p2, got %s", s.CurrentPlayerID)
	}
	p1 := s.Players[0]
	if p1.Tokens.White != 1 || p1.Tokens.Blue != 1 || p1.Tokens.Green != 1 {
		t.Fatalf("unexpected tokens after take action: %+v", p1.Tokens)
	}
}

func TestNotPlayerTurn(t *testing.T) {
	engine, err := New([]Seat{{ID: "p1", Name: "A"}, {ID: "p2", Name: "B"}})
	if err != nil {
		t.Fatalf("new game failed: %v", err)
	}

	err = engine.Apply("p2", Action{Type: "pass"})
	if err == nil {
		t.Fatal("expected not turn error")
	}
	if err != ErrNotPlayerTurn {
		t.Fatalf("expected ErrNotPlayerTurn, got %v", err)
	}
}
