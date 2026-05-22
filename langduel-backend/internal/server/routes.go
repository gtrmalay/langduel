package server

import (
	"fmt"
	"langduel/internal/ws"
	"net/http"
)

func (s *Server) routes() {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "LangDuel API running")
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
	s.mux.HandleFunc("/duels", authMiddleware(s.handleDuelDetails))

	// public endpoints
	s.mux.HandleFunc("/analysis", s.handleDuelAnalysisPublic)

	// public endpoints
	s.mux.HandleFunc("/api/leaderboard", s.handleLeaderboard)

	// achievement endpoints
	s.mux.HandleFunc("/me/achievements", authMiddleware(s.handleMyAchievements))
	s.mux.HandleFunc("/me/claim-coins", authMiddleware(s.handleClaimCoins))

	// avatar shop endpoints
	s.mux.HandleFunc("/me/buy-avatar", authMiddleware(s.handleBuyAvatar))

	// AI phrase generation endpoint (rate limited)
	s.mux.HandleFunc("/api/generate-phrases", rateLimitMiddleware(s.handleGeneratePhrases))
}
