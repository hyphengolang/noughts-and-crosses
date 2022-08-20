package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/google/uuid"
)

type contextKey struct{ string }

func (c *contextKey) String() string { return "context value " + c.string }

var upgradeKey = &contextKey{"upgrade-http"}

type pool struct {
	id uuid.UUID

	// msgs is a channel of messages to send to all connections
	msgs chan int

	// conns is a map of all connections in the pool
	conns map[uuid.UUID]*conn
}

func (p pool) listen() error {
	// send to all connections via goroutines
	for msg := range p.msgs {
		for _, c := range p.conns {
			go func(c *conn) error {
				log.Printf("Sending %d to %s", msg, c.id)
				if err := c.writeJSON(msg); err != nil {
					return err
				}
				return nil
			}(c)
		}
	}

	return nil
}

type conn struct {
	id uuid.UUID

	rwc *websocket.Conn

	p *pool
}

func (c conn) readJSON(v any) error { return c.rwc.ReadJSON(v) }

func (c conn) writeJSON(v any) error { return c.rwc.WriteJSON(v) }

func main() {
	mux := http.NewServeMux()
	s := NewService(mux)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", s))
}

func getConn[T any](r *http.Request) T {
	return r.Context().Value(upgradeKey).(T)
}

type Service struct {
	h  *http.ServeMux
	ps []*pool
}

func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.h.ServeHTTP(w, r)
}

func NewService(mux *http.ServeMux) *Service {
	s := &Service{
		h:  mux,
		ps: make([]*pool, 0),
	}
	s.h.HandleFunc("/ws", s.upgradeHTTP(s.handleEcho))
	return s
}

func (s Service) joinPool(c *conn) {
	s.ps[0].conns[c.id] = c
}

func (s Service) upgradeHTTP(f http.HandlerFunc) http.HandlerFunc {
	var ws = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	p := &pool{
		id:    uuid.New(),
		msgs:  make(chan int),
		conns: make(map[uuid.UUID]*conn),
	}

	go p.listen()

	return func(w http.ResponseWriter, r *http.Request) {
		rwc, err := ws.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		c := &conn{uuid.New(), rwc, p}

		go s.joinPool(c) // add to pool

		r = r.WithContext(context.WithValue(r.Context(), upgradeKey, c))
		f(w, r)
	}
}

func (s Service) handleEcho(w http.ResponseWriter, r *http.Request) {
	c := getConn[*conn](r)

	go c.serve()
}

func (c *conn) serve() error {
	defer func() {
		delete(c.p.conns, c.id)
		c.rwc.Close()
	}()

	for {
		var n int
		if err := c.readJSON(&n); err != nil {
			return err
		}
		c.p.msgs <- fib(n)
	}
}

// fibinacci returns the nth fib number
func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}
