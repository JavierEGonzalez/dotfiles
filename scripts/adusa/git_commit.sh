#/usr/local/bin/bash
source ~/.scratch/scripts/session_ticket_functions.sh

tickets_file=~/.scratch/.currentTickets.txt

if ! test -f $tickets_file; then
  touch $tickets_file
fi

read -d '' -ra ticketList <"$tickets_file"

prompt_to_stage_if_needed

echo 'Which ticket is this commit related to?'
select fname in ${ticketList[@]} "custom"; do
  if [[ $fname == "custom" ]]; then
    echo "Add ticket to $tickets_file"
    read -p ">" fname
    printf '%s\n' $fname $(cat $tickets_file) >$tickets_file
    cat $tickets_file
  fi
  if [[ -z $1 ]]; then
    echo "Please provide a commit message"
    read -p ">" msg
  else
    msg=$1
  fi
  commit_ticket "$fname" "$msg"
  break
done