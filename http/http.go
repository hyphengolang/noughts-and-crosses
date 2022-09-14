package http

import (
	"context"
	"net/http"

	"com.adoublef.websocket/http/ws"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	h "github.com/hyphengolang/prelude/http"
)

type contextKey struct{ string }

func (c *contextKey) String() string { return "context value " + c.string }

var (
	poolKey    = &contextKey{"ws-pool"}
	upgradeKey = &contextKey{"http-upgrade"}
)

func chain(hf http.HandlerFunc, mw ...h.MiddleWare) http.HandlerFunc { return h.Chain(hf, mw...) }

func (s Service) sessionPool(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			s.respond(w, r, err, http.StatusBadRequest)
			return
		}

		p, err := s.c.Get(uid)
		if err != nil {
			s.l.Println(err)
			return
		}

		s.l.Println("This session matches the ID", p.ID)

		r = r.WithContext(context.WithValue(r.Context(), poolKey, p))
		f(w, r)
	}
}

func (s Service) upgradeHTTP(f http.HandlerFunc) http.HandlerFunc {
	u := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	// p := ws.NewPool()

	return func(w http.ResponseWriter, r *http.Request) {
		p, ok := r.Context().Value(poolKey).(*ws.Pool)
		if !ok {
			s.l.Println("error: connection does not exist")
			return
		}

		if p.Size() == 2 {
			s.l.Println("maximum number of connections reached")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		c, err := p.Add(w, r, u)
		if err != nil {
			s.l.Println(err)
			// w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), upgradeKey, c))
		f(w, r)
	}
}
