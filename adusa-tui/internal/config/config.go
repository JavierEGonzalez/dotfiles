package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

var (
	HomeDir    = os.Getenv("HOME")
	ConfigDir  = filepath.Join(HomeDir, ".config", "adusa-tui")
	ConfigPath = filepath.Join(ConfigDir, "config.toml")
)

type Repo struct {
	Name          string `toml:"name"`
	Path          string `toml:"path"`
	WorktreeDir   string `toml:"worktree-dir"`
	DefaultBranch string `toml:"default-branch"`
}

// GetWorktreeDir returns the directory where worktrees are created.
// Falls back to the parent of Path if not explicitly set.
func (r Repo) GetWorktreeDir() string {
	if r.WorktreeDir != "" {
		return r.WorktreeDir
	}
	return filepath.Dir(r.Path)
}

type Jira struct {
	EmailPath string `toml:"email-path"`
	TokenPath string `toml:"token-path"`
	Domain    string `toml:"domain"`
}

type Config struct {
	DiffViewer string `toml:"diff-viewer"`
	ScratchDir string `toml:"scratch-dir"`
	TicketsDir string `toml:"tickets-dir"`
	Repos      []Repo `toml:"repo"`
	Jira       Jira   `toml:"jira"`
	AgentModel string `toml:"model"`
}

func GetRepoPaths() []string {
	var paths []string
	for _, repo := range AppConfig.Repos {
		paths = append(paths, repo.Path)
	}
	return paths
}

func GetRepoByPath(path string) *Repo {
	for i := range AppConfig.Repos {
		if AppConfig.Repos[i].Path == path {
			return &AppConfig.Repos[i]
		}
	}
	return nil
}

var AppConfig = loadConfig()

func loadConfig() Config {
	defaultConfig := Config{
		DiffViewer: "nvim -c DiffviewOpen",
		ScratchDir: "",
		TicketsDir: "",
		Repos: []Repo{
			{Name: "prism3", Path: filepath.Join(HomeDir, "workspace", "prism3", "nuclei"), WorktreeDir: filepath.Join(HomeDir, "workspace", "prism3"), DefaultBranch: "main"},
		},
		Jira: Jira{
			EmailPath: "",
			TokenPath: "",
			Domain:    "",
		},
		AgentModel: "opencode/minimax-m2.5-free",
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

	if len(cfg.Repos) == 0 {
		cfg.Repos = defaultConfig.Repos
	}

	for i := range cfg.Repos {
		if cfg.Repos[i].DefaultBranch == "" {
			cfg.Repos[i].DefaultBranch = "main"
		}
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

func GetScratchDir() string {
	if AppConfig.ScratchDir != "" {
		return AppConfig.ScratchDir
	}
	return filepath.Join(HomeDir, ".scratch")
}

func GetTicketsDir() string {
	if AppConfig.TicketsDir != "" {
		return AppConfig.TicketsDir
	}
	return filepath.Join(GetScratchDir(), "tickets")
}

func GetJiraEmailPath() string {
	if AppConfig.Jira.EmailPath != "" {
		return AppConfig.Jira.EmailPath
	}
	return filepath.Join(GetScratchDir(), "jira.email")
}

func GetJiraTokenPath() string {
	if AppConfig.Jira.TokenPath != "" {
		return AppConfig.Jira.TokenPath
	}
	return filepath.Join(GetScratchDir(), "jira.token")
}

func GetJiraDomain() string {
	if AppConfig.Jira.Domain != "" {
		return AppConfig.Jira.Domain
	}
	domainPath := filepath.Join(GetScratchDir(), "jira.domain")
	if data, err := os.ReadFile(domainPath); err == nil {
		return strings.TrimSpace(string(data))
	}
	return ""
}

func GetAgentModel() string {
	if AppConfig.AgentModel != "" {
		return AppConfig.AgentModel
	}
	return "opencode/minimax-m2.5-free"
}

func TicketFilePath(ticket string) string {
	return filepath.Join(GetTicketsDir(), fmt.Sprintf("%s.md", ticket))
}

func PlanFilePath(ticket string) string {
	return filepath.Join(GetTicketsDir(), fmt.Sprintf("%s_plan.md", ticket))
}

func WorktreeInfoPath(path string) string {
	return filepath.Join(path, ".worktree-info")
}

func LoadJiraCredentials() (email, token string, err error) {
	emailBytes, err := os.ReadFile(GetJiraEmailPath())
	if err != nil {
		return "", "", fmt.Errorf("failed to read jira email: %w", err)
	}
	tokenBytes, err := os.ReadFile(GetJiraTokenPath())
	if err != nil {
		return "", "", fmt.Errorf("failed to read jira token: %w", err)
	}
	return string(emailBytes), string(tokenBytes), nil
}

func EnsureDirs() error {
	dirs := []string{GetScratchDir(), GetTicketsDir()}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}
	}
	return nil
}
