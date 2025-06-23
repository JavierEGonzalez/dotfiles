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


if ! git branch | grep -q $ticket; then
  read -p "Do you want to create a branch with this ticket? [y/n] " create_branch

  if [[ $create_branch == "y" || $create_branch == "Y" ]]; then
    read -p "Enter a description for the branch: " description
    # ask for feature|bugfix|hotfix
    read -p "Is this a [f]eature, [b]ugfix, or [h]otfix? [f/b/h] " branch_type
    # read value and set prefix to be full word
    case $branch_type in
      f|F) prefix="feature" ;;
      b|B) prefix="bugfix" ;;
      h|H) prefix="hotfix" ;;
      *) echo "Invalid option, defaulting to feature"; prefix="feature" ;;
    esac

    branch_name="$prefix/$ticket-$(echo $description | sed 's/ /-/g')"

    git checkout -b $branch_name
    echo "Switched to new branch: $branch_name"
  fi
else
  echo "found these branches with ticket in name:"
  
  branchList=($(git branch | grep $ticket | sed 's/^[ *]*//'))
  select fname in ${branchList[@]} "Do not checkout branch"; do
    echo "selected $fname"
    if [[ $fname != "Do not checkout branch" ]]; then
      git checkout $fname
    fi
    break
  done
fi
