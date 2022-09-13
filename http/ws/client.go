package ws

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

type Client struct {
	mu sync.Mutex

	ps map[uuid.UUID]*Pool
}

func NewClient() *Client {
	c := &Client{
		ps: make(map[uuid.UUID]*Pool),
	}

	return c
}

func (c *Client) Size() int { return len(c.ps) }

func (c *Client) Close() error {
	return nil
}

func (c *Client) NewPool() (uuid.UUID, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	p := NewPool()

	c.ps[p.ID] = p

	return p.ID, nil
}

func (c *Client) Get(uid uuid.UUID) (*Pool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, p := range c.ps {
		if id == uid {
			return p, nil
		}
	}

	return nil, errors.New("ws: pool does not exist")
}

func (c *Client) Has(uid uuid.UUID) bool {
	if _, err := c.Get(uid); err != nil {
		return false
	}
	return true
}
