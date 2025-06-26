tickets_file=~/.scratch/.currentTickets.txt

if ! test -f $tickets_file; then
  touch $tickets_file
fi

if [[ $1 ]]; then
  ticket_name="CXPVSP-$1"
else
  echo "No ticket passed as param"
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

