function commit_ticket() {
  echo "RUNNING THE FOLLOWING COMMAND: git commit -m '[$ticket] $msg' --no-verify"
  git commit -m "[$1] $2" --no-verify
}

function check_commited_files() {
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
}
