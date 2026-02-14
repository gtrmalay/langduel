package duel

var GlobalManager = NewManager()

type Manager struct {
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

func (m *Manager) CreateRoom(id string) *Room {

	room := &Room{
		ID: id,
	}

	m.rooms[id] = room
	return room
}