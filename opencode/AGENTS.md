Always reference .github/copilot-instructions.md from the current project for project-specific guidelines.

After implementing a line of code, make a test and run tests and linting.

## JIRA Ticket Handling
Ticket Format: CXPVSP-{n} formalized as [a-zA-Z]+-[\d]+. Most will be CXPVSP-

If user gives you a ticket number use your jira finder skill.
If user simply mentions current ticket, look at the branch for the ticket name.

When users mention JIRA tickets (in format CXPVSP-<number> or full URLs), OpenCode should:
1. Check for existing local files (ticket_id.txt or ticket_id.md)
2. If found, use the existing content
3. If not found, fetch from JIRA and save locally
4. Reference the jira-fetcher skill for implementation details
