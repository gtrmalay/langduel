package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"langduel/internal/storage"

	"github.com/golang-jwt/jwt/v5"
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
		http.Error(w, "hash error", http.StatusInternalServerError)
		return
	}

	user, err := repo.CreateUser(r.Context(), req.Username, req.Email, string(hash))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
			"user_id":  u.ID,
			"username": u.Username,
			"avatar":   "default",
		})
		return
	}
	writeJSON(w, map[string]any{
		"user_id":  user.ID,
		"username": user.Username,
		"avatar":   user.Avatar,
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
