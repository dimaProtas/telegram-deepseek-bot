package logger_test

import (
	"bytes"
	"log"
	"os"
	"testing"

	"telegram-deepseek-bot/internal/logger"
)

// captureLogOutput redirects standard log output to a buffer and returns
// the buffer and a restore function.
func captureLogOutput() (*bytes.Buffer, func()) {
	var buf bytes.Buffer
	originalWriter := log.Writer()
	log.SetOutput(&buf)
	return &buf, func() {
		log.SetOutput(originalWriter)
	}
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantDebug bool // whether Debug should produce output
		wantInfo  bool // whether Info should produce output
		wantWarn  bool // whether Warn should produce output
		wantError bool // whether Error should produce output
	}{
		{
			name:      "debug_level",
			level:     "debug",
			wantDebug: true,
			wantInfo:  true,
			wantWarn:  true,
			wantError: true,
		},
		{
			name:      "info_level",
			level:     "info",
			wantDebug: false,
			wantInfo:  true,
			wantWarn:  true,
			wantError: true,
		},
		{
			name:      "warn_level",
			level:     "warn",
			wantDebug: false,
			wantInfo:  false,
			wantWarn:  true,
			wantError: true,
		},
		{
			name:      "error_level",
			level:     "error",
			wantDebug: false,
			wantInfo:  false,
			wantWarn:  false,
			wantError: true,
		},
		{
			name:      "invalid_level_defaults_to_info",
			level:     "invalid",
			wantDebug: false,
			wantInfo:  true,
			wantWarn:  true,
			wantError: true,
		},
		{
			name:      "empty_level_defaults_to_info",
			level:     "",
			wantDebug: false,
			wantInfo:  true,
			wantWarn:  true,
			wantError: true,
		},
		{
			name:      "uppercase_DEBUG_defaults_to_info",
			level:     "DEBUG",
			wantDebug: false,
			wantInfo:  true,
			wantWarn:  true,
			wantError: true,
		},
		{
			name:      "mixed_case_Debug_defaults_to_info",
			level:     "Debug",
			wantDebug: false,
			wantInfo:  true,
			wantWarn:  true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			l := logger.NewLogger(tt.level)
			if l == nil {
				t.Fatal("NewLogger returned nil")
			}

			// Verify behavior by checking log output
			buf, restore := captureLogOutput()
			defer restore()

			l.Debug("debug_msg")
			hasDebug := buf.Len() > 0
			buf.Reset()

			l.Info("info_msg")
			hasInfo := buf.Len() > 0
			buf.Reset()

			l.Warn("warn_msg")
			hasWarn := buf.Len() > 0
			buf.Reset()

			l.Error("error_msg")
			hasError := buf.Len() > 0

			if hasDebug != tt.wantDebug {
				t.Errorf("Debug output: got %v, want %v", hasDebug, tt.wantDebug)
			}
			if hasInfo != tt.wantInfo {
				t.Errorf("Info output: got %v, want %v", hasInfo, tt.wantInfo)
			}
			if hasWarn != tt.wantWarn {
				t.Errorf("Warn output: got %v, want %v", hasWarn, tt.wantWarn)
			}
			if hasError != tt.wantError {
				t.Errorf("Error output: got %v, want %v", hasError, tt.wantError)
			}
		})
	}
}

func TestLogOutputContainsPrefixes(t *testing.T) {
	l := logger.NewLogger("debug")

	buf, restore := captureLogOutput()
	defer restore()

	l.Debug("hello", "world")
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("[DEBUG]")) {
		t.Errorf("Debug output missing [DEBUG] prefix: %s", output)
	}

	buf.Reset()
	l.Info("test info")
	output = buf.String()
	if !bytes.Contains([]byte(output), []byte("[INFO]")) {
		t.Errorf("Info output missing [INFO] prefix: %s", output)
	}

	buf.Reset()
	l.Warn("test warn")
	output = buf.String()
	if !bytes.Contains([]byte(output), []byte("[WARN]")) {
		t.Errorf("Warn output missing [WARN] prefix: %s", output)
	}

	buf.Reset()
	l.Error("test error")
	output = buf.String()
	if !bytes.Contains([]byte(output), []byte("[ERROR]")) {
		t.Errorf("Error output missing [ERROR] prefix: %s", output)
	}
}

func TestLogger_NoOutputBelowLevel(t *testing.T) {
	l := logger.NewLogger("error")

	buf, restore := captureLogOutput()
	defer restore()

	l.Debug("should not appear")
	l.Info("should not appear")
	l.Warn("should not appear")

	if buf.Len() > 0 {
		t.Errorf("expected no output for levels below ERROR, got: %s", buf.String())
	}

	l.Error("should appear")
	if buf.Len() == 0 {
		t.Error("expected output for ERROR level")
	}
}

// TestLoggerGlobalState checks that log.SetOutput changes don't leak between tests
func TestLoggerGlobalState(t *testing.T) {
	originalWriter := log.Writer()
	defer log.SetOutput(originalWriter)

	l := logger.NewLogger("info")

	// First capture — should work
	buf1, restore1 := captureLogOutput()
	l.Info("first")
	_ = buf1.String()
	restore1()

	// Second capture — should also work independently
	buf2, restore2 := captureLogOutput()
	l.Info("second")
	output2 := buf2.String()
	restore2()

	if output2 == "" {
		t.Error("second log output was empty")
	}

	// Ensure original writer is still default
	if log.Writer() != os.Stderr {
		t.Log("warning: log.Writer() is not os.Stderr after restores; this may be expected in testing")
	}
}
