package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// Превращает HTTP в WebSocket
// CheckOrigin - Может подключиться любой сайт, на production переделать!
var upgrader = websocket.Upgrader{ 
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hub = NewHub() // Центральный диспетчер

// Запуск хаба в горутине при запуске программы
func init() { 
	go hub.Run()
}

func WsHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil) // Апгрейдим HTTP в WS
	if err != nil {
		return
	}

	client := &Client{ 
		conn: conn,
		send: make(chan []byte, 256),
		hub:  hub,
	}

	// Регистрация клиента
	hub.register <- client

	go readPump(client)
	go writePump(client)
}

func readPump(c *Client) { // Socket -> Hub
	defer func() { 
		c.hub.unregister <- c
		c.conn.Close()
	}() 

	for {
		_, msg, err := c.conn.ReadMessage() //Непрерывное чтение всех сообщений с соединения клиента
		if err != nil {
			break
		}

		c.hub.broadcast <- msg //Отправка полученного от клиента сообщения всем 
	}
}

func writePump(c *Client) { // Send Channel -> Socket
	for msg := range c.send {
		c.conn.WriteMessage(websocket.TextMessage, msg)
	}
}

//           (входящее)
// Браузер  →  Socket  →  readPump  →  Hub

//           (исходящее)
// Hub  →  send channel  →  writePump  →  Socket  →  Браузер
