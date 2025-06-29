if git branch --show-current | grep -q $ticket; then
  echo "In Ticket Branch '$(git branch --show-current)'"
  exit 0
fi

if [[ -z $ticket ]]; then
  echo "no ticket in environment"
  exit 0
fi

if ! git branch | grep -q $ticket; then
  echo "No branch found for: $ticket"
  read -p "Do you want to create a branch? [y/N] " create_branch

  if [[ $create_branch =~ ^[yY]$ ]]; then
    read -p "Is this a [f]eature, [b]ugfix, or [h]otfix? [f/b/h] " branch_type
    read -p "Enter a description for the branch or leave empty: " description
    if [[ $description ]]; then
      dash_description="-$(echo $description | sed 's/ /-/g')"
    fi

    case $branch_type in
      f|F) prefix="feature" ;;
      b|B) prefix="bugfix" ;;
      h|H) prefix="hotfix" ;;
      *) echo "Invalid option, defaulting to feature"; prefix="feature" ;;
    esac

    branch_name="$prefix/$ticket$dash_description"

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
