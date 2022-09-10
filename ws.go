package main

import (
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type pool struct {
	mu sync.Mutex

	id uuid.UUID

	// msgs is a channel of messages to send to all connections
	msgs chan int

	// conns is a map of all connections in the pool
	conns map[uuid.UUID]*conn
}

func (p *pool) listen() error {
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

func (p *pool) remove(id uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.conns, id)
}

type conn struct {
	id  uuid.UUID
	rwc *websocket.Conn
	p   *pool
}

func (c *conn) serve() error {
	defer c.close()

	for {
		var n int
		if err := c.readJSON(&n); err != nil {
			return err
		}
		c.p.msgs <- fib(n)
	}
}

func (c conn) close() error {
	c.p.remove(c.id)
	c.rwc.Close()
	return nil
}

func (c conn) readJSON(v any) error { return c.rwc.ReadJSON(v) }

func (c conn) writeJSON(v any) error { return c.rwc.WriteJSON(v) }

func (c *conn) join(p *pool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.conns[c.id] = c

	msg := struct {
		PoolID string
		ConnID string
	}{
		PoolID: p.id.String(),
		ConnID: c.id.String(),
	}

	return c.writeJSON(msg)
}
