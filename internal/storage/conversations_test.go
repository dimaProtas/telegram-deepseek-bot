package storage

import (
	"sync"
	"testing"
	"time"
)

func TestNewMemoryStorage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ttlHours int
		wantTTL  time.Duration
	}{
		{
			name:     "zero_ttl",
			ttlHours: 0,
			wantTTL:  0,
		},
		{
			name:     "positive_ttl_24h",
			ttlHours: 24,
			wantTTL:  24 * time.Hour,
		},
		{
			name:     "large_ttl_168h",
			ttlHours: 168,
			wantTTL:  168 * time.Hour,
		},
		{
			name:     "negative_ttl",
			ttlHours: -1,
			wantTTL:  -1 * time.Hour,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := NewMemoryStorage(tt.ttlHours)

			if s == nil {
				t.Fatal("NewMemoryStorage returned nil")
			}
			if s.ttl != tt.wantTTL {
				t.Errorf("ttl = %v, want %v", s.ttl, tt.wantTTL)
			}
			if s.conversations == nil {
				t.Error("conversations map is nil")
			}
			if len(s.conversations) != 0 {
				t.Errorf("conversations map not empty, got %d entries", len(s.conversations))
			}
		})
	}
}

func TestAddMessage_NewConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.AddMessage(12345, "user", "Hello, AI!")

	conv, ok := s.conversations[12345]
	if !ok {
		t.Fatal("conversation was not created")
	}
	if len(conv.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(conv.Messages))
	}
	if conv.Messages[0].Role != "user" {
		t.Errorf("role = %q, want %q", conv.Messages[0].Role, "user")
	}
	if conv.Messages[0].Content != "Hello, AI!" {
		t.Errorf("content = %q, want %q", conv.Messages[0].Content, "Hello, AI!")
	}
	if conv.LastSeen.IsZero() {
		t.Error("LastSeen was not set")
	}
	if conv.Messages[0].Time.IsZero() {
		t.Error("Message.Time was not set")
	}
}

func TestAddMessage_ExistingConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.AddMessage(1, "user", "Message 1")
	s.AddMessage(1, "assistant", "Message 2")
	s.AddMessage(1, "user", "Message 3")

	conv, ok := s.conversations[1]
	if !ok {
		t.Fatal("conversation was not created")
	}
	if len(conv.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(conv.Messages))
	}

	expected := []struct {
		role    string
		content string
	}{
		{"user", "Message 1"},
		{"assistant", "Message 2"},
		{"user", "Message 3"},
	}

	for i, exp := range expected {
		if conv.Messages[i].Role != exp.role {
			t.Errorf("message %d: role = %q, want %q", i, conv.Messages[i].Role, exp.role)
		}
		if conv.Messages[i].Content != exp.content {
			t.Errorf("message %d: content = %q, want %q", i, conv.Messages[i].Content, exp.content)
		}
	}
}

func TestAddMessage_MultipleChats(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.AddMessage(1, "user", "Chat 1 message")
	s.AddMessage(2, "user", "Chat 2 message")

	if len(s.conversations) != 2 {
		t.Fatalf("expected 2 conversations, got %d", len(s.conversations))
	}
	if s.conversations[1] == nil || s.conversations[2] == nil {
		t.Fatal("one of the conversations is nil")
	}
	if s.conversations[1].Messages[0].Content != "Chat 1 message" {
		t.Error("chat 1 has wrong content")
	}
	if s.conversations[2].Messages[0].Content != "Chat 2 message" {
		t.Error("chat 2 has wrong content")
	}
}

func TestGetMessages_ExistingConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.AddMessage(1, "user", "Q1")
	s.AddMessage(1, "assistant", "A1")

	messages := s.GetMessages(1)
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
	if messages[0].Role != "user" || messages[0].Content != "Q1" {
		t.Error("first message is wrong")
	}
	if messages[1].Role != "assistant" || messages[1].Content != "A1" {
		t.Error("second message is wrong")
	}
}

func TestGetMessages_NonExistentConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	messages := s.GetMessages(999)

	if messages != nil {
		t.Errorf("expected nil for non-existent conversation, got %v", messages)
	}
}

func TestAddTokens_NewConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.AddTokens(1, 100)

	if s.conversations[1].Tokens != 100 {
		t.Errorf("tokens = %d, want 100", s.conversations[1].Tokens)
	}
}

func TestAddTokens_ExistingConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.AddTokens(1, 50)
	s.AddTokens(1, 30)
	s.AddTokens(1, 20)

	if s.conversations[1].Tokens != 100 {
		t.Errorf("tokens = %d, want 100", s.conversations[1].Tokens)
	}
}

func TestGetTokens_ExistingConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.AddTokens(1, 150)

	tokens := s.GetTokens(1)
	if tokens != 150 {
		t.Errorf("tokens = %d, want 150", tokens)
	}
}

func TestGetTokens_NonExistentConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	tokens := s.GetTokens(999)

	if tokens != 0 {
		t.Errorf("tokens = %d, want 0", tokens)
	}
}

func TestGetTokens_Accumulation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.AddTokens(1, 10)
	s.AddTokens(1, 20)
	s.AddTokens(1, 30)

	tokens := s.GetTokens(1)
	if tokens != 60 {
		t.Errorf("tokens = %d, want 60", tokens)
	}
}

func TestGetModel_Default(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	model := s.GetModel(1, "default-model")

	if model != "default-model" {
		t.Errorf("model = %q, want %q", model, "default-model")
	}
}

func TestGetModel_Set(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.SetModel(1, "deepseek-coder")
	model := s.GetModel(1, "default-model")

	if model != "deepseek-coder" {
		t.Errorf("model = %q, want %q", model, "deepseek-coder")
	}
}

func TestGetModel_EmptySet(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	// Add a message to create the conversation (model will be empty string)
	s.AddMessage(1, "user", "hello")
	model := s.GetModel(1, "deepseek-chat")

	if model != "deepseek-chat" {
		t.Errorf("model = %q, want %q (default)", model, "deepseek-chat")
	}
}

func TestSetModel_NewConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.SetModel(10, "reasoner")

	conv := s.conversations[10]
	if conv == nil {
		t.Fatal("conversation was not created")
	}
	if conv.Model != "reasoner" {
		t.Errorf("model = %q, want %q", conv.Model, "reasoner")
	}
	if conv.LastSeen.IsZero() {
		t.Error("LastSeen was not updated")
	}
}

func TestSetModel_ExistingConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.SetModel(1, "chat")
	s.SetModel(1, "coder")

	if s.conversations[1].Model != "coder" {
		t.Errorf("model = %q, want %q", s.conversations[1].Model, "coder")
	}
}

func TestClear_ExistingConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.AddMessage(1, "user", "hello")
	s.AddTokens(1, 50)

	s.Clear(1)

	if _, ok := s.conversations[1]; ok {
		t.Error("conversation still exists after Clear")
	}
}

func TestClear_NonExistentConversation(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	// Should not panic
	s.Clear(999)

	if len(s.conversations) != 0 {
		t.Errorf("conversations map not empty after clearing non-existent, got %d entries", len(s.conversations))
	}
}

func TestClear_Twice(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(24)
	s.AddMessage(1, "user", "hello")
	s.Clear(1)
	s.Clear(1) // Should not panic

	if len(s.conversations) != 0 {
		t.Error("conversations map not empty after double clear")
	}
}

func TestCleanup_ExpiredConversations(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(1) // 1 hour TTL

	// Add conversations with different ages
	s.AddMessage(1, "user", "recent")
	s.AddMessage(2, "user", "old")
	s.AddMessage(3, "user", "very old")

	// Manually set LastSeen
	s.mu.Lock()
	s.conversations[2].LastSeen = time.Now().Add(-2 * time.Hour)     // expired
	s.conversations[3].LastSeen = time.Now().Add(-10 * time.Hour)    // expired
	s.conversations[1].LastSeen = time.Now()                         // fresh
	s.mu.Unlock()

	s.Cleanup()

	if len(s.conversations) != 1 {
		t.Errorf("expected 1 conversation after cleanup, got %d", len(s.conversations))
	}
	if _, ok := s.conversations[1]; !ok {
		t.Error("fresh conversation was incorrectly cleaned up")
	}
}

func TestCleanup_NoExpiredConversations(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(1)
	s.AddMessage(1, "user", "recent")
	s.AddMessage(2, "user", "also recent")

	s.Cleanup()

	if len(s.conversations) != 2 {
		t.Errorf("expected 2 conversations, got %d", len(s.conversations))
	}
}

func TestCleanup_AllExpired(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(1)
	s.AddMessage(1, "user", "old")
	s.AddMessage(2, "user", "old")

	s.mu.Lock()
	s.conversations[1].LastSeen = time.Now().Add(-2 * time.Hour)
	s.conversations[2].LastSeen = time.Now().Add(-3 * time.Hour)
	s.mu.Unlock()

	s.Cleanup()

	if len(s.conversations) != 0 {
		t.Errorf("expected 0 conversations, got %d", len(s.conversations))
	}
}

func TestCleanup_EmptyStorage(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(1)
	s.Cleanup() // Should not panic

	if len(s.conversations) != 0 {
		t.Error("expected empty conversations map")
	}
}

func TestCleanup_ZeroTTL(t *testing.T) {
	t.Parallel()

	s := NewMemoryStorage(0) // 0 TTL means immediate expiry
	s.AddMessage(1, "user", "message")

	// Even freshly added should expire with 0 TTL because
	// now.Sub(now) = 0, and 0 > 0 is false... 
	// Actually 0 > 0 is false, so nothing expires with 0 TTL.
	// Let's test with negative TTL.
	s2 := NewMemoryStorage(-1)
	s2.AddMessage(1, "user", "message")
	s2.Cleanup()

	if len(s2.conversations) != 0 {
		t.Errorf("expected 0 conversations with negative TTL, got %d", len(s2.conversations))
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := NewMemoryStorage(24)

	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 2) // readers + writers

	// Writers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				chatID := int64(id % 5) // 5 different chats
				s.AddMessage(chatID, "user", "message")
				s.AddTokens(chatID, 1)
			}
		}(i)
	}

	// Readers
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				chatID := int64(id % 5)
				_ = s.GetMessages(chatID)
				_ = s.GetTokens(chatID)
				_ = s.GetModel(chatID, "default")
			}
		}(i)
	}

	wg.Wait()

	// Verify no data corruption — each chat should have tokens
	for chatID := int64(0); chatID < 5; chatID++ {
		tokens := s.GetTokens(chatID)
		if tokens <= 0 {
			t.Errorf("chat %d has %d tokens, expected > 0", chatID, tokens)
		}
	}
}

func TestConcurrentClearAndAccess(t *testing.T) {
	s := NewMemoryStorage(24)

	s.AddMessage(1, "user", "initial")

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			s.AddMessage(1, "user", "msg")
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			s.GetMessages(1)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			s.Clear(1)
			s.AddMessage(1, "user", "after clear")
		}
	}()

	wg.Wait()
	// Should not panic — just verifies no data races
}
