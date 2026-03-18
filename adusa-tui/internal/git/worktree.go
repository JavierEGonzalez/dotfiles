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
	"unicode"

	"github.com/javiergonzalez/adusa-tui/internal/config"
	"github.com/javiergonzalez/adusa-tui/internal/types"
)

type FileStatus struct {
	Path   string
	Status string // M, A, D, ?
	Staged bool
}

func ListWorktrees() ([]types.Worktree, error) {
	var allWorktrees []types.Worktree

	for _, repo := range config.AppConfig.Repos {
		gitWorktrees, err := listGitWorktrees(repo.Path)
		if err != nil {
			continue
		}

		for _, gw := range gitWorktrees {
			dirName := filepath.Base(gw.Path)
			infoPath := config.WorktreeInfoPath(gw.Path)
			wt := types.Worktree{
				Path:   gw.Path,
				Branch: gw.Branch,
				Repo:   repo.Name,
			}

			if info, err := os.ReadFile(infoPath); err == nil {
				parsed := parseWorktreeInfo(string(info), dirName, repo.Path)
				wt = parsed
				wt.Path = gw.Path
				wt.Branch = gw.Branch
				wt.Repo = repo.Name
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

			allWorktrees = append(allWorktrees, wt)
		}
	}

	return allWorktrees, nil
}

// GitWorktree represents a worktree from git worktree list
type GitWorktree struct {
	Path      string
	Branch    string
	CreatedAt time.Time
}

// listGitWorktrees uses git worktree list to get actual worktrees
func listGitWorktrees(repoPath string) ([]GitWorktree, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("git worktree list timed out")
		}
		return listWorktreesByDirectory(repoPath)
	}

	var worktrees []GitWorktree
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "worktree ") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		path := parts[1]
		branch := "unknown"
		createdAt := time.Now()

		// Read remaining lines of this worktree block until blank line
		for scanner.Scan() {
			nextLine := scanner.Text()
			if nextLine == "" {
				break
			}
			if strings.HasPrefix(nextLine, "branch ") {
				branchRef := strings.TrimPrefix(nextLine, "branch ")
				branch = strings.TrimPrefix(branchRef, "refs/heads/")
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
func listWorktreesByDirectory(repoPath string) ([]GitWorktree, error) {
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read worktree base: %w", err)
	}

	var worktrees []GitWorktree
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(repoPath, entry.Name())

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

func ResolveRepoRoot(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--show-toplevel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to resolve repo root from %s: %w, output: %s", repoPath, err, string(out))
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", fmt.Errorf("failed to resolve repo root from %s: empty result", repoPath)
	}
	return root, nil
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

func parseWorktreeInfo(content, dirName, basePath string) types.Worktree {
	wt := types.Worktree{
		Path: filepath.Join(basePath, dirName),
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

func CreateWorktree(ticket, branchType, description, repoPath string) (*types.Worktree, error) {
	repoRoot, err := ResolveRepoRoot(repoPath)
	if err != nil {
		return nil, err
	}

	repo := config.GetRepoByPath(repoPath)
	if repo == nil {
		return nil, fmt.Errorf("repo not found in config for path: %s", repoPath)
	}
	repoName := repo.Name
	if repoName == "" {
		repoName = filepath.Base(repoRoot)
	}

	dirName := ticket
	worktreeBase := repo.GetWorktreeDir()
	worktreePath := filepath.Join(worktreeBase, dirName)

	prefix := map[string]string{
		"f": "feature",
		"b": "bugfix",
		"h": "hotfix",
	}[branchType]

	branch := ""
	if description == "" {
		branch = fmt.Sprintf("%s/%s", prefix, dirName)
	} else {
		slug := SlugifyDescription(description)
		if slug == "" {
			branch = fmt.Sprintf("%s/%s", prefix, dirName)
		} else {
			branch = fmt.Sprintf("%s/%s-%s", prefix, dirName, slug)
		}
	}

	if err := os.MkdirAll(worktreeBase, 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktree base: %w", err)
	}

	cmd := exec.Command("git", "worktree", "add", "-b", branch, worktreePath)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create worktree: %w, output: %s", err, string(out))
	}

	envSecretsSrc := filepath.Join(repoRoot, ".env.secrets")
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
		Repo:      repoName,
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

	// Find the main repo root by resolving the common git dir from the worktree.
	// This is needed because `git worktree remove` must run from within the repo.
	gitDirCmd := exec.Command("git", "-C", wtPath, "rev-parse", "--git-common-dir")
	gitDirOut, err := gitDirCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to find repo for worktree: %w, output: %s", err, string(gitDirOut))
	}
	repoRoot := filepath.Dir(strings.TrimSpace(string(gitDirOut)))

	cmd := exec.Command("git", "worktree", "remove", wtPath)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove git worktree: %w, output: %s", err, string(out))
	}

	if err := os.RemoveAll(wtPath); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	return nil
}

func SlugifyDescription(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}

	var b strings.Builder
	prevDash := false
	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(unicode.ToLower(r))
			prevDash = false
			continue
		}
		if prevDash {
			continue
		}
		b.WriteByte('-')
		prevDash = true
	}

	result := strings.Trim(b.String(), "-")
	return result
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

func DifftoolCmd(path string) *exec.Cmd {
	cmd := exec.Command("git", "difftool", "--no-prompt", "HEAD")
	cmd.Dir = path
	return cmd
}

func CustomDiffViewerCmd(path string, files []string) *exec.Cmd {
	parts := strings.Fields(config.AppConfig.DiffViewer)
	if len(parts) == 0 {
		return nil
	}

	args := append(parts[1:], files...)
	cmd := exec.Command(parts[0], args...)
	cmd.Dir = path
	return cmd
}

func LazyGitCmd(path string) *exec.Cmd {
	cmd := exec.Command("lazygit")
	cmd.Dir = path
	return cmd
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

func GetCurrentWorktree() (*types.Worktree, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	for _, repo := range config.AppConfig.Repos {
		gitWorktrees, err := listGitWorktrees(repo.Path)
		if err != nil {
			continue
		}

		for _, gw := range gitWorktrees {
			absPath, err := filepath.Abs(gw.Path)
			if err != nil {
				continue
			}

			if absPath == cwd || gw.Path == cwd {
				infoPath := config.WorktreeInfoPath(gw.Path)
				wt := types.Worktree{
					Path:   gw.Path,
					Branch: gw.Branch,
					Repo:   repo.Name,
				}

				if info, err := os.ReadFile(infoPath); err == nil {
					parsed := parseWorktreeInfo(string(info), filepath.Base(gw.Path), repo.Path)
					wt = parsed
					wt.Path = gw.Path
					wt.Branch = gw.Branch
					wt.Repo = repo.Name
				} else {
					dirName := filepath.Base(gw.Path)
					if strings.HasPrefix(dirName, "CXPVSP-") {
						wt.Ticket = dirName
					}
				}

				return &wt, nil
			}
		}
	}

	return nil, nil
}
