#/usr/local/bin/bash
source ~/.scratch/scripts/session_ticket_functions.sh

if [[ -z $ticket || $ticket == 'CXPVSP-' ]]; then
  echo "Ticket is not set"
  echo "Setup session ticket:"
  read -p ">" ticket
  export ticket=$ticket && tmux setenv ticket $ticket
fi

check_commited_files

echo "Committing to $ticket"

if [[ -z $1 ]]; then
  echo "Message:"
  read -p ">" msg
else
  msg=$1
  echo "Message: '$msg'"
fi

commit_ticket $ticket $msg
