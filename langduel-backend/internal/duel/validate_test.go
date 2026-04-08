package duel

import "testing"

func TestValidateRoomID(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"abc", true},
		{"room-123", true},
		{"a", true},
		{"ab", true},                            // min 1 now
		{"", false},                             // empty
		{"-abc", false},                         // starts with hyphen
		{"abc-", false},                         // ends with hyphen
		{"ab--cd", true},                        // double hyphen is allowed by regex
		{"room 123", false},                     // space
		{"room.123", false},                     // dot
		{"room@123", false},                     // special char
		{"a" + string(make([]byte, 60)), false}, // too long
	}

	for _, tc := range tests {
		err := ValidateRoomID(tc.input)
		if tc.valid && err != nil {
			t.Errorf("ValidateRoomID(%q) expected valid, got %v", tc.input, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("ValidateRoomID(%q) expected error, got valid", tc.input)
		}
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"ab", true},
		{"User123", true},
		{"user_name", true},
		{"user-name", true},
		{"a", false},                            // too short
		{"", false},                             // empty
		{"123user", false},                      // starts with number
		{"user name", false},                    // space
		{"user@name", false},                    // special char
		{"_user", false},                        // starts with underscore
		{"-user", false},                        // starts with hyphen
		{"a" + string(make([]byte, 40)), false}, // too long
	}

	for _, tc := range tests {
		err := ValidateUsername(tc.input)
		if tc.valid && err != nil {
			t.Errorf("ValidateUsername(%q) expected valid, got %v", tc.input, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("ValidateUsername(%q) expected error, got valid", tc.input)
		}
	}
}

func TestValidateAnswer(t *testing.T) {
	if err := ValidateAnswer("hello"); err != nil {
		t.Errorf("short answer should be valid")
	}

	longAnswer := string(make([]byte, 250))
	if err := ValidateAnswer(longAnswer); err == nil {
		t.Errorf("answer over 200 chars should be invalid")
	}
}

func TestValidateDifficulty(t *testing.T) {
	if err := ValidateDifficulty("intermediate"); err != nil {
		t.Errorf("intermediate should be valid")
	}
	if err := ValidateDifficulty("invalid"); err == nil {
		t.Errorf("invalid difficulty should be rejected")
	}
}

func TestValidateAvatar(t *testing.T) {
	if err := ValidateAvatar("knight"); err != nil {
		t.Errorf("knight should be valid")
	}
	if err := ValidateAvatar("hacker"); err == nil {
		t.Errorf("invalid avatar should be rejected")
	}
}

func TestValidateLang(t *testing.T) {
	if err := ValidateLang("en"); err != nil {
		t.Errorf("en should be valid")
	}
	if err := ValidateLang("ru"); err != nil {
		t.Errorf("ru should be valid")
	}
	if err := ValidateLang("fr"); err == nil {
		t.Errorf("fr should be invalid")
	}
}

func TestSanitizeAnswer(t *testing.T) {
	trimmed := SanitizeAnswer("  hello  ")
	if trimmed != "hello" {
		t.Errorf("expected 'hello', got %q", trimmed)
	}

	trimmed = SanitizeAnswer("  hello world  ")
	if trimmed != "hello world" {
		t.Errorf("expected 'hello world', got %q", trimmed)
	}
}
