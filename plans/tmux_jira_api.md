# TMux Jira Popup - Implementation Plan

## Overview
Add a tmux key binding (Shift+J) that opens a popup to fetch and display Jira ticket information.

## Components

### 1. Script: `~/.scripts/ticket-popup.sh`

**Behavior:**
1. Check if `$TICKET` environment variable is set
   - If set → use it directly
   - If NOT set → prompt user for ticket in the popup

**Input handling:**
- If input contains `-` → use as-is (e.g., `PROJ-123`)
- If input is just numbers → prepend `CXPVSP-` (e.g., `8236` → `CXPVSP-8236`)

**API Call:**
- Endpoint: `GET https://jira-us-aholddelhaize.atlassian.net/rest/api/2/issue/{TICKET}`
- Auth: Basic auth using credentials from:
  - `~/.scratch/jira.email` (email)
  - `~/.scratch/jira.key` (API token)
- Headers: `Accept: application/json`

**Output:**
- Display summary and description in the popup

### 2. Tmux Config Update

**File:** `~/.tmux.conf`

**Add binding:**
```
bind -r J display-popup -w 80% -h 80% -E "~/.scripts/ticket-popup.sh"
```

## Dependencies
- `curl` for API calls
- Credentials files must exist:
  - `~/.scratch/jira.email`
  - `~/.scratch/jira.key`

## Testing
1. Test with `$TICKET` environment variable set
2. Test without `$TICKET` (should prompt)
3. Test with full ticket key (e.g., `PROJ-123`)
4. Test with just number (e.g., `8236` → should become `CXPVSP-8236`)
