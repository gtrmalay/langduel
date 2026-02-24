package ws

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	// Через неё читаем входящие сообщения и отправляем ответы.
	conn *websocket.Conn

	// Hub кладёт сообщения сюда, а writePump отправляет их в сокет.
	// Буфер нужен, чтобы медленный клиент не блокировал весь сервер.
	send chan []byte

	// Нужен для регистрации, удаления клиента и рассылки сообщений.
	hub  *Hub

	// displayName реальное имя для игры/UI.
	displayName string

	// authUserID UUID из JWT (ид в БД).
	authUserID string

	roomID string
}
