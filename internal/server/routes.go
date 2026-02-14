package server

import (
	"fmt"
	"net/http"
	"langduel/internal/ws"
)

func (s *Server) routes() {

	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "LangDuel API running")
	})

	// websocket endpoint
	s.mux.HandleFunc("/ws", ws.WsHandler)
}
