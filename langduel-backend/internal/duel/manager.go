package duel

import (
	"errors"
	"strings"
	"sync"
)

// Manager управляет всеми комнатами и игровым потоком.
// Здесь единственное место, где изменяется состояние комнат.
var (
	GlobalManager   = NewManager()
	ErrRoomFull     = errors.New("room is full")
	ErrRoomNotFound = errors.New("room not found")
	ErrNotInRoom    = errors.New("player not in room")
	ErrNotStarted   = errors.New("room not started")
)

const (
	MaxPlayers = 2
	StartingHP = 100
)

type Event struct {
	Type       string         `json:"type"`
	RoomID     string         `json:"room_id"`
	Round      int            `json:"round,omitempty"`
	RoundToken int            `json:"round_token,omitempty"`
	Topic      string         `json:"topic,omitempty"`
	Lang       string         `json:"lang,omitempty"`
	Prompt     string         `json:"prompt,omitempty"`
	Players    []string       `json:"players,omitempty"`
	HP         map[string]int `json:"hp,omitempty"`
	AttackerID string         `json:"attacker_id,omitempty"`
	DefenderID string         `json:"defender_id,omitempty"`
	Damage     int            `json:"damage,omitempty"`
	Correct    bool           `json:"correct,omitempty"`
	Speed      int            `json:"speed,omitempty"`
	WinnerID   string         `json:"winner_id,omitempty"`
	Reason     string         `json:"reason,omitempty"`
	Error      string         `json:"error,omitempty"`
}

type Manager struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

// Join добавляет игрока в комнату и запускает первый раунд,
// когда в комнате появляются два игрока.
// Возвращает список событий для рассылки по комнате.
func (m *Manager) Join(roomID, userID, topic, lang string) ([]Event, error) {
	if roomID == "" || userID == "" {
		return nil, errors.New("room_id and user_id are required")
	}

	room := m.getOrCreateRoom(roomID)

	room.mu.Lock()
	defer room.mu.Unlock()

	if _, exists := room.Players[userID]; exists {
		// Повторный вход: возвращаем текущее состояние, чтобы клиент синхронизировался.
		return room.snapshotEventsLocked(), nil
	}
	if len(room.Players) >= MaxPlayers {
		return nil, ErrRoomFull
	}

	if room.Topic == "" && topic != "" {
		room.Topic = topic
	}
	if room.Lang == "" && lang != "" {
		room.Lang = lang
	}

	room.Players[userID] = &Player{ID: userID, HP: StartingHP}

	var events []Event
	events = append(events, Event{
		Type:    "player_joined",
		RoomID:  room.ID,
		Topic:   room.Topic,
		Lang:    room.Lang,
		Players: room.playerListLocked(),
		HP:      room.hpMapLocked(),
	})
	events = append(events, room.snapshotEventLocked())

	if len(room.Players) == MaxPlayers && !room.Started {
		// Стартуем первый раунд, когда в комнате два игрока.
		room.startRoundLocked()
		events = append(events, room.roundStartEventLocked())
	}

	return events, nil
}

// SubmitAnswer проверяет ответ, применяет урон и
// возвращает события для рассылки.
func (m *Manager) SubmitAnswer(roomID, userID, answer string, speed int) ([]Event, error) {
	room, err := m.getRoom(roomID)
	if err != nil {
		return nil, err
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if !room.Started {
		return nil, ErrNotStarted
	}

	attacker, ok := room.Players[userID]
	if !ok {
		return nil, ErrNotInRoom
	}

	defender := room.otherPlayerLocked(userID)
	if defender == nil {
		return nil, ErrNotInRoom
	}

	// Нормализуем обе строки, чтобы избежать проблем с регистром и пробелами.
	correct := normalize(answer) == normalize(room.Expected)
	damage := ProcessAnswer(attacker, defender, correct, speed)

	events := []Event{{
		Type:       "update",
		RoomID:     room.ID,
		Round:      room.Round,
		HP:         room.hpMapLocked(),
		AttackerID: attacker.ID,
		DefenderID: defender.ID,
		Damage:     damage,
		Correct:    correct,
		Speed:      speed,
	}}

	if defender.HP <= 0 {
		events = append(events, Event{
			Type:     "game_over",
			RoomID:   room.ID,
			HP:       room.hpMapLocked(),
			WinnerID: attacker.ID,
		})
		return events, nil
	}

	if correct {
		// Новый раунд запускаем только при правильном ответе.
		room.startRoundLocked()
		events = append(events, room.roundStartEventLocked())
	}

	return events, nil
}

// Leave удаляет игрока из комнаты и возвращает событие для рассылки.
func (m *Manager) Leave(roomID, userID string) ([]Event, error) {
	if roomID == "" || userID == "" {
		return nil, nil
	}

	room, err := m.getRoom(roomID)
	if err != nil {
		return nil, nil
	}

	room.mu.Lock()
	_, exists := room.Players[userID]
	if exists {
		delete(room.Players, userID)
		room.Started = false
		room.Prompt = ""
		room.Expected = ""
	}
	players := room.playerListLocked()
	hp := room.hpMapLocked()
	room.mu.Unlock()

	if !exists {
		return nil, nil
	}

	events := []Event{{
		Type:    "player_left",
		RoomID:  roomID,
		Players: players,
		HP:      hp,
		Reason:  "disconnect",
	}}

	// Если в комнате никого не осталось — удаляем ее.
	if len(players) == 0 {
		m.mu.Lock()
		delete(m.rooms, roomID)
		m.mu.Unlock()
	}

	return events, nil
}

// RoundTimeout завершает раунд по таймауту и запускает следующий.
func (m *Manager) RoundTimeout(roomID string, expectedToken int) ([]Event, error) {
	room, err := m.getRoom(roomID)
	if err != nil {
		return nil, err
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if !room.Started {
		return nil, ErrNotStarted
	}
	if room.RoundToken != expectedToken {
		return nil, nil
	}

	events := []Event{{
		Type:       "round_end",
		RoomID:     room.ID,
		Round:      room.Round,
		RoundToken: room.RoundToken,
		Prompt:     room.Prompt,
		HP:         room.hpMapLocked(),
		Reason:     "timeout",
	}}

	room.startRoundLocked()
	events = append(events, room.roundStartEventLocked())

	return events, nil
}

func (m *Manager) getOrCreateRoom(id string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[id]
	if ok {
		return room
	}

	room = &Room{
		ID:      id,
		Players: make(map[string]*Player),
	}
	m.rooms[id] = room
	return room
}

func (m *Manager) getRoom(id string) (*Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[id]
	if !ok {
		return nil, ErrRoomNotFound
	}
	return room, nil
}

// HasPlayer проверяет, есть ли игрок в комнате.
func (m *Manager) HasPlayer(roomID, userID string) bool {
	if roomID == "" || userID == "" {
		return false
	}
	room, err := m.getRoom(roomID)
	if err != nil {
		return false
	}
	room.mu.Lock()
	defer room.mu.Unlock()
	_, ok := room.Players[userID]
	return ok
}

// normalize: trim + lowercase перед сравнением.
func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
