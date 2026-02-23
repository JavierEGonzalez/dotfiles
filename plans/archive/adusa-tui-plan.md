# ADUSA AI Workflow TUI - Detailed Plan

## Initial user prompt mermaid diagram
flowchart TD
    A[Call Cli] -->|Without Args| B
    A -->|With Ticket Arg| C[Start]
    B[All Worktrees Screen] -->|Choose| B1
        B1[All Worktrees] --> C
    B -->|Choose| B2
        B2[Make New Worktree] --> C
    C[Individual Worktree Screen] -->|Press Esc| B
    C -->|Tab One| C1[View Changes]
    C -->|Tab Two| C2[View Agent Session]
    C -->|Tab Three| C3[View Ticket Summary]
    C3[View Ticket Summary] -->|Tab One| C3a
        C3a[View Ticket File] -->|Action| C3C{Append Info to Ticket File}
        C3a[View Ticket File] -->|Action| C3D{Refetch Ticket File}
    C3 -->|Tab Two| D
        D[Plan File] -->|Opt 1| D1
        D1{Create Plan File From Ticket File Info} --> D2
        D -->|Opt 2| D2{View or Edit Plan File} --> D2a
        D2a[Editor] -->|On Close| D
        D -->|Opt 3 if file exists| D3{Execute Plan File}
        D3 -->|Navigate To| C2

## Overview

A Bubble Tea TUI application that provides a unified interface for:
- Managing git worktrees
- Fetching and caching Jira ticket information
- Generating AI-powered implementation plans
- Running OpenCode/Ralph agent sessions
- Viewing changes and committing

## Architecture

### CLI Entry Point

```
adusa-tui [ticket]
```

- **With ticket**: Opens directly to Individual Worktree screen for that ticket
- **Without ticket**: Shows All Worktrees list screen

### Screen Flow

```
┌─────────────────┐     ┌──────────────────────────┐
│ All Worktrees  │────▶│ Individual Worktree      │
│    Screen      │     │        Screen           │
└─────────────────┘     └──────────────────────────┘
       │                          │
       │ ◀─────────────────────────│ (Esc key)
       │                          │
       │                    ┌─────┴─────┐
       │                    │            │
       ▼                    ▼            ▼
┌─────────────┐    ┌───────────┐  ┌─────────┐
│ New Worktree │    │ Changes   │  │  Agent  │
│   Prompt     │    │   Tab     │  │   Tab   │
└─────────────┘    └───────────┘  └─────────┘
                          │            │
                          ▼            ▼
                    ┌───────────┐  ┌─────────┐
                    │  Ticket   │  │  Plan   │
                    │   Tab     │  │   Tab   │
                    └───────────┘  └─────────┘
```

---

## Screen: All Worktrees (Landing)

### Purpose
Display list of all worktrees in `~/workspace/prism3/` and allow selection or creation.

### Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  ADUSA Worktrees                                        [n: New] [q: Quit] │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ > CXPVSP-123  feature/123-add-login           ✗ (2 files)          │ │
│  │   CXPVSP-122  bugfix/122-fix-validation      ✓ (clean)             │ │
│  │   CXPVSP-121  feature/121-api-integration    ✓ (clean)             │ │
│  │   CXPVSP-120  hotfix/120-security-patch      ✗ (5 files)           │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│  Branch: feature/123-add-login  │  Worktree: ~/workspace/prism3/123     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Data Source
- Scan `~/workspace/prism3/` for directories
- Each directory: check for `.worktree-info` file containing:
  ```
  TICKET=CXPVSP-123
  BRANCH=feature/123-add-login
  ```

### State
- `worktrees`: []Worktree - list of all worktrees
- `selectedIndex`: int - currently selected worktree
- `cursor`: int - navigation cursor (0 = first item)

### Interactions

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Enter` | Select worktree → go to Individual Worktree screen |
| `n` | New worktree prompt |
| `d` | Delete selected worktree (with confirmation) |
| `r` | Refresh worktree list |
| `q` | Quit |

### New Worktree Prompt

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  New Worktree                                                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Ticket: [CXPVSP-123________________]                                       │
│                                                                             │
│  Branch Type: (f)eature  (b)ugfix  (h)otfix                               │
│                                                                             │
│  Description (optional): [add-login________________]                       │
│                                                                             │
│  Worktree: ~/workspace/prism3/123                                          │
│                                                                             │
│  [Enter: Create]  [Esc: Cancel]                                            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Delete Confirmation

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Delete Worktree?                                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  CXPVSP-123 at ~/workspace/prism3/123                                     │
│                                                                             │
│  This will:                                                                 │
│  - Remove git worktree                                                     │
│  - Delete directory                                                        │
│  - NOT delete branch                                                      │
│                                                                             │
│  [y: Delete]  [n: Cancel]                                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Screen: Individual Worktree

### Purpose
Manage a specific worktree with 4 tabs: Changes, Agent, Ticket, Plan.

### Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  CXPVSP-123: Add user login feature                     [Esc: Back]       │
│  Branch: feature/123-add-login  │  ~/workspace/prism3/123                 │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  [g: Changes]  [a: Agent]  [t: Ticket]  [p: Plan]                        │
│                                                                             │
│  ═══════════════════════════════════════════════════════════════════════    │
│                                                                             │
│  (Tab content displayed here)                                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### State

```go
type WorktreeView struct {
    Ticket    string
    Branch    string
    Path      string
    ActiveTab int // 0=Changes, 1=Agent, 2=Ticket, 3=Plan
    
    // Changes tab
    GitStatus   GitStatus
    DiffOutput  string
    
    // Agent tab
    AgentStatus AgentStatus // idle, running, done, error
    LastRun     string // timestamp
    
    // Ticket tab
    TicketInfo  TicketInfo
    TicketError string
    
    // Plan tab
    PlanContent string
    PlanExists  bool
}
```

### Tab Navigation

| Key | Action |
|-----|--------|
| `g` | Switch to Changes tab |
| `a` | Switch to Agent tab |
| `t` | Switch to Ticket tab |
| `p` | Switch to Plan tab |
| `h` / `←` | Previous tab |
| `l` / `→` | Next tab |

---

## Tab: Changes

### Purpose
View git status, staged/unstaged changes, and commit.

### Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  CXPVSP-123: Add user login feature                     [Esc: Back]       │
│  Branch: feature/123-add-login  │  ~/workspace/prism3/123                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  [g: Changes]  [a: Agent]  [t: Ticket]  [p: Plan]                        │
│                                                                             │
│  M apps/prism/components/Login.vue                                         │
│  M apps/prism/stores/auth.js                                                │
│  A apps/prism/utils/crypto.js                                              │
│  ? apps/prism/utils/newfile.js                                             │
│                                                                             │
│  diff --stat: 3 files changed, 150 insertions                             │
│                                                                             │
│  [s: Stage All]  [c: Commit]  [v: View Full Diff]                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Git Status Display
- **M** = Modified (staged if in green, unstaged if in red)
- **A** = Added
- **D** = Deleted
- **?** = Untracked

### Interactions

| Key | Action |
|-----|--------|
| `s` | Stage all changes (`git add -A`) |
| `c` | Commit prompt → enter message → `git commit -m "[TICKET]: message"` |
| `v` | View full diff (scrollable) |
| `r` | Refresh status |

### Commit Prompt

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Commit Message                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Ticket: CXPVSP-123                                                        │
│                                                                             │
│  Message: [Add OAuth login component________________________]              │
│                                                                             │
│  [Enter: Commit]  [Esc: Cancel]                                            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Full Diff View

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Full Diff (scroll with j/k)                              [Esc: Back]     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  diff --git a/apps/prism/components/Login.vue b/apps/prism/components...  │
│  --- a/apps/prism/components/Login.vue                                     │
│  +++ b/apps/prism/components/Login.vue                                     │
│  @@ -1,10 +1,15 @@                                                         │
│   <template>                                                               │
│  +  <div class="login">                                                    │
│  +    <button @click="loginWithGoogle">Google</button>                    │
│  +    <button @click="loginWithGitHub">GitHub</button>                    │
│  +  </div>                                                                │
│   </template>                                                               │
│                                                                             │
│  ...                                                                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Tab: Agent

### Purpose
Run OpenCode or Ralph agent on the worktree.

### Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  CXPVSP-123: Add user login feature                     [Esc: Back]       │
│  Branch: feature/123-add-login  │  ~/workspace/prism3/123                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  [g: Changes]  [a: Agent]  [t: Ticket]  [p: Plan]                        │
│                                                                             │
│  ═══════════════════════════════════════════════════════════════════════    │
│                                                                             │
│  Agent Status: IDLE                                                        │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                                                                       │ │
│  │  [o: Run OpenCode]  [r: Run Ralph]                                  │ │
│  │                                                                       │ │
│  │  Model: opencode/minimax-m2.5-free  [change]                        │ │
│  │                                                                       │ │
│  │  Ralph iterations: 10  [change]                                      │ │
│  │                                                                       │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Running State

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  CXPVSP-123: Add user login feature                     [Esc: Back]       │
│  Branch: feature/123-add-login  │  ~/workspace/prism3/123                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  [g: Changes]  [a: Agent]  [t: Ticket]  [p: Plan]                        │
│                                                                             │
│  ═══════════════════════════════════════════════════════════════════════    │
│                                                                             │
│  Agent Status: RUNNING (Iteration 3/10)                                   │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │ Running in tmux session: CXPVSP-123                                  │ │
│  │                                                                       │ │
│  │ Check tmux for output:                                               │ │
│  │   tmux attach -t CXPVSP-123                                          │ │
│  │                                                                       │ │
│  │ [s: Stop]  [v: View Diff]                                           │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Done State

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  CXPVSP-123: Add user login feature                     [Esc: Back]       │
│  Branch: feature/123-add-login  │  ~/workspace/prism3/123                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  [g: Changes]  [a: Agent]  [t: Ticket]  [p: Plan]                        │
│                                                                             │
│  ═══════════════════════════════════════════════════════════════════════    │
│                                                                             │
│  Agent Status: DONE (completed in 5 iterations)                           │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                                                                       │ │
│  │  Completed: 5 iterations                                            │ │
│  │  Last run: 2026-02-19 14:30                                         │ │
│  │                                                                       │ │
│  │  [v: View Diff]  [c: Commit]  [r: Run Again]                       │ │
│  │                                                                       │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Interactions

| Key | Action |
|-----|--------|
| `o` | Run OpenCode (sends to tmux) |
| `r` | Run Ralph (sends to tmux) |
| `s` | Stop running agent |
| `v` | View diff (switches to Changes tab) |
| `c` | Commit (switches to Changes tab) |
| `m` | Change model (prompt) |
| `i` | Change iterations (prompt, Ralph only) |

### Tmux Integration
- Session name = ticket number (e.g., `CXPVSP-123`)
- Working directory = worktree path
- Commands:
  - OpenCode: `opencode run "use ralph-implementer skill" -m <model>`
  - Ralph: `ralph.sh <iterations> -m <model>`

---

## Tab: Ticket

### Purpose
Display Jira ticket information, refetch, append notes.

### Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  CXPVSP-123: Add user login feature                     [Esc: Back]       │
│  Branch: feature/123-add-login  │  ~/workspace/prism3/123                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  [g: Changes]  [a: Agent]  [t: Ticket]  [p: Plan]                        │
│                                                                             │
│  [View Ticket File]                                                        │
│                                                                             │
│  ═══════════════════════════════════════════════════════════════════════    │
│                                                                             │
│  Summary: Add OAuth login functionality                                    │
│  ─────────────────────────────────────────────────                         │
│  Status: 🔵 In Progress  │  Assignee: J. Gonzalez  │  Priority: 🔴 High │
│                                                                             │
│  Description:                                                             │
│  ─────────────────────────────────────────────────                         │
│  Implement OAuth 2.0 login flow with Google and GitHub providers.         │
│  The implementation should:                                               │
│  - Use existing auth infrastructure                                       │
│  - Support both providers                                                  │
│  - Handle token refresh                                                    │
│  - Store tokens securely                                                   │
│                                                                             │
│  ─────────────────────────────────────────────────                         │
│                                                                             │
│  [r: Refetch from Jira]  [a: Append Notes]                               │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Ticket File Format (`~/.scratch/tickets/{TICKET}.md`)

```markdown
# CXPVSP-123

## Summary
Add OAuth login functionality

## Status
In Progress

## Assignee
J. Gonzalez

## Priority
High

## Description
Implement OAuth 2.0 login flow with Google and GitHub providers.
The implementation should:
- Use existing auth infrastructure
- Support both providers
- Handle token refresh
- Store tokens securely

## Notes
- Created: 2026-02-19
- Updated: 2026-02-19

<!-- APPENDED NOTES -->
```

### Interactions

| Key | Action |
|-----|--------|
| `r` | Refetch ticket from Jira → update cache |
| `a` | Append notes to ticket file (opens $EDITOR) |
| `e` | Edit entire ticket file (opens $EDITOR) |

### Error States

**No credentials:**
```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Error: Missing Jira Credentials                                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Please create:                                                            │
│    ~/.scratch/jira.email                                                   │
│    ~/.scratch/jira.token                                                  │
│                                                                             │
│  [Press any key to go back]                                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Ticket not found:**
```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Error: Ticket CXPVSP-999 not found                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  The ticket may not exist or you don't have access.                       │
│                                                                             │
│  [Press any key to go back]                                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Tab: Plan

### Purpose
Create, edit, and execute implementation plans from ticket information.

### Layout (No Plan)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  CXPVSP-123: Add user login feature                     [Esc: Back]       │
│  Branch: feature/123-add-login  │  ~/workspace/prism3/123                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  [g: Changes]  [a: Agent]  [t: Ticket]  [p: Plan]                        │
│                                                                             │
│  [View Ticket File]  [Plan File]                                           │
│                                                                             │
│  ═══════════════════════════════════════════════════════════════════════    │
│                                                                             │
│  No plan file found.                                                       │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │                                                                       │ │
│  │  [c: Create Plan from Ticket]                                        │ │
│  │                                                                       │ │
│  │  This will use AI to generate a plan based on the ticket info.     │ │
│  │                                                                       │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Layout (With Plan)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  CXPVSP-123: Add user login feature                     [Esc: Back]       │
│  Branch: feature/123-add-login  │  ~/workspace/prism3/123                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  [g: Changes]  [a: Agent]  [t: Ticket]  [p: Plan]                        │
│                                                                             │
│  [View Ticket File]  [Plan File]                                           │
│                                                                             │
│  ═══════════════════════════════════════════════════════════════════════    │
│                                                                             │
│  ## Plan: Add OAuth login                                                  │
│                                                                             │
│  ### Task 1: Create Login Component                                        │
│  - Create `apps/prism/components/Login.vue`                                │
│  - Add Google/GitHub OAuth buttons                                         │
│  - Add form validation                                                     │
│                                                                             │
│  ### Task 2: Implement Auth Store                                          │
│  - Create `apps/prism/stores/auth.js`                                      │
│  - Add OAuth callback handling                                             │
│  - Add token storage                                                       │
│                                                                             │
│  ### Task 3: Add Crypto Utils                                              │
│  - Create `apps/prism/utils/crypto.js`                                    │
│  - Implement token encryption                                             │
│                                                                             │
│  ─────────────────────────────────────────────────                         │
│                                                                             │
│  [c: Recreate]  [e: Edit]  [x: Execute with OpenCode]                     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Interactions

| Key | Action |
|-----|--------|
| `c` | Create plan (AI prompt → generates markdown) |
| `e` | Edit plan in $EDITOR |
| `x` | Execute plan → run OpenCode in tmux, switch to Agent tab |
| `v` | View plan file (read-only) |

### Plan File Format (`~/.scratch/tickets/{TICKET}_plan.md`)

```markdown
# Plan: CXPVSP-123 - Add OAuth login

## Overview
Implement OAuth 2.0 login flow with Google and GitHub providers.

## Tasks

### Task 1: Create Login Component
- Create `apps/prism/components/Login.vue`
- Add Google/GitHub OAuth buttons
- Add form validation

### Task 2: Implement Auth Store
- Create `apps/prism/stores/auth.js`
- Add OAuth callback handling
- Add token storage

### Task 3: Add Crypto Utils
- Create `apps/prism/utils/crypto.js`
- Implement token encryption
```

### AI Plan Generation Prompt

```
You are a senior software engineer. Create a detailed implementation plan 
for the following Jira ticket.

=== TICKET ===
Key: {TICKET_KEY}
Summary: {SUMMARY}
Description: {DESCRIPTION}
Priority: {PRIORITY}

=== REQUIREMENTS ===
1. Break down into small, actionable tasks (5-10 tasks max)
2. Each task should be implementable in 10-30 minutes
3. Include specific file paths where changes are needed
4. Output ONLY valid markdown

=== OUTPUT FORMAT ===
# Plan: {TICKET_KEY} - {SUMMARY}

## Overview
{Brief description}

## Tasks

### Task 1: {Title}
- {Subtask}
- {Subtask}

### Task 2: {Title}
...
```

---

## Configuration

### File Locations

| Path | Purpose |
|------|---------|
| `~/workspace/prism3/` | Worktree base directory |
| `~/workspace/prism3/{ticket}/` | Individual worktree |
| `~/workspace/prism3/{ticket}/.worktree-info` | Worktree metadata |
| `~/.scratch/jira.email` | Jira email |
| `~/.scratch/jira.token` | Jira API token |
| `~/.scratch/tickets/{TICKET}.md` | Cached ticket info |
| `~/.scratch/tickets/{TICKET}_plan.md` | Plan file |
| `~/.scratch/.currentTickets.txt` | Recent tickets list |

### Worktree Info File

```bash
# ~/workspace/prism3/123/.worktree-info
TICKET=CXPVSP-123
BRANCH=feature/123-add-login
CREATED=2026-02-19T14:30:00Z
```

### Jira API

- **URL**: `https://jira-us-aholddelhaize.atlassian.net/rest/api/2/issue/{TICKET}`
- **Auth**: Basic Auth (email:api_token)
- **Fields**:
  - `.fields.summary`
  - `.fields.description`
  - `.fields.status.name`
  - `.fields.assignee.displayName`
  - `.fields.priority.name`

---

## Keyboard Shortcuts Summary

### Global

| Key | Action |
|-----|--------|
| `Esc` | Go back / Cancel |
| `q` | Quit |
| `?` | Show help |

### All Worktrees Screen

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select worktree |
| `n` | New worktree |
| `d` | Delete worktree |
| `r` | Refresh |

### Individual Worktree Screen

| Key | Action |
|-----|--------|
| `g` | Changes tab |
| `a` | Agent tab |
| `t` | Ticket tab |
| `p` | Plan tab |
| `h` / `←` | Previous tab |
| `l` / `→` | Next tab |

### Changes Tab

| Key | Action |
|-----|--------|
| `s` | Stage all |
| `c` | Commit |
| `v` | View full diff |
| `r` | Refresh |

### Agent Tab

| Key | Action |
|-----|--------|
| `o` | Run OpenCode |
| `r` | Run Ralph |
| `s` | Stop |
| `v` | View diff |
| `c` | Commit |
| `m` | Change model |
| `i` | Change iterations |

### Ticket Tab

| Key | Action |
|-----|--------|
| `r` | Refetch from Jira |
| `a` | Append notes |
| `e` | Edit ticket file |

### Plan Tab

| Key | Action |
|-----|--------|
| `c` | Create plan |
| `e` | Edit plan |
| `x` | Execute plan |
| `v` | View plan |

---

## Implementation Phases

### Phase 1: Project Setup
1. Create Go module
2. Install Bubble Tea dependencies
3. Set up main.go with tea.Program
4. Create basic model and view

### Phase 2: Core Infrastructure
1. Config management
2. Jira client
3. Git operations
4. Worktree discovery

### Phase 3: All Worktrees Screen
1. List worktrees
2. New worktree flow
3. Delete worktree flow
4. Navigation

### Phase 4: Individual Worktree Screen
1. Tab navigation
2. Changes tab
3. Ticket tab
4. Plan tab
5. Agent tab

### Phase 5: Integration
1. Connect plan creation to AI
2. Connect agent execution to tmux
3. Error handling
4. Edge cases

### Phase 6: Polish
1. Vim bindings
2. Styles and theming
3. Help screens
4. Testing

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| No worktrees | Show "No worktrees found. Press n to create one." |
| Worktree deleted externally | Refresh list, show error |
| Jira API error | Show error inline, offer retry |
| Invalid ticket format | Show format hint |
| Git error | Show error message, offer retry |
| OpenCode/Ralph not installed | Show installation instructions |
| No plan to execute | Disable execute button |
| No ticket info for plan | Prompt to fetch ticket first |
