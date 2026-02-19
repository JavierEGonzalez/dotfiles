#!/bin/bash
# Ralph Wiggum - Long-running AI agent loop
set -e

# Parse arguments
tool="opencode"
max_iterations=10
model="opencode/minimax-m2.5-free"

while [[ $# -gt 0 ]]; do
  case $1 in
    -m|--model)
      model="$2"
      shift 2
      ;;
    --model=*)
      model="${1#*=}"
      shift
      ;;
    -h|--help)
      echo "Usage: ralph.sh [-m|--model <model>] [max_iterations]"
      echo ""
      echo "Options:"
      echo "  -m, --model   Model to pass to opencode (default: $model)"
      echo ""
      echo "Args:"
      echo "  max_iterations  Number of iterations to run (default: $max_iterations)"
      exit 0
      ;;
    *)
      # Assume it's max_iterations if it's a number
      if [[ "$1" =~ ^[0-9]+$ ]]; then
        max_iterations="$1"
      fi
      shift
      ;;
  esac
done

# Validate tool choice
if [[ "$tool" != "copilot" && "$tool" != "opencode" ]]; then
  echo "Error: Invalid tool '$tool'. Must be 'copilot' or 'opencode'."
  exit 1
fi
# Determine project root
# First check if we're already in a directory with .scratch/ralph/prd.json
if [ -f "./.scratch/ralph/prd.json" ]; then
  project_root="$(pwd)"
  echo $project_root
else
  echo "No prd to run a ralph loop on"
  exit 1
fi
script_dir="$project_root/.scratch/ralph"
prd_file="$script_dir/prd.json"
progress_file="$script_dir/progress.txt"
archive_dir="$script_dir/archive"
last_branch_file="$script_dir/.last-branch"

# Track current branch
if [ -f "$prd_file" ]; then
  current_branch=$(jq -r '.branchName // empty' "$prd_file" 2>/dev/null || echo "")
  if [ -n "$current_branch" ]; then
    echo "$current_branch" > "$last_branch_file"
  fi
fi

# Initialize progress file if it doesn't exist
if [ ! -f "$progress_file" ]; then
  echo "# Ralph Progress Log" > "$progress_file"
  echo "Started: $(date)" >> "$progress_file"
  echo "---" >> "$progress_file"
fi

echo "Starting Ralph - Tool: $tool - Model: $model - Max iterations: $max_iterations"
echo "Project root: $project_root"
echo "PRD location: $prd_file"

# Change to project root so opencode runs in the right directory
cd "$project_root"

for i in $(seq 1 $max_iterations); do
  echo ""
  echo "==============================================================="
  echo "  Ralph Iteration $i of $max_iterations ($tool)"
  echo "==============================================================="

  OUTPUT=$(opencode run 'use ralph-implementer skill to work on a SINGLE task of the current prd. Do not do more than one task.' -m "$model" 2>&1 | tee /dev/stderr) || true
  
  # Check for completion signal
  if echo "$OUTPUT" | grep -q "<promise>COMPLETE</promise>"; then
    echo ""
    echo "Ralph completed all tasks!"
    echo "Completed at iteration $i of $max_iterations"
    exit 0
  fi
  
  echo "Iteration $i complete. Continuing..."
  sleep 2
done

echo ""
echo "Ralph reached max iterations ($max_iterations) without completing all tasks."
echo "Check $progress_file for status."
exit 1
