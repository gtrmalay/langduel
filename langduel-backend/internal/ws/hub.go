package ws

type registration struct {
	client *Client
	roomID string
}

type broadcastMsg struct {
	roomID string
	data   []byte
}

type Hub struct {
	rooms      map[string]map[*Client]bool
	register   chan registration
	unregister chan *Client
	broadcast  chan broadcastMsg
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan registration),
		unregister: make(chan *Client),
		broadcast:  make(chan broadcastMsg),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case reg := <-h.register:
			if reg.roomID == "" {
				continue
			}
			room := h.rooms[reg.roomID]
			if room == nil {
				room = make(map[*Client]bool)
				h.rooms[reg.roomID] = room
			}
			reg.client.roomID = reg.roomID
			room[reg.client] = true

		case client := <-h.unregister:
			if client.roomID == "" {
				continue
			}
			room := h.rooms[client.roomID]
			if room == nil {
				continue
			}
			if _, ok := room[client]; ok {
				delete(room, client)
				close(client.send)
			}
			if len(room) == 0 {
				delete(h.rooms, client.roomID)
			}

		case msg := <-h.broadcast:
			room := h.rooms[msg.roomID]
			for client := range room {
				select {
				case client.send <- msg.data:
				default:
					close(client.send)
					delete(room, client)
				}
			}
		}
	}
}
