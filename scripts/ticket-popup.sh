#!/bin/bash

JIRA_URL="https://jira-us-aholddelhaize.atlassian.net/rest/api/2/issue"
CRED_DIR="$HOME/.scratch"
TICKET_DIR="$HOME/workspace/Work Brain/Work/tickets"

mkdir -p "$TICKET_DIR"

if [[ -z "$TICKET" ]]; then
    echo "Enter ticket (e.g., PROJ-123 or 8236):"
    read -r TICKET
fi

if [[ -z "$TICKET" ]]; then
    exit 0
fi

if [[ "$TICKET" == *"-"* ]]; then
    TICKET_KEY="$TICKET"
else
    TICKET_KEY="CXPVSP-$TICKET"
fi

CACHE_FILE="$TICKET_DIR/${TICKET_KEY}.txt"

refresh_ticket() {
    EMAIL=$(cat "$CRED_DIR/jira.email" 2>/dev/null)
    API_KEY=$(cat "$CRED_DIR/jira.token" 2>/dev/null)

    if [[ -z "$EMAIL" || -z "$API_KEY" ]]; then
        echo "Error: Missing Jira credentials" >&2
        exit 1
    fi

    RESPONSE=$(curl -s -u "$EMAIL:$API_KEY" \
        -H "Accept: application/json" \
        "$JIRA_URL/$TICKET_KEY")

    if echo "$RESPONSE" | grep -q '"errorMessages"'; then
        ERROR_MSG=$(echo "$RESPONSE" | jq -r '.errorMessages[0]' 2>/dev/null)
        echo "Error: $ERROR_MSG"
        exit 1
    fi

    SUMMARY=$(echo "$RESPONSE" | jq -r '.fields.summary // "No summary"')
    DESCRIPTION=$(echo "$RESPONSE" | jq -r '.fields.description // "No description"')
    STATUS=$(echo "$RESPONSE" | jq -r '.fields.status.name // "Unknown"')
    ASSIGNEE=$(echo "$RESPONSE" | jq -r '.fields.assignee.displayName // "Unassigned"')
    PRIORITY=$(echo "$RESPONSE" | jq -r '.fields.priority.name // "None"')

    {
        echo "=== $TICKET_KEY ==="
        echo ""
        echo "Summary: $SUMMARY"
        echo "Status: $STATUS"
        echo "Assignee: $ASSIGNEE"
        echo "Priority: $PRIORITY"
        echo ""
        echo "Description:"
        echo "$DESCRIPTION"
    } > "$CACHE_FILE"
}

if [[ -f "$CACHE_FILE" && "$REFRESH" != "1" ]]; then
    cat "$CACHE_FILE"
else
    refresh_ticket
    cat "$CACHE_FILE"
fi

echo ""
echo "Press R to refresh, E to edit notes, O to OpenCode, Enter to close..."
read -r -n 1 key
echo ""

if [[ "$key" == "r" || "$key" == "R" ]]; then
    REFRESH=1 "$0"
elif [[ "$key" == "e" || "$key" == "E" ]]; then
    "$EDITOR" "$CACHE_FILE"
    "$0"
elif [[ "$key" == "o" || "$key" == "O" ]]; then
    if [[ -n "$SESSION_ONLY" ]]; then
        SESSION_TICKET=$(head -1 "$CRED_DIR/.currentTickets.txt" 2>/dev/null)
        if [[ -z "$SESSION_TICKET" ]]; then
            echo "No session ticket found."
            exit 1
        fi
        TICKET_KEY="$SESSION_TICKET"
        CACHE_FILE="$TICKET_DIR/${TICKET_KEY}.txt"
        if [[ ! -f "$CACHE_FILE" ]]; then
            REFRESH=1 "$0"
        fi
    fi
    PLAN_FILE="$TICKET_DIR/${TICKET_KEY}_plan.md"
    TICKET_INFO="$(cat "$CACHE_FILE" | sed 's/"/\\"/g')"
    opencode --prompt "Hey, I'm working on this task, make a plan to address these issues and write it to $PLAN_FILE. Ticket info:"$'\n'"$TICKET_INFO" -m "opencode/minimax-m2.5-free"
    echo "Press Enter to close..."
    read -r
fi
