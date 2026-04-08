package ws

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn        *websocket.Conn
	send        chan []byte
	hub         *Hub
	displayName string
	authUserID  string
	roomID      string

	mu          sync.Mutex
	lastMsgTime time.Time
	msgCount    int
}

const (
	rateLimitWindow  = 1 * time.Second
	rateLimitMaxMsgs = 10
)

func (c *Client) checkRateLimit() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if now.Sub(c.lastMsgTime) > rateLimitWindow {
		c.lastMsgTime = now
		c.msgCount = 1
		return true
	}

	c.msgCount++
	if c.msgCount > rateLimitMaxMsgs {
		return false
	}
	return true
}
