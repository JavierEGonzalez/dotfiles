---
name: pr-review
description: Pull and address GitHub PR review comments. Use when asked to pull review comments, address pr comments, fix pr comments, handle pr review.
---

# PR Review Comments Skill

Fetch PR review comments from the current branch's PR and help address them one by one.

## Workflow

1. **Detect PR**: Get PR number from current branch using `gh pr view --json number,title,headRefName`
2. **Fetch Comments**: Use `gh api` to get review threads (better than `gh pr view --json` for threads)
3. **Filter Resolved**: Skip any thread where `is_resolved: true`
4. **Present Comments**: Display in clear format, grouped by file
5. **Ask User**: For each comment, ask to address or skip

## Fetching Data

Use this GraphQL query for best thread display:

```bash
gh api graphql -f query='
query($owner: String!, $repo: String!, $prNumber: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $prNumber) {
      title
      url
      reviewThreads(first: 100) {
        nodes {
          isResolved
          isCollapsed
          comments(first: 10) {
            nodes {
              body
              author { login }
              path
              line
              diffSide
              createdAt
            }
          }
        }
      }
    }
  }
}
' -f owner={owner} -f repo={repo} -f prNumber={number}
```

Also get regular issue comments:

```bash
gh api repos/{owner}/{repo}/issues/{prNumber}/comments --jq '.[].{body: .body, author: .user.login, createdAt: .created_at}'
```

## Display Format

Present each thread like this:

```
📁 src/components/Foo.vue:42
👤 @reviewer • 2 days ago
┌─────────────────────────────────────────┐
│ Consider using a computed property here │
│ instead of a method for better perf.    │
└─────────────────────────────────────────┘
```

Group by: File path, then General Comments

## Handling Comments

For each unresolved comment:

1. **Check if it's a nit**: Look for prefixes like "nit:", "nitpick:", "typo:", "small:", "minor:"
   - Ask user: "This is a nitpick. Address it? [y/n/bulk]"
   - If "bulk", collect all nits and ask once at end

2. **For fixable comments**:
   - Read the relevant file
   - Propose the code change
   - Show diff: `git diff --no-color`
   - Ask user: "Apply this fix? [y/n/skip]"
   - If yes: `git add -A && git commit -m "fix: addressed PR review comment"`

3. **For questions/explanations**:
   - Ask user: "Reply with explanation? [y/n]"
   - If yes, draft reply, let user edit, then post via `gh api`

## Replying to Comments

Post a comment reply:

```bash
gh api repos/{owner}/{repo}/pulls/{prNumber}/comments/{comment_id}/replies -f body='Your reply here'
```

Or issue comment:

```bash
gh api repos/{owner}/{repo}/issues/{prNumber}/comments -f body='Your reply here'
```

## Bulk Nit Handling

If user chooses "bulk" for nits:
1. List all nits
2. Ask "Fix all X nits? [y/n]"
3. If yes, apply all fixes in one commit or multiple commits

## Important

- NEVER auto-commit without explicit user approval
- Always show the diff before committing
- Skip resolved threads (check `isResolved` field)
- Use `gh repo` to get current owner/repo
