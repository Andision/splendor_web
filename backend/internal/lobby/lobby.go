package lobby

import (
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"splendor/backend/internal/game"
)

var (
	ErrRoomNotFound       = errors.New("room not found")
	ErrRoomFull           = errors.New("room is full")
	ErrPlayerDuplicate    = errors.New("player already in room")
	ErrPlayerNotFound     = errors.New("player not found")
	ErrOnlyHostCanStart   = errors.New("only host can start")
	ErrInvalidStartState  = errors.New("cannot start game in current room state")
	ErrGameNotStarted     = errors.New("game not started")
	ErrGameAlreadyStarted = errors.New("game already started")
)

const MaxPlayers = 4

type RoomStatus string

const (
	RoomWaiting  RoomStatus = "waiting"
	RoomPlaying  RoomStatus = "playing"
	RoomFinished RoomStatus = "finished"
)

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Room struct {
	ID         string      `json:"id"`
	Code       string      `json:"code"`
	HostID     string      `json:"hostId"`
	Status     RoomStatus  `json:"status"`
	Players    []Player    `json:"players"`
	CreatedAt  time.Time   `json:"createdAt"`
	StartedAt  *time.Time  `json:"startedAt,omitempty"`
	FinishedAt *time.Time  `json:"finishedAt,omitempty"`
	Game       *game.State `json:"game,omitempty"`
}

type roomEntity struct {
	ID         string
	Code       string
	HostID     string
	Status     RoomStatus
	Players    []Player
	CreatedAt  time.Time
	StartedAt  *time.Time
	FinishedAt *time.Time
	Engine     *game.Engine
}

type Store struct {
	mu       sync.RWMutex
	rooms    map[string]*roomEntity
	codeToID map[string]string
}

func NewStore() *Store {
	return &Store{
		rooms:    make(map[string]*roomEntity),
		codeToID: make(map[string]string),
	}
}

func (s *Store) CreateRoom(hostName string) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	roomID := randomCode(6)
	roomCode := randomRoomCode(s.codeToID)
	host := Player{ID: randomCode(8), Name: strings.TrimSpace(hostName)}
	room := &roomEntity{
		ID:        roomID,
		Code:      roomCode,
		HostID:    host.ID,
		Status:    RoomWaiting,
		Players:   []Player{host},
		CreatedAt: time.Now().UTC(),
	}

	s.rooms[roomID] = room
	s.codeToID[roomCode] = roomID
	return snapshotRoom(room)
}

func (s *Store) JoinRoom(roomRef, playerName string) (*Room, Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.resolveRoomLocked(roomRef)
	if !ok {
		return nil, Player{}, ErrRoomNotFound
	}
	if room.Status != RoomWaiting {
		return nil, Player{}, ErrGameAlreadyStarted
	}

	for _, p := range room.Players {
		if strings.EqualFold(p.Name, playerName) {
			return nil, Player{}, ErrPlayerDuplicate
		}
	}

	if len(room.Players) >= MaxPlayers {
		return nil, Player{}, ErrRoomFull
	}

	player := Player{ID: randomCode(8), Name: strings.TrimSpace(playerName)}
	room.Players = append(room.Players, player)
	return snapshotRoom(room), player, nil
}

func (s *Store) StartGame(roomRef, playerID string) (*Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.resolveRoomLocked(roomRef)
	if !ok {
		return nil, ErrRoomNotFound
	}
	if room.Status != RoomWaiting {
		return nil, ErrInvalidStartState
	}
	if room.HostID != strings.TrimSpace(playerID) {
		return nil, ErrOnlyHostCanStart
	}
	if len(room.Players) < 2 {
		return nil, ErrInvalidStartState
	}

	seats := make([]game.Seat, 0, len(room.Players))
	for _, p := range room.Players {
		seats = append(seats, game.Seat{ID: p.ID, Name: p.Name})
	}

	engine, err := game.New(seats)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	room.Engine = engine
	room.StartedAt = &now
	room.Status = RoomPlaying
	return snapshotRoom(room), nil
}

func (s *Store) ApplyAction(roomRef, playerID string, action game.Action) (*Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.resolveRoomLocked(roomRef)
	if !ok {
		return nil, ErrRoomNotFound
	}
	if room.Engine == nil {
		return nil, ErrGameNotStarted
	}
	if !containsPlayer(room.Players, playerID) {
		return nil, ErrPlayerNotFound
	}

	if err := room.Engine.Apply(playerID, action); err != nil {
		return nil, err
	}

	snapshot := room.Engine.Snapshot()
	if snapshot.Status == game.StatusFinished && room.Status != RoomFinished {
		now := time.Now().UTC()
		room.Status = RoomFinished
		room.FinishedAt = &now
	}

	return snapshotRoom(room), nil
}

func (s *Store) SetConnected(roomRef, playerID string, connected bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.resolveRoomLocked(roomRef)
	if !ok {
		return ErrRoomNotFound
	}
	if room.Engine == nil {
		return nil
	}

	room.Engine.SetConnected(playerID, connected)
	return nil
}

func (s *Store) GetRoom(roomRef string) (*Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.resolveRoomLocked(roomRef)
	if !ok {
		return nil, ErrRoomNotFound
	}
	return snapshotRoom(room), nil
}

func (s *Store) resolveRoomLocked(roomRef string) (*roomEntity, bool) {
	key := strings.TrimSpace(roomRef)
	if key == "" {
		return nil, false
	}
	if room, ok := s.rooms[key]; ok {
		return room, true
	}
	if id, ok := s.codeToID[strings.ToLower(key)]; ok {
		room, ok := s.rooms[id]
		return room, ok
	}
	return nil, false
}

func containsPlayer(players []Player, playerID string) bool {
	for _, p := range players {
		if p.ID == playerID {
			return true
		}
	}
	return false
}

func snapshotRoom(room *roomEntity) *Room {
	out := &Room{
		ID:         room.ID,
		Code:       room.Code,
		HostID:     room.HostID,
		Status:     room.Status,
		Players:    append([]Player(nil), room.Players...),
		CreatedAt:  room.CreatedAt,
		StartedAt:  room.StartedAt,
		FinishedAt: room.FinishedAt,
	}
	if room.Engine != nil {
		s := room.Engine.Snapshot()
		out.Game = &s
	}
	return out
}

func randomCode(length int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	out := make([]byte, length)
	for i := 0; i < length; i++ {
		out[i] = chars[rand.Intn(len(chars))]
	}
	return string(out)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomRoomCode(used map[string]string) string {
	words := []string{
		"apple", "beach", "bread", "cloud", "coffee", "dance", "dream", "earth", "flame", "forest",
		"garden", "globe", "grape", "green", "happy", "honey", "hotel", "house", "island", "jelly",
		"juice", "light", "lucky", "magic", "mango", "maple", "melon", "metal", "money", "moon",
		"music", "night", "ocean", "olive", "party", "pearl", "piano", "pilot", "pizza", "plain",
		"plant", "queen", "quick", "radio", "river", "robot", "salad", "scale", "sheep", "smile",
		"snow", "sound", "spark", "spice", "sport", "star", "stone", "storm", "sugar", "sunny",
		"sweet", "table", "tiger", "toast", "tower", "train", "tree", "union", "urban", "vivid",
		"water", "whale", "white", "wind", "world", "yacht", "young", "zebra",
	}

	for i := 0; i < 256; i++ {
		code := words[rand.Intn(len(words))]
		if _, exists := used[code]; !exists {
			return code
		}
	}

	return "room" + strings.ToLower(randomCode(4))
}
