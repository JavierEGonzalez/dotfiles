#!/usr/bin/env bash
# ai_bash.sh - Natural language to runnable shell command assistant (EPHEMERAL SESSIONS)
# User Stories Implemented:
# 1. One-line natural language -> shell command (ready to run on Enter)
# 1b. Options if output not correct: conversation refine, manual edit, docs mode, exit
#     Docs mode preference order: man page -> --help/-h -> model explanation
# 1c. Redo command (:redo / redo) re-runs last executed command
#
# Ephemeral design: NO persistence across separate invocations. History lives only
# for process lifetime. Exiting menu clears everything.
#
# NOTE: Requires Cloudflare Workers AI credentials (ACCOUNT_FILE / TOKEN_FILE)
#
set -euo pipefail

DEFAULT_MODEL="@cf/meta/llama-4-scout-17b-16e-instruct"
DEFAULT_MAX_TOKENS=512
DEFAULT_HISTORY_LIMIT=20

ACCOUNT_FILE="${WORKERS_AI_ACCOUNT_FILE:-$HOME/.scratch/workers_ai_account}"
TOKEN_FILE="${WORKERS_AI_TOKEN_FILE:-$HOME/.scratch/workers_ai_api.key}"
SESSIONS_DIR="${AI_BASH_SESSIONS_DIR:-$HOME/.scratch/ai_bash_sessions}"
PROMPT_FILE="${AI_BASH_PROMPT_FILE:-}" # Optional external system prompt override

model="$DEFAULT_MODEL"
max_tokens="$DEFAULT_MAX_TOKENS"
history_limit="$DEFAULT_HISTORY_LIMIT"
session_name="main"
interactive=false
auto_run=false
system_prompt=""
user_message=""
dry_run=false

current_command=""
last_executed_command=""

# --- Helpers -----------------------------------------------------------------
die(){ echo "ERROR: $*" >&2; exit 1; }
have(){ command -v "$1" >/dev/null 2>&1; }

bat_cmd="cat"
if have bat; then bat_cmd="bat --style grid -l bash"; fi

show_help(){ cat <<EOF
Usage: $0 [options] "natural language request"
Options:
  -i, --interactive          Start directly in conversation refine mode
  -s, --session NAME         Session name (default: main)
  -m, --model MODEL          Model id (default: $DEFAULT_MODEL)
      --system-file PATH     Use external system prompt file
      --system "TEXT"        Inline system prompt
      --max-tokens N         Max tokens (default: $DEFAULT_MAX_TOKENS)
      --history-limit N      History size (default: $DEFAULT_HISTORY_LIMIT)
      --auto-run             Auto-run first generated command (still shows menu)
      --dry-run              Show payloads; skip API calls
  -h, --help                 Show help

Menu (single-shot; exits after run):
  [Enter] Run command & exit
  r        Run command & exit
  c        Conversation refine mode (stay until :run/:redo/:q)
  e        Manual edit (returns to menu)
  d        Docs for command (returns to menu)
  redo     Re-run last executed command & exit
  x        Exit without running

Conversation Commands:
  :q       Exit conversation to menu
  :run     Run current command
  :edit    Manual edit
  :doc     Docs for current command
  :redo    Re-run last executed command
  :help    Show conversation help
  :reset   Reset conversation history (keep system prompt)
EOF
}

# --- Arg Parsing --------------------------------------------------------------
while [[ $# -gt 0 ]]; do
  case "$1" in
    -i|--interactive) interactive=true ;;
    -s|--session) session_name="${2:-}"; shift ;;
    -m|--model) model="${2:-}"; shift ;;
    --system-file) PROMPT_FILE="${2:-}"; shift ;;
    --system) system_prompt="${2:-}"; shift ;;
    --max-tokens) max_tokens="${2:-}"; shift ;;
    --history-limit) history_limit="${2:-}"; shift ;;
    --auto-run) auto_run=true ;;
    --dry-run) dry_run=true ;;
    -h|--help) show_help; exit 0 ;;
    --) shift; break ;;
    -*) die "Unknown flag $1" ;;
    *)
      if [[ -z "$user_message" ]]; then
        user_message="$1"
      else
        user_message+=" $1"
      fi
      ;;
  esac
  shift || true
done

# --- Credential & Setup -------------------------------------------------------
have jq || die "jq required"
have curl || die "curl required"
[[ -r "$ACCOUNT_FILE" ]] || die "Account id file missing: $ACCOUNT_FILE"
[[ -r "$TOKEN_FILE" ]] || die "API token file missing: $TOKEN_FILE"
ACCOUNT_ID="$(<"$ACCOUNT_FILE")"
API_TOKEN="$(<"$TOKEN_FILE")"

if [[ -n "$PROMPT_FILE" ]]; then
  [[ -r "$PROMPT_FILE" ]] || die "System prompt file unreadable: $PROMPT_FILE"
  system_prompt="$(<"$PROMPT_FILE")"
fi

if [[ -z "$system_prompt" ]]; then
  # Default prompt focused on command generation only
  system_prompt="You translate user natural language intents into SAFE, concise, POSIX-compliant shell commands.\nRules:\n- Output ONLY the runnable command(s); no explanation.\n- Prefer a single line if feasible.\n- NEVER assume destructive operations unless explicitly stated (rm, mv overwrite, chmod 777).\n- If intent is ambiguous, output a single line starting with '# clarify:' followed by a short clarification question; no command.\n- Use portable POSIX syntax; avoid aliases.\n- If user asks for complex multi-step, chain with '&&' or provide multiple lines.\n- Avoid unneeded subshells and UUOC.\n- For file safety, default to listing or dry-run style if user asks vaguely for deletion."
fi

# --- Session (Ephemeral) ------------------------------------------------------
session_file="/tmp/ai_bash_session.$$"
: > "$session_file" || die "Cannot create session file"
printf '%s\n' "$(jq -n --arg c "$system_prompt" '{role:"system",content:$c}')" >> "$session_file"
trap 'rm -f "$session_file"' EXIT

append_message(){
  local role="$1"; shift; local content="$*"
  printf '%s\n' "$(jq -n --arg r "$role" --arg c "$content" '{role:$r,content:$c}')" >> "$session_file"
}

trim_history(){
  local total; total="$(wc -l <"$session_file")"
  if (( total > history_limit )); then
    grep '"role":"system"' "$session_file" | head -1 > /tmp/ai_bash_head.$$
    tail -n $((history_limit - 1)) "$session_file" > /tmp/ai_bash_tail.$$
    cat /tmp/ai_bash_head.$$ /tmp/ai_bash_tail.$$ > /tmp/ai_bash_new.$$
    mv /tmp/ai_bash_new.$$ "$session_file"
    rm -f /tmp/ai_bash_head.$$ /tmp/ai_bash_tail.$$
  fi
}

build_payload(){ jq -s --argjson mt "$max_tokens" '{messages: ., max_tokens: $mt}' "$session_file"; }

raw_to_command(){
  local raw="$1" cmd
  # If clarification requested
  if [[ "$raw" =~ ^#\ clarify: ]]; then
    echo "$raw"; return 0
  fi
  if grep -q '```' <<<"$raw"; then
    cmd="$(echo "$raw" | awk 'BEGIN{in=0} /```/{if(in==0){in=1;next}else{in=0}} in==1 {print}')"
    # Strip leading bash/lang hint if present
    cmd="$(echo "$cmd" | sed -E '1s/^[[:space:]]*(bash|sh)[[:space:]]*//')"
    echo "$cmd"
    return 0
  fi
  # Otherwise assume raw is the command(s) already
  echo "$raw"
}

call_api(){
  local payload; payload="$(build_payload)"
  if [[ "$dry_run" == true ]]; then
    echo "=== Dry Run Payload ==="; echo "$payload"; return 0
  fi
  echo "$payload" > /tmp/ai_bash_last_request.json
  local url="https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/ai/run/$model"
  local http_code
  http_code=$(curl -s -o /tmp/ai_bash_raw.json -w '%{http_code}' \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    "$url" -d "$payload")
  if [[ "$http_code" != 200 ]]; then die "API HTTP $http_code (see /tmp/ai_bash_raw.json)"; fi
  local assistant_raw
  assistant_raw="$(jq -r '.result.response // empty' /tmp/ai_bash_raw.json)"
  [[ -z "$assistant_raw" ]] && die "Empty assistant response"
  append_message assistant "$assistant_raw"
  current_command="$(raw_to_command "$assistant_raw")"
}

generate_command(){
  local msg="$1"
  [[ -z "$msg" ]] && return 1
  append_message user "$msg"
  trim_history
  call_api
}

print_current(){
  if [[ -z "$current_command" ]]; then echo "(no command)"; return; fi
  echo "$current_command" | $bat_cmd
}

run_current(){
  if [[ -z "$current_command" ]]; then echo "No command to run"; return 1; fi
  if [[ "$current_command" =~ ^#\ clarify: ]]; then
    echo "Cannot run clarification line. Refine first."; return 1; fi
  echo "=== Running ==="
  echo "$current_command" | $bat_cmd -l sh
  if [[ "$dry_run" == true ]]; then echo "(dry-run: not executed)"; return 0; fi
  set +e
  eval "$current_command"
  local ec=$?
  set -e
  last_executed_command="$current_command"
  echo "=== Exit Code: $ec ==="
  return $ec
}

redo_last(){
  if [[ -z "$last_executed_command" ]]; then echo "No last command"; return 1; fi
  current_command="$last_executed_command"
  run_current
}

edit_command(){
  if [[ -z "$current_command" ]]; then echo "No command to edit"; return 1; fi
  local tmp; tmp="$(mktemp)"
  printf '%s\n' "$current_command" > "$tmp"
  if [[ -n "${EDITOR:-}" ]] && have "$EDITOR"; then
    "$EDITOR" "$tmp"
  else
    echo "Enter new command. End with single '.' line.";
    local lines=()
    while true; do
      printf 'edit> '
      IFS= read -r l || break
      [[ "$l" == "." ]] && break
      lines+=("$l")
    done
    printf '%s\n' "${lines[@]}" > "$tmp"
  fi
  local new; new="$(<"$tmp")"; rm -f "$tmp"
  if [[ -z "$new" ]]; then echo "(empty -> unchanged)"; return 0; fi
  current_command="$new"
  echo "Updated command:"; print_current
}

show_docs(){
  if [[ -z "$current_command" ]]; then echo "No command to document"; return 1; fi
  if [[ "$current_command" =~ ^#\ clarify: ]]; then echo "Clarification requested; refine first."; return 1; fi
  local primary
  primary="$(echo "$current_command" | sed -E 's/#.*$//' | sed -E 's/^[[:space:]]*sudo[[:space:]]+//' | sed -E 's/[|&;].*$//' | awk '{print $1}')"
  if [[ -z "$primary" ]]; then echo "Cannot determine primary command"; return 1; fi
  echo "=== Docs: $primary ==="
  if have man && man -w "$primary" >/dev/null 2>&1; then
    MANWIDTH=100 MANPAGER=cat man "$primary" | col -b | sed -n '1,120p'
    return 0
  fi
  if command -v "$primary" >/dev/null 2>&1; then
    if "$primary" --help >/dev/null 2>&1; then
      "$primary" --help 2>&1 | sed -n '1,90p'; return 0
    elif "$primary" -h >/dev/null 2>&1; then
      "$primary" -h 2>&1 | sed -n '1,90p'; return 0
    fi
  fi
  echo "(local docs unavailable; asking model)"
  local tmp="$(mktemp)"
  printf '%s\n' "$(jq -n --arg c "You are a concise shell docs assistant. Explain succinct usage for '$primary'. Limit to 15 lines." '{role:"system",content:$c}')" > "$tmp"
  printf '%s\n' "$(jq -n --arg c "Explain command '$primary' usage and common flags." '{role:"user",content:$c}')" >> "$tmp"
  local payload; payload="$(jq -s --argjson mt 512 '{messages: ., max_tokens: $mt}' "$tmp")"
  if [[ "$dry_run" == true ]]; then echo "(dry-run: would query model)"; cat "$tmp"; rm -f "$tmp"; return 0; fi
  local url="https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/ai/run/$model"
  local http_code
  http_code=$(curl -s -o /tmp/ai_bash_doc.json -w '%{http_code}' -H "Authorization: Bearer $API_TOKEN" -H "Content-Type: application/json" "$url" -d "$payload")
  if [[ "$http_code" == 200 ]]; then jq -r '.result.response' /tmp/ai_bash_doc.json | sed -n '1,120p'; else echo "Doc lookup failed (HTTP $http_code)"; fi
  rm -f "$tmp"
}

conversation_loop(){
  echo "--- Conversation Mode (refine command) ---"
  echo "Type instructions to refine. :help for commands."
  while true; do
    printf 'conv> '
    IFS= read -r line || break
    case "$line" in
      :q) echo "Leaving conversation"; break ;;
      :help) echo ":q :run :edit :doc :redo :reset"; continue ;;
      :run) run_current; exit $? ;;
      :edit) edit_command; continue ;;
      :doc) show_docs; continue ;;
      :redo) redo_last; exit $? ;;
      :reset)
        echo "Resetting conversation history"; : > "$session_file";
        printf '%s\n' "$(jq -n --arg c "$system_prompt" '{role:"system",content:$c}')" >> "$session_file"; continue ;;
      "") continue ;;
    esac
    generate_command "$line"; print_current
  done
}

menu_loop(){
  while true; do
    echo "Options: [Enter] run & exit | r run & exit | c refine | e edit | d docs | redo run & exit | x exit"
    printf 'Option> '
    IFS= read -r opt || break
    case "$opt" in
      ""|r) run_current; break ;;
      c) conversation_loop; break ;;
      e) edit_command ;;
      d) show_docs ;;
      redo|:redo) redo_last; break ;;
      x|:q) echo "Bye"; break ;;
      *) echo "Unknown option: $opt" ;;
    esac
  done
}

# --- Execution ----------------------------------------------------------------
if [[ "$interactive" == true && -z "$user_message" ]]; then
  echo "Interactive conversation (no initial request). Type natural language."
  conversation_loop
  exit 0
fi

if [[ -z "$user_message" ]]; then
  die "No natural language request provided (or use -i for conversation)"
fi

echo "Generating: '$user_message'\n"
generate_command "$user_message"
print_current
if [[ "$auto_run" == true ]]; then run_current; fi
menu_loop

