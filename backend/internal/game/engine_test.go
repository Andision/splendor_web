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

func TestTakeSingleTokenAllowed(t *testing.T) {
	engine, err := New([]Seat{{ID: "p1", Name: "A"}, {ID: "p2", Name: "B"}})
	if err != nil {
		t.Fatalf("new game failed: %v", err)
	}

	beforeWhite := engine.state.Bank.White
	err = engine.Apply("p1", Action{Type: "take_tokens", Payload: ActionInput{Colors: []string{"white"}}})
	if err != nil {
		t.Fatalf("single take failed: %v", err)
	}

	s := engine.Snapshot()
	p1 := s.Players[0]
	if p1.Tokens.White != 1 {
		t.Fatalf("expected 1 white token, got %+v", p1.Tokens)
	}
	if s.Bank.White != beforeWhite-1 {
		t.Fatalf("expected bank white %d, got %d", beforeWhite-1, s.Bank.White)
	}
}

func TestTakeTwoDifferentTokensAllowed(t *testing.T) {
	engine, err := New([]Seat{{ID: "p1", Name: "A"}, {ID: "p2", Name: "B"}})
	if err != nil {
		t.Fatalf("new game failed: %v", err)
	}

	beforeWhite := engine.state.Bank.White
	beforeBlue := engine.state.Bank.Blue
	err = engine.Apply("p1", Action{Type: "take_tokens", Payload: ActionInput{Colors: []string{"white", "blue"}}})
	if err != nil {
		t.Fatalf("take two different failed: %v", err)
	}

	s := engine.Snapshot()
	p1 := s.Players[0]
	if p1.Tokens.White != 1 || p1.Tokens.Blue != 1 {
		t.Fatalf("unexpected tokens after take: %+v", p1.Tokens)
	}
	if s.Bank.White != beforeWhite-1 || s.Bank.Blue != beforeBlue-1 {
		t.Fatalf("unexpected bank after take: %+v", s.Bank)
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

func TestDiscardTokens(t *testing.T) {
	engine, err := New([]Seat{{ID: "p1", Name: "A"}, {ID: "p2", Name: "B"}})
	if err != nil {
		t.Fatalf("new game failed: %v", err)
	}

	engine.state.Players[0].Tokens.White = 2
	engine.state.Players[0].Tokens.Blue = 1
	beforeWhite := engine.state.Bank.White
	beforeBlue := engine.state.Bank.Blue

	err = engine.Apply("p1", Action{Type: "discard_tokens", Payload: ActionInput{Colors: []string{"white", "blue"}}})
	if err != nil {
		t.Fatalf("discard failed: %v", err)
	}

	s := engine.Snapshot()
	p1 := s.Players[0]
	if p1.Tokens.White != 1 || p1.Tokens.Blue != 0 {
		t.Fatalf("unexpected tokens after discard: %+v", p1.Tokens)
	}
	if s.Bank.White != beforeWhite+1 || s.Bank.Blue != beforeBlue+1 {
		t.Fatalf("bank not updated after discard: %+v", s.Bank)
	}
	if s.CurrentPlayerID != "p2" {
		t.Fatalf("expected turn switch to p2, got %s", s.CurrentPlayerID)
	}
}

func TestDiscardTokensInsufficient(t *testing.T) {
	engine, err := New([]Seat{{ID: "p1", Name: "A"}, {ID: "p2", Name: "B"}})
	if err != nil {
		t.Fatalf("new game failed: %v", err)
	}

	err = engine.Apply("p1", Action{Type: "discard_tokens", Payload: ActionInput{Colors: []string{"white"}}})
	if err == nil {
		t.Fatal("expected insufficient token discard error")
	}
}

func TestAdjustTokensMixedTakeAndDiscard(t *testing.T) {
	engine, err := New([]Seat{{ID: "p1", Name: "A"}, {ID: "p2", Name: "B"}})
	if err != nil {
		t.Fatalf("new game failed: %v", err)
	}
	engine.state.Players[0].Tokens.White = 1

	beforeBlue := engine.state.Bank.Blue
	beforeWhite := engine.state.Bank.White
	err = engine.Apply("p1", Action{Type: "adjust_tokens", Payload: ActionInput{
		Adjust: map[string]int{
			"blue":  1,
			"white": -1,
		},
	}})
	if err != nil {
		t.Fatalf("adjust failed: %v", err)
	}

	s := engine.Snapshot()
	p1 := s.Players[0]
	if p1.Tokens.Blue != 1 || p1.Tokens.White != 0 {
		t.Fatalf("unexpected tokens after adjust: %+v", p1.Tokens)
	}
	if s.Bank.Blue != beforeBlue-1 || s.Bank.White != beforeWhite+1 {
		t.Fatalf("unexpected bank after adjust: %+v", s.Bank)
	}
}
