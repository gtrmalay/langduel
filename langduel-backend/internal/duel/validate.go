package duel

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	MinRoomIDLen   = 1
	MaxRoomIDLen   = 50
	MinUsernameLen = 2
	MaxUsernameLen = 30
	MaxAnswerLen   = 200
	MaxTopicLen    = 100
)

var (
	ErrRoomIDTooShort    = errors.New("room_id: minimum 1 character")
	ErrRoomIDTooLong     = errors.New("room_id: maximum 50 characters")
	ErrRoomIDInvalid     = errors.New("room_id: only letters, numbers and hyphens allowed")
	ErrUsernameTooShort  = errors.New("username: minimum 2 characters")
	ErrUsernameTooLong   = errors.New("username: maximum 30 characters")
	ErrUsernameInvalid   = errors.New("username: only letters, numbers, underscores and hyphens allowed")
	ErrAnswerTooLong     = errors.New("answer: maximum 200 characters")
	ErrTopicTooLong      = errors.New("topic: maximum 100 characters")
	ErrTopicInvalid      = errors.New("topic: contains invalid characters")
	ErrDifficultyInvalid = errors.New("difficulty: invalid value")
	ErrAvatarInvalid     = errors.New("avatar: invalid value")
	ErrLangInvalid       = errors.New("lang: invalid value")
)

var (
	roomIDPattern   = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)
	usernamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]*$`)
)

var validDifficulties = map[string]bool{
	"beginner":     true,
	"intermediate": true,
	"advanced":     true,
}

var validAvatars = map[string]bool{
	"default":   true,
	"knight":    true,
	"wizard":    true,
	"archer":    true,
	"dragon":    true,
	"skull":     true,
	"fire":      true,
	"ice":       true,
	"lightning": true,
	"sword":     true,
	"shield":    true,
	"potion":    true,
	"crown":     true,
	"star":      true,
	"moon":      true,
	"guest":     true,
}

var validLangs = map[string]bool{
	"en":    true,
	"ru":    true,
	"en-ru": true,
	"ru-en": true,
}

func ValidateRoomID(roomID string) error {
	if len(roomID) < MinRoomIDLen {
		return ErrRoomIDTooShort
	}
	if len(roomID) > MaxRoomIDLen {
		return ErrRoomIDTooLong
	}
	if !roomIDPattern.MatchString(roomID) {
		return ErrRoomIDInvalid
	}
	return nil
}

func ValidateUsername(username string) error {
	if utf8.RuneCountInString(username) < MinUsernameLen {
		return ErrUsernameTooShort
	}
	if utf8.RuneCountInString(username) > MaxUsernameLen {
		return ErrUsernameTooLong
	}
	if !usernamePattern.MatchString(username) {
		return ErrUsernameInvalid
	}
	return nil
}

func ValidateAnswer(answer string) error {
	if len(answer) > MaxAnswerLen {
		return ErrAnswerTooLong
	}
	return nil
}

func ValidateTopic(topic string) error {
	if topic == "" || topic == "custom" {
		return nil
	}
	if utf8.RuneCountInString(topic) > MaxTopicLen {
		return ErrTopicTooLong
	}
	sanitized := sanitizeTopic(topic)
	if sanitized != topic {
		return ErrTopicInvalid
	}
	return nil
}

func ValidateCustomTopic(topic string) error {
	if utf8.RuneCountInString(topic) > MaxTopicLen {
		return ErrTopicTooLong
	}
	sanitized := sanitizeTopic(topic)
	if sanitized == "" {
		return ErrTopicInvalid
	}
	return nil
}

func ValidateDifficulty(difficulty string) error {
	if difficulty == "" {
		return nil
	}
	if !validDifficulties[difficulty] {
		return ErrDifficultyInvalid
	}
	return nil
}

func ValidateAvatar(avatar string) error {
	if avatar == "" {
		return nil
	}
	if !validAvatars[avatar] {
		return ErrAvatarInvalid
	}
	return nil
}

func ValidateLang(lang string) error {
	if lang == "" {
		return nil
	}
	if !validLangs[lang] {
		return ErrLangInvalid
	}
	return nil
}

func SanitizeAnswer(answer string) string {
	trimmed := strings.TrimSpace(answer)
	if len(trimmed) > MaxAnswerLen {
		trimmed = trimmed[:MaxAnswerLen]
	}
	return trimmed
}

func SanitizeUsername(username string) string {
	return strings.TrimSpace(username)
}

func sanitizeTopic(topic string) string {
	return strings.TrimSpace(topic)
}
