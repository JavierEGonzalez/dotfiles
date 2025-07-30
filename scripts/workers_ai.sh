#!/bin/bash

USER_MESSAGE="$1"
CUSTOM_PROMPT_FILE="$2"

ACCOUNT_ID=$(<~/.scratch/workers_ai_account)
API_TOKEN=$(<~/.scratch/workers_ai_api.key)

USE_DEFAULT_BASH_ASSISTANT_PROMPT=true
if [[ -n "$CUSTOM_PROMPT_FILE" && -f "$CUSTOM_PROMPT_FILE" ]]; then
  SYSTEM_PROMPT=$(<"$CUSTOM_PROMPT_FILE")
  USE_DEFAULT_BASH_ASSISTANT_PROMPT=false
else
  SYSTEM_PROMPT=$(<~/.scratch/scripts/workers_ai_bash_assistant_prompt)
fi

curl -s \
  "https://api.cloudflare.com/client/v4/accounts/$ACCOUNT_ID/ai/run/@cf/meta/llama-3-8b-instruct" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$(jq -n --arg sys "$SYSTEM_PROMPT" --arg user "$USER_MESSAGE" \
    '{messages: [{role: "system", content: $sys}, {role: "user", content: $user}]}')" \
  | jq -r '.result.response' > /tmp/workers_ai_response.sh

bat /tmp/workers_ai_response.sh -l zsh -n

if [ "$USE_DEFAULT_BASH_ASSISTANT_PROMPT" = true ]; then
  echo "Do you want to execute the command? (y/n)"
  read -r CONFIRMATION
  if [[ "$CONFIRMATION" == "y" ]]; then
    COMMAND=$(< /tmp/workers_ai_response.sh)
    echo "Executing command: $COMMAND"
    eval "$COMMAND"
  else
    echo "Cancelled"
  fi
fi

