package duel

type Message struct {
	Type   string `json:"type"`
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
	Topic  string `json:"topic,omitempty"`
	Lang   string `json:"lang,omitempty"`
	Answer string `json:"answer"`
	Speed  int    `json:"speed"`
}
