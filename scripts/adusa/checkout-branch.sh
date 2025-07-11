#/usr/local/bin/bash
if [[ -z $1 ]] ; then
  tickets_file=~/.scratch/.currentTickets.txt

  if ! test -f $tickets_file; then
    touch $tickets_file
  fi

  read -d '' -ra ticketList <"$tickets_file"
  echo 'Which ticket to checkout branch for?'
  select fname in ${ticketList[@]} "custom"; do
    if [[ $fname == "custom" ]]; then
      read -p ">" fname
      printf '%s\n' $fname $(cat $tickets_file) >$tickets_file
      echo "Adding $fname to ticket list"
    fi
    echo "Checking out branch with ticket $fname in name"
    git branch | grep $fname | xargs git checkout
    break
  done
else
  echo "Checking out branch with ticket $1 in name"
  git branch | grep $1 | xargs git checkout
fi
