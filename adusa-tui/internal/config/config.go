package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

var (
	HomeDir       = os.Getenv("HOME")
	ScratchDir    = filepath.Join(HomeDir, ".scratch")
	WorktreeBase  = filepath.Join(HomeDir, "workspace", "prism3")
	TicketsDir    = filepath.Join(ScratchDir, "tickets")
	JiraEmailPath = filepath.Join(ScratchDir, "jira.email")
	JiraTokenPath = filepath.Join(ScratchDir, "jira.token")
	ConfigDir     = filepath.Join(HomeDir, ".config", "adusa-tui")
	ConfigPath    = filepath.Join(ConfigDir, "config.toml")
)

type Config struct {
	DiffViewer string `toml:"diff-viewer"`
}

var AppConfig = loadConfig()

func loadConfig() Config {
	defaultConfig := Config{
		DiffViewer: "nvim -c DiffviewOpen",
	}

	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		if err := os.MkdirAll(ConfigDir, 0755); err != nil {
			return defaultConfig
		}
		if err := saveConfig(ConfigPath, defaultConfig); err != nil {
			return defaultConfig
		}
		return defaultConfig
	}

	var cfg Config
	if _, err := toml.DecodeFile(ConfigPath, &cfg); err != nil {
		return defaultConfig
	}

	if cfg.DiffViewer == "" {
		cfg.DiffViewer = defaultConfig.DiffViewer
	}

	return cfg
}

func saveConfig(path string, cfg Config) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

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
