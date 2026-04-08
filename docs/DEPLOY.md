# LangDuel Deployment Guide for Ihor.cloud

## Prerequisites
- Docker and Docker Compose installed
- Domain configured (optional)

## Quick Start with Docker Compose

```bash
# Clone repository
git clone <repo-url>
cd langduel

# Set environment variables
cp langduel-backend/.env.example langduel-backend/.env
nano langduel-backend/.env  # Edit with your values

# Start services
docker-compose up -d

# Check logs
docker-compose logs -f
```

## Environment Variables

### Backend (.env)
```
DATABASE_URL=postgres://postgres:postgres@db:5432/langduel?sslmode=disable
JWT_SECRET=your_random_secret_here
OPENROUTER_API_KEY=  # Optional, leave empty to use fallback phrases
```

## Database Setup

The database tables are created automatically via migrations. For initial setup:

```bash
# Run migrations manually if needed
docker-compose exec backend ./server
```

## Building Frontend

```bash
cd langduel-frontend
npm install
npm run build
```

The built files will be in `.output/public/`

## Manual Deployment (without Docker)

### Backend
```bash
cd langduel-backend
go build -o server ./cmd/server
./server
```

### Database Migrations
```bash
# Apply migrations manually using psql
psql -U postgres -d langduel -f migrations/0001_init.sql
psql -U postgres -d langduel -f migrations/0002_seed_phrases.sql
# ... apply remaining migrations
```

## Nginx Configuration

See `docs/nginx.example.conf` for production nginx config.

## Troubleshooting

### "failed to create ai_phrases table"
The table creation uses separate EXEC calls. This warning can be ignored if the table already exists.

### WebSocket connection fails
Ensure nginx is configured with WebSocket proxy support (see nginx.example.conf)

### Database connection issues
- Check DATABASE_URL format
- Ensure PostgreSQL is running
- Check container logs: `docker-compose logs db`
