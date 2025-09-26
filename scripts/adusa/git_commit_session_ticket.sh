#/usr/local/bin/bash
source ~/.scratch/scripts/session_ticket_functions.sh

if [[ -z $ticket || $ticket == 'CXPVSP-' ]]; then
  echo "Ticket is not set"
  branch=$(git rev-parse --abbrev-ref HEAD)
  if [[ $branch =~ (CXPVSP-[0-9]+) ]]; then
    branchTicket=${BASH_REMATCH[1]}
    echo "Found ticket in branch name: $branchTicket"
    set_session_ticket $branchTicket
  else
    echo "Setup session ticket:"
    read -p ">" newSessionTicket
    set_session_ticket $newSessionTicket
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
