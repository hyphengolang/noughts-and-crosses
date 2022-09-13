package ws

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Conn struct {
	ID uuid.UUID

	rwc *websocket.Conn
	p   *Pool
}

func (c Conn) Close() error {
	c.p.Remove(c.ID)
	return c.rwc.Close()
}

func (c Conn) ReadJSON(v any) error { return c.rwc.ReadJSON(v) }

func (c Conn) WriteJSON(v any) error { return c.rwc.WriteJSON(v) }

func (c Conn) Send(v any) { c.p.msgs <- v }

func (c *Conn) join(p *Pool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cs[c.ID] = c

	msg := struct {
		PoolID string
		ConnID string
	}{
		PoolID: p.ID.String(),
		ConnID: c.ID.String(),
	}

	return c.WriteJSON(msg)
}
