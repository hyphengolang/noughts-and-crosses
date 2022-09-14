package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Pool struct {
	mu sync.Mutex

	ID uuid.UUID

	msgs chan any
	cs   map[uuid.UUID]*Conn
}

func NewPool() *Pool {
	p := &Pool{
		ID:   uuid.New(),
		cs:   make(map[uuid.UUID]*Conn),
		msgs: make(chan any),
	}

	go p.listen()

	return p
}

func (p *Pool) Close() error {
	for uid, c := range p.cs {
		p.Remove(uid)
		c.Close()
	}
	log.Println("closed")
	return nil
}

func (p *Pool) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.cs)
}

func (p *Pool) Add(w http.ResponseWriter, r *http.Request, u *websocket.Upgrader) (*Conn, error) {
	// p.mu.Lock()
	// defer p.mu.Unlock()

	rwc, err := u.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	c := &Conn{
		ID:  uuid.New(),
		rwc: rwc,
		p:   p,
	}

	p.cs[c.ID] = c

	log.Println("New size", p.Size())

	return c, nil
}

func (p *Pool) Remove(uid uuid.UUID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.cs, uid)
}

func (p *Pool) listen() error {
	// send to all connections via goroutines

	for msg := range p.msgs {
		for _, c := range p.cs {
			go c.WriteJSON(msg)
		}
	}

	return nil
}
