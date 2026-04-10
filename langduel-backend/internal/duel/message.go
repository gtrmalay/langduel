package duel

type Message struct {
	Type       string `json:"type"`
	RoomID     string `json:"room_id"`
	UserID     string `json:"user_id"`
	Topic      string `json:"topic,omitempty"`
	Difficulty string `json:"difficulty,omitempty"`
	Lang       string `json:"lang,omitempty"`
	Answer     string `json:"answer"`
	Speed      int    `json:"speed"`
	Avatar     string `json:"avatar,omitempty"`
	Ts         int64  `json:"ts,omitempty"`
}
