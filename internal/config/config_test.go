package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		setEnv   bool
		fallback string
		want     string
	}{
		{
			name:     "env_set",
			key:      "TEST_GETENV_1",
			envValue: "actual_value",
			setEnv:   true,
			fallback: "fallback",
			want:     "actual_value",
		},
		{
			name:     "env_not_set_fallback",
			key:      "TEST_GETENV_2_NONEXISTENT",
			envValue: "",
			setEnv:   false,
			fallback: "fallback_value",
			want:     "fallback_value",
		},
		{
			name:     "env_empty_string",
			key:      "TEST_GETENV_3",
			envValue: "",
			setEnv:   true,
			fallback: "fallback",
			want:     "",
		},
		{
			name:     "empty_fallback",
			key:      "TEST_GETENV_4_NONEXISTENT",
			envValue: "",
			setEnv:   false,
			fallback: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before and after
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnv(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("getEnv() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		setEnv   bool
		fallback int
		want     int
	}{
		{
			name:     "valid_int_positive",
			key:      "TEST_INT_1",
			envValue: "42",
			setEnv:   true,
			fallback: 0,
			want:     42,
		},
		{
			name:     "valid_int_negative",
			key:      "TEST_INT_2",
			envValue: "-10",
			setEnv:   true,
			fallback: 100,
			want:     -10,
		},
		{
			name:     "valid_int_zero",
			key:      "TEST_INT_3",
			envValue: "0",
			setEnv:   true,
			fallback: 1,
			want:     0,
		},
		{
			name:     "invalid_int_returns_fallback",
			key:      "TEST_INT_4",
			envValue: "not_a_number",
			setEnv:   true,
			fallback: 99,
			want:     99,
		},
		{
			name:     "float_string_returns_fallback",
			key:      "TEST_INT_5",
			envValue: "3.14",
			setEnv:   true,
			fallback: 10,
			want:     10,
		},
		{
			name:     "empty_env_returns_fallback",
			key:      "TEST_INT_6",
			envValue: "",
			setEnv:   true,
			fallback: 50,
			want:     50,
		},
		{
			name:     "env_not_set_returns_fallback",
			key:      "TEST_INT_7_NONEXISTENT",
			envValue: "",
			setEnv:   false,
			fallback: 77,
			want:     77,
		},
		{
			name:     "large_int",
			key:      "TEST_INT_8",
			envValue: "2147483647",
			setEnv:   true,
			fallback: 0,
			want:     2147483647,
		},
		{
			name:     "overflow_returns_fallback",
			key:      "TEST_INT_9",
			envValue: "99999999999999999999",
			setEnv:   true,
			fallback: 100,
			want:     100,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsInt(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("getEnvAsInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsFloat(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		setEnv   bool
		fallback float64
		want     float64
	}{
		{
			name:     "valid_float_positive",
			key:      "TEST_FLOAT_1",
			envValue: "3.14",
			setEnv:   true,
			fallback: 0.0,
			want:     3.14,
		},
		{
			name:     "valid_float_negative",
			key:      "TEST_FLOAT_2",
			envValue: "-2.5",
			setEnv:   true,
			fallback: 1.0,
			want:     -2.5,
		},
		{
			name:     "valid_float_zero",
			key:      "TEST_FLOAT_3",
			envValue: "0.0",
			setEnv:   true,
			fallback: 1.0,
			want:     0.0,
		},
		{
			name:     "valid_float_integer_form",
			key:      "TEST_FLOAT_4",
			envValue: "7",
			setEnv:   true,
			fallback: 0.0,
			want:     7.0,
		},
		{
			name:     "invalid_float_returns_fallback",
			key:      "TEST_FLOAT_5",
			envValue: "not_a_float",
			setEnv:   true,
			fallback: 0.7,
			want:     0.7,
		},
		{
			name:     "empty_env_returns_fallback",
			key:      "TEST_FLOAT_6",
			envValue: "",
			setEnv:   true,
			fallback: 0.5,
			want:     0.5,
		},
		{
			name:     "env_not_set_returns_fallback",
			key:      "TEST_FLOAT_7_NONEXISTENT",
			envValue: "",
			setEnv:   false,
			fallback: 0.9,
			want:     0.9,
		},
		{
			name:     "scientific_notation",
			key:      "TEST_FLOAT_8",
			envValue: "1e-3",
			setEnv:   true,
			fallback: 0.0,
			want:     0.001,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsFloat(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("getEnvAsFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsIntSlice(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		setEnv   bool
		fallback []int64
		want     []int64
	}{
		{
			name:     "valid_slice_single",
			key:      "TEST_SLICE_1",
			envValue: "42",
			setEnv:   true,
			fallback: nil,
			want:     []int64{42},
		},
		{
			name:     "valid_slice_multiple",
			key:      "TEST_SLICE_2",
			envValue: "1,2,3,4,5",
			setEnv:   true,
			fallback: nil,
			want:     []int64{1, 2, 3, 4, 5},
		},
		{
			name:     "slice_with_spaces",
			key:      "TEST_SLICE_3",
			envValue: " 10 , 20 , 30 ",
			setEnv:   true,
			fallback: nil,
			want:     []int64{10, 20, 30},
		},
		{
			name:     "empty_env_returns_fallback",
			key:      "TEST_SLICE_4",
			envValue: "",
			setEnv:   true,
			fallback: []int64{1, 2, 3},
			want:     []int64{1, 2, 3},
		},
		{
			name:     "env_not_set_returns_fallback",
			key:      "TEST_SLICE_5_NONEXISTENT",
			envValue: "",
			setEnv:   false,
			fallback: []int64{99},
			want:     []int64{99},
		},
		{
			name:     "mixed_valid_and_invalid",
			key:      "TEST_SLICE_6",
			envValue: "1,abc,3,xyz,5",
			setEnv:   true,
			fallback: nil,
			want:     []int64{1, 3, 5},
		},
		{
			name:     "negative_numbers",
			key:      "TEST_SLICE_7",
			envValue: "-1,-2,-3",
			setEnv:   true,
			fallback: nil,
			want:     []int64{-1, -2, -3},
		},
		{
			name:     "all_invalid_returns_empty",
			key:      "TEST_SLICE_8",
			envValue: "abc,xyz,def",
			setEnv:   true,
			fallback: nil,
			want:     []int64{},
		},
		{
			name:     "empty_fallback_nil",
			key:      "TEST_SLICE_9_NONEXISTENT",
			envValue: "",
			setEnv:   false,
			fallback: nil,
			want:     nil,
		},
		{
			name:     "trailing_comma",
			key:      "TEST_SLICE_10",
			envValue: "1,2,",
			setEnv:   true,
			fallback: nil,
			want:     []int64{1, 2},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.key)

			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnvAsIntSlice(tt.key, tt.fallback)
			if len(got) != len(tt.want) {
				t.Errorf("getEnvAsIntSlice() len = %d, want len %d; got = %v, want = %v",
					len(got), len(tt.want), got, tt.want)
				return
			}
			if len(got) == 0 && len(tt.want) == 0 {
				return // both empty, ok
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getEnvAsIntSlice()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Ensure no .env file is loaded by clearing relevant env vars
	relevantVars := []string{
		"TELEGRAM_BOT_TOKEN", "DEEPSEEK_API_KEY", "DEEPSEEK_MODEL",
		"DEEPSEEK_API_URL", "DEEPSEEK_MAX_TOKENS", "DEEPSEEK_TEMPERATURE",
		"ALLOWED_USER_IDS", "ALLOWED_CHAT_IDS", "LOG_LEVEL",
		"CONVERSATION_TTL_HOURS", "OPENCODE_ENDPOINT", "OPENCODE_TIMEOUT",
		"OPENCODE_WORKSPACE",
	}

	// Unset all relevant vars to test defaults
	saved := make(map[string]string)
	for _, k := range relevantVars {
		saved[k] = os.Getenv(k)
		os.Unsetenv(k)
	}
	defer func() {
		for k, v := range saved {
			if v != "" {
				os.Setenv(k, v)
			}
		}
	}()

	cfg := LoadConfig()

	if cfg.DeepSeekModel != "deepseek-chat" {
		t.Errorf("DeepSeekModel = %q, want %q", cfg.DeepSeekModel, "deepseek-chat")
	}
	if cfg.DeepSeekAPIURL != "https://api.deepseek.com/v1" {
		t.Errorf("DeepSeekAPIURL = %q, want %q", cfg.DeepSeekAPIURL, "https://api.deepseek.com/v1")
	}
	if cfg.DeepSeekMaxTokens != 4096 {
		t.Errorf("DeepSeekMaxTokens = %d, want 4096", cfg.DeepSeekMaxTokens)
	}
	if cfg.DeepSeekTemp != 0.7 {
		t.Errorf("DeepSeekTemp = %v, want 0.7", cfg.DeepSeekTemp)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.ConversationTTL != 24 {
		t.Errorf("ConversationTTL = %d, want 24", cfg.ConversationTTL)
	}
	if cfg.OpenCodeEndpoint != "http://localhost:8080" {
		t.Errorf("OpenCodeEndpoint = %q, want %q", cfg.OpenCodeEndpoint, "http://localhost:8080")
	}
	if cfg.OpenCodeTimeout != 600 {
		t.Errorf("OpenCodeTimeout = %d, want 600", cfg.OpenCodeTimeout)
	}
	if cfg.OpenCodeWorkspace != "." {
		t.Errorf("OpenCodeWorkspace = %q, want %q", cfg.OpenCodeWorkspace, ".")
	}
	if cfg.TelegramToken != "" {
		t.Errorf("TelegramToken should be empty by default, got %q", cfg.TelegramToken)
	}
	if cfg.DeepSeekAPIKey != "" {
		t.Errorf("DeepSeekAPIKey should be empty by default, got %q", cfg.DeepSeekAPIKey)
	}
}

func TestLoadConfig_FromEnv(t *testing.T) {
	// Set custom values
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("DEEPSEEK_API_KEY", "test-api-key")
	os.Setenv("DEEPSEEK_MODEL", "deepseek-coder")
	os.Setenv("DEEPSEEK_MAX_TOKENS", "2048")
	os.Setenv("DEEPSEEK_TEMPERATURE", "0.5")
	os.Setenv("ALLOWED_USER_IDS", "100,200,300")
	os.Setenv("ALLOWED_CHAT_IDS", "-100,-200")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("CONVERSATION_TTL_HOURS", "48")
	os.Setenv("OPENCODE_TIMEOUT", "300")

	defer func() {
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		os.Unsetenv("DEEPSEEK_API_KEY")
		os.Unsetenv("DEEPSEEK_MODEL")
		os.Unsetenv("DEEPSEEK_MAX_TOKENS")
		os.Unsetenv("DEEPSEEK_TEMPERATURE")
		os.Unsetenv("ALLOWED_USER_IDS")
		os.Unsetenv("ALLOWED_CHAT_IDS")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("CONVERSATION_TTL_HOURS")
		os.Unsetenv("OPENCODE_TIMEOUT")
	}()

	cfg := LoadConfig()

	if cfg.TelegramToken != "test-token" {
		t.Errorf("TelegramToken = %q, want %q", cfg.TelegramToken, "test-token")
	}
	if cfg.DeepSeekAPIKey != "test-api-key" {
		t.Errorf("DeepSeekAPIKey = %q, want %q", cfg.DeepSeekAPIKey, "test-api-key")
	}
	if cfg.DeepSeekModel != "deepseek-coder" {
		t.Errorf("DeepSeekModel = %q, want %q", cfg.DeepSeekModel, "deepseek-coder")
	}
	if cfg.DeepSeekMaxTokens != 2048 {
		t.Errorf("DeepSeekMaxTokens = %d, want 2048", cfg.DeepSeekMaxTokens)
	}
	if cfg.DeepSeekTemp != 0.5 {
		t.Errorf("DeepSeekTemp = %v, want 0.5", cfg.DeepSeekTemp)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
	if cfg.ConversationTTL != 48 {
		t.Errorf("ConversationTTL = %d, want 48", cfg.ConversationTTL)
	}
	if cfg.OpenCodeTimeout != 300 {
		t.Errorf("OpenCodeTimeout = %d, want 300", cfg.OpenCodeTimeout)
	}
	if len(cfg.AllowedUserIDs) != 3 {
		t.Errorf("AllowedUserIDs len = %d, want 3", len(cfg.AllowedUserIDs))
	}
	if len(cfg.AllowedChatIDs) != 2 {
		t.Errorf("AllowedChatIDs len = %d, want 2", len(cfg.AllowedChatIDs))
	}
}
