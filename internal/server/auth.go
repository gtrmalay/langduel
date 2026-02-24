package server

import (
	"encoding/json"
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
	writeJSON(w, map[string]any{
		"user_id":  u.ID,
		"username": u.Username,
	})
}

func (s *Server) handleMyStats(w http.ResponseWriter, r *http.Request, u authedUser) {
	if repo == nil {
		http.Error(w, "db not configured", http.StatusServiceUnavailable)
		return
	}
	stats, err := repo.GetUserStats(r.Context(), u.ID)
	if err != nil {
		http.Error(w, "stats not found", http.StatusNotFound)
		return
	}
	writeJSON(w, stats)
}
