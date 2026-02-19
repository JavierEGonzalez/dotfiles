#!/usr/bin/env bash
set -euo pipefail

DEFAULT_BASE="~/workspace/prism3"
SECRETS_SOURCE="$HOME/.scratch/.env.secrets"

get_ticket() {
  if [[ -n "${1:-}" ]]; then
    echo "$1"
  else
    read -p "Enter ticket number: " ticket
    echo "$ticket"
  fi
}

get_base_dir() {
  local default_base expanded_default
  expanded_default=$(eval echo "$DEFAULT_BASE")
  read -p "Worktree base directory [$expanded_default]: " base_dir
  if [[ -z "$base_dir" ]]; then
    echo "$expanded_default"
  else
    echo "$base_dir"
  fi
}

find_branch() {
  local ticket=$1
  git branch --list "*${ticket}*" | sed 's/^[ *+]*//'
}

find_remote_branch() {
  local ticket=$1
  git branch -r --list "*${ticket}*" | sed 's/^[ *+]*origin\///' | sort -u
}

choose_branch() {
  local ticket=$1
  shift
  local branches=("$@")
  local count=${#branches[@]}

  if [[ $count -eq 0 ]]; then
    return 1
  elif [[ $count -eq 1 ]]; then
    echo "${branches[0]}"
    return 0
  fi

  echo "Multiple branches found for '$ticket':"
  local i=1
  for branch in "${branches[@]}"; do
    echo "  [$i] $branch"
    ((i++))
  done
  echo "  [$i] Create new branch"

  read -p "Choose branch [$i]: " choice
  choice=${choice:-$i}

  if [[ "$choice" -eq "$i" ]]; then
    create_branch "$ticket"
  elif [[ "$choice" -ge 1 && "$choice" -le $count ]]; then
    echo "${branches[$((choice-1))]}"
  else
    echo "Invalid choice"
    exit 1
  fi
}

create_branch() {
  local ticket=$1
  read -p "Is this a [f]eature, [b]ugfix, or [h]otfix? [f/b/h] " branch_type
  read -p "Enter a description (optional): " description

  local prefix
  case "$branch_type" in
    f|F) prefix="feature" ;;
    b|B) prefix="bugfix" ;;
    h|H) prefix="hotfix" ;;
    *) prefix="feature" ;;
  esac

  local branch_name="$prefix/$ticket"
  if [[ -n "$description" ]]; then
    branch_name="$branch_name-$(echo "$description" | sed 's/ /-/g')"
  fi

  git branch "$branch_name"
  echo "$branch_name"
}

main() {
  local ticket base_dir worktree_dir branch

  ticket=$(get_ticket "${1:-}")
  if [[ -z "$ticket" ]]; then
    echo "Error: Ticket number is required"
    exit 1
  fi

  base_dir=$(get_base_dir)
  worktree_dir="$base_dir/$ticket"

  if [[ -d "$worktree_dir" ]]; then
    echo "Worktree already exists at $worktree_dir"
    exit 0
  fi

  echo "Looking for branches matching '$ticket'..."
  local branch_list
  branch_list=$(find_branch "$ticket")
  IFS=$'\n' read -d '' -r -a branches <<< "$branch_list" || true

  if [[ ${#branches[@]} -eq 0 || -z "${branches[0]:-}" ]]; then
    echo "No local branch found for '$ticket', searching remote..."
    local remote_list
    remote_list=$(find_remote_branch "$ticket")
    IFS=$'\n' read -d '' -r -a remote_branches <<< "$remote_list" || true

    if [[ ${#remote_branches[@]} -gt 0 && -n "${remote_branches[0]:-}" ]]; then
      echo "Found ${#remote_branches[@]} remote branch(es): ${remote_branches[*]}"
      if [[ ${#remote_branches[@]} -eq 1 ]]; then
        echo "Checking out remote branch: ${remote_branches[0]}"
        branch="${remote_branches[0]}"
      else
        echo "Note: Multiple remote branches found. Please choose or create new."
        read -p "Create a new branch? [y/N] " create
        if [[ "$create" =~ ^[yY]$ ]]; then
          branch=$(create_branch "$ticket")
        else
          exit 1
        fi
      fi
    else
      echo "No branch found for '$ticket'"
      read -p "Create a new branch? [y/N] " create
      if [[ "$create" =~ ^[yY]$ ]]; then
        branch=$(create_branch "$ticket")
      else
        exit 1
      fi
    fi
  elif [[ ${#branches[@]} -eq 1 ]]; then
    echo "Using existing branch: ${branches[0]}"
    branch="${branches[0]}"
  else
    echo "Found ${#branches[@]} branch(es): ${branches[*]}"
    branch=$(choose_branch "$ticket" "${branches[@]}")
  fi

  if [[ -z "$branch" ]]; then
    echo "Failed to determine branch"
    exit 1
  fi

  echo "Creating worktree at $worktree_dir for branch '$branch'"
  
  if git show-ref --verify --quiet "refs/heads/$branch" 2>/dev/null; then
    git worktree add "$worktree_dir" "$branch"
  else
    echo "Branch '$branch' does not exist locally, creating from remote..."
    git worktree add -b "$branch" "$worktree_dir" "origin/$branch"
  fi

  if [[ ! -f "$SECRETS_SOURCE" ]]; then
    echo "Error: Secrets file not found at $SECRETS_SOURCE"
    exit 1
  fi

  local secrets_dest="$worktree_dir/apps/prism/.env.secrets"
  mkdir -p "$(dirname "$secrets_dest")"
  cp "$SECRETS_SOURCE" "$secrets_dest"
  echo "Copied secrets to $secrets_dest"

  local env_src="$base_dir/apps/prism/.env"
  local env_dest="$worktree_dir/apps/prism/.env"
  if [[ -f "$env_src" ]]; then
    cp "$env_src" "$env_dest"
    echo "Copied .env to $env_dest"
  else
    echo "Warning: .env not found at $env_src"
  fi

  cat > "$worktree_dir/mise.toml" << EOF
[env]
TICKET = "$ticket"
EOF
  echo "Created mise.toml with TICKET=$ticket"

  cd "$worktree_dir"
  mise trust
  echo "Ran mise trust"

  echo "Creating tmux session '$ticket' and running pnpm install..."
  tmux new-session -d -s "$ticket" -c "$worktree_dir"
  tmux send-keys -t "$ticket" "pnpm install" C-m

  cd "$worktree_dir"
  tms bookmark

  echo "Done! Tmux session '$ticket' created at $worktree_dir"
}

main "$@"
