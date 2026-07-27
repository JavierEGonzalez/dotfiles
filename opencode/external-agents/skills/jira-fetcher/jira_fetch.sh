#!/bin/bash

# JIRA Ticket Fetcher Script
# Fetches JIRA ticket information and saves it locally

# Configuration
JIRA_BASE_URL="https://jira-us-aholddelhaize.atlassian.net"
JIRA_API="${JIRA_BASE_URL}/rest/api/2/issue"
OUTPUT_DIR="$HOME/.scratch/tickets"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Function to extract ticket ID from various formats
extract_ticket_id() {
    local input="$1"
    
    # If it's already in the format PROJECT-NUMBER
    if [[ "$input" =~ ^[A-Z]+-[0-9]+$ ]]; then
        echo "$input"
        return 0
    fi
    
    # If it's a full URL
    if [[ "$input" =~ https?://[^/]+/browse/([A-Z]+-[0-9]+) ]]; then
        echo "${BASH_REMATCH[1]}"
        return 0
    fi
    
    # If it contains the pattern somewhere
    if [[ "$input" =~ ([A-Z]+-[0-9]+) ]]; then
        echo "${BASH_REMATCH[1]}"
        return 0
    fi
    
    # If none matched
    echo "Error: Could not extract ticket ID from '$input'" >&2
    return 1
}

# Function to check if file exists for ticket
file_exists_for_ticket() {
    local ticket_id="$1"
    
    if [[ -f "${OUTPUT_DIR}/${ticket_id}.txt" ]] || [[ -f "${OUTPUT_DIR}/${ticket_id}.md" ]]; then
        return 0
    else
        return 1
    fi
}

# Function to read existing ticket file
read_existing_ticket() {
    local ticket_id="$1"
    
    # Prefer .md if both exist, otherwise use .txt
    if [[ -f "${OUTPUT_DIR}/${ticket_id}.md" ]]; then
        cat "${OUTPUT_DIR}/${ticket_id}.md"
    elif [[ -f "${OUTPUT_DIR}/${ticket_id}.txt" ]]; then
        cat "${OUTPUT_DIR}/${ticket_id}.txt"
    fi
}

# Function to fetch ticket from JIRA
fetch_ticket_from_jira() {
    local ticket_id="$1"
    
    # Load credentials from ~/.scratch/jira.* files if not already set
    if [[ -z "$JIRA_USERNAME" && -f "$HOME/.scratch/jira.email" ]]; then
        JIRA_USERNAME=$(cat "$HOME/.scratch/jira.email")
    fi
    if [[ -z "$JIRA_API_TOKEN" && -f "$HOME/.scratch/jira.token" ]]; then
        JIRA_API_TOKEN=$(cat "$HOME/.scratch/jira.token")
    fi
    if [[ -f "$HOME/.scratch/jira.domain" ]]; then
        local raw_domain
        raw_domain=$(cat "$HOME/.scratch/jira.domain" | tr -d '[:space:]')
        # If it doesn't start with http, construct the full URL
        if [[ "$raw_domain" != http* ]]; then
            JIRA_BASE_URL="https://${raw_domain}.atlassian.net"
        else
            JIRA_BASE_URL="$raw_domain"
        fi
        JIRA_API="${JIRA_BASE_URL}/rest/api/2/issue"
    fi

    # Check if we have credentials
    if [[ -z "$JIRA_USERNAME" || -z "$JIRA_API_TOKEN" ]]; then
        echo "Error: JIRA_USERNAME and JIRA_API_TOKEN environment variables must be set, or stored in ~/.scratch/jira.email and ~/.scratch/jira.token" >&2
        return 1
    fi
    
    # Fetch from JIRA API
    local response
    response=$(curl -s -u "$JIRA_USERNAME:$JIRA_API_TOKEN" \
        -H "Content-Type: application/json" \
        "${JIRA_API}/${ticket_id}")
    
    if [[ $? -ne 0 ]]; then
        echo "Error: Failed to fetch ticket ${ticket_id} from JIRA" >&2
        return 1
    fi
    
    # Extract relevant fields using jq (if available) or basic parsing
    if command -v jq >/dev/null 2>&1; then
        local summary description status issue_type
        summary=$(echo "$response" | jq -r '.fields.summary // "No summary"')
        description=$(echo "$response" | jq -r '.fields.description // "No description"')
        status=$(echo "$response" | jq -r '.fields.status.name // "No status"')
        issue_type=$(echo "$response" | jq -r '.fields.issuetype.name // "No type"')
        
        # Create plain text version
        cat > "${OUTPUT_DIR}/${ticket_id}.txt" <<EOF
Ticket: ${ticket_id}
Summary: ${summary}
Status: ${status}
Type: ${issue_type}

Description:
${description}
EOF
        
        # Create markdown version
        cat > "${OUTPUT_DIR}/${ticket_id}.md" <<EOF
# ${ticket_id}: ${summary}

**Status:** ${status}
**Type:** ${issue_type}

## Description

${description}

---

*Fetched from JIRA on $(date)*
EOF
    else
        # Fallback without jq - just save raw JSON
        echo "Warning: jq not available, saving raw JSON response" >&2
        echo "$response" > "${OUTPUT_DIR}/${ticket_id}.txt"
        echo "# ${ticket_id}\n\nRaw JIRA response:\n\n\`\`\`json\n${response}\n\`\`\`" > "${OUTPUT_DIR}/${ticket_id}.md"
    fi
    
    echo "Successfully fetched and saved ticket ${ticket_id}"
    return 0
}

# Main function
main() {
    if [[ $# -lt 1 ]]; then
        echo "Usage: $0 <ticket-id-or-url>" >&2
        echo "Example: $0 CXPVSP-123" >&2
        echo "Example: $0 https://jira-us-aholddelhaize.atlassian.net/browse/CXPVSP-123" >&2
        exit 1
    fi
    
    local input="$1"
    local ticket_id
    
    # Extract ticket ID
    ticket_id=$(extract_ticket_id "$input")
    if [[ $? -ne 0 ]]; then
        exit 1
    fi
    
    echo "Processing ticket: ${ticket_id}"
    
    # Check if file already exists
    if file_exists_for_ticket "$ticket_id"; then
        echo "Found existing ticket file for ${ticket_id}"
        read_existing_ticket "$ticket_id"
    else
        echo "No existing file found for ${ticket_id}, fetching from JIRA..."
        if fetch_ticket_from_jira "$ticket_id"; then
            # Read and output the newly created file
            read_existing_ticket "$ticket_id"
        else
            exit 1
        fi
    fi
}

# Execute main function with all arguments
main "$@"