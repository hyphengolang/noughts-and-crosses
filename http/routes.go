package http

import (
	"net/http"

	"com.adoublef.websocket/http/ws"
	"com.adoublef.websocket/internal"

	v4 "github.com/lithammer/shortuuid/v4"
)

func (s Service) apiCreateGame(w http.ResponseWriter, r *http.Request) {
	uid, _ := s.c.NewPool()

	s.l.Printf("create uid: %s\n", uid)

	s.respond(w, r, v4.DefaultEncoder.Encode(uid), http.StatusOK)
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

		c.Send(internal.Fibonacci(n))
		// if err := c.WriteJSON(internal.Fibonacci(n)); err != nil {
		// 	s.l.Println(err)
		// 	return
		// }
	}

}
