#!/usr/bin/env bash
# workers_ai.sh - Cloudflare Workers AI CLI helper with multi-turn chat
# Supports one-shot and interactive sessions, persistent history, bash assistant execution
# Backward compatible: positional 1=user message, positional 2=custom prompt file

set -euo pipefail

DEFAULT_MODEL="@cf/meta/llama-4-scout-17b-16e-instruct"
DEFAULT_MAX_TOKENS=1024
DEFAULT_HISTORY_LIMIT=30

ACCOUNT_FILE="${WORKERS_AI_ACCOUNT_FILE:-$HOME/.scratch/workers_ai_account}"
TOKEN_FILE="${WORKERS_AI_TOKEN_FILE:-$HOME/.scratch/workers_ai_api.key}"
PROMPT_DIR="$HOME/.scratch/scripts"
SESSIONS_DIR="$HOME/.scratch/workers_ai_sessions"

model="$DEFAULT_MODEL"
max_tokens="$DEFAULT_MAX_TOKENS"
history_limit="$DEFAULT_HISTORY_LIMIT"
session_name="default"
interactive=false
auto_exec=false
dry_run=false
system_prompt=""
system_file=""
user_message=""
bash_assistant=false

# For collecting any extra positional args (ignored for now)
declare -a user_extra_args

# --- Helpers -----------------------------------------------------------------

die(){ echo "ERROR: $*" >&2; exit 1; }
have(){ command -v "$1" >/dev/null 2>&1; }

show_help() {
  cat <<EOF
Usage: $0 [options] [message] [legacy_prompt_file]
Options:
  -i, --interactive          Interactive multi-turn chat loop
  -s, --session NAME         Session name (default: default)
  -m, --model MODEL          Model id (default: $DEFAULT_MODEL)
      --system-file PATH     System prompt file
      --system "TEXT"        Inline system prompt text
      --max-tokens N         Max tokens (default: $DEFAULT_MAX_TOKENS)
      --history-limit N      Keep last N messages (default: $DEFAULT_HISTORY_LIMIT)
      --auto-exec            Auto execute assistant replies (bash assistant mode)
      --dry-run              Build payload; skip API call
  -h, --help                 Show help

Interactive Commands:
  :q        Quit session
  :reset    Clear history (retain system prompt)
  :save NAME Save current session under NAME
  :exec     Toggle command execution (bash assistant mode only)
  :show     Show last 5 messages

Legacy Usage Compatibility:
  $0 "message" /path/to/custom_prompt_file
EOF
}

# --- Arg Parsing -------------------------------------------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    -i|--interactive) interactive=true ;;
    -s|--session) session_name="${2:-}"; [[ -z "$session_name" ]] && die "Missing session name"; shift ;;
    -m|--model) model="${2:-}"; shift ;;
    --system-file) system_file="${2:-}"; shift ;;
    --system) system_prompt="${2:-}"; shift ;;
    --max-tokens) max_tokens="${2:-}"; shift ;;
    --history-limit) history_limit="${2:-}"; shift ;;
    --auto-exec) auto_exec=true ;;
    --dry-run) dry_run=true ;;
    -h|--help) show_help; exit 0 ;;
    --) shift; break ;;
    -*) die "Unknown flag $1" ;;
    *)
       if [[ -z "$user_message" ]]; then
         user_message="$1"
       elif [[ -z "$system_file" && -z "$system_prompt" && -f "$1" ]]; then
         # Legacy second positional argument as prompt file
         system_file="$1"
       else
         user_extra_args+=("$1")
       fi
       ;;
  esac
  shift || true
done

# --- Dependency & Credential Checks -----------------------------------------
have jq || die "jq required"
have curl || die "curl required"
bat_cmd="cat"
if have bat; then bat_cmd="bat --paging=never --plain"; fi

[[ -r "$ACCOUNT_FILE" ]] || die "Account id file missing: $ACCOUNT_FILE"
[[ -r "$TOKEN_FILE" ]] || die "API token file missing: $TOKEN_FILE"
ACCOUNT_ID="$(<"$ACCOUNT_FILE")"
API_TOKEN="$(<"$TOKEN_FILE")"

if [[ -n "$system_file" ]]; then
  [[ -r "$system_file" ]] || die "System file unreadable: $system_file"
  system_prompt="$(<"$system_file")"
fi

# --- System Prompt Selection -------------------------------------------------
if [[ -z "$system_prompt" ]]; then
  default_bash_prompt="$PROMPT_DIR/workers_ai_bash_assistant_prompt"
  general_prompt="$PROMPT_DIR/workers_ai_general_question"
  spanish_prompt="$PROMPT_DIR/workers_ai_pregunta"
  if [[ -r "$default_bash_prompt" && "$interactive" == "true" ]]; then
    system_prompt="$(<"$default_bash_prompt")"
    bash_assistant=true
  elif [[ -r "$default_bash_prompt" && "$user_message" == code* ]]; then
    system_prompt="$(<"$default_bash_prompt")"
    bash_assistant=true
  elif [[ -r "$general_prompt" ]]; then
    system_prompt="$(<"$general_prompt")"
  elif [[ -r "$spanish_prompt" ]]; then
    system_prompt="$(<"$spanish_prompt")"
  else
    die "No available system prompt files in $PROMPT_DIR"
  fi
fi

# Adjust tokens if bash assistant (shorter default) or custom prompt (longer)
if [[ "$bash_assistant" == "true" ]]; then
  max_tokens=512
else
  max_tokens=${max_tokens:-2048}
fi

# --- Session Management ------------------------------------------------------
if [[ ! -d "$SESSIONS_DIR" ]]; then
  mkdir -p "$SESSIONS_DIR" || die "Cannot create sessions directory $SESSIONS_DIR"
fi
session_file="$SESSIONS_DIR/session_${session_name}.jsonl"
# Create if absent
if [[ ! -f "$session_file" ]]; then
  : > "$session_file" || die "Cannot create session file $session_file"
fi

# Seed system message if not present
if ! grep -q '"role":"system"' "$session_file"; then
  printf '%s\n' "$(jq -n --arg c "$system_prompt" '{role:"system",content:$c}')" >> "$session_file"
fi

append_message() {
  local role="$1"; shift
  local content="$*"
  printf '%s\n' "$(jq -n --arg r "$role" --arg c "$content" '{role:$r,content:$c}')" >> "$session_file"
}

trim_history() {
  local total
  total="$(wc -l <"$session_file")"
  if (( total > history_limit )); then
    # Keep system + last (history_limit -1) messages
    grep '"role":"system"' "$session_file" | head -1 > /tmp/session_head.$$
    tail -n $((history_limit - 1)) "$session_file" > /tmp/session_tail.$$
    cat /tmp/session_head.$$ /tmp/session_tail.$$ > /tmp/session_new.$$
    mv /tmp/session_new.$$ "$session_file"
    rm -f /tmp/session_head.$$ /tmp/session_tail.$$
  fi
}

build_payload() {
  jq -s --argjson mt "$max_tokens" '{messages: ., max_tokens: $mt}' "$session_file"
}

call_api() {
  local payload
  payload="$(build_payload)"
  if [[ "$dry_run" == "true" ]]; then
    echo "=== Dry Run Payload ==="
    echo "$payload"
    return 0
  fi
  echo "$payload" > /tmp/workers_ai_last_request.json
  local url="https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/ai/run/$model"
  local http_code
  http_code=$(curl -s -o /tmp/workers_ai_raw.json -w '%{http_code}' \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    "$url" -d "$payload")
  if [[ "$http_code" != "200" ]]; then
    die "API HTTP $http_code (see /tmp/workers_ai_raw.json)"
  fi
  local assistant_raw
  assistant_raw="$(jq -r '.result.response // empty' /tmp/workers_ai_raw.json)"
  [[ -z "$assistant_raw" ]] && die "Empty assistant response"
  append_message assistant "$assistant_raw"
  if [[ "$bash_assistant" == "true" ]]; then
    echo "=== Assistant (code) ==="
    echo "$assistant_raw" | $bat_cmd -l sh
    if [[ "$auto_exec" == "true" ]]; then
      echo "Executing..."
      eval "$assistant_raw"
    else
      echo "(execution disabled; use :exec to toggle)"
    fi
  else
    echo "=== Assistant ==="
    echo "$assistant_raw"
  fi
}

do_turn() {
  local msg="$1"
  [[ -z "$msg" ]] && return 0
  append_message user "$msg"
  trim_history
  call_api
}

# --- Interactive Loop --------------------------------------------------------
interactive_loop() {
  echo "Interactive session: $session_name (model: $model)"
  echo "Commands: :q :reset :save <name> :exec :show"
  [[ -n "$user_message" ]] && do_turn "$user_message"
  while true; do
    printf 'You> '
    IFS= read -r line || break
    case "$line" in
      :q) echo "Bye"; break ;;
      :reset)
        echo "Resetting session"
        : > "$session_file"
        printf '%s\n' "$(jq -n --arg c "$system_prompt" '{role:"system",content:$c}')" >> "$session_file"
        continue ;;
      :exec)
        if [[ "$bash_assistant" == "true" ]]; then
          if [[ "$auto_exec" == "true" ]]; then auto_exec=false; else auto_exec=true; fi
          echo "auto_exec=$auto_exec"
        else
          echo "Not in bash assistant mode"
        fi
        continue ;;
      :show)
        echo "Messages: $(wc -l <"$session_file")"
        tail -n 5 "$session_file"
        continue ;;
      :save\ *)
        local new_name="${line#*:save }"
        [[ -z "$new_name" ]] && { echo "Name required"; continue; }
        local new_file="$SESSIONS_DIR/session_${new_name}.jsonl"
        cp "$session_file" "$new_file" && session_name="$new_name" && session_file="$new_file" && echo "Saved as $session_name"
        continue ;;
      "") continue ;;
    esac
    do_turn "$line"
  done
}

# --- Main Execution ----------------------------------------------------------
if [[ "$interactive" == "true" ]]; then
  interactive_loop
else
  [[ -z "$user_message" ]] && die "No message provided (use -i for interactive or supply a message)"
  do_turn "$user_message"
fi

# End of script
