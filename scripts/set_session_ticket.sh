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

if git branch --show-current | grep -q $ticket; then
  echo "In Ticket Branch '$(git branch --show-current)'"
  exit 0
fi

if ! git branch | grep -q $ticket; then
  echo "No branches for this ticket"
  read -p "Do you want to create a branch with this ticket? [y/n] " create_branch

  if [[ $crete_branch =~ ^[yY]$ ]]; then
    read -p "Is this a [f]eature, [b]ugfix, or [h]otfix? [f/b/h] " branch_type
    read -p "Enter a description for the branch or leave empty: " description
    if [[ $description ]]; then
      description = "-$description"
    fi

    case $branch_type in
      f|F) prefix="feature" ;;
      b|B) prefix="bugfix" ;;
      h|H) prefix="hotfix" ;;
      *) echo "Invalid option, defaulting to feature"; prefix="feature" ;;
    esac

    branch_name="$prefix/$ticket$(echo $description | sed 's/ /-/g')"

    git checkout -b $branch_name
  fi
else
  echo "Branches with ticket, choose to checkout:"
  branchList=($(git branch | grep $ticket | sed 's/^[ *]*//'))
  select fname in ${branchList[@]} "Do not checkout branch"; do
    echo "selected $fname"
    if [[ $fname != "Do not checkout branch" ]]; then
      git checkout $fname
    fi
    break
  done
fi
