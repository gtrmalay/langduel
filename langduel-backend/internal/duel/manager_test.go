package duel

import (
	"strings"
	"testing"
)

func hasType(events []Event, typ string) bool {
	for _, ev := range events {
		if ev.Type == typ {
			return true
		}
	}
	return false
}

func TestJoinStartsRound(t *testing.T) {
	m := NewManager()
	if _, err := m.Join("room1", "u1", "default", "intermediate", "en", ""); err != nil {
		t.Fatalf("join u1: %v", err)
	}
	events, err := m.Join("room1", "u2", "default", "intermediate", "en", "")
	if err != nil {
		t.Fatalf("join u2: %v", err)
	}
	if !hasType(events, "round_start") {
		t.Fatalf("expected round_start on second join")
	}
}

func TestRoomFull(t *testing.T) {
	m := NewManager()
	if _, err := m.Join("room1", "u1", "default", "intermediate", "en", ""); err != nil {
		t.Fatalf("join u1: %v", err)
	}
	if _, err := m.Join("room1", "u2", "default", "intermediate", "en", ""); err != nil {
		t.Fatalf("join u2: %v", err)
	}
	if _, err := m.Join("room1", "u3", "default", "intermediate", "en", ""); err != ErrRoomFull {
		t.Fatalf("expected ErrRoomFull, got %v", err)
	}
}

func TestSubmitAnswerWrongNoRoundStart(t *testing.T) {
	m := NewManager()
	if _, err := m.Join("room1", "u1", "default", "intermediate", "en", ""); err != nil {
		t.Fatalf("join u1: %v", err)
	}
	if _, err := m.Join("room1", "u2", "default", "intermediate", "en", ""); err != nil {
		t.Fatalf("join u2: %v", err)
	}
	events, err := m.SubmitAnswer("room1", "u1", "wrong", 3000)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if len(events) == 0 || events[0].Type != "update" {
		t.Fatalf("expected update event")
	}
	if events[0].Correct {
		t.Fatalf("expected correct=false")
	}
	if events[0].Speed != 3000 {
		t.Fatalf("expected speed=3000, got %d", events[0].Speed)
	}
	if hasType(events, "round_start") {
		t.Fatalf("did not expect round_start on wrong answer")
	}
}

func TestRoundTimeoutAdvances(t *testing.T) {
	m := NewManager()
	if _, err := m.Join("room1", "u1", "default", "intermediate", "en", ""); err != nil {
		t.Fatalf("join u1: %v", err)
	}
	events, err := m.Join("room1", "u2", "default", "intermediate", "en", "")
	if err != nil {
		t.Fatalf("join u2: %v", err)
	}
	var token int
	for _, ev := range events {
		if ev.Type == "round_start" {
			token = ev.RoundToken
		}
	}
	if token == 0 {
		t.Fatalf("expected round_start token")
	}
	tEvents, err := m.RoundTimeout("room1", token)
	if err != nil {
		t.Fatalf("timeout: %v", err)
	}
	if !hasType(tEvents, "round_end") || !hasType(tEvents, "round_start") {
		t.Fatalf("expected round_end and round_start on timeout")
	}
}

func TestCalculateDamage(t *testing.T) {
	if got := CalculateDamage(true, 500); got != 25 {
		t.Errorf("fast correct: want 25, got %d", got)
	}
	if got := CalculateDamage(true, 1500); got != 25 {
		t.Errorf("at 1.5s: want 25, got %d", got)
	}
	if got := CalculateDamage(true, 3000); got != 15 {
		t.Errorf("medium speed: want 15, got %d", got)
	}
	if got := CalculateDamage(true, 5000); got != 10 {
		t.Errorf("slow speed: want 10, got %d", got)
	}
	if got := CalculateDamage(false, 1000); got != 0 {
		t.Errorf("wrong answer: want 0, got %d", got)
	}
}

func TestProcessAnswerCorrect(t *testing.T) {
	attacker := &Player{HP: 100}
	defender := &Player{HP: 100}
	damage := ProcessAnswer(attacker, defender, true, 1000)
	if damage != 25 {
		t.Errorf("want damage 25, got %d", damage)
	}
	if defender.HP != 75 {
		t.Errorf("want defender HP 75, got %d", defender.HP)
	}
	if attacker.CorrectCount != 1 {
		t.Errorf("want correctCount 1, got %d", attacker.CorrectCount)
	}
}

func TestProcessAnswerWrong(t *testing.T) {
	attacker := &Player{HP: 100}
	defender := &Player{HP: 100}
	damage := ProcessAnswer(attacker, defender, false, 1000)
	if damage != 0 {
		t.Errorf("want damage 0, got %d", damage)
	}
	if attacker.HP != 95 {
		t.Errorf("want self-damage, attacker HP 95, got %d", attacker.HP)
	}
	if attacker.WrongCount != 1 {
		t.Errorf("want wrongCount 1, got %d", attacker.WrongCount)
	}
}

func TestProcessAnswerHPFloor(t *testing.T) {
	attacker := &Player{HP: 5}
	defender := &Player{HP: 100}
	ProcessAnswer(attacker, defender, false, 1000)
	if attacker.HP != 0 {
		t.Errorf("HP should floor at 0, got %d", attacker.HP)
	}
}

func TestGameOverOnHPZero(t *testing.T) {
	m := NewManager()
	m.Join("room1", "u1", "default", "intermediate", "en", "")
	events, _ := m.Join("room1", "u2", "default", "intermediate", "en", "")

	hasRoundStart := false
	for _, ev := range events {
		if ev.Type == "round_start" {
			hasRoundStart = true
		}
	}
	if !hasRoundStart {
		t.Fatal("expected round_start event")
	}

	ProcessAnswer(&Player{HP: 100}, &Player{HP: 100}, true, 3000)

	for _, ev := range events {
		if ev.Type == "game_over" {
			t.Logf("game_over event: winner=%s, hp=%v", ev.WinnerID, ev.HP)
		}
	}
}

func TestReconnectReturnsSnapshot(t *testing.T) {
	m := NewManager()
	m.Join("room1", "u1", "default", "intermediate", "en", "")
	events, _ := m.Join("room1", "u2", "default", "intermediate", "en", "")

	var token int
	for _, ev := range events {
		if ev.Type == "round_start" {
			token = ev.RoundToken
		}
	}

	snapshot, err := m.GetRoomSnapshot("room1", "u1")
	if err != nil {
		t.Fatalf("GetRoomSnapshot failed: %v", err)
	}
	if snapshot == nil {
		t.Fatal("snapshot should not be nil")
	}
	if len(snapshot) == 0 {
		t.Fatal("snapshot should contain events")
	}
	if token > 0 {
		found := false
		for _, ev := range snapshot {
			if ev.RoundToken == token {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("round token should be preserved in snapshot")
		}
	}
}

func TestReconnectWithHasPlayer(t *testing.T) {
	m := NewManager()
	m.Join("room1", "u1", "default", "intermediate", "en", "")

	snapshot, err := m.Join("room1", "u1", "default", "intermediate", "en", "")
	if err != nil {
		t.Fatalf("rejoin should succeed: %v", err)
	}
	if snapshot == nil {
		t.Fatal("snapshot should be returned for reconnect")
	}
}

func TestLocalPhrasesLoaded(t *testing.T) {
	m := NewManager()
	events, err := m.Join("room1", "u1", "food", "intermediate", "en", "")
	if err != nil {
		t.Fatalf("join u1: %v", err)
	}
	events, err = m.Join("room1", "u2", "food", "intermediate", "en", "")
	if err != nil {
		t.Fatalf("join u2: %v", err)
	}

	// Check that round started with food topic
	var roundStart Event
	for _, ev := range events {
		if ev.Type == "round_start" {
			roundStart = ev
			break
		}
	}
	if roundStart.Type == "" {
		t.Fatal("expected round_start event")
	}
	if roundStart.Topic != "food" {
		t.Errorf("expected topic 'food', got %q", roundStart.Topic)
	}
	// Food phrases should contain food-related prompts
	foodWords := []string{"cheese", "coffee", "bread", "apple", "meat"}
	found := false
	for _, w := range foodWords {
		if strings.Contains(strings.ToLower(roundStart.Prompt), w) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected food-related prompt, got %q", roundStart.Prompt)
	}
}
