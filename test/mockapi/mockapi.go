package mockapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// Request is one recorded call, captured before dispatch so tests can assert
// on exactly what the CLI sent.
type Request struct {
	Method   string
	Path     string
	Token    string
	TenantID string
	Body     []byte
}

// Server is a canned-response Management API double. Unrouted paths return a
// FastAPI-style 404 ({"detail": "Not Found"}), matching the real backend.
type Server struct {
	*httptest.Server

	mu       sync.Mutex
	requests []Request
	routes   map[string]http.HandlerFunc
}

func New() *Server {
	s := &Server{routes: map[string]http.HandlerFunc{}}
	s.Server = httptest.NewServer(http.HandlerFunc(s.dispatch))
	return s
}

// Handle registers a canned JSON response for METHOD path.
func (s *Server) Handle(method, path string, status int, body any) {
	s.HandleFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, status, body)
	})
}

// HandleFunc registers a custom handler for METHOD path.
func (s *Server) HandleFunc(method, path string, h http.HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes[method+" "+path] = h
}

// Requests returns a copy of all recorded requests.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Request(nil), s.requests...)
}

// RequestsTo returns the recorded requests matching METHOD path.
func (s *Server) RequestsTo(method, path string) []Request {
	var out []Request
	for _, r := range s.Requests() {
		if r.Method == method && r.Path == path {
			out = append(out, r)
		}
	}
	return out
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (s *Server) dispatch(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	s.mu.Lock()
	s.requests = append(s.requests, Request{
		Method:   r.Method,
		Path:     r.URL.Path,
		Token:    strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "),
		TenantID: r.Header.Get("X-Tenant-ID"),
		Body:     body,
	})
	h := s.routes[r.Method+" "+r.URL.Path]
	s.mu.Unlock()

	if h == nil {
		WriteJSON(w, http.StatusNotFound, map[string]string{"detail": "Not Found"})
		return
	}
	h(w, r)
}
