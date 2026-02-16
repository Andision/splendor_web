package game

import "strings"

const (
	StatusPlaying  = "playing"
	StatusFinished = "finished"
)

const (
	GemWhite = "white"
	GemBlue  = "blue"
	GemGreen = "green"
	GemRed   = "red"
	GemBlack = "black"
	GemGold  = "gold"
)

var ColoredGems = []string{GemWhite, GemBlue, GemGreen, GemRed, GemBlack}

type TokenSet struct {
	White int `json:"white"`
	Blue  int `json:"blue"`
	Green int `json:"green"`
	Red   int `json:"red"`
	Black int `json:"black"`
	Gold  int `json:"gold"`
}

func (t TokenSet) Get(color string) int {
	switch strings.ToLower(color) {
	case GemWhite:
		return t.White
	case GemBlue:
		return t.Blue
	case GemGreen:
		return t.Green
	case GemRed:
		return t.Red
	case GemBlack:
		return t.Black
	case GemGold:
		return t.Gold
	default:
		return 0
	}
}

func (t *TokenSet) Add(color string, n int) {
	switch strings.ToLower(color) {
	case GemWhite:
		t.White += n
	case GemBlue:
		t.Blue += n
	case GemGreen:
		t.Green += n
	case GemRed:
		t.Red += n
	case GemBlack:
		t.Black += n
	case GemGold:
		t.Gold += n
	}
}

func (t *TokenSet) Sub(color string, n int) {
	t.Add(color, -n)
}

func (t TokenSet) Total() int {
	return t.White + t.Blue + t.Green + t.Red + t.Black + t.Gold
}

type Card struct {
	ID     string   `json:"id"`
	Tier   int      `json:"tier"`
	Bonus  string   `json:"bonus"`
	Points int      `json:"points"`
	Cost   TokenSet `json:"cost"`
}

type Noble struct {
	ID          string   `json:"id"`
	Points      int      `json:"points"`
	Requirement TokenSet `json:"requirement"`
}

type PlayerState struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Tokens         TokenSet  `json:"tokens"`
	Bonuses        TokenSet  `json:"bonuses"`
	Reserved       []Card    `json:"reserved"`
	PurchasedCount int       `json:"purchasedCount"`
	Points         int       `json:"points"`
	Nobles         []Noble   `json:"nobles"`
	IsConnected    bool      `json:"isConnected"`
	LastAction     string    `json:"lastAction"`
}

type State struct {
	Status          string        `json:"status"`
	Turn            int           `json:"turn"`
	CurrentPlayerID string        `json:"currentPlayerId"`
	Bank            TokenSet      `json:"bank"`
	Tier1           []Card        `json:"tier1"`
	Tier2           []Card        `json:"tier2"`
	Tier3           []Card        `json:"tier3"`
	Deck1Count      int           `json:"deck1Count"`
	Deck2Count      int           `json:"deck2Count"`
	Deck3Count      int           `json:"deck3Count"`
	Nobles          []Noble       `json:"nobles"`
	Players         []PlayerState `json:"players"`
	WinnerIDs       []string      `json:"winnerIds"`
	FinalRound      bool          `json:"finalRound"`
	FinalTurnsLeft  int           `json:"finalTurnsLeft"`
}

type Seat struct {
	ID   string
	Name string
}

type Action struct {
	Type    string      `json:"type"`
	Payload ActionInput `json:"payload"`
}

type ActionInput struct {
	Colors []string `json:"colors,omitempty"`
	CardID string   `json:"cardId,omitempty"`
	Source string   `json:"source,omitempty"`
}
