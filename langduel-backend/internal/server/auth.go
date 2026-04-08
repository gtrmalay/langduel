package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"langduel/internal/ai"
	"langduel/internal/duel"
	"langduel/internal/storage"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var repo *storage.DuelRepo

// SetRepo wires storage for HTTP handlers.
func SetRepo(r *storage.DuelRepo) {
	repo = r
}

type registerReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type authResp struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, "password too short", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "hash error", http.StatusBadRequest)
		return
	}

	var user *storage.User
	var errMsg string

	user, err = repo.ConvertGuestToUser(r.Context(), req.Username, req.Email, string(hash))
	if err != nil {
		if err == storage.ErrNotFound {
			user, err = repo.CreateUser(r.Context(), req.Username, req.Email, string(hash))
			if err != nil {
				errMsg = err.Error()
			}
		} else {
			errMsg = err.Error()
		}
	}
	if errMsg != "" {
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	token, err := buildToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, authResp{UserID: user.ID, Username: user.Username, Token: token})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Login == "" || req.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	user, err := repo.GetAuthUserByUsernameOrEmail(r.Context(), req.Login)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := buildToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, authResp{UserID: user.ID, Username: user.Username, Token: token})
}

func buildToken(userID, username string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errMissingSecret
	}
	claims := jwt.MapClaims{
		"sub": userID,
		"usr": username,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

var errMissingSecret = jwt.ErrTokenInvalidClaims

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

type authedUser struct {
	ID       string
	Username string
}

func authMiddleware(next func(http.ResponseWriter, *http.Request, authedUser)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		tokenStr := authHeader[len(prefix):]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(w, "server misconfig", http.StatusInternalServerError)
			return
		}
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		sub, _ := claims["sub"].(string)
		usr, _ := claims["usr"].(string)
		if sub == "" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		next(w, r, authedUser{ID: sub, Username: usr})
	}
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request, u authedUser) {
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}
	user, err := repo.GetUserByID(r.Context(), u.ID)
	if err != nil {
		writeJSON(w, map[string]any{
			"user_id":          u.ID,
			"username":         u.Username,
			"avatar":           "default",
			"coins":            0,
			"unlocked_avatars": []string{"default"},
		})
		return
	}
	writeJSON(w, map[string]any{
		"user_id":          user.ID,
		"username":         user.Username,
		"avatar":           user.Avatar,
		"coins":            user.Coins,
		"unlocked_avatars": user.UnlockedAvatars,
	})
}

func (s *Server) handleMyStats(w http.ResponseWriter, r *http.Request, u authedUser) {
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}
	log.Printf("GET /me/stats user_id=%s", u.ID)
	stats, err := repo.GetUserStats(r.Context(), u.ID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Printf("GET /me/stats not found, returning zeros user_id=%s", u.ID)
			writeJSON(w, map[string]any{
				"user_id":             u.ID,
				"total_duels_played":  0,
				"total_duels_won":     0,
				"total_score":         0,
				"overall_accuracy":    0,
				"best_win_streak":     0,
				"total_play_time_min": 0,
				"updated_at":          "",
			})
			return
		}
		log.Printf("GET /me/stats error: %v", err)
		http.Error(w, "stats not found", http.StatusNotFound)
		return
	}
	log.Printf("GET /me/stats ok user_id=%s", u.ID)
	writeJSON(w, stats)
}

func (s *Server) handleMyDuels(w http.ResponseWriter, r *http.Request, u authedUser) {
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}
	items, err := repo.GetRecentDuels(r.Context(), u.ID, 0)
	if err != nil {
		log.Printf("GET /me/duels error: %v", err)
		writeJSON(w, []any{})
		return
	}
	writeJSON(w, items)
}

func (s *Server) handleDuelDetails(w http.ResponseWriter, r *http.Request, u authedUser) {
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	duelID := r.URL.Query().Get("id")
	if duelID == "" {
		http.Error(w, "duel id required", http.StatusBadRequest)
		return
	}

	detail, err := repo.GetDuelDetails(r.Context(), duelID, u.ID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "duel not found", http.StatusNotFound)
			return
		}
		log.Printf("GET /duels error: %v", err)
		http.Error(w, "failed to get duel details", http.StatusInternalServerError)
		return
	}

	writeJSON(w, detail)
}

func (s *Server) handleDuelAnalysisPublic(w http.ResponseWriter, r *http.Request) {
	duelID := r.URL.Query().Get("id")
	if duelID == "" {
		http.Error(w, "duel id required", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(duelID); err != nil {
		http.Error(w, "invalid duel id format", http.StatusBadRequest)
		return
	}

	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	analysis, err := repo.GetDuelAnalysis(r.Context(), duelID)
	if err != nil {
		log.Printf("GET /analysis error: %v", err)
		http.Error(w, "failed to get analysis", http.StatusInternalServerError)
		return
	}

	writeJSON(w, analysis)
}

func (s *Server) handleDuelAnalysis(w http.ResponseWriter, r *http.Request, u authedUser) {
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	duelID := r.URL.Query().Get("id")
	if duelID == "" {
		http.Error(w, "duel id required", http.StatusBadRequest)
		return
	}

	analysis, err := repo.GetDuelAnalysis(r.Context(), duelID)
	if err != nil {
		log.Printf("GET /analysis error: %v", err)
		http.Error(w, "failed to get analysis", http.StatusInternalServerError)
		return
	}

	writeJSON(w, analysis)
}

type updateUsernameReq struct {
	Username string `json:"username"`
}

func (s *Server) handleUpdateUsername(w http.ResponseWriter, r *http.Request, u authedUser) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	var req updateUsernameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}
	if len(req.Username) < 3 || len(req.Username) > 30 {
		http.Error(w, "username must be 3-30 characters", http.StatusBadRequest)
		return
	}

	err := repo.UpdateUsername(r.Context(), u.ID, req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{"username": req.Username})
}

type updateAvatarReq struct {
	Avatar string `json:"avatar"`
}

func (s *Server) handleUpdateAvatar(w http.ResponseWriter, r *http.Request, u authedUser) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	var req updateAvatarReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Avatar == "" {
		req.Avatar = "default"
	}

	err := repo.UpdateAvatar(r.Context(), u.ID, req.Avatar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"avatar": req.Avatar})
}

func (s *Server) handleMyRating(w http.ResponseWriter, r *http.Request, u authedUser) {
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	rating, err := repo.GetUserRating(r.Context(), u.ID)
	if err != nil {
		log.Printf("GET /me/rating error: %v", err)
		http.Error(w, "rating not found", http.StatusNotFound)
		return
	}

	writeJSON(w, rating)
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	entries, err := repo.GetLeaderboard(r.Context(), 100)
	if err != nil {
		log.Printf("GET /leaderboard error: %v", err)
		http.Error(w, "leaderboard error", http.StatusInternalServerError)
		return
	}

	if entries == nil {
		entries = []storage.LeaderboardEntry{}
	}

	writeJSON(w, entries)
}

func (s *Server) handleMyAchievements(w http.ResponseWriter, r *http.Request, u authedUser) {
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	achievements, err := repo.GetAllAchievements(r.Context(), u.ID)
	if err != nil {
		log.Printf("GET /me/achievements error: %v", err)
		http.Error(w, "achievements error", http.StatusInternalServerError)
		return
	}

	if achievements == nil {
		achievements = []storage.Achievement{}
	}

	writeJSON(w, achievements)
}

// handleClaimCoins retroactively claims coins for already unlocked achievements
func (s *Server) handleClaimCoins(w http.ResponseWriter, r *http.Request, u authedUser) {
	log.Printf("POST /me/claim-coins called for user: %s", u.ID)
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	coins, err := repo.AwardCoinsForUnlockedAchievements(r.Context(), u.ID)
	if err != nil {
		log.Printf("POST /me/claim-coins error: %v", err)
		http.Error(w, "claim coins error", http.StatusInternalServerError)
		return
	}

	log.Printf("POST /me/claim-coins result: %d coins for user %s", coins, u.ID)
	writeJSON(w, map[string]int{"coins_awarded": coins})
}

type buyAvatarReq struct {
	AvatarID string `json:"avatar_id"`
}

func (s *Server) handleBuyAvatar(w http.ResponseWriter, r *http.Request, u authedUser) {
	log.Printf("POST /me/buy-avatar called for user: %s", u.ID)
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	var req buyAvatarReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.AvatarID == "" {
		http.Error(w, "avatar_id required", http.StatusBadRequest)
		return
	}

	avatarPrices := map[string]int{
		"knight":    50,
		"wizard":    75,
		"archer":    75,
		"dragon":    100,
		"skull":     50,
		"fire":      60,
		"ice":       60,
		"lightning": 80,
		"sword":     50,
		"shield":    50,
		"potion":    60,
		"crown":     150,
		"star":      100,
		"moon":      80,
	}

	price, ok := avatarPrices[req.AvatarID]
	if !ok {
		http.Error(w, "invalid avatar", http.StatusBadRequest)
		return
	}

	userBefore, _ := repo.GetUserByID(r.Context(), u.ID)
	log.Printf("DEBUG: user %s has %d coins, avatar %s costs %d", u.ID, userBefore.Coins, req.AvatarID, price)

	err := repo.BuyAvatar(r.Context(), u.ID, req.AvatarID, price)
	if err != nil {
		log.Printf("POST /me/buy-avatar error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, _ := repo.GetUserByID(r.Context(), u.ID)
	writeJSON(w, map[string]interface{}{
		"success":          true,
		"avatar_id":        req.AvatarID,
		"price":            price,
		"coins":            user.Coins,
		"unlocked_avatars": user.UnlockedAvatars,
	})
}

type generatePhrasesReq struct {
	RoomID     string `json:"room_id"`
	Topic      string `json:"topic"`
	Difficulty string `json:"difficulty"`
	LangFrom   string `json:"lang_from"`
	LangTo     string `json:"lang_to"`
}

func difficultyToInt(d string) int {
	switch d {
	case "beginner":
		return 1
	case "intermediate":
		return 2
	case "advanced":
		return 3
	default:
		return 2
	}
}

func (s *Server) handleGeneratePhrases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}

	var req generatePhrasesReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.RoomID == "" || req.Topic == "" || req.Difficulty == "" {
		http.Error(w, "room_id, topic and difficulty required", http.StatusBadRequest)
		return
	}

	if err := duel.ValidateRoomID(req.RoomID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := duel.ValidateDifficulty(req.Difficulty); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.LangFrom == "" {
		req.LangFrom = "en"
	}
	if req.LangTo == "" {
		req.LangTo = "ru"
	}

	if req.LangFrom != "en" && req.LangFrom != "ru" {
		http.Error(w, "lang_from must be 'en' or 'ru'", http.StatusBadRequest)
		return
	}
	if req.LangTo != "en" && req.LangTo != "ru" {
		http.Error(w, "lang_to must be 'en' or 'ru'", http.StatusBadRequest)
		return
	}

	if req.Topic == "custom" || req.Topic == "" {
		http.Error(w, "topic cannot be 'custom' or empty", http.StatusBadRequest)
		return
	}
	if err := duel.ValidateTopic(req.Topic); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Try to get existing duel by room code first
	d, err := repo.GetDuelByRoomCode(r.Context(), req.RoomID)
	if err != nil {
		// No existing duel - generate phrases without saving to DB first
		// We'll create the duel when players actually join
		log.Printf("No duel exists for room %s, generating phrases without DB", req.RoomID)
		d = nil
	} else {
		log.Printf("Found existing duel %s for room %s", d.ID, req.RoomID)
	}

	log.Printf("Generating phrases for topic=%s difficulty=%s lang=%s->%s",
		req.Topic, req.Difficulty, req.LangFrom, req.LangTo)

	generator := ai.NewGenerator()

	// Create context with timeout for AI generation
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	phrases, err := generator.GeneratePhrases(ctx, req.Topic, req.Difficulty, req.LangFrom, req.LangTo, 20)

	if err != nil {
		log.Printf("AI generation error: %v", err)
		http.Error(w, "failed to generate phrases: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully generated %d phrases", len(phrases))

	duelID := ""
	if d != nil {
		duelID = d.ID
	}

	// Always save with room_code for easy lookup
	for _, p := range phrases {
		err := repo.SaveAIPhrase(r.Context(), duelID, req.RoomID, p.Prompt, p.Answers, req.Topic, req.Difficulty, req.LangFrom, req.LangTo)
		if err != nil {
			log.Printf("Failed to save phrase: %v", err)
			http.Error(w, "failed to save phrases", http.StatusInternalServerError)
			return
		}
	}

	if d != nil {
		log.Printf("Saved %d phrases for duel %s (room %s)", len(phrases), d.ID, req.RoomID)
	} else {
		log.Printf("Saved %d phrases for room %s (no duel yet)", len(phrases), req.RoomID)
	}

	writeJSON(w, map[string]any{
		"success": true,
		"count":   len(phrases),
		"duel_id": duelID,
		"phrases": phrases,
	})
}
