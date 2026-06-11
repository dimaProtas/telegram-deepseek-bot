package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken         string
	DeepSeekAPIKey        string
	DeepSeekModel         string
	DeepSeekAPIURL        string
	DeepSeekMaxTokens     int
	DeepSeekTemp          float64
	AllowedUserIDs        []int64
	AllowedChatIDs        []int64
	LogLevel              string
	ConversationTTL       int
	OpenCodeEndpoint      string
	OpenCodeTimeout       int
	OpenCodeWorkspace     string
	OpenCodeDefaultAgent  string
	ProxyConf
}

type ProxyConf struct {
	Addr     string
	Username string
	Password string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		TelegramToken:     getEnv("TELEGRAM_BOT_TOKEN", ""),
		DeepSeekAPIKey:    getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekModel:     getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
		DeepSeekAPIURL:    getEnv("DEEPSEEK_API_URL", "https://api.deepseek.com/v1"),
		DeepSeekMaxTokens: getEnvAsInt("DEEPSEEK_MAX_TOKENS", 4096),
		DeepSeekTemp:      getEnvAsFloat("DEEPSEEK_TEMPERATURE", 0.7),
		AllowedUserIDs:    getEnvAsIntSlice("ALLOWED_USER_IDS", []int64{}),
		AllowedChatIDs:    getEnvAsIntSlice("ALLOWED_CHAT_IDS", []int64{}),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		ConversationTTL:   getEnvAsInt("CONVERSATION_TTL_HOURS", 24),
		OpenCodeEndpoint:      getEnv("OPENCODE_ENDPOINT", "http://localhost:8080"),
		OpenCodeTimeout:       getEnvAsInt("OPENCODE_TIMEOUT", 600),
		OpenCodeWorkspace:     getEnv("OPENCODE_WORKSPACE", "."),
		OpenCodeDefaultAgent:  getEnv("OPENCODE_DEFAULT_AGENT", ""),
		ProxyConf: ProxyConf{
			Addr:     os.Getenv("PROXY_ADDR"),
			Username: os.Getenv("PROXY_USERNAME"),
			Password: os.Getenv("PROXY_PASSWORD"),
		},
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return fallback
}

func getEnvAsFloat(key string, fallback float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return fallback
}

func getEnvAsIntSlice(key string, fallback []int64) []int64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return fallback
	}

	var result []int64
	for _, s := range strings.Split(valueStr, ",") {
		if i, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
			result = append(result, i)
		}
	}
	return result
}
