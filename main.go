package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	"github.com/google/uuid"
)

type contextKey struct{ string }

func (c *contextKey) String() string { return "context value " + c.string }

var upgradeKey = &contextKey{"upgrade-http"}

type pool struct {
	id uuid.UUID

	msgs chan int

	conns map[uuid.UUID]*conn
}

func (p pool) listen() error {
	for {
		select {
		case msg := <-p.msgs:
			for _, c := range p.conns {
				log.Printf("Sending %d to %s", msg, c.id)
				if err := c.writeJSON(msg); err != nil {
					return err
				}
			}
		}
	}
	// return nil
}

func (p pool) broadcast(msg any) error {
	for _, c := range p.conns {
		log.Printf("Sending %d to %s", msg, c.id)
		if err := c.writeJSON(msg); err != nil {
			return err
		}
	}

	return nil
}

type conn struct {
	id uuid.UUID

	rwc net.Conn

	r *wsutil.Reader
	w *wsutil.Writer

	p *pool
}

func (c conn) readJSON(v any) error { return json.NewDecoder(c.r).Decode(v) }

func (c conn) writeJSON(v any) error { return json.NewEncoder(c.w).Encode(v) }

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
	h *http.ServeMux
}

func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.h.ServeHTTP(w, r)
}

func NewService(mux *http.ServeMux) *Service {
	s := &Service{
		h: mux,
	}
	s.h.HandleFunc("/ws", s.upgradeHTTP(s.handleEcho))
	return s
}

func (s Service) upgradeHTTP(f http.HandlerFunc) http.HandlerFunc {
	p := &pool{
		id:    uuid.New(),
		msgs:  make(chan int),
		conns: make(map[uuid.UUID]*conn),
	}

	// go p.listen()

	// north bridge, perth

	return func(w http.ResponseWriter, r *http.Request) {
		rwc, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		c := &conn{
			id:  uuid.New(),
			rwc: rwc,
			r:   wsutil.NewReader(rwc, ws.StateServerSide),
			w:   wsutil.NewWriter(rwc, ws.StateServerSide, ws.OpText),
			p:   p,
		}

		p.conns[c.id] = c // add to pool
		log.Printf("Added %s to pool %d", c.id, len(p.conns))

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
		h, err := c.r.NextFrame()
		if err != nil {
			return err
		}
		if h.OpCode == ws.OpClose {
			return io.EOF
		}

		var n int
		if err := c.readJSON(&n); err != nil {
			return err
		}

		if err := c.w.Flush(); err != nil {
			return err
		}

		c.p.broadcast(fib(n))
	}

}

// fibinacci returns the nth fib number
func fib(n int) int {
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}
