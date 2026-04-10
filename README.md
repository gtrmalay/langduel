# LangDuel 🥷

Browser-based 1v1 translation battle game. Translate faster than your opponent to deal damage and win!

## 🏗 Architecture

- **Backend**: Go (chi router, gorilla/websocket, PostgreSQL via pgx)
- **Frontend**: Svelte 5 with SvelteKit
- **Deployment**: Docker (PostgreSQL + Go + Nginx)

## 🎮 Game Rules

- 2 players per room
- 2 halves × 10 rounds (20 total)
- 5-second halftime break between halves
- 10 seconds per round
- "First blood" - game ends when any player reaches 0 HP
- ELO: +25 for win, -15 for loss

## 🚀 Quick Start (Docker)

```bash
# Clone and run
docker-compose up --build

# Access at http://localhost
```

## 🔧 Local Development

### Backend
```bash
cd langduel-backend

# Run (requires PostgreSQL)
go run ./cmd/server

# Test
go test ./...
```

### Frontend
```bash
cd langduel-frontend

# Dev server
npm run dev

# Production build
npm run build
npm run preview
```

## 📝 Environment Variables

Create `.env`:
```
JWT_SECRET=your-secret-key
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable
OPENROUTER_API_KEY=your-key  # Optional: AI phrase generation
```

## 🎯 Features

- Guest mode (no registration required)
- User authentication & stats
- ELO rating system
- Leaderboard
- Achievements
- Ping indicator
- Disconnect notifications
- Halftime system

## 📄 License

MIT