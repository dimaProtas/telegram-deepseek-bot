package deepseek

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testAPIKey = "sk-test-mock-key"

func newTestClient(serverURL string) *Client {
	return &Client{
		apiKey:  testAPIKey,
		baseURL: serverURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func TestNewClient_WithoutProxy(t *testing.T) {
	t.Parallel()

	client := NewClient("key-123", "https://api.example.com", "", "", "")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.apiKey != "key-123" {
		t.Errorf("apiKey = %q, want %q", client.apiKey, "key-123")
	}
	if client.baseURL != "https://api.example.com" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://api.example.com")
	}
	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}
}

func TestNewClient_WithProxy(t *testing.T) {
	t.Parallel()

	// With an invalid proxy address, it should fall back to direct connection
	client := NewClient("key-456", "https://api.example.com", "invalid:1080", "", "")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.apiKey != "key-456" {
		t.Errorf("apiKey = %q, want %q", client.apiKey, "key-456")
	}
	if client.httpClient == nil {
		t.Error("httpClient is nil even with invalid proxy (should fallback)")
	}
}

func TestChat_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected /chat/completions, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer "+testAPIKey {
			t.Errorf("expected Authorization Bearer %s, got %s", testAPIKey, r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "Hello! How can I help you?"
					},
					"finish_reason": "stop"
				}
			],
			"usage": {
				"prompt_tokens": 10,
				"completion_tokens": 7,
				"total_tokens": 17
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	req := ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "user", Content: "Hi!"},
		},
		Temp:      0.7,
		MaxTokens: 4096,
		Stream:    false,
	}

	resp, err := client.Chat(req)
	if err != nil {
		t.Fatalf("Chat() unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("Chat() returned nil response")
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Role != "assistant" {
		t.Errorf("role = %q, want %q", resp.Choices[0].Message.Role, "assistant")
	}
	if resp.Choices[0].Message.Content != "Hello! How can I help you?" {
		t.Errorf("content = %q, want %q", resp.Choices[0].Message.Content, "Hello! How can I help you?")
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Errorf("finish_reason = %q, want %q", resp.Choices[0].FinishReason, "stop")
	}
	if resp.Usage.TotalTokens != 17 {
		t.Errorf("total_tokens = %d, want 17", resp.Usage.TotalTokens)
	}
	if resp.Usage.PromptTokens != 10 {
		t.Errorf("prompt_tokens = %d, want 10", resp.Usage.PromptTokens)
	}
	if resp.Usage.CompletionTokens != 7 {
		t.Errorf("completion_tokens = %d, want 7", resp.Usage.CompletionTokens)
	}
}

func TestChat_Success_EmptyChoices(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [],
			"usage": {
				"prompt_tokens": 5,
				"completion_tokens": 0,
				"total_tokens": 5
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	resp, err := client.Chat(ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	})

	if err != nil {
		t.Fatalf("Chat() unexpected error: %v", err)
	}
	if len(resp.Choices) != 0 {
		t.Errorf("expected 0 choices, got %d", len(resp.Choices))
	}
}

func TestChat_APIError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "unauthorized_401",
			statusCode: http.StatusUnauthorized,
			body:       `{"error": "Invalid API key"}`,
		},
		{
			name:       "rate_limited_429",
			statusCode: http.StatusTooManyRequests,
			body:       `{"error": "Rate limit exceeded"}`,
		},
		{
			name:       "server_error_500",
			statusCode: http.StatusInternalServerError,
			body:       `{"error": "Internal server error"}`,
		},
		{
			name:       "bad_request_400",
			statusCode: http.StatusBadRequest,
			body:       `{"error": "Bad request"}`,
		},
		{
			name:       "not_found_404",
			statusCode: http.StatusNotFound,
			body:       `{"error": "Not found"}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := newTestClient(server.URL)
			_, err := client.Chat(ChatRequest{
				Model: "deepseek-chat",
				Messages: []Message{
					{Role: "user", Content: "Test"},
				},
			})

			if err == nil {
				t.Error("expected error for non-200 status, got nil")
			}
		})
	}
}

func TestChat_InvalidJSONResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.Chat(ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "user", Content: "Test"},
		},
	})

	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestChat_EmptyResponseBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Write empty body
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.Chat(ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "user", Content: "Test"},
		},
	})

	if err == nil {
		t.Error("expected error for empty body, got nil")
	}
}

func TestChat_NetworkError(t *testing.T) {
	t.Parallel()

	// Use a client pointing to a closed/non-existent server
	client := &Client{
		apiKey:  testAPIKey,
		baseURL: "http://127.0.0.1:1", // Unused port, will fail
		httpClient: &http.Client{
			Timeout: 10 * time.Millisecond,
		},
	}

	_, err := client.Chat(ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "user", Content: "Test"},
		},
	})

	if err == nil {
		t.Error("expected network error, got nil")
	}
}

func TestChat_ResponseWithMultipleChoices(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [
				{
					"message": {"role": "assistant", "content": "First response"},
					"finish_reason": "stop"
				},
				{
					"message": {"role": "assistant", "content": "Second response"},
					"finish_reason": "stop"
				}
			],
			"usage": {
				"prompt_tokens": 20,
				"completion_tokens": 10,
				"total_tokens": 30
			}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	resp, err := client.Chat(ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "user", Content: "Test"},
		},
	})

	if err != nil {
		t.Fatalf("Chat() unexpected error: %v", err)
	}
	if len(resp.Choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "First response" {
		t.Errorf("first choice = %q, want %q", resp.Choices[0].Message.Content, "First response")
	}
	if resp.Choices[1].Message.Content != "Second response" {
		t.Errorf("second choice = %q, want %q", resp.Choices[1].Message.Content, "Second response")
	}
}

func TestChat_PreservesRequestHeaders(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all expected headers
		checks := []struct {
			header string
			want   string
		}{
			{"Content-Type", "application/json"},
			{"Authorization", "Bearer " + testAPIKey},
		}
		for _, c := range checks {
			if r.Header.Get(c.header) != c.want {
				t.Errorf("header %s = %q, want %q", c.header, r.Header.Get(c.header), c.want)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"OK"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	_, err := client.Chat(ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "user", Content: "Test"},
		},
		Temp:      0.5,
		MaxTokens: 100,
	})

	if err != nil {
		t.Fatalf("Chat() unexpected error: %v", err)
	}
}

// TestChat_WithLongContent verifies handling of long message content
func TestChat_WithLongContent(t *testing.T) {
	t.Parallel()

	longText := make([]byte, 10000)
	for i := range longText {
		longText[i] = 'A'
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [{"message": {"role": "assistant", "content": "Got it!"}, "finish_reason": "stop"}],
			"usage": {"prompt_tokens": 2500, "completion_tokens": 3, "total_tokens": 2503}
		}`))
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	resp, err := client.Chat(ChatRequest{
		Model: "deepseek-chat",
		Messages: []Message{
			{Role: "user", Content: string(longText)},
		},
	})

	if err != nil {
		t.Fatalf("Chat() unexpected error: %v", err)
	}
	if resp.Usage.PromptTokens != 2500 {
		t.Errorf("prompt_tokens = %d, want 2500", resp.Usage.PromptTokens)
	}
}
