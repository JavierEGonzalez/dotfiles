#/usr/local/bin/bash
source ~/.scratch/scripts/session_ticket_functions.sh

if [[ -z $ticket || $ticket == 'CXPVSP-' ]]; then
  echo "Ticket is not set"
  # check if ticket is set in branch name
  branch=$(git rev-parse --abbrev-ref HEAD)
  if [[ $branch =~ (CXPVSP-[0-9]+) ]]; then
    ticket=${BASH_REMATCH[1]}
    echo "Found ticket in branch name: $ticket"
    export ticket=$ticket 
    #redirect to dev/null if there's error
    tmux setenv ticket $ticket 2>/dev/null
  else
    echo "Setup session ticket:"
    read -p ">" ticket
    export ticket=$ticket && tmux setenv ticket $ticket
  fi
fi

prompt_to_stage_if_needed

echo "Committing to $ticket"

if [[ -z $1 ]]; then
  echo "Message:"
  read -p ">" msg
else
  msg=$1
  echo "Message: '$msg'"
fi

commit_ticket $ticket "$msg"
