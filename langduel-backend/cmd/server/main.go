package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"langduel/internal/server"
	"langduel/internal/storage"
	"langduel/internal/ws"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Overload(".env")
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err := storage.Open(context.Background(), dsn)
		if err != nil {
			log.Fatalf("DB open error: %v", err)
		}
		defer db.Close()
		repo := storage.NewDuelRepo(db)

		if err := repo.EnsureSchemaFixes(context.Background()); err != nil {
			log.Printf("Warning: schema fixes failed: %v", err)
		}

		if err := repo.EnsureAIPhraseTable(context.Background()); err != nil {
			log.Printf("Warning: failed to create ai_phrases table: %v", err)
		}

		ws.SetRepo(repo)
		server.SetRepo(repo)
		log.Println("DB enabled")
		log.Println("DB URL (redacted):", redactDSN(dsn))

		if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
			log.Println("OpenRouter API key configured")
		} else {
			log.Println("Warning: OPENROUTER_API_KEY not set, AI generation disabled")
		}
	} else {
		log.Println("DB disabled")
	}

	r := server.NewRouter()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
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
