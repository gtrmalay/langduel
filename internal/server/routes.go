package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"langduel/internal/ws"
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
}
