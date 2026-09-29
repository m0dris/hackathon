package core_max_session

import "sync"

type Session struct {
	Step  string
	Draft map[string]string
}

type Store interface {
	Get(chatID int64) (Session, bool)
	Set(chatID int64, s Session)
	Clear(chatID int64)
}

type memoryStore struct {
	mu   sync.RWMutex
	data map[int64]Session
}

func NewMemoryStore() Store {
	return &memoryStore{data: make(map[int64]Session)}
}

func (s *memoryStore) Get(chatID int64) (Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, ok := s.data[chatID]

	return sess, ok
}

func (s *memoryStore) Set(chatID int64, sess Session) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[chatID] = sess
}

func (s *memoryStore) Clear(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, chatID)
}
