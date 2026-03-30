# LangDuel - Agent Guidelines

LangDuel is a browser-based 1v1 translation battle game:
- **Backend**: Go (chi router, gorilla/websocket, PostgreSQL via pgx)
- **Frontend**: Svelte 5 with SvelteKit

---

## Roadmap

| # | Этап | Статус |
|---|-------|--------|
| 0 | MVP (базовые баттлы, профиль, i18n) | ✅ Готово |
| 1 | ELO + Лидерборд | ✅ Готово |
| 2 | AI Генерация фраз | 📋 Запланирован |
| 3 | Ачивки | ✅ Готово |
| 4 | XP + Level | 📋 Запланирован |
| 5 | Косметика | 📋 Запланирован |
| 6 | Анимации | 📋 Запланирован |

### Ранги ELO
| Звание | ELO |
|--------|-----|
| 🥉 Newbie | 0-999 |
| 🥈 Apprentice | 1000-1999 |
| 🥇 Expert | 2000-2999 |
| 💎 Master | 3000+ |
| 😔 Struggler | 0 (если >10 поражений) |

### Формула ELO
- Победа: +25 ELO
- Поражение: -15 ELO
- Минимум: 0

---

## Build, Run & Test Commands

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

## Project Structure

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
│   │   │   ├── +page.svelte     # Home
│   │   │   ├── play/+page.svelte   # Create/Join room
│   │   │   ├── lobby/+page.svelte  # Waiting room
│   │   │   ├── battle/+page.svelte # Game (no header)
│   │   │   ├── profile/+page.svelte # User stats
│   │   │   ├── auth/+page.svelte    # Login/Register
│   │   │   └── leaderboard/+page.svelte # Leaderboard
│   │   └── lib/
│   │       ├── stores/duel.js       # Game state store
│   │       ├── i18n/                 # Translations (en.json, ru.json)
│   │       └── components/          # UI components
│   └── package.json
│
└── AGENTS.md
```

---

## Go Code Style

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

## Svelte Code Style

### Conventions (Svelte 5)
- Props via `export let` for Svelte 4 compatibility
- Use `$state()`, `$derived()`, `$effect()` runes when needed
- Components in `src/lib/components/`, pages in `src/routes/`
- Header hidden on `/battle` page (fullscreen game)

### Import Patterns
```javascript
import { goto } from '$app/navigation';
import { page } from '$app/stores';
import { duel } from '$lib/stores/duel.js';
import Button from '$lib/components/Button.svelte';
```

### Store Usage (duel.js)
```javascript
// Subscribe to store
let value = $duel.someField;

// Update store
duel.setField('fieldName', value);
duel.setAuthMode('guest');
duel.selectGuest();
duel.logout();
```

### Event Handling
```svelte
<button on:click={handler}>Click</button>
<button on:click={() => func(param)}>Click</button>
<input bind:value={text} on:keydown={(e) => e.key === 'Enter' && submit()} />
```

### CSS Variables (defined in +layout.svelte)
```css
:root {
    --bg: #0b1020;
    --card: #171f33;
    --text: #e9edf6;
    --accent: #25f4b7;     /* Green - success, HP */
    --accent-2: #f6c144;   /* Yellow - timer, warnings */
    --danger: #ff5c7a;      /* Red - low HP, errors */
    --outline: #2b344a;
}
```

---

## Auth Flow

- Guest mode is default (no registration required)
- Auth is optional for saving stats/history
- Store fields: `authMode` ('guest' | 'auth'), `authedUsername`, `jwtToken`
- On logout: always set `authMode: 'guest'`
- On first visit to `/play`: redirect to `/auth` to choose mode

---

## API Endpoints

### HTTP (REST)
```
POST /auth/register  {"username", "email", "password"}
POST /auth/login     {"login", "password"}
GET  /me             (Bearer token) -> user info
GET  /me/stats       (Bearer token) -> statistics
GET  /me/duels       (Bearer token) -> duel history
GET  /me/rating      (Bearer token) -> ELO rating
GET  /leaderboard    -> top 100 players
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
{"type": "game_over", "winner_id", "hp", "elo_change"}
{"type": "error", "error"}
```

---

## Game State

### Backend (Go)
- `Manager`: holds all rooms, manages creation/removal
- `Room`: players, current round, HP, game status
- `Player`: ID, name, HP, Elo

### Frontend (duel.js store)
- `authMode`: 'guest' or 'auth'
- `currentRoom`: active room ID
- `currentUser`: player ID
- `hp`: { playerId: hpValue }
- `elo`: { playerId: eloValue }
- `promptText`: phrase to translate
- `timerText`: countdown display

### Game Rules
- 2 players per room
- Game is "first blood" - ends when any player reaches 0 HP
- Round timer: 10 seconds

---

## Notes for Agents

- No ESLint/Prettier - format code manually
- Test files use standard Go testing package
- WebSocket is primary communication channel for game state
- Header not shown on battle page to avoid distraction
- Use confirmation dialogs for Leave/Logout actions
- Guests are excluded from leaderboard
- AI generation uses Grok or GPT-3.5 for phrase generation
