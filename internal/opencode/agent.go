package opencode

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"telegram-deepseek-bot/internal/config"
)

type Agent struct {
	cfg     *config.Config
	baseDir string
}

func NewAgent(cfg *config.Config) *Agent {
	return &Agent{
		cfg:     cfg,
		baseDir: ".",
	}
}

func (a *Agent) ReadFile(fileName string) (string, error) {
	fullPath := filepath.Join(a.baseDir, fileName)
	if !filepath.HasPrefix(fullPath, filepath.Clean(a.baseDir)+string(os.PathSeparator)) &&
		fullPath != filepath.Clean(a.baseDir) {
		return "", fmt.Errorf("access denied: path is outside base directory")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", fileName, err)
	}

	return string(content), nil
}

func (a *Agent) Execute(ctx context.Context, prompt string) (string, error) {
	if len(prompt) > 4096 {
		return "", fmt.Errorf("prompt is too long (max 4096 characters)")
	}

	args := []string{"run", prompt, "--dangerously-skip-permissions"}

	cmd := exec.CommandContext(ctx, "opencode", args...)
	cmd.Dir = a.cfg.OpenCodeWorkspace

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("timeout exceeded while executing opencode agent")
		}

		errMsg := stderr.String()
		if errMsg != "" {
			if len(errMsg) > 200 {
				errMsg = errMsg[:200] + "..."
			}
			return "", fmt.Errorf("opencode agent error: %s", errMsg)
		}
		return "", fmt.Errorf("opencode agent failed: %w", err)
	}

	return stdout.String(), nil
}
