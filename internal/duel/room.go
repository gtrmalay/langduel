package duel

import (
	"sync"
	"langduel/internal/storage"
)

type Room struct {
	mu         sync.Mutex
	ID         string
	Players    map[string]*Player
	Round      int
	RoundToken int
	Topic      string
	Lang       string
	Prompt     string
	Expected   string
	Started    bool
}

var phrasesByTopic = storage.PhraseSets

func (r *Room) startRoundLocked() {
	phrases := phrasesByTopic[r.Topic]
	if len(phrases) == 0 {
		phrases = phrasesByTopic["default"]
	}
	if len(phrases) == 0 {
		r.Prompt = ""
		r.Expected = ""
		r.Started = true
		return
	}

	r.Round++
	r.RoundToken++
	idx := (r.Round - 1) % len(phrases)
	r.Prompt = phrases[idx].Prompt
	r.Expected = phrases[idx].Expected
	r.Started = true
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

func (r *Room) snapshotEventLocked() Event {
	return Event{
		Type:       "room_state",
		RoomID:     r.ID,
		Round:      r.Round,
		RoundToken: r.RoundToken,
		Topic:      r.Topic,
		Lang:       r.Lang,
		Prompt:     r.Prompt,
		Players:    r.playerListLocked(),
		HP:         r.hpMapLocked(),
	}
}

func (r *Room) roundStartEventLocked() Event {
	return Event{
		Type:       "round_start",
		RoomID:     r.ID,
		Round:      r.Round,
		RoundToken: r.RoundToken,
		Topic:      r.Topic,
		Lang:       r.Lang,
		Prompt:     r.Prompt,
		HP:         r.hpMapLocked(),
	}
}

func (r *Room) snapshotEventsLocked() []Event {
	events := []Event{r.snapshotEventLocked()}
	if r.Started {
		events = append(events, r.roundStartEventLocked())
	}
	return events
}
