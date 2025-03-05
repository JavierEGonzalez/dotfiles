#/usr/local/bin/bash

if [[ -z $ticket || $ticket == 'CXPVSP-' ]]; then
  echo "Ticket is not set"
  echo "Setup session ticket:"
  read -p ">" ticket
  export ticket=$ticket && tmux setenv ticket $ticket
fi

if [[ $(git status | tail -n 1) == "nothing to commit, working tree clean" ]]; then
  echo "No changes to commit"
  exit 1
fi
if [[ $(git status | tail -n 1 | sed -e "s/ (.*//g") == "no changes added to commit" ]]; then
  echo "No changes added to commit, run git add .?"
  select shouldAdd in "yes" "cancel"; do
    case $shouldAdd in
    yes)
      git add .
      break
      ;;
    cancel)
      exit 1
      ;;
    esac
  done
fi

echo "Committing to $ticket"

if [[ -z $1 ]]; then
  echo "Message:"
  read -p ">" msg
else
  msg=$1
  echo "Message: '$msg'"
fi

echo "RUNNING THE FOLLOWING COMMAND: git commit -m '$ticket: $msg' --no-verify"
git commit -m "$ticket: $msg" --no-verify
