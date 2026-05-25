package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"langduel/internal/duel"
	"langduel/internal/storage"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

func difficultyToInt(d string) int {
	switch d {
	case "beginner":
		return 1
	case "intermediate":
		return 2
	case "advanced":
		return 3
	default:
		return 2 // default to intermediate
	}
}

// CheckOrigin: for MVP accept any origin. Tighten for production.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var hub = NewHub()
var mgr = duel.GlobalManager
var repo *storage.DuelRepo
var roomDuelID = struct {
	mu     sync.Mutex
	byRoom map[string]string
}{
	byRoom: make(map[string]string),
}
var roomParticipants = struct {
	mu         sync.Mutex
	byRoomUser map[string]map[string]string
}{
	byRoomUser: make(map[string]map[string]string),
}
var roomUserID = struct {
	mu         sync.Mutex
	byRoomUser map[string]map[string]string
}{
	byRoomUser: make(map[string]map[string]string),
}
var roomRoundID = struct {
	mu          sync.Mutex
	byRoomRound map[string]map[int]string
}{
	byRoomRound: make(map[string]map[int]string),
}

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

// SetRepo allows wiring the storage layer from main.
func SetRepo(r *storage.DuelRepo) {
	repo = r
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	// Optional JWT auth via query param: ?token=...
	authUserID, username, ok := parseUserFromToken(r)
	if !ok {
		// Allow guest for now if no token.
		authUserID = ""
		username = ""
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
		hub:  hub,
	}
	if authUserID != "" {
		client.authUserID = authUserID
		client.displayName = username
	}

	go readPump(client)
	go writePump(client)
}

func readPump(c *Client) {
	defer func() {
		// При отключении убираем игрока и оповещаем комнату.
		if c.roomID != "" && c.displayName != "" {
			events, rematchWasPending, _ := mgr.Leave(c.roomID, c.displayName)
			for _, ev := range events {
				broadcastRoom(c.hub, c.roomID, ev)
				if ev.Type == "game_over" {
					processGameOverDB(&ev)
				}
			}
			if rematchWasPending {
				broadcastRoom(c.hub, c.roomID, duel.Event{
					Type:   "rematch_cancelled",
					RoomID: c.roomID,
					Reason: "opponent_left",
				})
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

		if msg.Type == "answer" || msg.Type == "join" {
			if !c.checkRateLimit() {
				sendError(c, msg.RoomID, "rate limit exceeded")
				continue
			}
		}

		switch msg.Type {
		case "join":
			// If JWT provided, trust it over client-sent user_id.
			if c.displayName != "" {
				msg.UserID = c.displayName
			}
			log.Printf("WS join: room=%s user=%s", msg.RoomID, msg.UserID)
			handleJoin(c, msg)
		case "answer":
			if c.displayName != "" {
				msg.UserID = c.displayName
			}
			log.Printf("WS answer: room=%s user=%s", msg.RoomID, msg.UserID)
			handleAnswer(c, msg)
		case "ping":
			// Respond with pong
			pong := duel.Message{
				Type: "pong",
				Ts:   msg.Ts,
			}
			data, _ := json.Marshal(pong)
			select {
			case c.send <- data:
			default:
			}
		case "next_round":
			// Player requesting to continue after halftime
			log.Printf("WS next_round: room=%s user=%s", msg.RoomID, msg.UserID)
			handleNextRound(c, msg)
		case "rematch":
			// Player requesting a rematch in the same room
			log.Printf("WS rematch: room=%s user=%s", msg.RoomID, msg.UserID)
			handleRematch(c, msg)
		case "leave":
			// Player leaving the room
			log.Printf("WS leave: room=%s user=%s", msg.RoomID, msg.UserID)
			handleLeave(c, msg)
		default:
			sendError(c, msg.RoomID, "unknown message type")
		}
	}
}

func parseUserFromToken(r *http.Request) (string, string, bool) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		return "", "", false
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", "", false
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", false
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", "", false
	}
	usr, _ := claims["usr"].(string)
	if usr == "" {
		usr = sub
	}
	return sub, usr, true
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

	if err := duel.ValidateRoomID(msg.RoomID); err != nil {
		sendError(c, msg.RoomID, err.Error())
		return
	}

	if err := duel.ValidateUsername(msg.UserID); err != nil {
		sendError(c, msg.RoomID, err.Error())
		return
	}

	if err := duel.ValidateDifficulty(msg.Difficulty); err != nil {
		sendError(c, msg.RoomID, err.Error())
		return
	}

	if err := duel.ValidateAvatar(msg.Avatar); err != nil {
		sendError(c, msg.RoomID, err.Error())
		return
	}

	if err := duel.ValidateLang(msg.Lang); err != nil {
		sendError(c, msg.RoomID, err.Error())
		return
	}

	if msg.Topic != "" && msg.Topic != "custom" {
		if err := duel.ValidateTopic(msg.Topic); err != nil {
			sendError(c, msg.RoomID, err.Error())
			return
		}
	}

	if c.displayName != "" {
		msg.UserID = c.displayName
	}

	// Check if player already in room (might be reconnect)
	if mgr.HasPlayer(msg.RoomID, msg.UserID) {
		// Return current room state for reconnect
		events, err := mgr.GetRoomSnapshot(msg.RoomID, msg.UserID)
		if err != nil {
			sendError(c, msg.RoomID, err.Error())
			return
		}
		c.displayName = msg.UserID
		c.roomID = msg.RoomID
		c.hub.register <- registration{client: c, roomID: msg.RoomID}
		for _, ev := range events {
			_ = c.conn.WriteJSON(ev)
		}
		return
	}

	// Optional DB persistence: create guest user and duel if needed.
	if repo != nil {
		ctx := context.Background()
		const guestTTLHours = 72
		var user *storage.User
		var err error
		if c.authUserID != "" {
			// JWT path: user_id already exists in DB
			user, err = repo.GetUserByID(ctx, c.authUserID)
			if err != nil {
				log.Printf("DB GetUserByID error: %v", err)
				sendError(c, msg.RoomID, "db error")
				return
			}
		} else {
			// Guest path
			user, err = repo.GetUserByUsername(ctx, msg.UserID)
			if err != nil {
				if err != storage.ErrNotFound {
					log.Printf("DB GetUserByUsername error: %v", err)
					sendError(c, msg.RoomID, "db error")
					return
				}
				user, err = repo.CreateGuestUser(ctx, msg.UserID, guestTTLHours)
				if err != nil {
					log.Printf("DB CreateGuestUser error: %v", err)
					sendError(c, msg.RoomID, "failed to create user")
					return
				}
			}
		}
		d, err := repo.GetDuelByRoomCode(ctx, msg.RoomID)
		if err != nil {
			if err != storage.ErrNotFound {
				log.Printf("DB GetDuelByRoomCode error: %v", err)
				sendError(c, msg.RoomID, "db error")
				return
			}
			difficulty := difficultyToInt(msg.Difficulty)
			d, err = repo.CreateDuel(ctx, msg.RoomID, user.ID, msg.Topic, difficulty, msg.Lang, "ru")
			if err != nil {
				log.Printf("DB CreateDuel error: %v", err)
				sendError(c, msg.RoomID, "failed to create duel")
				return
			}
		}

		// store duel_id by room
		roomDuelID.mu.Lock()
		roomDuelID.byRoom[msg.RoomID] = d.ID
		roomDuelID.mu.Unlock()

		// create participant
		playerOrder := 1
		roomParticipants.mu.Lock()
		if roomParticipants.byRoomUser[msg.RoomID] == nil {
			roomParticipants.byRoomUser[msg.RoomID] = make(map[string]string)
		}
		if len(roomParticipants.byRoomUser[msg.RoomID]) == 1 {
			playerOrder = 2
		}
		roomParticipants.mu.Unlock()

		p, err := repo.EnsureParticipant(ctx, d.ID, user.ID, playerOrder)
		if err != nil {
			log.Printf("DB EnsureParticipant error: %v", err)
			sendError(c, msg.RoomID, "failed to create participant")
			return
		}

		roomParticipants.mu.Lock()
		roomParticipants.byRoomUser[msg.RoomID][msg.UserID] = p.ID
		roomParticipants.mu.Unlock()

		roomUserID.mu.Lock()
		if roomUserID.byRoomUser[msg.RoomID] == nil {
			roomUserID.byRoomUser[msg.RoomID] = make(map[string]string)
		}
		roomUserID.byRoomUser[msg.RoomID][msg.UserID] = user.ID
		roomUserID.mu.Unlock()
	}

	// Load AI phrases BEFORE joining the room
	var aiPhrases []duel.AIPhraseData
	if repo != nil {
		ctx := context.Background()
		log.Printf("Looking for AI phrases for room: %s", msg.RoomID)
		phrases, err := repo.GetAIPhrases(ctx, "", msg.RoomID)
		if err != nil {
			log.Printf("GetAIPhrases error: %v", err)
		} else if len(phrases) == 0 {
			log.Printf("No AI phrases found for room %s", msg.RoomID)
		} else {
			log.Printf("Found %d AI phrases for room %s", len(phrases), msg.RoomID)
			aiPhrases = make([]duel.AIPhraseData, len(phrases))
			for i, p := range phrases {
				aiPhrases[i] = duel.AIPhraseData{
					Prompt:  p.Prompt,
					Answers: p.Answers,
				}
			}
		}
	}

	events, err := mgr.Join(msg.RoomID, msg.UserID, msg.Topic, msg.Difficulty, msg.Lang, msg.Avatar)
	if err != nil {
		sendError(c, msg.RoomID, err.Error())
		return
	}

	// Set AI phrases after joining
	if len(aiPhrases) > 0 {
		mgr.SetAIPhrases(msg.RoomID, aiPhrases)
		log.Printf("Set %d AI phrases for room %s", len(aiPhrases), msg.RoomID)
	}

	// set duel_id in room (must be after mgr.Join creates the room)
	if repo != nil {
		roomDuelID.mu.Lock()
		duelID := roomDuelID.byRoom[msg.RoomID]
		roomDuelID.mu.Unlock()
		if duelID != "" {
			mgr.SetDuelID(msg.RoomID, duelID)
		}
	}

	c.displayName = msg.UserID
	c.roomID = msg.RoomID
	c.hub.register <- registration{client: c, roomID: msg.RoomID}

	for _, ev := range events {
		if ev.Type == "round_start" {
			if repo != nil {
				ctx := context.Background()
				roomDuelID.mu.Lock()
				duelID := roomDuelID.byRoom[ev.RoomID]
				roomDuelID.mu.Unlock()
				if duelID != "" {
					log.Printf("DB round_start: room=%s duel_id=%s round=%d", ev.RoomID, duelID, ev.Round)
					if err := repo.MarkDuelStarted(ctx, duelID); err != nil {
						log.Printf("DB MarkDuelStarted error: %v", err)
					}
					rnd, err := repo.CreateRound(ctx, duelID, ev.Round, ev.Prompt, ev.CorrectAnswer, ev.Lang, ev.Topic, int(roundTimeout/time.Millisecond), ev.ValidAnswers)
					if err != nil {
						log.Printf("DB CreateRound error: %v", err)
					} else {
						roomRoundID.mu.Lock()
						if roomRoundID.byRoomRound[ev.RoomID] == nil {
							roomRoundID.byRoomRound[ev.RoomID] = make(map[int]string)
						}
						roomRoundID.byRoomRound[ev.RoomID][ev.Round] = rnd.ID
						roomRoundID.mu.Unlock()
					}
				}
			}
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

	if err := duel.ValidateAnswer(msg.Answer); err != nil {
		sendError(c, c.roomID, err.Error())
		return
	}

	if c.displayName != "" {
		msg.UserID = c.displayName
	}

	sanitizedAnswer := duel.SanitizeAnswer(msg.Answer)
	events, err := mgr.SubmitAnswer(c.roomID, msg.UserID, sanitizedAnswer, msg.Speed, msg.RoundToken)
	if err != nil {
		sendError(c, c.roomID, err.Error())
		return
	}

	for _, ev := range events {
		if ev.Type == "round_start" {
			if repo != nil {
				ctx := context.Background()
				roomDuelID.mu.Lock()
				duelID := roomDuelID.byRoom[ev.RoomID]
				roomDuelID.mu.Unlock()
				if duelID != "" {
					log.Printf("DB round_start: room=%s duel_id=%s round=%d", ev.RoomID, duelID, ev.Round)
					if err := repo.MarkDuelStarted(ctx, duelID); err != nil {
						log.Printf("DB MarkDuelStarted error: %v", err)
					}
					rnd, err := repo.CreateRound(ctx, duelID, ev.Round, ev.Prompt, ev.CorrectAnswer, ev.Lang, ev.Topic, int(roundTimeout/time.Millisecond), ev.ValidAnswers)
					if err != nil {
						log.Printf("DB CreateRound error: %v", err)
					} else {
						roomRoundID.mu.Lock()
						if roomRoundID.byRoomRound[ev.RoomID] == nil {
							roomRoundID.byRoomRound[ev.RoomID] = make(map[int]string)
						}
						roomRoundID.byRoomRound[ev.RoomID][ev.Round] = rnd.ID
						roomRoundID.mu.Unlock()
					}
				}
			}
			scheduleRoundTimeout(ev.RoomID, ev.RoundToken)
		}
		if ev.Type == "game_over" {
			processGameOverDB(&ev)
			stopRoundTimer(ev.RoomID)
			cleanupRoomState(ev.RoomID)
		}
		if ev.Type == "update" && repo != nil {
			ctx := context.Background()

			// --- resolve roundID ---
			roomRoundID.mu.Lock()
			roundID := ""
			if m := roomRoundID.byRoomRound[ev.RoomID]; m != nil {
				roundID = m[ev.Round]
			}
			roomRoundID.mu.Unlock()
			if roundID == "" {
				// cache miss — query DB directly
				roomDuelID.mu.Lock()
				duelID := roomDuelID.byRoom[ev.RoomID]
				roomDuelID.mu.Unlock()
				if duelID != "" {
					if id, err := repo.GetRoundID(ctx, duelID, ev.Round); err == nil {
						roundID = id
						// warm the cache
						roomRoundID.mu.Lock()
						if roomRoundID.byRoomRound[ev.RoomID] == nil {
							roomRoundID.byRoomRound[ev.RoomID] = make(map[int]string)
						}
						roomRoundID.byRoomRound[ev.RoomID][ev.Round] = roundID
						roomRoundID.mu.Unlock()
					} else {
						log.Printf("DB GetRoundID failed room=%s round=%d duel=%s: %v", ev.RoomID, ev.Round, duelID, err)
					}
				}
			}

			// --- resolve participantID ---
			roomParticipants.mu.Lock()
			participantID := ""
			if m := roomParticipants.byRoomUser[ev.RoomID]; m != nil {
				participantID = m[ev.AttackerID]
			}
			roomParticipants.mu.Unlock()
			if participantID == "" {
				// cache miss — look up via roomUserID → DB
				roomDuelID.mu.Lock()
				duelID := roomDuelID.byRoom[ev.RoomID]
				roomDuelID.mu.Unlock()
				roomUserID.mu.Lock()
				dbUserID := ""
				if m := roomUserID.byRoomUser[ev.RoomID]; m != nil {
					dbUserID = m[ev.AttackerID]
				}
				roomUserID.mu.Unlock()
				if duelID != "" && dbUserID != "" {
					if id, err := repo.GetParticipantID(ctx, duelID, dbUserID); err == nil {
						participantID = id
						// warm the cache
						roomParticipants.mu.Lock()
						if roomParticipants.byRoomUser[ev.RoomID] == nil {
							roomParticipants.byRoomUser[ev.RoomID] = make(map[string]string)
						}
						roomParticipants.byRoomUser[ev.RoomID][ev.AttackerID] = participantID
						roomParticipants.mu.Unlock()
					} else {
						log.Printf("DB GetParticipantID failed room=%s user=%s duel=%s: %v", ev.RoomID, dbUserID, duelID, err)
					}
				} else {
					log.Printf("DB CreateAnswer skipped: participantID empty for room=%s attacker=%s (duelID=%q dbUserID=%q)", ev.RoomID, ev.AttackerID, duelID, dbUserID)
				}
			}

			if roundID != "" && participantID != "" {
				err := repo.CreateAnswer(ctx, roundID, participantID, sanitizedAnswer, ev.Correct, msg.Speed, ev.Damage)
				if err != nil {
					log.Printf("DB CreateAnswer error: %v", err)
				} else {
					log.Printf("DB CreateAnswer ok: room=%s round=%d participant=%s correct=%v", ev.RoomID, ev.Round, participantID, ev.Correct)
				}
			} else {
				log.Printf("DB CreateAnswer skipped: roundID=%q participantID=%q room=%s round=%d attacker=%s", roundID, participantID, ev.RoomID, ev.Round, ev.AttackerID)
			}
		}
		broadcastRoom(c.hub, c.roomID, ev)
	}
}

func handleNextRound(c *Client, msg duel.Message) {
	if c.roomID == "" {
		sendError(c, msg.RoomID, "join room first")
		return
	}

	if c.displayName != "" {
		msg.UserID = c.displayName
	}

	events, err := mgr.ContinueAfterHalftime(c.roomID, msg.UserID)
	if err != nil {
		sendError(c, c.roomID, err.Error())
		return
	}

	for _, ev := range events {
		if ev.Type == "round_start" {
			if repo != nil {
				ctx := context.Background()
				roomDuelID.mu.Lock()
				duelID := roomDuelID.byRoom[ev.RoomID]
				roomDuelID.mu.Unlock()
				if duelID != "" {
					log.Printf("DB round_start (halftime): room=%s duel_id=%s round=%d", ev.RoomID, duelID, ev.Round)
					rnd, err := repo.CreateRound(ctx, duelID, ev.Round, ev.Prompt, ev.CorrectAnswer, ev.Lang, ev.Topic, int(roundTimeout/time.Millisecond), ev.ValidAnswers)
					if err != nil {
						log.Printf("DB CreateRound error: %v", err)
					} else {
						roomRoundID.mu.Lock()
						if roomRoundID.byRoomRound[ev.RoomID] == nil {
							roomRoundID.byRoomRound[ev.RoomID] = make(map[int]string)
						}
						roomRoundID.byRoomRound[ev.RoomID][ev.Round] = rnd.ID
						roomRoundID.mu.Unlock()
					}
				}
			}
			scheduleRoundTimeout(ev.RoomID, ev.RoundToken)
		}
		broadcastRoom(c.hub, c.roomID, ev)
	}
}

func handleLeave(c *Client, msg duel.Message) {
	if c.roomID == "" {
		return
	}

	roomID := c.roomID
	userID := c.displayName
	if userID == "" {
		userID = msg.UserID
	}

	log.Printf("handleLeave: room=%s user=%s", roomID, userID)

	// Delete pending duel if exists
	if repo != nil {
		if err := repo.DeletePendingDuel(context.Background(), roomID); err != nil {
			log.Printf("DB DeletePendingDuel error: %v", err)
		} else {
			log.Printf("DB Deleted pending duel for room: %s", roomID)
		}
	}

	// Close client connection - this will trigger cleanup
	c.hub.unregister <- c
}

func handleRematch(c *Client, msg duel.Message) {
	if c.roomID == "" {
		sendError(c, msg.RoomID, "join room first")
		return
	}

	if c.displayName != "" {
		msg.UserID = c.displayName
	}

	if !mgr.HasPlayer(msg.RoomID, msg.UserID) {
		sendError(c, msg.RoomID, "player not in room")
		return
	}

	events, allReady, err := mgr.RequestRematch(msg.RoomID, msg.UserID)
	if err != nil {
		sendError(c, msg.RoomID, err.Error())
		return
	}

	if !allReady {
		ev := duel.Event{
			Type:   "rematch_waiting",
			RoomID: msg.RoomID,
			UserID: msg.UserID,
		}
		broadcastRoom(c.hub, msg.RoomID, ev)
		return
	}

	// Both players ready — set up new duel in DB.
	newDuelID := ""
	if repo != nil {
		ctx := context.Background()

		roomDuelID.mu.Lock()
		oldDuelID := roomDuelID.byRoom[msg.RoomID]
		roomDuelID.mu.Unlock()

		if oldDuelID != "" {
			_ = repo.FinishDuel(ctx, oldDuelID)
		}

		playerNames := mgr.GetPlayerNames(msg.RoomID)

		userIDMap := make(map[string]string)
		roomUserID.mu.Lock()
		if m := roomUserID.byRoomUser[msg.RoomID]; m != nil {
			for displayName, uid := range m {
				userIDMap[displayName] = uid
			}
		}
		roomUserID.mu.Unlock()

		for _, displayName := range playerNames {
			uid, ok := userIDMap[displayName]
			if !ok || uid == "" {
				user, err := repo.GetUserByUsername(ctx, displayName)
				if err != nil {
					if err == storage.ErrNotFound {
						user, err = repo.CreateGuestUser(ctx, displayName, 72)
						if err != nil {
							log.Printf("handleRematch: CreateGuestUser error for %s: %v", displayName, err)
							continue
						}
					} else {
						log.Printf("handleRematch: GetUserByUsername error for %s: %v", displayName, err)
						continue
					}
				}
				uid = user.ID
			}
			userIDMap[displayName] = uid
		}

		creatorUID := ""
		for _, displayName := range playerNames {
			if uid, ok := userIDMap[displayName]; ok && uid != "" {
				creatorUID = uid
				break
			}
		}

		// Determine topic/difficulty/lang from room state (fallback to message values)
		rematchTopic := msg.Topic
		rematchLang := msg.Lang
		rematchDifficulty := msg.Difficulty
		if rTopic, rLang, rDiff, ok := mgr.GetRoomSettings(msg.RoomID); ok {
			if rTopic != "" {
				rematchTopic = rTopic
			}
			if rLang != "" {
				rematchLang = rLang
			}
			if rDiff != "" {
				rematchDifficulty = rDiff
			}
		}

		difficulty := difficultyToInt(rematchDifficulty)
		d, err := repo.CreateDuel(ctx, msg.RoomID, creatorUID, rematchTopic, difficulty, rematchLang, "ru")
		if err != nil {
			log.Printf("handleRematch: CreateDuel error: %v", err)
			// DB failed but manager already reset — broadcast error to both players
			broadcastRoom(c.hub, msg.RoomID, duel.Event{
				Type:   "error",
				RoomID: msg.RoomID,
				Error:  "rematch failed, please rejoin",
			})
			return
		}
		newDuelID = d.ID

		roomDuelID.mu.Lock()
		roomDuelID.byRoom[msg.RoomID] = d.ID
		roomDuelID.mu.Unlock()

		mgr.SetDuelID(msg.RoomID, d.ID)

		roomParticipants.mu.Lock()
		roomParticipants.byRoomUser[msg.RoomID] = make(map[string]string)
		roomParticipants.mu.Unlock()

		roomUserID.mu.Lock()
		roomUserID.byRoomUser[msg.RoomID] = make(map[string]string)
		roomUserID.mu.Unlock()

		for i, displayName := range playerNames {
			uid := userIDMap[displayName]
			if uid == "" {
				continue
			}
			p, err := repo.EnsureParticipant(ctx, d.ID, uid, i+1)
			if err != nil {
				log.Printf("handleRematch: EnsureParticipant error for %s: %v", displayName, err)
				continue
			}

			roomParticipants.mu.Lock()
			roomParticipants.byRoomUser[msg.RoomID][displayName] = p.ID
			roomParticipants.mu.Unlock()

			roomUserID.mu.Lock()
			roomUserID.byRoomUser[msg.RoomID][displayName] = uid
			roomUserID.mu.Unlock()
		}
	}

	// Update DuelID in all events to the new duel (events were generated before SetDuelID).
	if newDuelID != "" {
		for i := range events {
			events[i].DuelID = newDuelID
		}
	}

	for _, ev := range events {
		if ev.Type == "round_start" {
			if repo != nil {
				ctx := context.Background()
				roomDuelID.mu.Lock()
				duelID := roomDuelID.byRoom[ev.RoomID]
				roomDuelID.mu.Unlock()
				if duelID != "" {
					_ = repo.MarkDuelStarted(ctx, duelID)
					rnd, err := repo.CreateRound(ctx, duelID, ev.Round, ev.Prompt, ev.CorrectAnswer, ev.Lang, ev.Topic, int(roundTimeout/time.Millisecond), ev.ValidAnswers)
					if err != nil {
						log.Printf("DB CreateRound error: %v", err)
					} else {
						roomRoundID.mu.Lock()
						if roomRoundID.byRoomRound[ev.RoomID] == nil {
							roomRoundID.byRoomRound[ev.RoomID] = make(map[int]string)
						}
						roomRoundID.byRoomRound[ev.RoomID][ev.Round] = rnd.ID
						roomRoundID.mu.Unlock()
					}
				}
			}
			scheduleRoundTimeout(ev.RoomID, ev.RoundToken)
		}
		broadcastRoom(c.hub, msg.RoomID, ev)
	}
}

func processGameOverDB(ev *duel.Event) {
	if repo == nil {
		return
	}
	ctx := context.Background()
	roomDuelID.mu.Lock()
	duelID := roomDuelID.byRoom[ev.RoomID]
	roomDuelID.mu.Unlock()
	if duelID == "" {
		return
	}
	if err := repo.FinishDuel(ctx, duelID); err != nil {
		log.Printf("DB FinishDuel error: %v", err)
	}

	roomUserID.mu.Lock()
	winnerUserID := ""
	loserUserID := ""
	if m := roomUserID.byRoomUser[ev.RoomID]; m != nil {
		winnerUserID = m[ev.WinnerID]
		for uid, uid2 := range m {
			if uid != ev.WinnerID {
				loserUserID = uid2
			}
		}
	}
	roomUserID.mu.Unlock()

	if winnerUserID != "" {
		if err := repo.SetDuelWinner(ctx, duelID, winnerUserID); err != nil {
			log.Printf("DB SetDuelWinner error: %v", err)
		}
		if loserUserID != "" {
			if err := repo.UpdateRating(ctx, winnerUserID, loserUserID); err != nil {
				log.Printf("DB UpdateRating error: %v", err)
			} else {
				winnerRating, _ := repo.GetUserRating(ctx, winnerUserID)
				loserRating, _ := repo.GetUserRating(ctx, loserUserID)
				ev.EloChange = map[string]int{
					winnerUserID: 25,
					loserUserID:  -15,
				}
				winnerElo := 1000
				loserElo := 1000
				if winnerRating != nil {
					winnerElo = winnerRating.Elo
				}
				if loserRating != nil {
					loserElo = loserRating.Elo
				}
				ev.Elo = map[string]int{
					winnerUserID: winnerElo,
					loserUserID:  loserElo,
				}
				log.Printf("Rating updated: winner=%s +25=%d, loser=%s -15=%d",
					winnerUserID, winnerElo, loserUserID, loserElo)
			}
		}
	} else {
		log.Printf("DB SetDuelWinner skipped: winner user_id not found for %s", ev.WinnerID)
	}

	roomParticipants.mu.Lock()
	pmap := roomParticipants.byRoomUser[ev.RoomID]
	roomParticipants.mu.Unlock()
	if pmap != nil {
		for uid, pid := range pmap {
			finalHP := 0
			if ev.HP != nil {
				finalHP = ev.HP[uid]
			}
			if err := repo.SetParticipantFinalHP(ctx, pid, finalHP); err != nil {
				log.Printf("DB SetParticipantFinalHP error: %v", err)
			} else {
				log.Printf("DB SetParticipantFinalHP ok: participant=%s final_hp=%d", pid, finalHP)
			}
			roomUserID.mu.Lock()
			userID := ""
			if m := roomUserID.byRoomUser[ev.RoomID]; m != nil {
				userID = m[uid]
			}
			roomUserID.mu.Unlock()
			if userID != "" {
				isWinner := uid == ev.WinnerID

				if err := repo.UpdateUserStats(ctx, userID, isWinner); err != nil {
					log.Printf("DB UpdateUserStats error: %v", err)
				}

				xpAmount := 10
				if !isWinner {
					xpAmount = 5
				}
				log.Printf("Awarding %d XP to user %s (winner: %v)", xpAmount, userID, isWinner)
				oldLevel, newLevel, err := repo.AwardXP(ctx, userID, xpAmount)
				if err != nil {
					log.Printf("AwardXP error: %v", err)
				} else {
					log.Printf("XP awarded: user=%s, amount=%d, oldXP=%d, newLevel=%d", userID, xpAmount, oldLevel, newLevel)
				}

				rating, _ := repo.GetUserRating(ctx, userID)
				currentStreak := 0
				if rating != nil {
					currentStreak = rating.CurrentStreak
					if isWinner && currentStreak < 1 {
						currentStreak = 1
					} else if !isWinner {
						currentStreak = -1
					}
				}
				unlocked, _ := repo.CheckAndUnlockAchievements(ctx, userID, isWinner, currentStreak)
				if len(unlocked) > 0 {
					for _, a := range unlocked {
						log.Printf("Achievement unlocked: user=%s achievement=%s (%s)", userID, a.ID, a.Name)
					}
				}
			}
		}
	}
}

func cleanupRoomState(roomID string) {
	roomRoundID.mu.Lock()
	delete(roomRoundID.byRoomRound, roomID)
	roomRoundID.mu.Unlock()

	roomParticipants.mu.Lock()
	delete(roomParticipants.byRoomUser, roomID)
	roomParticipants.mu.Unlock()

	roomUserID.mu.Lock()
	delete(roomUserID.byRoomUser, roomID)
	roomUserID.mu.Unlock()

	roomDuelID.mu.Lock()
	delete(roomDuelID.byRoom, roomID)
	roomDuelID.mu.Unlock()
}

func broadcastRoom(h *Hub, roomID string, ev duel.Event) {
	payload, _ := json.Marshal(ev)
	h.broadcast <- broadcastMsg{roomID: roomID, data: payload}
}

func sendError(c *Client, roomID, msg string) {
	log.Printf("WS error: room=%s user=%s err=%s", roomID, c.displayName, msg)
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
		// Save timeout (no-answer) entries for the round that just expired.
		// round_end is always the first event and carries the timed-out round number.
		if repo != nil && len(events) > 0 && events[0].Type == "round_end" {
			timedOutRound := events[0].Round
			ctx := context.Background()
			roomRoundID.mu.Lock()
			timedOutRoundID := ""
			if m := roomRoundID.byRoomRound[roomID]; m != nil {
				timedOutRoundID = m[timedOutRound]
			}
			roomRoundID.mu.Unlock()
			if timedOutRoundID != "" {
				roomParticipants.mu.Lock()
				pmap := make(map[string]string)
				if m := roomParticipants.byRoomUser[roomID]; m != nil {
					for k, v := range m {
						pmap[k] = v
					}
				}
				roomParticipants.mu.Unlock()
				for _, participantID := range pmap {
					// ON CONFLICT DO NOTHING — won't overwrite a real answer submitted in time.
					if err := repo.CreateAnswer(ctx, timedOutRoundID, participantID, "", false, 0, 0); err != nil {
						log.Printf("DB SaveTimeoutAnswer error round=%s participant=%s: %v", timedOutRoundID, participantID, err)
					}
				}
			}
		}

		for i := range events {
			ev := &events[i]
			if ev.Type == "round_start" {
				if repo != nil {
					ctx := context.Background()
					roomDuelID.mu.Lock()
					duelID := roomDuelID.byRoom[ev.RoomID]
					roomDuelID.mu.Unlock()
					if duelID != "" {
						if err := repo.MarkDuelStarted(ctx, duelID); err != nil {
							log.Printf("DB MarkDuelStarted error (timeout): %v", err)
						}
						rnd, err := repo.CreateRound(ctx, duelID, ev.Round, ev.Prompt, ev.CorrectAnswer, ev.Lang, ev.Topic, int(roundTimeout/time.Millisecond), ev.ValidAnswers)
						if err != nil {
							log.Printf("DB CreateRound error (timeout): %v", err)
						} else {
							roomRoundID.mu.Lock()
							if roomRoundID.byRoomRound[ev.RoomID] == nil {
								roomRoundID.byRoomRound[ev.RoomID] = make(map[int]string)
							}
							roomRoundID.byRoomRound[ev.RoomID][ev.Round] = rnd.ID
							roomRoundID.mu.Unlock()
						}
					}
				}
				scheduleRoundTimeout(ev.RoomID, ev.RoundToken)
			}
			if ev.Type == "game_over" {
				processGameOverDB(ev)
				stopRoundTimer(roomID)
				cleanupRoomState(roomID)
			}
			broadcastRoom(hub, roomID, *ev)
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
