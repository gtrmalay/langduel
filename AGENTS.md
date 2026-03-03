# LangDuel - Agent Guidelines

LangDuel is a browser-based 1v1 translation battle game:
- **Backend**: Go (chi router, gorilla/websocket, PostgreSQL via pgx)
- **Frontend**: Svelte 5 with SvelteKit

---

## 1. Build, Run & Test Commands

### Frontend
```bash
cd langduel-frontend

npm run dev        # Development server
npm run build      # Production build
npm run preview    # Preview production build
npm run prepare    # Generate types (svelte-kit sync)
```

### Backend (Go)
```bash
cd langduel-backend

go run ./cmd/server                    # Run server (requires PostgreSQL)
go test ./...                          # Run all tests
go test -run TestJoinStartsRound ./... # Run single test
go test -v ./...                       # Run with verbose output
go test -cover ./...                  # Run with coverage
go fmt ./...                           # Format code
go vet ./...                           # Check for errors
golangci-lint run                      # Lint (requires golangci-lint)
```

---

## 2. Project Structure

```
langduel/
├── langduel-backend/
│   ├── cmd/server/main.go    # Entry point
│   ├── internal/
│   │   ├── duel/             # Game logic (Manager, Room, Player)
│   │   ├── server/           # HTTP routes, auth
│   │   ├── storage/          # PostgreSQL repos
│   │   └── ws/               # WebSocket hub
│   ├── migrations/           # SQL migrations
│   └── docs/                 # API docs
│
├── langduel-frontend/
│   ├── src/
│   │   ├── routes/           # SvelteKit pages
│   │   └── lib/
│   │       ├── stores/       # Svelte stores (duel.js)
│   │       └── components/  # UI components
│   └── package.json
│
└── AGENTS.md
```

---

## 3. Go Code Style

### Conventions
- One package per directory
- `camelCase` for functions/variables, `PascalCase` for exported
- `snake_case` for DB columns
- Use `context.Context` for request timeouts

### Error Handling
```go
// Define sentinel errors
var (
    ErrRoomFull    = errors.New("room is full")
    ErrRoomNotFound = errors.New("room not found")
)

// Return with context
if room.Full() {
    return nil, ErrRoomFull
}
```

### Testing
- Test files: `*_test.go` in same package
- Use `t.Fatalf` for fatal errors, `t.Error` for failures

---

## 4. Svelte Code Style

### Conventions (Svelte 5)
- Props via `export let` for compatibility
- Use `$state()`, `$derived()`, `$effect()` runes when needed
- Components in `src/lib/components/`, pages in `src/routes/`

### Import Patterns
```javascript
import { goto } from '$app/navigation';
import { duel } from '$lib/stores/duel.js';
import Button from '$lib/components/Button.svelte';
```

### CSS Variables (defined in +layout.svelte)
```css
:root {
  --bg: #0b1020;
  --card: #171f33;
  --text: #e9edf6;
  --accent: #25f4b7;
  --accent-2: #f6c144;
  --danger: #ff5c7a;
  --outline: #2b344a;
}
```

---

## 5. API Endpoints

### HTTP (REST)
```
POST /auth/register  {"username", "email", "password"}
POST /auth/login     {"login", "password"}
GET  /me             (Bearer token) -> user info
GET  /me/stats       (Bearer token) -> statistics
GET  /me/duels      (Bearer token) -> duel history
```

### WebSocket (ws://localhost:8080/ws)
```json
// Client -> Server
{"type": "join", "room_id", "user_id", "lang", "topic"}
{"type": "answer", "room_id", "user_id", "answer", "speed"}

// Server -> Client
{"type": "room_state", "room_id", "round", "prompt", "players", "hp"}
{"type": "player_joined", "players", "hp"}
{"type": "round_start", "round", "prompt", "hp"}
{"type": "update", "attacker_id", "defender_id", "damage", "correct", "hp"}
{"type": "game_over", "winner_id", "hp"}
{"type": "error", "error"}
```

---

## 6. Game State

### Backend (Go)
- `Manager`: holds all rooms, manages creation/removal
- `Room`: players, current round, HP, game status
- `Player`: ID, name, HP

### Frontend (duel.js store)
- `authMode`: 'guest' or 'auth'
- `currentRoom`: active room ID
- `currentUser`: player ID
- `hp`: { playerId: hpValue }
- `promptText`: phrase to translate
- `timerText`: countdown display

---

## 7. Notes for Agents

- Guest mode is default; auth is optional for saving stats
- Game is "first blood" - ends when any player reaches 0 HP
- WebSocket is primary communication channel for game state
- No ESLint/Prettier - format code manually
- Test files use standard Go testing package
