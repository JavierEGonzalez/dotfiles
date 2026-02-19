# ADUSA TUI - Implementation Chunks

## Chunk 1: Project Setup

**Goal**: Initialize Go project with Bubble Tea and create basic app that compiles and runs.

### Steps

1. Create directory structure
2. Initialize Go module
3. Install Bubble Tea dependencies
4. Create main.go with tea.Program skeleton
5. Create basic model and view
6. Build and test

### Acceptance Criteria

- [ ] `go mod init github.com/javiergonzalez/adusa-tui` succeeds
- [ ] `go get github.com/charmbracelet/bubbletea` succeeds
- [ ] `go get github.com/charmbracelet/bubbles` succeeds
- [ ] `go get github.com/charmbracelet/lipgloss` succeeds
- [ ] `go build -o adusa-tui` produces binary
- [ ] Running `./adusa-tui` shows blank window that responds to Ctrl+C
- [ ] No linter errors

---

## Chunk 2: Core Types and Config

**Goal**: Define core data types and configuration management.

### Steps

1. Create `internal/config/config.go` - paths, credentials
2. Create `internal/types/worktree.go` - Worktree struct
3. Create `internal/types/ticket.go` - TicketInfo struct
4. Create `internal/types/agent.go` - AgentStatus enum
5. Create helper functions for file paths

### Acceptance Criteria

- [ ] Config returns correct paths (worktree base, tickets dir, jira creds)
- [ ] Worktree struct has: Ticket, Branch, Path, CreatedAt
- [ ] TicketInfo struct has: Key, Summary, Description, Status, Assignee, Priority
- [ ] AgentStatus enum: Idle, Running, Done, Error
- [ ] Functions for: TicketFilePath(ticket), PlanFilePath(ticket), WorktreeInfoPath(path)
- [ ] `go build` succeeds

---

## Chunk 3: Git Operations

**Goal**: Implement git command wrappers for worktree operations.

### Steps

1. Create `internal/git/worktree.go`
2. Implement: ListWorktrees() - scan directory, parse .worktree-info
3. Implement: CreateWorktree(ticket, branchType, description) - git worktree add
4. Implement: DeleteWorktree(path) - git worktree remove, rm directory
5. Implement: GetStatus(path) - git status --short
6. Implement: GetDiff(path) - git diff
7. Implement: StageAll(path) - git add -A
8. Implement: Commit(path, message) - git commit with [TICKET]: prefix

### Acceptance Criteria

- [ ] ListWorktrees() returns []Worktree from ~/workspace/prism3/
- [ ] CreateWorktree creates branch and worktree directory
- [ ] CreateWorktree creates .worktree-info file
- [ ] DeleteWorktree removes worktree but NOT branch
- [ ] GetStatus returns formatted status (M, A, D, ?)
- [ ] StageAll stages all files
- [ ] Commit creates commit with "[TICKET]: message" format
- [ ] All functions return error on failure
- [ ] `go build` succeeds

---

## Chunk 4: Jira Client

**Goal**: Fetch ticket information from Jira API.

### Steps

1. Create `internal/jira/client.go`
2. Implement: LoadCredentials() - read ~/.scratch/jira.email and jira.token
3. Implement: FetchTicket(key) - call REST API
4. Implement: ParseTicketInfo(response) - extract fields
5. Implement: SaveTicketCache(info) - write to ~/.scratch/tickets/{key}.md
6. Implement: LoadTicketCache(key) - read from cache

### Acceptance Criteria

- [ ] LoadCredentials returns email and token
- [ ] FetchTicket returns TicketInfo for valid ticket
- [ ] FetchTicket returns error for invalid ticket
- [ ] SaveTicketCache writes markdown file
- [ ] LoadTicketCache reads from file if exists
- [ ] Cache includes: Key, Summary, Description, Status, Assignee, Priority
- [ ] `go build` succeeds

---

## Chunk 5: All Worktrees Screen

**Goal**: Display list of worktrees and allow selection.

### Steps

1. Create `internal/ui/screens/worktrees.go`
2. Implement: WorktreesModel struct
3. Implement: Init() - load worktrees
4. Implement: Update() - handle keys (j/k/Enter/q/n)
5. Implement: View() - render list
6. Implement: NewWorktreesScreen() - constructor
7. Update main.go to use WorktreesModel

### Acceptance Criteria

- [ ] Shows list of all worktrees from ~/workspace/prism3/
- [ ] Each row shows: ticket, branch, dirty/clean status
- [ ] j/k or arrows move cursor
- [ ] Enter selects worktree (sets selected path)
- [ ] n opens new worktree prompt
- [ ] q quits application
- [ ] Empty state: "No worktrees found. Press n to create one."
- [ ] `go build` succeeds

---

## Chunk 6: New Worktree Flow

**Goal**: Create new worktree from within TUI.

### Steps

1. Add new screen state for prompt
2. Input for ticket number
3. Branch type selection (f/b/h = feature/bugfix/hotfix)
4. Input for description (optional)
5. Show preview of worktree path
6. Create worktree on confirm
7. Handle errors

### Acceptance Criteria

- [ ] Press n shows ticket input
- [ ] Validates ticket format (123 or CXPVSP-123)
- [ ] Shows branch type selection (f/b/h keys)
- [ ] Description is optional
- [ ] Shows preview: "Worktree: ~/workspace/prism3/123"
- [ ] Enter creates worktree
- [ ] Creates branch with correct prefix
- [ ] Copies .env.secrets to worktree
- [ ] Creates tmux session with ticket name
- [ ] On success: goes to worktree view
- [ ] On error: shows error, can retry
- [ ] Esc cancels and returns to list

---

## Chunk 7: Delete Worktree Flow

**Goal**: Delete worktree with confirmation.

### Steps

1. Add delete confirmation screen
2. Show what will be deleted
3. Confirm with y/n
4. Execute git worktree remove
5. Remove directory

### Acceptance Criteria

- [ ] Press d on selected worktree shows confirmation
- [ ] Shows: "This will remove git worktree and delete directory"
- [ ] Shows: "Branch will NOT be deleted"
- [ ] y deletes worktree
- [ ] n cancels
- [ ] Refreshes list after delete
- [ ] Shows error if deletion fails

---

## Chunk 8: Individual Worktree Screen (Tabs)

**Goal**: Create base screen with 4-tab navigation.

### Steps

1. Create `internal/ui/screens/worktree.go`
2. Implement: WorktreeModel struct with tabs
3. Implement: Tab navigation (g/a/t/p keys)
4. Implement: Tab 0 = Changes, 1 = Agent, 2 = Ticket, 3 = Plan
5. Implement: Header shows ticket, branch, path
6. Implement: Esc returns to worktrees list
7. Integrate with worktrees screen

### Acceptance Criteria

- [ ] Header shows: ticket title, branch, path
- [ ] Tab bar shows: [g: Changes] [a: Agent] [t: Ticket] [p: Plan]
- [ ] g switches to Changes tab
- [ ] a switches to Agent tab
- [ ] t switches to Ticket tab
- [ ] p switches to Plan tab
- [ ] h/left goes to previous tab
- [ ] l/right goes to next tab
- [ ] Esc goes back to worktrees list
- [ ] `go build` succeeds

---

## Chunk 9: Changes Tab

**Goal**: View git status, diff, stage, and commit.

### Steps

1. Render git status output
2. Show staged vs unstaged
3. Add stage all action (s)
4. Add commit action (c) - prompt for message
5. Add view full diff (v)
6. Add refresh (r)

### Acceptance Criteria

- [ ] Shows git status --short output
- [ ] M = modified (color coded)
- [ ] A = added
- [ ] ? = untracked
- [ ] s stages all files, shows success
- [ ] c opens commit prompt
- [ ] Commit message pre-filled with "[TICKET]: "
- [ ] Enter commits successfully
- [ ] v shows full diff in viewport
- [ ] j/k scrolls diff
- [ ] Esc exits diff view
- [ ] r refreshes status

---

## Chunk 10: Ticket Tab

**Goal**: Display Jira ticket info, refetch, append notes.

### Steps

1. Load ticket info from cache
2. Render summary, status, assignee, priority
3. Render description
4. Add refetch (r) - calls Jira API
5. Add append notes (a) - opens $EDITOR
6. Add edit (e) - opens $EDITOR

### Acceptance Criteria

- [ ] Shows "Loading..." if cache miss
- [ ] Shows ticket info: Summary, Status, Assignee, Priority
- [ ] Shows Description (with scroll if long)
- [ ] r fetches from Jira and updates cache
- [ ] r shows loading indicator
- [ ] r shows error on failure (no creds, not found)
- [ ] a appends to notes section in cache file
- [ ] e opens entire file in $EDITOR
- [ ] Esc returns to tab view
- [ ] `go build` succeeds

---

## Chunk 11: Plan Tab

**Goal**: Create, edit, and execute implementation plans.

### Steps

1. Check if plan file exists
2. Render plan markdown
3. Add create (c) - AI generates plan
4. Add edit (e) - opens $EDITOR
5. Add execute (x) - runs OpenCode in tmux

### Acceptance Criteria

- [ ] Shows "No plan file found" if missing
- [ ] Shows plan content if exists
- [ ] Scrollable if content is long
- [ ] c prompts for confirmation first
- [ ] c calls AI with ticket info
- [ ] c saves to ~/.scratch/tickets/{ticket}_plan.md
- [ ] c shows "Generating..." during AI call
- [ ] e opens plan in $EDITOR
- [ ] e reloads after editor closes
- [ ] x shows confirmation
- [ ] x creates/uses tmux session
- [ ] x sends OpenCode command to tmux
- [ ] x switches to Agent tab

---

## Chunk 12: Agent Tab

**Goal**: Run OpenCode/Ralph and show status.

### Steps

1. Show idle state with run options
2. Add run OpenCode (o)
3. Add run Ralph (r)
4. Add change model (m)
5. Add change iterations (i)
6. Show running state
7. Add stop (s) - kill tmux session
8. Show done state

### Acceptance Criteria

- [ ] Idle: shows [o: Run OpenCode] [r: Run Ralph]
- [ ] o sends OpenCode to tmux session
- [ ] r sends Ralph to tmux session
- [ ] m prompts for model name
- [ ] i prompts for iterations (Ralph only)
- [ ] Running: shows "Running in tmux session: {ticket}"
- [ ] Running: shows "tmux attach -t {ticket}" instruction
- [ ] s stops/kills the tmux session
- [ ] Done: shows completion status
- [ ] v switches to Changes tab
- [ ] c switches to Changes tab with commit ready
- [ ] r runs again

---

## Chunk 13: Integration & Polish

**Goal**: Final integration and polish.

### Steps

1. Vim bindings (j/k/h/l) throughout
2. Help overlay (? key)
3. Error handling - all edge cases
4. Loading indicators (spinners)
5. Confirmation prompts
6. Build binary to ~/go/bin/

### Acceptance Criteria

- [ ] j/k navigate lists everywhere
- [ ] h/l navigate tabs
- [ ] ? shows help overlay with all keys
- [ ] Esc cancels prompts
- [ ] All errors show user-friendly messages
- [ ] Loading states for async operations
- [ ] Confirm before destructive actions
- [ ] Binary installs to ~/go/bin/adusa-tui
- [ ] `adusa-tui` runs from any directory

---

## Summary

| Chunk | Name | Est. Complexity |
|-------|------|-----------------|
| 1 | Project Setup | Low |
| 2 | Core Types & Config | Low |
| 3 | Git Operations | Medium |
| 4 | Jira Client | Medium |
| 5 | All Worktrees Screen | Medium |
| 6 | New Worktree Flow | Medium |
| 7 | Delete Worktree Flow | Low |
| 8 | Individual Worktree (Tabs) | Medium |
| 9 | Changes Tab | Medium |
| 10 | Ticket Tab | Low |
| 11 | Plan Tab | High |
| 12 | Agent Tab | High |
| 13 | Integration & Polish | Medium |

**Total**: 13 chunks
