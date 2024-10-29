#/usr/local/bin/bash

tickets_file=~/.scratch/.currentTickets.txt

if ! test -f $tickets_file; then
  touch $tickets_file
fi

read -d '' -ra ticketList <"$tickets_file"

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
  echo "Committing to $fname"
  echo "RUNNING THE FOLLOWING COMMAND: git commit -m '$fname: $msg' $2"
  git commit -m "$fname: $msg" $2 --no-verify
  break
done
