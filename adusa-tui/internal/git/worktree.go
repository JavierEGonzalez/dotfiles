package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/javiergonzalez/adusa-tui/internal/config"
	"github.com/javiergonzalez/adusa-tui/internal/types"
)

type FileStatus struct {
	Path   string
	Status string // M, A, D, ?
	Staged bool
}

func ListWorktrees() ([]types.Worktree, error) {
	entries, err := os.ReadDir(config.WorktreeBase)
	if err != nil {
		return nil, fmt.Errorf("failed to read worktree base: %w", err)
	}

	var worktrees []types.Worktree
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		infoPath := config.WorktreeInfoPath(filepath.Join(config.WorktreeBase, entry.Name()))
		info, err := os.ReadFile(infoPath)
		if err != nil {
			continue
		}

		wt := parseWorktreeInfo(string(info), entry.Name())
		wt.Path = filepath.Join(config.WorktreeBase, entry.Name())
		worktrees = append(worktrees, wt)
	}

	return worktrees, nil
}

func parseWorktreeInfo(content, dirName string) types.Worktree {
	wt := types.Worktree{
		Path: filepath.Join(config.WorktreeBase, dirName),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "TICKET=") {
			wt.Ticket = strings.TrimPrefix(line, "TICKET=")
		} else if strings.HasPrefix(line, "BRANCH=") {
			wt.Branch = strings.TrimPrefix(line, "BRANCH=")
		} else if strings.HasPrefix(line, "CREATED=") {
			ts := strings.TrimPrefix(line, "CREATED=")
			t, err := time.Parse(time.RFC3339, ts)
			if err == nil {
				wt.CreatedAt = t
			}
		}
	}

	return wt
}

func CreateWorktree(ticket, branchType, description string) (*types.Worktree, error) {
	dirName := strings.TrimPrefix(ticket, "CXPVSP-")
	worktreePath := filepath.Join(config.WorktreeBase, dirName)

	prefix := map[string]string{
		"f": "feature",
		"b": "bugfix",
		"h": "hotfix",
	}[branchType]

	branchDesc := description
	if branchDesc == "" {
		branchDesc = dirName
	}
	branch := fmt.Sprintf("%s/%s-%s", prefix, dirName, branchDesc)

	if err := os.MkdirAll(config.WorktreeBase, 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktree base: %w", err)
	}

	cmd := exec.Command("git", "worktree", "add", "-b", branch, worktreePath)
	cmd.Dir = config.WorktreeBase
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w, output: %s", err, string(out))
	}

	envSecretsSrc := filepath.Join(config.WorktreeBase, ".env.secrets")
	envSecretsDst := filepath.Join(worktreePath, ".env.secrets")
	if _, err := os.Stat(envSecretsSrc); err == nil {
		if err := copyFile(envSecretsSrc, envSecretsDst); err != nil {
			return nil, fmt.Errorf("failed to copy .env.secrets: %w", err)
		}
	}

	infoContent := fmt.Sprintf("TICKET=%s\nBRANCH=%s\nCREATED=%s\n", ticket, branch, time.Now().Format(time.RFC3339))
	infoPath := config.WorktreeInfoPath(worktreePath)
	if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write worktree info: %w", err)
	}

	return &types.Worktree{
		Ticket:    ticket,
		Branch:    branch,
		Path:      worktreePath,
		CreatedAt: time.Now(),
	}, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func DeleteWorktree(path string) error {
	wtPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	cmd := exec.Command("git", "worktree", "remove", wtPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove git worktree: %w, output: %s", err, string(out))
	}

	if err := os.RemoveAll(wtPath); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	return nil
}

func GetStatus(path string) ([]FileStatus, error) {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w, output: %s", err, string(out))
	}

	var status []FileStatus
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		statusChar := line[0]
		staged := statusChar != '?' && statusChar != ' '

		path := strings.TrimSpace(line[3:])
		status = append(status, FileStatus{
			Path:   path,
			Status: string(statusChar),
			Staged: staged,
		})
	}

	return status, nil
}

func GetDiff(path string) (string, error) {
	cmd := exec.Command("git", "diff")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff: %w, output: %s", err, string(out))
	}

	return string(out), nil
}

func GetDiffStat(path string) (string, error) {
	cmd := exec.Command("git", "diff", "--stat")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff stat: %w, output: %s", err, string(out))
	}

	return string(out), nil
}

func StageAll(path string) error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = path
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stage files: %w, output: %s", err, string(out))
	}
	return nil
}

func Commit(path, ticket, message string) error {
	commitMsg := fmt.Sprintf("[%s]: %s", ticket, message)
	cmd := exec.Command("git", "commit", "-m", commitMsg)
	cmd.Dir = path
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to commit: %w, output: %s", err, string(out))
	}
	return nil
}

func GetCurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w, output: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}
