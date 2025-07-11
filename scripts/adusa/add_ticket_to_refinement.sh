if [[ $1 == https://* ]]; then
  ticket=$(echo $1 | awk -F'/' '{print $NF}')
  printf "$(date +%Y-%m-%d) // $ticket\n[ ] - $1\n$(cat $script_dir/refinement/tickets.txt)" > $script_dir/refinement/tickets.txt
  echo "Ticket $ticket added to refinement"
else
  printf "$(date +%Y-%m-%d) // $1\n[ ] - https://jira-us-aholddelhaize.atlassian.net/browse/$1\n$(cat $script_dir/refinement/tickets.txt)" > $script_dir/refinement/tickets.txt
  echo "Ticket $1 added to refinement"
fi

