package main

import (
	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type contextKey struct{ string }

func (c *contextKey) String() string { return "context value " + c.string }

var upgradeKey = &contextKey{"upgrade-http"}

func upgradeHTTP(f http.HandlerFunc) http.HandlerFunc {
	ws := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	p := &pool{
		id:    uuid.New(),
		msgs:  make(chan int),
		conns: make(map[uuid.UUID]*conn),
	}

	// fmt.Println("create pool") log
	go p.listen()

	return func(w http.ResponseWriter, r *http.Request) {
		if len(p.conns) == 2 {
			log.Println("maximum number of connections reached")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rwc, err := ws.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			// w.WriteHeader(http.StatusInternalServerError)
			return
		}

		c := &conn{uuid.New(), rwc, p}

		go c.join(p)

		r = r.WithContext(context.WithValue(r.Context(), upgradeKey, c))
		f(w, r)
	}
}

func (s service) handleEcho(w http.ResponseWriter, r *http.Request) {
	go r.Context().Value(upgradeKey).(*conn).serve()
}
