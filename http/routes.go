package http

import (
	"net/http"

	"com.adoublef.websocket/http/ws"
	"com.adoublef.websocket/internal"
)

func (s Service) apiCreateGame(w http.ResponseWriter, r *http.Request) {
	uid, _ := s.c.NewPool()

	s.respond(w, r, uid, http.StatusOK)
}

func (s Service) handlePlayGame(w http.ResponseWriter, r *http.Request) {
	c, ok := r.Context().Value(upgradeKey).(*ws.Conn)
	if !ok {
		s.l.Println("error: connection does not exist")
		return
	}

	defer c.Close()

	for {
		var n int
		if err := c.ReadJSON(&n); err != nil {
			s.l.Println(err)
			return
		}

		s.l.Println("Value starts at 0")

		c.WriteJSON(internal.Fibonacci(n))
	}

}
