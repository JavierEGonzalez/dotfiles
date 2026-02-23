# ADUSA TUI - Architecture Reference

## CLI Entry Point

```
adusa-tui [ticket]
```

- **With ticket**: Opens directly to Individual Worktree screen for that ticket
- **Without ticket**: Shows All Worktrees list screen

## Screen Flow

```
flowchart TD
    A[Call CLI] -->|Without Args| B
    A -->|With Ticket Arg| C[Start]
    B[All Worktrees Screen] -->|Choose| B1
        B1[All Worktrees] --> C
    B -->|Choose| B2
        B2[Make New Worktree] --> C
    C[Individual Worktree Screen] -->|Press Esc| B
    C -->|Tab One| C1[View Changes]
    C -->|Tab Two| C2[View Agent Session]
    C -->|Tab Three| C3[View Ticket Summary]
    C3 -->|Action| C3C{Append Info to Ticket File}
    C3 -->|Action| C3D{Refetch Ticket File}
    C -->|Tab Four| D
        D[Plan File] -->|Opt 1| D1
        D1{Create Plan From Ticket Info} --> D2
        D -->|Opt 2| D2{View or Edit Plan File} --> D2a
        D2a[Editor] -->|On Close| D
        D -->|Opt 3 if file exists| D3{Execute Plan File}
        D3 -->|Navigate To| C2
```

## File Locations

| Path | Purpose |
|------|---------|
| `~/workspace/prism3/` | Worktree base directory |
| `~/workspace/prism3/{ticket}/` | Individual worktree |
| `~/workspace/prism3/{ticket}/.worktree-info` | Worktree metadata |
| `~/.scratch/jira.email` | Jira email |
| `~/.scratch/jira.token` | Jira API token |
| `~/.scratch/jira.domain` | Jira domain (optional, defaults to atlassian.com) |
| `~/.scratch/tickets/{TICKET}.md` | Cached ticket info |
| `~/.scratch/tickets/{TICKET}_plan.md` | Plan file |

## Jira API

- **URL**: `https://{domain}.atlassian.net/rest/api/3/issue/{TICKET}`
- **Auth**: Basic Auth (email:api_token)
- **Description format**: Atlassian Document Format (ADF) - nested JSON

## Keyboard Shortcuts

### Global
| Key | Action |
|-----|--------|
| `Esc` | Go back / Cancel |
| `q` | Quit |
| `?` | Show help |

### All Worktrees Screen
| Key | Action |
|-----|--------|
| `j/k` / arrows | Navigate |
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
| `h/l` / arrows | Prev/next tab |

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

## Project Structure

```
adusa-tui/
├── main.go                         # Entry point, screen routing
├── internal/
│   ├── config/config.go            # Paths, credentials
│   ├── git/worktree.go             # Git operations (list, create, delete, status, diff, commit)
│   ├── jira/client.go              # Jira REST API v3 client, caching
│   ├── types/
│   │   ├── agent.go                # AgentStatus enum
│   │   ├── ticket.go               # TicketInfo struct
│   │   └── worktree.go             # Worktree struct
│   └── ui/screens/
│       ├── worktrees.go            # All Worktrees list screen
│       ├── create_worktree.go      # New worktree wizard
│       ├── delete_worktree.go      # Delete confirmation
│       └── worktree.go             # Individual worktree (4 tabs, 924 lines)
```
