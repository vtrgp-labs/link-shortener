package main

import (
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

//go:embed static/index.html
var staticFS embed.FS

var indexTmpl = template.Must(template.ParseFS(staticFS, "static/index.html"))

// Server holds the HTTP handlers.
type Server struct {
	store *Store
}

// NewServer returns an http.Handler with all routes registered.
func NewServer(store *Store) http.Handler {
	s := &Server{store: store}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("POST /shorten", s.handleShorten)
	mux.HandleFunc("GET /{code}", s.handleRedirect)
	return mux
}

type indexData struct {
	ShortURL string
	Error    string
	Links    []Link
	BaseURL  string
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, status int, data indexData) {
	links, err := s.store.List()
	if err != nil {
		log.Printf("list links: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data.Links = links
	data.BaseURL = baseURL(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := indexTmpl.Execute(w, data); err != nil {
		log.Printf("render index: %v", err)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, http.StatusOK, indexData{})
}

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(r.FormValue("url"))
	if !validURL(raw) {
		s.render(w, r, http.StatusBadRequest, indexData{Error: "Please enter a valid http(s) URL."})
		return
	}
	code, err := s.store.Create(raw)
	if err != nil {
		log.Printf("create link: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	s.render(w, r, http.StatusOK, indexData{ShortURL: baseURL(r) + "/" + code})
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	target, err := s.store.Visit(r.PathValue("code"))
	if errors.Is(err, ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("visit link: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}

func validURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
