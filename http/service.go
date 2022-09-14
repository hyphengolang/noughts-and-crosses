package http

import (
	"log"
	"net/http"

	"com.adoublef.websocket/http/ws"

	"github.com/go-chi/chi/v5"
	h "github.com/hyphengolang/prelude/http"
	t "github.com/hyphengolang/prelude/template"
	"github.com/lithammer/shortuuid/v4"
)

type Service struct {
	m chi.Router
	c *ws.Client

	l *log.Logger
}

func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.m.ServeHTTP(w, r) }

func NewService(mux chi.Router) *Service {
	s := &Service{mux, ws.NewClient(), log.Default()}

	go s.routes()

	return s
}

func (s Service) routes() {
	// http
	s.m.Handle("/assets/*", s.handleFiles("/assets/", "assets"))
	s.m.HandleFunc("/", s.viewIndex("pages/index.html"))
	s.m.HandleFunc("/play/{id}", s.viewPlay("pages/play.html"))

	// api
	s.m.Get("/api/game/create", s.apiCreateGame)

	// ws
	s.m.HandleFunc("/api/game/play/{id}", chain(s.handlePlayGame, s.upgradeHTTP, s.sessionPool))
}

func (s Service) respond(w http.ResponseWriter, r *http.Request, data any, status int) {
	h.Respond(w, r, data, status)
}

func (s Service) handleFiles(prefix string, dirname string) http.Handler {
	return h.FileServer(prefix, dirname)
}

func (s Service) viewIndex(path string) http.HandlerFunc {
	render, err := t.Render(path)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, nil)
	}
}

func (s Service) viewPlay(path string) http.HandlerFunc {
	render, err := t.Render(path)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := shortuuid.DefaultEncoder.Decode(chi.URLParam(r, "id"))
		if err != nil {
			// can render page with error infomation
			s.respond(w, r, err, http.StatusBadRequest)
			return
		}

		if _, err = s.c.Get(uid); err != nil {
			// can render page with error infomation
			s.respond(w, r, err, http.StatusInternalServerError)
			return
		}

		render(w, r, nil)
	}
}
