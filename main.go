package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store, err := NewStore("links.db")
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer store.Close()

	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, NewServer(store)))
}
