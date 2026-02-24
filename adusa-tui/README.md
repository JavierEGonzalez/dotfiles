# adusa-tui

## Config File

Configuration is stored at `~/.config/adusa-tui/config.toml`.

```toml
diff-viewer = "nvim -c DiffviewOpen"

scratch-dir = ""                      # defaults to ~/.scratch
tickets-dir = ""                      # defaults to scratch-dir/tickets

[[repo]]
name = "prism3"
path = "/Users/javier/workspace/prism3"
default-branch = "main"

# Add more repos as needed:
# [[repo]]
# name = "other-repo"
# path = "/Users/javier/workspace/other"
# default-branch = "develop"

[[jira]]
email-path = ""                       # defaults to scratch-dir/jira.email
token-path = ""                       # defaults to scratch-dir/jira.token
domain = "yourcompany"                # e.g., "prism3" for prism3.atlassian.net
```

The config file is created automatically on first run with default values if it doesn't exist.

## Usage

### Open a specific worktree

Pass a worktree path or ticket ID as an argument to open directly to that worktree:

```bash
./adusa-tui /path/to/worktree
./adusa-tui CXPVSP-8423
```

The app will first try to match by path, then fall back to matching by ticket ID.
