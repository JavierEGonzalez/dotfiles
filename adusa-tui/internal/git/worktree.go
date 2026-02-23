package git

import (
	"bufio"
	"context"
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
	gitWorktrees, err := listGitWorktrees()
	if err != nil {
		return nil, err
	}

	var worktrees []types.Worktree
	for _, gw := range gitWorktrees {
		dirName := filepath.Base(gw.Path)
		infoPath := config.WorktreeInfoPath(gw.Path)
		wt := types.Worktree{
			Path:   gw.Path,
			Branch: gw.Branch,
		}

		if info, err := os.ReadFile(infoPath); err == nil {
			parsed := parseWorktreeInfo(string(info), dirName)
			wt = parsed
			wt.Path = gw.Path
			wt.Branch = gw.Branch
		} else {
			ticket := ""
			if strings.HasPrefix(dirName, "CXPVSP-") {
				ticket = dirName
			}

			wt.Ticket = ticket
			wt.CreatedAt = gw.CreatedAt

			infoContent := fmt.Sprintf("TICKET=%s\nBRANCH=%s\nCREATED=%s\n", ticket, gw.Branch, gw.CreatedAt.Format(time.RFC3339))
			if err := os.WriteFile(infoPath, []byte(infoContent), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to write worktree info for %s: %v\n", dirName, err)
			}
		}

		worktrees = append(worktrees, wt)
	}

	return worktrees, nil
}

// GitWorktree represents a worktree from git worktree list
type GitWorktree struct {
	Path      string
	Branch    string
	CreatedAt time.Time
}

// listGitWorktrees uses git worktree list to get actual worktrees
func listGitWorktrees() ([]GitWorktree, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "worktree", "list", "--porcelain")
	cmd.Dir = config.WorktreeBase
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("git worktree list timed out")
		}
		return listWorktreesByDirectory()
	}

	var worktrees []GitWorktree
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "worktree") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		path := parts[1]
		branch := "unknown"
		createdAt := time.Now()

		// Read the next line to get branch info
		if scanner.Scan() {
			nextLine := scanner.Text()
			if strings.HasPrefix(nextLine, "branch") {
				parts := strings.Split(nextLine, " ")
				if len(parts) >= 2 {
					// Extract branch name from reference path (e.g., "refs/heads/feature/123")
					branchRef := parts[1]
					branch = strings.TrimPrefix(branchRef, "refs/heads/")
				}
			}
		}

		worktrees = append(worktrees, GitWorktree{
			Path:      path,
			Branch:    branch,
			CreatedAt: createdAt,
		})
	}

	return worktrees, nil
}

// listWorktreesByDirectory falls back to directory listing if git command fails
func listWorktreesByDirectory() ([]GitWorktree, error) {
	entries, err := os.ReadDir(config.WorktreeBase)
	if err != nil {
		return nil, fmt.Errorf("failed to read worktree base: %w", err)
	}

	var worktrees []GitWorktree
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(config.WorktreeBase, entry.Name())

		gitDir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		branch := getBranchFromPath(path)
		if branch == "unknown" {
			continue
		}

		worktrees = append(worktrees, GitWorktree{
			Path:      path,
			Branch:    branch,
			CreatedAt: info.ModTime(),
		})
	}

	return worktrees, nil
}

// getBranchFromPath tries to determine the current branch in a directory
func getBranchFromPath(path string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
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

	createTmuxCmd := exec.Command("tmux", "new-session", "-d", "-s", ticket, "-c", worktreePath)
	if err := createTmuxCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create tmux session: %w", err)
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
