package main

import (
	"log"
	"net/http"

	"langduel/internal/server"
)

func main() {
	r := server.NewRouter()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}