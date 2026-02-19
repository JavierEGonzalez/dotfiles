package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	HomeDir       = os.Getenv("HOME")
	ScratchDir    = filepath.Join(HomeDir, ".scratch")
	WorktreeBase  = filepath.Join(HomeDir, "workspace", "prism3")
	TicketsDir    = filepath.Join(ScratchDir, "tickets")
	JiraEmailPath = filepath.Join(ScratchDir, "jira.email")
	JiraTokenPath = filepath.Join(ScratchDir, "jira.token")
)

func TicketFilePath(ticket string) string {
	return filepath.Join(TicketsDir, fmt.Sprintf("%s.md", ticket))
}

func PlanFilePath(ticket string) string {
	return filepath.Join(TicketsDir, fmt.Sprintf("%s_plan.md", ticket))
}

func WorktreeInfoPath(path string) string {
	return filepath.Join(path, ".worktree-info")
}

func LoadJiraCredentials() (email, token string, err error) {
	emailBytes, err := os.ReadFile(JiraEmailPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read jira email: %w", err)
	}
	tokenBytes, err := os.ReadFile(JiraTokenPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read jira token: %w", err)
	}
	return string(emailBytes), string(tokenBytes), nil
}

func EnsureDirs() error {
	dirs := []string{ScratchDir, TicketsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}
	}
	return nil
}
