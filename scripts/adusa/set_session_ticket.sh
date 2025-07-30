tickets_file=~/.scratch/.currentTickets.txt

if ! test -f $tickets_file; then
  touch $tickets_file
fi

if [[ $1 ]]; then
  if [[ "$1" =~ ^[a-zA-Z0-9]+-[0-9]+$ ]]; then
    ticket_name="$1"
  elif [[ "$1" =~ ^[0-9]+$ ]]; then
    ticket_name="CXPVSP-$1"
  else
    echo "Invalid ticket format. Please use a number (e.g., 1919) or the format [alphanumeric]-[number] (e.g., CXPVSP-1919)." >&2
    exit 1
  fi

  export ticket="$ticket_name" && tmux setenv ticket "$ticket_name"

  echo "Added $ticket to session"
  if ! grep -q $ticket $tickets_file; then
    echo "Added $ticket to $tickets_file"
    printf '%s\n' $ticket $(cat $tickets_file) > $tickets_file
  else
    echo "Ticket already in ticket file"
  fi
else
  echo "No ticket passed as param"
fi
