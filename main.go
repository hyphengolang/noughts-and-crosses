package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	s := NewService(mux)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", s))
}

// fibinacci returns the nth fib number
func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}
