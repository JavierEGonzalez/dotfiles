---
name: jira-fetcher
description: Fetch JIRA ticket info. Use when user mentions CXPVSP-<number> or JIRA URLs.
---

# JIRA Ticket Fetcher Skill

When user mentions `CXPVSP-<number>` or a JIRA URL:

1. Check for existing `jira_tickets/{ticket_id}.md`
2. If exists, read it
3. If not, fetch via script and save

## Script Usage

```bash
~/.config/opencode/skill/jira-fetcher/jira_fetch.sh CXPVSP-123
```
