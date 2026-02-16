package game

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"
)

var (
	ErrInvalidPlayerCount = errors.New("invalid player count")
	ErrNotPlayerTurn      = errors.New("not player's turn")
	ErrGameFinished       = errors.New("game already finished")
	ErrUnknownAction      = errors.New("unknown action")
	ErrInvalidAction      = errors.New("invalid action")
)

type Engine struct {
	state *State
	deck1 []Card
	deck2 []Card
	deck3 []Card
}

func New(seats []Seat) (*Engine, error) {
	if len(seats) < 2 || len(seats) > 4 {
		return nil, ErrInvalidPlayerCount
	}

	rand.Seed(time.Now().UnixNano())
	deck1, deck2, deck3 := initDecks()
	nobles := noblesDataset()
	rand.Shuffle(len(nobles), func(i, j int) { nobles[i], nobles[j] = nobles[j], nobles[i] })

	players := make([]PlayerState, 0, len(seats))
	for _, seat := range seats {
		players = append(players, PlayerState{
			ID:          seat.ID,
			Name:        seat.Name,
			IsConnected: true,
		})
	}

	bankCount := tokenCountByPlayers(len(seats))
	state := &State{
		Status:          StatusPlaying,
		Turn:            1,
		CurrentPlayerID: seats[0].ID,
		Bank: TokenSet{
			White: bankCount,
			Blue:  bankCount,
			Green: bankCount,
			Red:   bankCount,
			Black: bankCount,
			Gold:  5,
		},
		Tier1: append([]Card(nil), draw(&deck1, 4)...),
		Tier2: append([]Card(nil), draw(&deck2, 4)...),
		Tier3: append([]Card(nil), draw(&deck3, 4)...),
		Nobles: append([]Noble(nil),
			nobles[:len(seats)+1]...,
		),
		Players: players,
	}
	state.Deck1Count = len(deck1)
	state.Deck2Count = len(deck2)
	state.Deck3Count = len(deck3)

	return &Engine{state: state, deck1: deck1, deck2: deck2, deck3: deck3}, nil
}

func tokenCountByPlayers(n int) int {
	switch n {
	case 2:
		return 4
	case 3:
		return 5
	default:
		return 7
	}
}

func (e *Engine) Snapshot() State {
	copy := *e.state
	copy.Players = append([]PlayerState(nil), e.state.Players...)
	for i := range copy.Players {
		copy.Players[i].Reserved = append([]Card(nil), e.state.Players[i].Reserved...)
		copy.Players[i].Nobles = append([]Noble(nil), e.state.Players[i].Nobles...)
	}
	copy.Tier1 = append([]Card(nil), e.state.Tier1...)
	copy.Tier2 = append([]Card(nil), e.state.Tier2...)
	copy.Tier3 = append([]Card(nil), e.state.Tier3...)
	copy.Nobles = append([]Noble(nil), e.state.Nobles...)
	copy.WinnerIDs = append([]string(nil), e.state.WinnerIDs...)
	return copy
}

func (e *Engine) SetConnected(playerID string, connected bool) {
	idx := e.playerIndex(playerID)
	if idx == -1 {
		return
	}
	e.state.Players[idx].IsConnected = connected
}

func (e *Engine) Apply(playerID string, action Action) error {
	if e.state.Status == StatusFinished {
		return ErrGameFinished
	}
	if e.state.CurrentPlayerID != playerID {
		return ErrNotPlayerTurn
	}

	switch strings.ToLower(strings.TrimSpace(action.Type)) {
	case "take_tokens":
		if err := e.applyTakeTokens(playerID, action.Payload.Colors); err != nil {
			return err
		}
	case "reserve_card":
		if err := e.applyReserveCard(playerID, action.Payload.CardID); err != nil {
			return err
		}
	case "buy_card":
		if err := e.applyBuyCard(playerID, action.Payload.CardID, action.Payload.Source); err != nil {
			return err
		}
	case "pass":
		// no-op
	default:
		return ErrUnknownAction
	}

	e.endTurn(playerID, action.Type)
	return nil
}

func (e *Engine) applyTakeTokens(playerID string, colors []string) error {
	idx := e.playerIndex(playerID)
	if idx == -1 {
		return ErrInvalidAction
	}
	if len(colors) == 0 {
		return fmt.Errorf("%w: colors is required", ErrInvalidAction)
	}

	normalized := make([]string, 0, len(colors))
	for _, c := range colors {
		c = strings.ToLower(strings.TrimSpace(c))
		if !slices.Contains(ColoredGems, c) {
			return fmt.Errorf("%w: unsupported gem color", ErrInvalidAction)
		}
		normalized = append(normalized, c)
	}

	p := &e.state.Players[idx]
	projected := p.Tokens.Total() + len(normalized)
	if projected > 10 {
		return fmt.Errorf("%w: token limit exceeded", ErrInvalidAction)
	}

	if len(normalized) == 2 && normalized[0] == normalized[1] {
		color := normalized[0]
		if e.state.Bank.Get(color) < 4 {
			return fmt.Errorf("%w: bank needs at least 4 of the same color", ErrInvalidAction)
		}
		e.state.Bank.Sub(color, 2)
		p.Tokens.Add(color, 2)
		return nil
	}

	if len(normalized) != 3 {
		return fmt.Errorf("%w: take must be 3 different or 2 same", ErrInvalidAction)
	}

	seen := map[string]struct{}{}
	for _, c := range normalized {
		if _, ok := seen[c]; ok {
			return fmt.Errorf("%w: colors must be unique", ErrInvalidAction)
		}
		seen[c] = struct{}{}
		if e.state.Bank.Get(c) < 1 {
			return fmt.Errorf("%w: bank has no token for color %s", ErrInvalidAction, c)
		}
	}

	for _, c := range normalized {
		e.state.Bank.Sub(c, 1)
		p.Tokens.Add(c, 1)
	}
	return nil
}

func (e *Engine) applyReserveCard(playerID, cardID string) error {
	idx := e.playerIndex(playerID)
	if idx == -1 {
		return ErrInvalidAction
	}
	cardID = strings.TrimSpace(cardID)
	if cardID == "" {
		return fmt.Errorf("%w: cardId is required", ErrInvalidAction)
	}

	p := &e.state.Players[idx]
	if len(p.Reserved) >= 3 {
		return fmt.Errorf("%w: reserved card limit is 3", ErrInvalidAction)
	}

	card, ok := e.takeTableauCardByID(cardID)
	if !ok {
		return fmt.Errorf("%w: card not found in tableau", ErrInvalidAction)
	}

	p.Reserved = append(p.Reserved, card)

	if e.state.Bank.Gold > 0 && p.Tokens.Total() < 10 {
		e.state.Bank.Gold--
		p.Tokens.Gold++
	}
	return nil
}

func (e *Engine) applyBuyCard(playerID, cardID, source string) error {
	idx := e.playerIndex(playerID)
	if idx == -1 {
		return ErrInvalidAction
	}
	cardID = strings.TrimSpace(cardID)
	if cardID == "" {
		return fmt.Errorf("%w: cardId is required", ErrInvalidAction)
	}
	source = strings.ToLower(strings.TrimSpace(source))
	if source == "" {
		source = "tableau"
	}

	p := &e.state.Players[idx]
	var card Card
	var ok bool

	switch source {
	case "tableau":
		card, ok = e.takeTableauCardByID(cardID)
		if !ok {
			return fmt.Errorf("%w: card not found in tableau", ErrInvalidAction)
		}
	case "reserved":
		card, ok = takeReservedCardByID(p, cardID)
		if !ok {
			return fmt.Errorf("%w: card not found in reserved", ErrInvalidAction)
		}
	default:
		return fmt.Errorf("%w: source must be tableau or reserved", ErrInvalidAction)
	}

	payment, canPay := calculatePayment(card.Cost, p.Bonuses, p.Tokens)
	if !canPay {
		// restore card if we already removed it.
		e.restoreCard(card, source, p)
		return fmt.Errorf("%w: not enough tokens to buy card", ErrInvalidAction)
	}

	applyPayment(&p.Tokens, &e.state.Bank, payment)
	grantCard(p, card)
	e.tryClaimNoble(p)
	return nil
}

func (e *Engine) restoreCard(card Card, source string, p *PlayerState) {
	switch source {
	case "tableau":
		switch card.Tier {
		case 1:
			e.state.Tier1 = append(e.state.Tier1, card)
		case 2:
			e.state.Tier2 = append(e.state.Tier2, card)
		case 3:
			e.state.Tier3 = append(e.state.Tier3, card)
		}
	case "reserved":
		p.Reserved = append(p.Reserved, card)
	}
}

func calculatePayment(cost TokenSet, bonuses TokenSet, tokens TokenSet) (TokenSet, bool) {
	payment := TokenSet{}
	goldNeed := 0

	for _, color := range ColoredGems {
		need := cost.Get(color) - bonuses.Get(color)
		if need < 0 {
			need = 0
		}
		have := tokens.Get(color)
		useColored := min(need, have)
		payment.Add(color, useColored)
		short := need - useColored
		if short > 0 {
			goldNeed += short
		}
	}

	if goldNeed > tokens.Gold {
		return TokenSet{}, false
	}
	payment.Gold = goldNeed
	return payment, true
}

func applyPayment(player *TokenSet, bank *TokenSet, pay TokenSet) {
	for _, color := range append(append([]string(nil), ColoredGems...), GemGold) {
		n := pay.Get(color)
		if n == 0 {
			continue
		}
		player.Sub(color, n)
		bank.Add(color, n)
	}
}

func grantCard(p *PlayerState, card Card) {
	p.PurchasedCount++
	p.Points += card.Points
	p.Bonuses.Add(card.Bonus, 1)
}

func takeReservedCardByID(p *PlayerState, cardID string) (Card, bool) {
	for i := range p.Reserved {
		if p.Reserved[i].ID == cardID {
			card := p.Reserved[i]
			p.Reserved = append(p.Reserved[:i], p.Reserved[i+1:]...)
			return card, true
		}
	}
	return Card{}, false
}

func (e *Engine) takeTableauCardByID(cardID string) (Card, bool) {
	cardID = strings.TrimSpace(cardID)
	if cardID == "" {
		return Card{}, false
	}

	if c, ok := takeCardFromPile(&e.state.Tier1, &e.deck1, cardID); ok {
		e.state.Deck1Count = len(e.deck1)
		return c, true
	}
	if c, ok := takeCardFromPile(&e.state.Tier2, &e.deck2, cardID); ok {
		e.state.Deck2Count = len(e.deck2)
		return c, true
	}
	if c, ok := takeCardFromPile(&e.state.Tier3, &e.deck3, cardID); ok {
		e.state.Deck3Count = len(e.deck3)
		return c, true
	}
	return Card{}, false
}

func takeCardFromPile(tableau *[]Card, deck *[]Card, cardID string) (Card, bool) {
	for i := range *tableau {
		if (*tableau)[i].ID == cardID {
			card := (*tableau)[i]
			*tableau = append((*tableau)[:i], (*tableau)[i+1:]...)
			if len(*deck) > 0 {
				drawn := draw(deck, 1)
				*tableau = append(*tableau, drawn[0])
			}
			return card, true
		}
	}
	return Card{}, false
}

func (e *Engine) tryClaimNoble(p *PlayerState) {
	for i := range e.state.Nobles {
		n := e.state.Nobles[i]
		if hasNobleRequirement(p.Bonuses, n.Requirement) {
			p.Nobles = append(p.Nobles, n)
			p.Points += n.Points
			e.state.Nobles = append(e.state.Nobles[:i], e.state.Nobles[i+1:]...)
			return
		}
	}
}

func hasNobleRequirement(bonuses TokenSet, req TokenSet) bool {
	for _, color := range ColoredGems {
		if bonuses.Get(color) < req.Get(color) {
			return false
		}
	}
	return true
}

func (e *Engine) endTurn(playerID, actionType string) {
	idx := e.playerIndex(playerID)
	if idx == -1 {
		return
	}
	e.state.Players[idx].LastAction = actionType

	if !e.state.FinalRound && e.state.Players[idx].Points >= 15 {
		e.state.FinalRound = true
		e.state.FinalTurnsLeft = len(e.state.Players) - 1
	} else if e.state.FinalRound {
		if e.state.FinalTurnsLeft == 0 {
			e.finishGame()
			return
		}
		e.state.FinalTurnsLeft--
	}

	next := (idx + 1) % len(e.state.Players)
	e.state.CurrentPlayerID = e.state.Players[next].ID
	e.state.Turn++
}

func (e *Engine) finishGame() {
	e.state.Status = StatusFinished
	e.state.WinnerIDs = computeWinners(e.state.Players)
}

func computeWinners(players []PlayerState) []string {
	maxPoint := -1
	for _, p := range players {
		if p.Points > maxPoint {
			maxPoint = p.Points
		}
	}

	contenders := make([]PlayerState, 0)
	for _, p := range players {
		if p.Points == maxPoint {
			contenders = append(contenders, p)
		}
	}

	if len(contenders) == 1 {
		return []string{contenders[0].ID}
	}

	minCards := int(^uint(0) >> 1)
	for _, p := range contenders {
		if p.PurchasedCount < minCards {
			minCards = p.PurchasedCount
		}
	}

	winners := make([]string, 0)
	for _, p := range contenders {
		if p.PurchasedCount == minCards {
			winners = append(winners, p.ID)
		}
	}
	return winners
}

func (e *Engine) playerIndex(playerID string) int {
	for i := range e.state.Players {
		if e.state.Players[i].ID == playerID {
			return i
		}
	}
	return -1
}
