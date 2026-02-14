package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
