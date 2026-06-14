package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type record struct {
	UserID    int64
	Email     string
	ExpiresAt time.Time
}

var (
	mu    sync.RWMutex
	store = map[string]record{}
)

func Create(userID int64, email string) string {
	id := uuid.NewString()

	mu.Lock()
	store[id] = record{
		UserID:    userID,
		Email:     email,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	mu.Unlock()

	return id
}

func Validate(id string) (int64, string, bool) {
	mu.RLock()
	r, ok := store[id]
	mu.RUnlock()

	if !ok {
		return 0, "", false
	}

	if time.Now().After(r.ExpiresAt) {
		mu.Lock()
		delete(store, id)
		mu.Unlock()
		return 0, "", false
	}

	return r.UserID, r.Email, true
}

func Destroy(id string) {
	mu.Lock()
	delete(store, id)
	mu.Unlock()
}
