package ws

import (
	"log"

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

func (c Conn) Send(v any) {
	log.Println("size of pool is", c.p.Size())
	c.p.msgs <- v
}
