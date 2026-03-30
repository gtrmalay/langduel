package server

import (
	"fmt"
	"langduel/internal/ws"
	"net/http"
	"os"
	"path/filepath"
)

func (s *Server) routes() {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "LangDuel API running")
	})
	s.mux.HandleFunc("/battle", func(w http.ResponseWriter, r *http.Request) {
		cwd, _ := os.Getwd()
		http.ServeFile(w, r, filepath.Join(cwd, "battle.html"))
	})

	// websocket endpoint
	s.mux.HandleFunc("/ws", ws.WsHandler)

	// auth endpoints
	s.mux.HandleFunc("/auth/register", s.handleRegister)
	s.mux.HandleFunc("/auth/login", s.handleLogin)

	// protected endpoints
	s.mux.HandleFunc("/me", authMiddleware(s.handleMe))
	s.mux.HandleFunc("/me/stats", authMiddleware(s.handleMyStats))
	s.mux.HandleFunc("/me/duels", authMiddleware(s.handleMyDuels))
	s.mux.HandleFunc("/me/username", authMiddleware(s.handleUpdateUsername))
	s.mux.HandleFunc("/me/avatar", authMiddleware(s.handleUpdateAvatar))
	s.mux.HandleFunc("/me/rating", authMiddleware(s.handleMyRating))

	// public endpoints
	s.mux.HandleFunc("/leaderboard", s.handleLeaderboard)

	// achievement endpoints
	s.mux.HandleFunc("/me/achievements", authMiddleware(s.handleMyAchievements))
	s.mux.HandleFunc("/me/claim-coins", authMiddleware(s.handleClaimCoins))
}
