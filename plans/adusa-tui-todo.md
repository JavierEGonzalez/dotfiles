# ADUSA TUI - TODO

> Reference: [adusa-tui-reference.md](./adusa-tui-reference.md) for architecture and user flow.
> Archive: [archive/](./archive/) for original plan docs and progress history.

## Priority 1: Bugs & Correctness

- [ ] **Fix async loading** - `loadStatus()`, `loadTicket()`, `fetchTicket()`, `loadDiff()` all block
      the UI thread. The `isLoading`/`loadingMsg` fields exist but spinners never render because
      calls are synchronous inside `Update()`. Convert to `tea.Cmd` pattern like `generatePlanCmd()`.
      Files: `internal/ui/screens/worktree.go`

- [ ] **Fix `appendToTicketNotes()`** - Currently identical to `editTicket()` (both just open
      `$EDITOR` on the whole file). Should either: open editor positioned at the Notes section,
      or append via a TUI text input. File: `internal/ui/screens/worktree.go`

- [ ] **Use `AgentStatus` type** - `types/agent.go` defines `AgentIdle/Running/Done/Error` enum
      but `WorktreeModel` uses raw booleans (`agentRunning`, `agentDone`). Replace booleans with
      the proper enum. File: `internal/ui/screens/worktree.go`

- [ ] **Fix duplicate `showingHelp` check** - Same condition checked twice consecutively at
      lines ~69-75. File: `internal/ui/screens/worktrees.go`

## Priority 2: Missing Features (from plan)

- [ ] **Add CLI argument support** - Plan specifies `adusa-tui [ticket]` should jump directly to
      that worktree's detail screen. Currently `main()` ignores all args.
      File: `main.go`

- [ ] **Add tmux session creation on new worktree** - Plan (Chunk 6) says "Creates tmux session
      with ticket name" during worktree creation. Currently only creates git worktree and copies
      `.env.secrets`. File: `internal/ui/screens/create_worktree.go` or `internal/git/worktree.go`

- [ ] **Show diff stats in Changes tab** - `git.GetDiffStat()` is implemented but never called
      by the UI. Plan shows `diff --stat: 3 files changed, 150 insertions` in the Changes layout.
      File: `internal/ui/screens/worktree.go`

- [ ] **Agent completion detection** - Once an agent is launched in tmux, there's no way to detect
      when it finishes. User must manually press `s`. Consider polling tmux pane for idle state or
      checking for a sentinel file. File: `internal/ui/screens/worktree.go`

## Priority 3: Refactoring

- [ ] **Split `worktree.go` (924 lines)** - Extract each tab into its own file:
      `changes_tab.go`, `agent_tab.go`, `ticket_tab.go`, `plan_tab.go`.
      Keep `worktree.go` as the coordinator with header/tab bar/routing.

- [ ] **Consolidate state flags into mode enum** - Replace `promptModel`, `promptIter`,
      `selectingModel`, `commitMode`, `showingDiff`, `showingHelp` booleans with a single
      `viewMode` enum (Normal, DiffView, CommitInput, ModelSelect, IterInput, Help).

- [ ] **Use constants for view routing** - `main.go` uses string literals (`"create"`, `"delete"`,
      `"worktree"`) for `m.view`. Replace with `const` block.

- [ ] **Remove dead code** - `config.LoadJiraCredentials()` is never called (jira package has its
      own). `GetDiffStat()` unused. Bubbles `list.Model` initialized but rendering bypassed.

- [ ] **Update hardcoded model list** - 8 models in `worktree.go` include stale identifiers
      (old Claude 3 names). Make configurable or update to current model names.

## Priority 4: Polish & Enhancements

- [ ] **Add tests** - Zero test files exist. At minimum: `git/worktree_test.go`,
      `jira/client_test.go`, `config/config_test.go`.

- [ ] **Resolve `r` key conflict on Agent tab** - `r` is mapped to both "refresh" (global per-tab)
      and "Run Ralph" (agent-specific). Clarify precedence or remap.

- [ ] **Improve Jira description parsing** - ADF parser only handles one level of nesting. Deeper
      formatting (tables, code blocks, nested lists) is lost.

## Done

_Move items here as they're completed._
