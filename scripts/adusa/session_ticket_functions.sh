function commit_ticket() {
  echo "RUNNING THE FOLLOWING COMMAND: git commit -m '[$1]: $2' --no-verify"
  git commit -m "[$1]: $2" --no-verify
}

function prompt_to_stage_if_needed() {
    if ! git diff --staged --quiet; then
        return 0 # Staged changes exist, proceed with commit.
    fi

    if ! git diff --quiet; then
        echo "No changes added to commit. Stage all?"
        select shouldAdd in "yes" "cancel"; do
            case $shouldAdd in
            yes)
                git add .
                echo "All changes staged."
                break
                ;;
            cancel)
                echo "Commit cancelled."
                exit 1
                ;;
            esac
        done
    else
        echo "No changes to commit, working tree clean."
        exit 1
    fi
}