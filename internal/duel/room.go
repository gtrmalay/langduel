package duel

type Player struct {
	ID string
	HP int
}

type Room struct {
	ID string
	Players map[string]*Player
}

room := &Room{
	ID: id,
	Players: make(map[string]*Player),
}