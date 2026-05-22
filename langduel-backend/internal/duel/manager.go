package duel

import (
	"errors"
	"log"
	"strings"
	"sync"
)

// Manager управляет всеми комнатами и игровым потоком.
// Здесь единственное место, где изменяется состояние комнат.
var (
	GlobalManager        = NewManager()
	ErrRoomFull          = errors.New("room is full")
	ErrRoomNotFound      = errors.New("room not found")
	ErrNotInRoom         = errors.New("player not in room")
	ErrNotStarted        = errors.New("room not started")
	ErrRoundTokenMismatch = errors.New("round token mismatch")
)

const (
	MaxPlayers          = 2
	StartingHP          = 100
	MaxRounds           = 20
	HalfSize            = 10
	HalftimeDurationSec = 8
)

type Event struct {
	Type          string            `json:"type"`
	RoomID        string            `json:"room_id"`
	DuelID        string            `json:"duel_id,omitempty"`
	Round         int               `json:"round"`
	Half          int               `json:"half,omitempty"`
	RoundToken    int               `json:"round_token,omitempty"`
	TotalPhrases  int               `json:"total_phrases,omitempty"`
	Topic         string            `json:"topic,omitempty"`
	Difficulty    string            `json:"difficulty,omitempty"`
	Lang          string            `json:"lang,omitempty"`
	Prompt        string            `json:"prompt,omitempty"`
	CorrectAnswer string            `json:"correct_answer,omitempty"`
	ValidAnswers  []string          `json:"valid_answers,omitempty"`
	Players       []string          `json:"players,omitempty"`
	HP            map[string]int    `json:"hp"`
	Avatars       map[string]string `json:"avatars,omitempty"`
	Elo           map[string]int    `json:"elo,omitempty"`
	AttackerID    string            `json:"attacker_id"`
	DefenderID    string            `json:"defender_id"`
	Damage        int               `json:"damage"`
	SelfDamage    int               `json:"self_damage"`
	Correct       bool              `json:"correct"`
	Speed         int               `json:"speed"`
	WinnerID      string            `json:"winner_id,omitempty"`
	Reason        string            `json:"reason,omitempty"`
	Error         string            `json:"error,omitempty"`
	EloChange     map[string]int    `json:"elo_change,omitempty"`
	CorrectCount  map[string]int    `json:"correct_count,omitempty"`
	WrongCount    map[string]int    `json:"wrong_count,omitempty"`
	UserID        string            `json:"user_id,omitempty"`
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
func (m *Manager) Join(roomID, userID, topic, difficulty, lang, avatar string) ([]Event, error) {
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
	if room.Difficulty == "" && difficulty != "" {
		room.Difficulty = difficulty
	}
	if room.Lang == "" && lang != "" {
		room.Lang = lang
	}

	// Set local topic phrases (fallback if no AI phrases)
	if len(room.LocalPhrases) == 0 && room.Topic != "" && room.Difficulty != "" {
		room.SetLocalPhrases(room.Topic, room.Difficulty)
	}

	if avatar == "" {
		avatar = "default"
	}
	room.Players[userID] = &Player{ID: userID, HP: StartingHP, Avatar: avatar}

	var events []Event
	events = append(events, Event{
		Type:    "player_joined",
		RoomID:  room.ID,
		Topic:   room.Topic,
		Lang:    room.Lang,
		Players: room.playerListLocked(),
		HP:      room.hpMapLocked(),
		Avatars: room.avatarMapLocked(),
	})
	events = append(events, room.snapshotEventLocked())

	if len(room.Players) == MaxPlayers && !room.Started {
		// Стартуем первый раунд, когда в комнате два игрока.
		room.startRoundLocked()
		events = append(events, room.roundStartEventLocked())
	}

	return events, nil
}

// GetRoom returns the room if the user is a member
func (m *Manager) GetRoom(roomID, userID string) (*Room, error) {
	room, err := m.getRoom(roomID)
	if err != nil {
		return nil, ErrRoomNotFound
	}
	room.mu.Lock()
	defer room.mu.Unlock()
	if _, ok := room.Players[userID]; !ok {
		return nil, ErrNotInRoom
	}
	return room, nil
}

// GetRoomSnapshot возвращает текущее состояние комнаты для reconnect
func (m *Manager) GetRoomSnapshot(roomID, userID string) ([]Event, error) {
	room, err := m.getRoom(roomID)
	if err != nil {
		return nil, ErrRoomNotFound
	}
	if _, ok := room.Players[userID]; !ok {
		return nil, ErrNotInRoom
	}
	return room.snapshotEventsLocked(), nil
}

// SubmitAnswer проверяет ответ, применяет урон и
// возвращает события для рассылки.
func (m *Manager) SubmitAnswer(roomID, userID, answer string, speed int, roundToken int) ([]Event, error) {
	room, err := m.getRoom(roomID)
	if err != nil {
		return nil, err
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if !room.Started {
		return nil, ErrNotStarted
	}

	if roundToken != room.RoundToken {
		return nil, ErrRoundTokenMismatch
	}

	attacker, ok := room.Players[userID]
	if !ok {
		return nil, ErrNotInRoom
	}

	defender := room.otherPlayerLocked(userID)
	if defender == nil {
		return nil, ErrNotInRoom
	}

	// Проверяем ответ по всем допустимым вариантам
	correct := room.IsAnswerCorrect(answer)
	damage := ProcessAnswer(attacker, defender, correct, speed)

	correctCount, wrongCount := room.GetPlayerStats()

	selfDamage := 0
	if !correct && damage == 0 {
		selfDamage = SelfDamageOnWrong
	}

	events := []Event{{
		Type:       "update",
		RoomID:     room.ID,
		Round:      room.Round,
		RoundToken: room.RoundToken,
		HP:         room.hpMapLocked(),
		AttackerID: attacker.ID,
		DefenderID: defender.ID,
		Damage:     damage,
		SelfDamage: selfDamage,
		Correct:    correct,
		Speed:      speed,
	}}

	if defender.HP <= 0 {
		events = append(events, Event{
			Type:         "game_over",
			RoomID:       room.ID,
			DuelID:       room.DuelID,
			HP:           room.hpMapLocked(),
			Elo:          room.eloMapLocked(),
			WinnerID:     attacker.ID,
			CorrectCount: correctCount,
			WrongCount:   wrongCount,
			Reason:       "hp_zero",
		})
		return events, nil
	}

	if attacker.HP <= 0 {
		events = append(events, Event{
			Type:         "game_over",
			RoomID:       room.ID,
			DuelID:       room.DuelID,
			HP:           room.hpMapLocked(),
			Elo:          room.eloMapLocked(),
			WinnerID:     defender.ID,
			CorrectCount: correctCount,
			WrongCount:   wrongCount,
			Reason:       "hp_zero",
		})
		return events, nil
	}

	if correct {
		hasNext := room.HasMorePhrasesLocked()

		// Check if we need halftime (BEFORE starting next round - after round 10)
		needsHalftime := room.Round == HalfSize && room.Round < MaxRounds && hasNext

		log.Printf("HALFTIME CHECK: Round=%d HalfSize=%d MaxRounds=%d hasNext=%v needsHalftime=%v",
			room.Round, HalfSize, MaxRounds, hasNext, needsHalftime)

		if !hasNext {
			events = append(events, Event{
				Type:         "game_over",
				RoomID:       room.ID,
				DuelID:       room.DuelID,
				HP:           room.hpMapLocked(),
				Elo:          room.eloMapLocked(),
				WinnerID:     attacker.ID,
				CorrectCount: correctCount,
				WrongCount:   wrongCount,
				Reason:       "phrases_exhausted",
			})
			return events, nil
		}

		room.startRoundLocked()

		if needsHalftime {
			// Send halftime event instead of immediate round start
			events = append(events, Event{
				Type:          "halftime",
				RoomID:        room.ID,
				DuelID:        room.DuelID,
				Round:         room.Round, // now 11
				Half:          2,
				TotalPhrases:  room.GetTotalPhrasesLocked(),
				HP:            room.hpMapLocked(),
				Prompt:        "⏸ HALF TIME ⏸",
				CorrectAnswer: "",
			})
			return events, nil
		}

		events = append(events, room.roundStartEventLocked())
	}
	// Wrong answer: just apply self-damage, don't start new round
	// Round continues until both answer or timeout

	return events, nil
}

// RequestRematch marks a player as ready for rematch.
// When both players are ready, the room is reset and a new round starts.
func (m *Manager) RequestRematch(roomID, userID string) ([]Event, bool, error) {
	room, err := m.getRoom(roomID)
	if err != nil {
		return nil, false, err
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if len(room.Players) != MaxPlayers {
		return nil, false, errors.New("need 2 players for rematch")
	}
	if _, ok := room.Players[userID]; !ok {
		return nil, false, ErrNotInRoom
	}

	room.RematchReady[userID] = true

	readyCount := 0
	for _, ready := range room.RematchReady {
		if ready {
			readyCount++
		}
	}

	if readyCount < MaxPlayers {
		return nil, false, nil
	}

	room.Round = 0
	room.RoundToken = 0
	room.Started = false
	room.Prompt = ""
	room.Expected = ""
	room.ValidAnswers = []string{}
	room.AICurrent = 0
	room.LocalCurrent = 0
	room.RematchReady = make(map[string]bool)

	for _, p := range room.Players {
		p.HP = StartingHP
		p.CorrectCount = 0
		p.WrongCount = 0
	}

	room.startRoundLocked()

	events := []Event{room.snapshotEventLocked(), room.roundStartEventLocked()}
	return events, true, nil
}

// ResetRoom resets a finished room for a new match (rematch with same players).
func (m *Manager) ResetRoom(roomID string) ([]Event, error) {
	room, err := m.getRoom(roomID)
	if err != nil {
		return nil, err
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if len(room.Players) != MaxPlayers {
		return nil, errors.New("need 2 players for rematch")
	}

	room.Round = 0
	room.RoundToken = 0
	room.Started = false
	room.Prompt = ""
	room.Expected = ""
	room.ValidAnswers = []string{}
	room.AICurrent = 0
	room.LocalCurrent = 0

	for _, p := range room.Players {
		p.HP = StartingHP
		p.CorrectCount = 0
		p.WrongCount = 0
	}

	// Start first round immediately
	room.startRoundLocked()

	events := []Event{room.snapshotEventLocked(), room.roundStartEventLocked()}
	return events, nil
}

// Leave удаляет игрока из комнаты и возвращает событие для рассылки.
// Второй возвращаемый bool = true если другой игрок ожидал реванша.
func (m *Manager) Leave(roomID, userID string) ([]Event, bool, error) {
	if roomID == "" || userID == "" {
		return nil, false, nil
	}

	room, err := m.getRoom(roomID)
	if err != nil {
		return nil, false, nil
	}

	room.mu.Lock()
	_, exists := room.Players[userID]
	rematchWasPending := false
	if exists {
		// Check if any player had requested a rematch before removing
		if len(room.RematchReady) > 0 {
			rematchWasPending = true
		}
		delete(room.Players, userID)
		delete(room.RematchReady, userID)
		room.Started = false
		room.Prompt = ""
		room.Expected = ""
	}
	players := room.playerListLocked()
	hp := room.hpMapLocked()
	room.mu.Unlock()

	if !exists {
		return nil, false, nil
	}

	events := []Event{{
		Type:    "player_left",
		RoomID:  roomID,
		UserID:  userID,
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

	return events, rematchWasPending, nil
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

	hasNext := room.HasMorePhrasesLocked()

	// Check if we need halftime (BEFORE starting next round - after round 10)
	needsHalftime := room.Round == HalfSize && room.Round < MaxRounds && hasNext

	if !hasNext {
		winnerID := room.determineWinnerByHP()
		correctCount, wrongCount := room.GetPlayerStats()
		events = append(events, Event{
			Type:         "game_over",
			RoomID:       room.ID,
			DuelID:       room.DuelID,
			HP:           room.hpMapLocked(),
			Elo:          room.eloMapLocked(),
			WinnerID:     winnerID,
			CorrectCount: correctCount,
			WrongCount:   wrongCount,
			Reason:       "phrases_exhausted",
		})
		return events, nil
	}

	room.startRoundLocked()

	if needsHalftime {
		events = append(events, Event{
			Type:          "halftime",
			RoomID:        room.ID,
			DuelID:        room.DuelID,
			Round:         room.Round,
			Half:          2,
			TotalPhrases:  room.GetTotalPhrasesLocked(),
			HP:            room.hpMapLocked(),
			Prompt:        "⏸ HALF TIME ⏸",
			CorrectAnswer: "",
		})
		return events, nil
	}

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
		ID:           id,
		Players:      make(map[string]*Player),
		RematchReady: make(map[string]bool),
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

func (m *Manager) GetPlayerNames(roomID string) []string {
	room, err := m.getRoom(roomID)
	if err != nil {
		return nil
	}
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.playerListLocked()
}

// GetRoomSettings returns topic, lang, difficulty for a room.
func (m *Manager) GetRoomSettings(roomID string) (topic, lang, difficulty string, ok bool) {
	room, err := m.getRoom(roomID)
	if err != nil {
		return "", "", "", false
	}
	room.mu.Lock()
	defer room.mu.Unlock()
	return room.Topic, room.Lang, room.Difficulty, true
}

// SetAIPhrases устанавливает AI сгенерированные фразы для комнаты
func (m *Manager) SetAIPhrases(roomID string, phrases []AIPhraseData) bool {
	room, err := m.getRoom(roomID)
	if err != nil {
		return false
	}
	room.SetAIPhrases(phrases)
	return true
}

func (m *Manager) SetDuelID(roomID, duelID string) bool {
	room, err := m.getRoom(roomID)
	if err != nil {
		return false
	}
	room.SetDuelID(duelID)
	return true
}

// ContinueAfterHalftime starts the next round after halftime pause
func (m *Manager) ContinueAfterHalftime(roomID, userID string) ([]Event, error) {
	room, err := m.getRoom(roomID)
	if err != nil {
		return nil, ErrRoomNotFound
	}
	if _, ok := room.Players[userID]; !ok {
		return nil, ErrNotInRoom
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if !room.Started {
		return nil, ErrNotStarted
	}

	// Check if we're actually at the right point (round 11, waiting for halftime)
	if room.Round != HalfSize+1 {
		return nil, errors.New("not at halftime")
	}

	// Start next round
	room.startRoundLocked()

	events := []Event{room.roundStartEventLocked()}
	return events, nil
}

// normalize: trim + lowercase перед сравнением.
func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
