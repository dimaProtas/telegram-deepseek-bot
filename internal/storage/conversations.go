package storage

import (
	"sync"
	"time"
)

type Message struct {
	Role    string
	Content string
	Time    time.Time
}

type Conversation struct {
	Messages []Message
	LastSeen time.Time
	Model    string
	Tokens   int
}

func (s *MemoryStorage) AddTokens(chatID int64, tokens int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv, ok := s.conversations[chatID]
	if !ok {
		conv = &Conversation{
			Messages: []Message{},
		}
		s.conversations[chatID] = conv
	}
	conv.Tokens += tokens
}

func (s *MemoryStorage) GetTokens(chatID int64) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conv, ok := s.conversations[chatID]
	if !ok {
		return 0
	}
	return conv.Tokens
}


type MemoryStorage struct {
	mu            sync.RWMutex
	conversations map[int64]*Conversation
	ttl           time.Duration
}

func NewMemoryStorage(ttlHours int) *MemoryStorage {
	return &MemoryStorage{
		conversations: make(map[int64]*Conversation),
		ttl:           time.Duration(ttlHours) * time.Hour,
	}
}

func (s *MemoryStorage) AddMessage(chatID int64, role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv, ok := s.conversations[chatID]
	if !ok {
		conv = &Conversation{
			Messages: []Message{},
		}
		s.conversations[chatID] = conv
	}

	conv.Messages = append(conv.Messages, Message{
		Role:    role,
		Content: content,
		Time:    time.Now(),
	})
	conv.LastSeen = time.Now()
}

func (s *MemoryStorage) GetModel(chatID int64, defaultModel string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conv, ok := s.conversations[chatID]
	if !ok {
		return defaultModel
	}
	if conv.Model == "" {
		return defaultModel
	}
	return conv.Model
}

func (s *MemoryStorage) SetModel(chatID int64, model string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv, ok := s.conversations[chatID]
	if !ok {
		conv = &Conversation{
			Messages: []Message{},
		}
		s.conversations[chatID] = conv
	}
	conv.Model = model
	conv.LastSeen = time.Now()
}

func (s *MemoryStorage) GetMessages(chatID int64) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conv, ok := s.conversations[chatID]
	if !ok {
		return nil
	}

	return conv.Messages
}

func (s *MemoryStorage) Clear(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conversations, chatID)
}

func (s *MemoryStorage) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, conv := range s.conversations {
		if now.Sub(conv.LastSeen) > s.ttl {
			delete(s.conversations, id)
		}
	}
}
