package duel

import "testing"

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
