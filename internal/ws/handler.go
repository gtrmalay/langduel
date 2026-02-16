package ws

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"langduel/internal/duel"

	"github.com/gorilla/websocket"
)

// CheckOrigin: for MVP accept any origin. Tighten for production.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var hub = NewHub()
var mgr = duel.GlobalManager

const roundTimeout = 10 * time.Second

var roundTimers = struct {
	mu     sync.Mutex
	timers map[string]*time.Timer
}{
	timers: make(map[string]*time.Timer),
}

func init() {
	go hub.Run()
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
		hub:  hub,
	}

	go readPump(client)
	go writePump(client)
}

func readPump(c *Client) {
	defer func() {
		// При отключении убираем игрока и оповещаем комнату.
		if c.roomID != "" && c.userID != "" {
			if events, _ := mgr.Leave(c.roomID, c.userID); len(events) > 0 {
				for _, ev := range events {
					broadcastRoom(c.hub, c.roomID, ev)
				}
			}
			stopRoundTimer(c.roomID)
		}
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg duel.Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			sendError(c, "", "invalid json")
			continue
		}

		switch msg.Type {
		case "join":
			handleJoin(c, msg)
		case "answer":
			handleAnswer(c, msg)
		default:
			sendError(c, msg.RoomID, "unknown message type")
		}
	}
}

func writePump(c *Client) {
	for msg := range c.send {
		_ = c.conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func handleJoin(c *Client, msg duel.Message) {
	if c.roomID != "" {
		sendError(c, msg.RoomID, "already in room")
		return
	}

	events, err := mgr.Join(msg.RoomID, msg.UserID, msg.Topic, msg.Lang)
	if err != nil {
		sendError(c, msg.RoomID, err.Error())
		return
	}

	c.userID = msg.UserID
	c.roomID = msg.RoomID
	c.hub.register <- registration{client: c, roomID: msg.RoomID}

	for _, ev := range events {
		if ev.Type == "round_start" {
			scheduleRoundTimeout(ev.RoomID, ev.RoundToken)
		}
		broadcastRoom(c.hub, msg.RoomID, ev)
	}
}

func handleAnswer(c *Client, msg duel.Message) {
	if c.roomID == "" {
		sendError(c, msg.RoomID, "join room first")
		return
	}

	events, err := mgr.SubmitAnswer(c.roomID, c.userID, msg.Answer, msg.Speed)
	if err != nil {
		sendError(c, c.roomID, err.Error())
		return
	}

	for _, ev := range events {
		if ev.Type == "round_start" {
			scheduleRoundTimeout(ev.RoomID, ev.RoundToken)
		}
		if ev.Type == "game_over" {
			stopRoundTimer(ev.RoomID)
		}
		broadcastRoom(c.hub, c.roomID, ev)
	}
}

func broadcastRoom(h *Hub, roomID string, ev duel.Event) {
	payload, _ := json.Marshal(ev)
	h.broadcast <- broadcastMsg{roomID: roomID, data: payload}
}

func sendError(c *Client, roomID, msg string) {
	ev := duel.Event{
		Type:   "error",
		RoomID: roomID,
		Error:  msg,
	}
	payload, _ := json.Marshal(ev)
	select {
	case c.send <- payload:
	default:
	}
}

func scheduleRoundTimeout(roomID string, token int) {
	roundTimers.mu.Lock()
	if t := roundTimers.timers[roomID]; t != nil {
		t.Stop()
	}
	roundTimers.timers[roomID] = time.AfterFunc(roundTimeout, func() {
		events, err := mgr.RoundTimeout(roomID, token)
		if err != nil || len(events) == 0 {
			return
		}
		for _, ev := range events {
			broadcastRoom(hub, roomID, ev)
		}
	})
	roundTimers.mu.Unlock()
}

func stopRoundTimer(roomID string) {
	roundTimers.mu.Lock()
	if t := roundTimers.timers[roomID]; t != nil {
		t.Stop()
		delete(roundTimers.timers, roomID)
	}
	roundTimers.mu.Unlock()
}
