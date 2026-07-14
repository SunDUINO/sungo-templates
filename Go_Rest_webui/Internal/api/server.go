package api

import (
	
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"Go_REST_webui/src/Internal/store"
)

// Server grupuje zależności potrzebne handlerom (magazyn danych, czas startu, liczniki).
type Server struct {
	store        *store.Store
	startedAt    time.Time
	requestCount uint64
}

func NewServer(st *store.Store, startedAt time.Time) *Server {
	return &Server{store: st, startedAt: startedAt}
}

// ---------- Middleware ----------

// WithRequestCounter zlicza każde zapytanie HTTP (widoczne potem w /api/status).
func WithRequestCounter(s *Server, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&s.requestCount, 1)
		next.ServeHTTP(w, r)
	})
}

// WithLogging loguje metodę, ścieżkę i czas obsługi zapytania.
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

// ---------- Pomocnicze ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ---------- Health & Status ----------

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) Status(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	writeJSON(w, http.StatusOK, map[string]any{
		"uptimeSeconds": int(time.Since(s.startedAt).Seconds()),
		"startedAt":     s.startedAt.Format(time.RFC3339),
		"requestCount":  atomic.LoadUint64(&s.requestCount),
		"itemCount":     len(s.store.List()),
		"goroutines":    runtime.NumGoroutine(),
		"memAllocMB":    mem.Alloc / 1024 / 1024,
		"goVersion":     runtime.Version(),
	})
}

// ---------- CRUD /api/items ----------

type itemPayload struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (s *Server) ListItems(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

func (s *Server) CreateItem(w http.ResponseWriter, r *http.Request) {
	var p itemPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "nieprawidłowe dane JSON")
		return
	}
	if p.Name == "" {
		writeError(w, http.StatusBadRequest, "pole 'name' jest wymagane")
		return
	}
	if p.Status == "" {
		p.Status = "nowy"
	}
	it := s.store.Create(p.Name, p.Status)
	writeJSON(w, http.StatusCreated, it)
}

func (s *Server) GetItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	it, err := s.store.Get(id)
	if err != nil {
		s.handleStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (s *Server) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var p itemPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "nieprawidłowe dane JSON")
		return
	}
	it, err := s.store.Update(id, p.Name, p.Status)
	if err != nil {
		s.handleStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (s *Server) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.store.Delete(id); err != nil {
		s.handleStoreErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStoreErr(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "nie znaleziono elementu")
		return
	}
	writeError(w, http.StatusInternalServerError, "błąd wewnętrzny serwera")
}