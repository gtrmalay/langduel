package duel

import (
	"strings"
	"sync"

	"langduel/internal/storage"
)

type AIPhraseData struct {
	Prompt  string
	Answers []string
}

var fallbackPhrases = []struct {
	Prompt   string
	Expected string
}{
	{"hello", "привет"},
	{"world", "мир"},
	{"love", "любовь"},
	{"time", "время"},
	{"water", "вода"},
	{"fire", "огонь"},
	{"earth", "земля"},
	{"sun", "солнце"},
	{"moon", "луна"},
	{"star", "звезда"},
	{"cat", "кот"},
	{"dog", "собака"},
	{"bird", "птица"},
	{"fish", "рыба"},
	{"tree", "дерево"},
	{"book", "книга"},
	{"house", "дом"},
	{"car", "машина"},
	{"friend", "друг"},
	{"family", "семья"},
}

type Room struct {
	mu           sync.Mutex
	ID           string
	Players      map[string]*Player
	Round        int
	RoundToken   int
	Topic        string
	Difficulty   string
	Lang         string
	Prompt       string
	Expected     string
	ValidAnswers []string
	Started      bool
	DuelID       string
	AIPhrases    []AIPhraseData
	AICurrent    int
	LocalPhrases []storage.Phrase
	LocalCurrent int
}

func (r *Room) startRoundLocked() {
	r.Round++
	r.RoundToken++

	// 1. Try AI phrases first (highest priority)
	if len(r.AIPhrases) > 0 && r.AICurrent < len(r.AIPhrases) {
		phrase := r.AIPhrases[r.AICurrent]
		r.Prompt = phrase.Prompt
		r.ValidAnswers = phrase.Answers
		if len(phrase.Answers) > 0 {
			r.Expected = phrase.Answers[0]
		} else {
			r.Expected = ""
		}
		r.AICurrent++
		r.Started = true
		return
	}

	// 2. Try local topic packages (from local_test_phrases.go)
	if len(r.LocalPhrases) > 0 && r.LocalCurrent < len(r.LocalPhrases) {
		phrase := r.LocalPhrases[r.LocalCurrent]
		r.Prompt = phrase.Prompt
		r.Expected = phrase.Expected
		r.ValidAnswers = []string{phrase.Expected}
		r.LocalCurrent++
		r.Started = true
		return
	}

	// 3. Use fallback phrases (ultimate fallback - 20 basic words)
	if len(fallbackPhrases) > 0 {
		idx := r.Round % len(fallbackPhrases)
		phrase := fallbackPhrases[idx]
		r.Prompt = phrase.Prompt
		r.Expected = phrase.Expected
		r.ValidAnswers = []string{phrase.Expected}
	} else {
		r.Prompt = ""
		r.Expected = ""
		r.ValidAnswers = []string{}
	}
	r.Started = true
}

func (r *Room) HasMorePhrasesLocked() bool {
	if r.Round >= MaxRounds {
		return false
	}
	// Check AI phrases
	if len(r.AIPhrases) > 0 {
		return r.AICurrent < len(r.AIPhrases)
	}
	// Check local phrases
	if len(r.LocalPhrases) > 0 {
		return r.LocalCurrent < len(r.LocalPhrases)
	}
	// Fallback always has phrases
	return true
}

func (r *Room) GetTotalPhrasesLocked() int {
	// AI phrases have highest priority
	if len(r.AIPhrases) > 0 {
		return len(r.AIPhrases)
	}
	// Local topic packages
	if len(r.LocalPhrases) > 0 {
		return len(r.LocalPhrases)
	}
	// Fallback
	return MaxRounds
}

func (r *Room) GetPlayerStats() (map[string]int, map[string]int) {
	correct := make(map[string]int)
	wrong := make(map[string]int)
	for id, p := range r.Players {
		correct[id] = p.CorrectCount
		wrong[id] = p.WrongCount
	}
	return correct, wrong
}

func (r *Room) determineWinnerByHP() string {
	var winnerID string
	var maxHP int
	for id, p := range r.Players {
		if p.HP > maxHP {
			maxHP = p.HP
			winnerID = id
		}
	}
	return winnerID
}

func (r *Room) SetAIPhrases(phrases []AIPhraseData) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.AIPhrases = phrases
	r.AICurrent = 0
}

func (r *Room) SetLocalPhrases(topic, difficulty string) {
	phrases := storage.GetPhrases(topic, difficulty)
	if phrases != nil && len(phrases) > 0 {
		r.LocalPhrases = phrases
		r.LocalCurrent = 0
	}
}

func (r *Room) SetDuelID(duelID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.DuelID = duelID
}

func (r *Room) IsAnswerCorrect(answer string) bool {
	normalized := strings.ToLower(strings.TrimSpace(answer))
	for _, valid := range r.ValidAnswers {
		if strings.ToLower(strings.TrimSpace(valid)) == normalized {
			return true
		}
	}
	return false
}

func (r *Room) otherPlayerLocked(userID string) *Player {
	for id, p := range r.Players {
		if id != userID {
			return p
		}
	}
	return nil
}

func (r *Room) hpMapLocked() map[string]int {
	hp := make(map[string]int, len(r.Players))
	for id, p := range r.Players {
		hp[id] = p.HP
	}
	return hp
}

func (r *Room) playerListLocked() []string {
	players := make([]string, 0, len(r.Players))
	for id := range r.Players {
		players = append(players, id)
	}
	return players
}

func (r *Room) avatarMapLocked() map[string]string {
	avatars := make(map[string]string, len(r.Players))
	for id, p := range r.Players {
		avatars[id] = p.Avatar
	}
	return avatars
}

func (r *Room) eloMapLocked() map[string]int {
	elo := make(map[string]int, len(r.Players))
	for id, p := range r.Players {
		elo[id] = p.Elo
	}
	return elo
}

func (r *Room) snapshotEventLocked() Event {
	return Event{
		Type:       "room_state",
		RoomID:     r.ID,
		DuelID:     r.DuelID,
		Round:      r.Round,
		RoundToken: r.RoundToken,
		Topic:      r.Topic,
		Difficulty: r.Difficulty,
		Lang:       r.Lang,
		Prompt:     r.Prompt,
		Players:    r.playerListLocked(),
		HP:         r.hpMapLocked(),
	}
}

func (r *Room) roundStartEventLocked() Event {
	totalPhrases := len(r.AIPhrases)
	if totalPhrases == 0 {
		totalPhrases = MaxRounds
	}
	return Event{
		Type:          "round_start",
		RoomID:        r.ID,
		Round:         r.Round,
		RoundToken:    r.RoundToken,
		TotalPhrases:  totalPhrases,
		Topic:         r.Topic,
		Difficulty:    r.Difficulty,
		Lang:          r.Lang,
		Prompt:        r.Prompt,
		CorrectAnswer: r.Expected,
		HP:            r.hpMapLocked(),
	}
}

func (r *Room) snapshotEventsLocked() []Event {
	events := []Event{r.snapshotEventLocked()}
	if r.Started {
		events = append(events, r.roundStartEventLocked())
	}
	return events
}
