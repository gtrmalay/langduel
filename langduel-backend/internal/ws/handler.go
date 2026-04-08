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
			if events, _ := mgr.Leave(c.roomID, c.displayName); len(events) > 0 {
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
					rnd, err := repo.CreateRound(ctx, duelID, ev.Round, ev.Prompt, ev.CorrectAnswer, ev.Lang, ev.Topic, int(roundTimeout/time.Millisecond))
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
	events, err := mgr.SubmitAnswer(c.roomID, msg.UserID, sanitizedAnswer, msg.Speed)
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
					rnd, err := repo.CreateRound(ctx, duelID, ev.Round, ev.Prompt, ev.CorrectAnswer, ev.Lang, ev.Topic, int(roundTimeout/time.Millisecond))
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
			if repo != nil {
				ctx := context.Background()
				roomDuelID.mu.Lock()
				duelID := roomDuelID.byRoom[ev.RoomID]
				roomDuelID.mu.Unlock()
				if duelID != "" {
					if err := repo.FinishDuel(ctx, duelID); err != nil {
						log.Printf("DB FinishDuel error: %v", err)
					}
					// winner
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
						// Update ratings
						if loserUserID != "" {
							if err := repo.UpdateRating(ctx, winnerUserID, loserUserID); err != nil {
								log.Printf("DB UpdateRating error: %v", err)
							} else {
								// Get updated ratings
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

					// final hp + stats
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

								// Award XP: +10 for win, +5 for loss
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

								// Check and unlock achievements
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
			}
			stopRoundTimer(ev.RoomID)
		}
		if ev.Type == "update" && repo != nil {
			ctx := context.Background()
			roomRoundID.mu.Lock()
			roundID := ""
			if m := roomRoundID.byRoomRound[ev.RoomID]; m != nil {
				roundID = m[ev.Round]
			}
			roomRoundID.mu.Unlock()
			roomParticipants.mu.Lock()
			participantID := ""
			if m := roomParticipants.byRoomUser[ev.RoomID]; m != nil {
				participantID = m[ev.AttackerID]
			}
			roomParticipants.mu.Unlock()
			if roundID != "" && participantID != "" {
				err := repo.CreateAnswer(ctx, roundID, participantID, msg.Answer, ev.Correct, msg.Speed, ev.Damage)
				if err != nil {
					log.Printf("DB CreateAnswer error: %v", err)
				} else {
					log.Printf("DB CreateAnswer ok: round=%s participant=%s correct=%v damage=%d", roundID, participantID, ev.Correct, ev.Damage)
				}
			}
		}
		broadcastRoom(c.hub, c.roomID, ev)
	}
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
		for _, ev := range events {
			// If round_start is produced by timeout, arm the next timer.
			if ev.Type == "round_start" {
				scheduleRoundTimeout(roomID, ev.RoundToken)
			}
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
