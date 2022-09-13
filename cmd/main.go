package main

import (
	"log"
	"net/http"

	h "com.adoublef.websocket/http"
	"github.com/go-chi/chi/v5"
)

func main() {
	mux := chi.NewMux()
	s := h.NewService(mux)
	log.Println("Starting server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", s))
}
