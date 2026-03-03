package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"langduel/internal/server"
	"langduel/internal/storage"
	"langduel/internal/ws"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env explicitly and override existing env vars.
	_ = godotenv.Overload(".env")
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := storage.Open(context.Background(), dsn)
		if err != nil {
			log.Fatalf("DB open error: %v", err)
		}
		defer db.Close()
		repo := storage.NewDuelRepo(db)
		ws.SetRepo(repo)
		server.SetRepo(repo)
		log.Println("DB enabled")
		log.Println("DB URL (redacted):", redactDSN(dsn))
	} else {
		log.Println("DB disabled")
	}

	r := server.NewRouter()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func redactDSN(dsn string) string {
	// Hide password in logs: postgres://user:***@host:port/db
	const mask = "***"
	i := strings.Index(dsn, "://")
	if i == -1 {
		return dsn
	}
	rest := dsn[i+3:]
	at := strings.Index(rest, "@")
	if at == -1 {
		return dsn
	}
	cred := rest[:at]
	colon := strings.Index(cred, ":")
	if colon == -1 {
		return dsn
	}
	return dsn[:i+3+colon+1] + mask + dsn[i+3+at:]
}
