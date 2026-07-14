package store

import (
	"errors"
	"sync"
	"time"
)

// ErrNotFound zwracany, gdy element o podanym ID nie istnieje.
var ErrNotFound = errors.New("element nie znaleziony")

// Item reprezentuje pojedynczy rekord zarządzany przez usługę.
// To generyczny model — w realnym projekcie zastąp go swoją domeną
// (np. Task, Device, Job, Ticket...).
type Item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Store to bezpieczny wątkowo magazyn danych w pamięci.
type Store struct {
	mu     sync.RWMutex
	items  map[string]Item
	nextID int
}

func New() *Store {
	return &Store{
		items: make(map[string]Item),
	}
}

func (s *Store) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Item, 0, len(s.items))
	for _, it := range s.items {
		out = append(out, it)
	}
	return out
}

func (s *Store) Get(id string) (Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	it, ok := s.items[id]
	if !ok {
		return Item{}, ErrNotFound
	}
	return it, nil
}

func (s *Store) Create(name, status string) Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	id := itoa(s.nextID)
	now := time.Now()
	it := Item{
		ID:        id,
		Name:      name,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.items[id] = it
	return it
}

func (s *Store) Update(id, name, status string) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	it, ok := s.items[id]
	if !ok {
		return Item{}, ErrNotFound
	}
	if name != "" {
		it.Name = name
	}
	if status != "" {
		it.Status = status
	}
	it.UpdatedAt = time.Now()
	s.items[id] = it
	return it, nil
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}