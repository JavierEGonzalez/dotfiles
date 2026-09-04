#!/usr/bin/env bash
set -euo pipefail

VERSION="0.1.0"
SCRIPT_NAME="worktree-checkout.sh"

DEFAULT_BASE="~/workspace/prism3"

NON_INTERACTIVE=false
if [[ ! -t 0 ]]; then
  NON_INTERACTIVE=true
fi

COMMIT_REF=""
BRANCH_NAME_OVERRIDE=""
BASE_DIR_OVERRIDE=""

usage() {
  cat << EOF
$SCRIPT_NAME v$VERSION

Usage: worktree [options] <ticket> [name]

Arguments:
  ticket                  Ticket number (e.g. 11580) or ticket ID (e.g. CXPVSP-11580).
  name                    Optional worktree directory name. Defaults to the ticket.

Options:
  -c, --commit <ref>      Base the new worktree on this specific commit/ref instead of
                          resolving an existing branch for the ticket. Creates a new
                          branch (see --branch) pointing at <ref>.
  -b, --branch <name>     Explicit branch name to create. Defaults to
                          "review/<ticket>" when used with --commit, otherwise
                          follows the normal feature/bugfix/hotfix prompt flow.
  -d, --dir <path>        Worktree base directory. Defaults to $DEFAULT_BASE.
  -y, --yes, --non-interactive
                          Run without prompts, using defaults for anything not
                          supplied via flags. Auto-enabled when stdin is not a
                          TTY (e.g. when invoked by an agent/script).
  -v, --version           Print version ($VERSION) and exit.
  -h, --help              Show this help and exit.

Examples:
  worktree 11580                          # interactive, resolves existing branch
  worktree -c 8a36c1aba CXPVSP-11580-ssr  # worktree pinned to a specific commit
  worktree -y -c HEAD~3 11580 review-x    # fully headless, base off HEAD~3
EOF
}

positional=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    -v|--version)
      echo "$SCRIPT_NAME v$VERSION"
      exit 0
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    -c|--commit)
      COMMIT_REF="${2:?--commit requires a git ref/commit argument}"
      shift 2
      ;;
    -b|--branch)
      BRANCH_NAME_OVERRIDE="${2:?--branch requires a branch name argument}"
      shift 2
      ;;
    -d|--dir)
      BASE_DIR_OVERRIDE="${2:?--dir requires a path argument}"
      shift 2
      ;;
    -y|--yes|--non-interactive)
      NON_INTERACTIVE=true
      shift
      ;;
    --)
      shift
      positional+=("$@")
      break
      ;;
    -*)
      echo "Unknown option: $1" >&2
      usage >&2
      exit 1
      ;;
    *)
      positional+=("$1")
      shift
      ;;
  esac
done
set -- "${positional[@]:-}"
[[ ${#positional[@]} -eq 0 ]] && set --

get_ticket() {
  local input
  if [[ -n "${1:-}" ]]; then
    input="$1"
  elif [[ "$NON_INTERACTIVE" == true ]]; then
    echo "Error: ticket number is required in non-interactive mode" >&2
    exit 1
  else
    read -p "Enter ticket number: " input
  fi

  if [[ "$input" =~ ^[0-9]+$ ]]; then
    echo "CXPVSP-${input}"
  else
    echo "$input"
  fi
}

get_base_dir() {
  if [[ -n "$BASE_DIR_OVERRIDE" ]]; then
    eval echo "$BASE_DIR_OVERRIDE"
    return
  fi

  local expanded_default
  expanded_default=$(eval echo "$DEFAULT_BASE")

  if [[ "$NON_INTERACTIVE" == true ]]; then
    echo "$expanded_default"
    return
  fi

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

  if [[ "$NON_INTERACTIVE" == true ]]; then
    echo "Error: multiple branches found for '$ticket' and running non-interactively." >&2
    printf '  %s\n' "${branches[@]}" >&2
    echo "Pass -b/--branch to disambiguate, or -c/--commit to base on a specific ref." >&2
    exit 1
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
  local branch_type description

  if [[ "$NON_INTERACTIVE" == true ]]; then
    branch_type="f"
    description=""
  else
    read -p "Is this a [f]eature, [b]ugfix, or [h]otfix? [f/b/h] " branch_type
    read -p "Enter a description (optional): " description
  fi

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

  local current_branch base_choice
  current_branch=$(git rev-parse --abbrev-ref HEAD)

  if [[ "$NON_INTERACTIVE" == true ]]; then
    base_choice="m"
  else
    read -p "Branch off [m]ain or [c]urrent branch ($current_branch)? [m/c] " base_choice
  fi

  local base_ref
  case "$base_choice" in
    c|C)
      base_ref="$current_branch"
      echo "Using current branch: $current_branch" >&2
      ;;
    *)
      echo "Fetching origin/main..." >&2
      git fetch origin main >&2
      base_ref="origin/main"
      ;;
  esac

  git branch "$branch_name" "$base_ref" >&2
  echo "$branch_name"
}

# Creates a new branch pointed at an arbitrary commit/ref (e.g. a specific
# historical SHA), bypassing ticket-based branch resolution entirely.
create_branch_from_commit() {
  local ticket=$1
  local commit_ref=$2
  local branch_name="${BRANCH_NAME_OVERRIDE:-review/$ticket}"

  if ! git rev-parse --verify --quiet "$commit_ref^{commit}" > /dev/null; then
    echo "Error: '$commit_ref' is not a valid commit/ref" >&2
    exit 1
  fi

  if git show-ref --verify --quiet "refs/heads/$branch_name"; then
    echo "Error: branch '$branch_name' already exists. Pass -b/--branch to choose another name." >&2
    exit 1
  fi

  git branch "$branch_name" "$commit_ref" >&2
  echo "$branch_name"
}

main() {
  local ticket name base_dir worktree_dir branch

  ticket=$(get_ticket "${1:-}")
  if [[ -z "$ticket" ]]; then
    echo "Error: Ticket number is required"
    exit 1
  fi

  name="${2:-$ticket}"

  base_dir=$(get_base_dir)
  worktree_dir="$base_dir/$name"

  if [[ -d "$worktree_dir" ]]; then
    echo "Worktree already exists at $worktree_dir"
    exit 0
  fi

  if [[ -n "$COMMIT_REF" ]]; then
    echo "Basing worktree on commit/ref '$COMMIT_REF' (bypassing branch resolution)..."
    branch=$(create_branch_from_commit "$ticket" "$COMMIT_REF")
  else
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
        elif [[ "$NON_INTERACTIVE" == true ]]; then
          echo "Error: multiple remote branches found and running non-interactively." >&2
          printf '  %s\n' "${remote_branches[@]}" >&2
          exit 1
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
        if [[ "$NON_INTERACTIVE" == true ]]; then
          echo "Creating new branch (non-interactive default: feature/$ticket off origin/main)..."
          branch=$(create_branch "$ticket")
        else
          read -p "Create a new branch? [y/N] " create
          if [[ "$create" =~ ^[yY]$ ]]; then
            branch=$(create_branch "$ticket")
          else
            exit 1
          fi
        fi
      fi
    elif [[ ${#branches[@]} -eq 1 ]]; then
      echo "Using existing branch: ${branches[0]}"
      branch="${branches[0]}"
    else
      echo "Found ${#branches[@]} branch(es): ${branches[*]}"
      branch=$(choose_branch "$ticket" "${branches[@]}")
    fi
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

  local turbo_src="$base_dir/nuclei/.turbo"
  local turbo_dest="$worktree_dir/.turbo"
  if [[ -d "$turbo_src" ]]; then
    cp -r "$turbo_src" "$turbo_dest"
    echo "Copied .turbo to $turbo_dest"
  else
    echo "Warning: .turbo not found at $turbo_src"
  fi

  cat > "$worktree_dir/mise.toml" << EOF
[env]
TICKET = "$ticket"
EOF
  echo "Created mise.toml with TICKET=$ticket"

  cd "$worktree_dir"
  mise trust
  echo "Ran mise trust"

  echo "\n\nCreating tmux session '$name' and running pnpm install... \n\n"
  tmux new-session -d -s "$name" -c "$worktree_dir"

  cd "$worktree_dir"
  tms bookmark

  tmux send-keys -t "$name" "mise run setup" C-m

  echo "Done! Tmux session '$name' created at $worktree_dir running setup script in mise"
}

main "$@"
