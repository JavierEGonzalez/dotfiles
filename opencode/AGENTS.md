Always reference .github/copilot-instructions.md from the current project for project-specific guidelines.

After implementing a line of code, make a test and run tests and linting.

## JIRA Ticket Handling
Ticket Format: CXPVSP-{n} formalized as [a-zA-Z]+-[\d]+. Most will be CXPVSP-

If user gives you a ticket number use your jira finder skill.
If user simply mentions current ticket, look at the branch for the ticket name.

When users mention JIRA tickets (in format CXPVSP-<number> or full URLs), OpenCode should:
1. Check for existing local files at `~/.scratch/tickets/{ticket_id}.md` or `~/.scratch/tickets/{ticket_id}.txt`
2. If found, use the existing content
3. If not found, fetch from JIRA and save locally (the script saves to `~/.scratch/tickets/`)
4. Reference the jira-fetcher skill for implementation details

## Planning Mode

If there's a ticket associated with the current plan, name the plan {ticket}_plan.md and save it to `./.scratch`
otherwise for planning, always store the plan in a (new if doesn't exist) `./.scratch` directory

When making a plan, provide context at the top, where we are in the implementation/what has been done, and code 
examples with references to the proposed solution. 
