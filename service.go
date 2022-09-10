package main

import (
	"net/http"

	"github.com/google/uuid"

	h "github.com/hyphengolang/prelude/http"
	t "github.com/hyphengolang/prelude/template"
)

type service struct {
	m *http.ServeMux
	p *pool
}

func (s service) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.m.ServeHTTP(w, r) }

func NewService(mux *http.ServeMux) *service {
	p := &pool{
		id:    uuid.New(),
		msgs:  make(chan int),
		conns: make(map[uuid.UUID]*conn),
	}

	s := &service{mux, p}

	go s.routes()

	return s
}

func (s service) routes() {
	s.m.Handle("/assets/", h.FileServer("/assets/", "assets"))
	s.m.HandleFunc("/", s.fileHandler("index.html"))
	s.m.HandleFunc("/ws", upgradeHTTP(s.handleEcho))
}

func (s service) fileHandler(path string) http.HandlerFunc {
	render, err := t.Render(path)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, nil)
	}
}
