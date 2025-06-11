tickets_file=~/.scratch/.currentTickets.txt

if ! test -f $tickets_file; then
  touch $tickets_file
fi

export ticket=CXPVSP-$1 && tmux setenv ticket CXPVSP-$1

echo "Added $ticket to session"
if ! grep -q $ticket $tickets_file; then
  echo "Added $ticket to $tickets_file"
  printf '%s\n' $ticket $(cat $tickets_file) > $tickets_file
else
  echo "Ticket already in ticket file"
fi
