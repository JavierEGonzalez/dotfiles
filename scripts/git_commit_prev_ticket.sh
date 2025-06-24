#/usr/local/bin/bash

fname=$(git log | head -n 5 | tail -n 1 | sed -e 's/^[[:space:]]*//; s/:.*//')

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

echo "Committing to $fname"
echo "Message:"
read -p ">" msg

commit_ticket $fname $msg
