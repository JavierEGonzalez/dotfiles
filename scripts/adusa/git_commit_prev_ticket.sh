#/usr/local/bin/bash
source ~/.scratch/scripts/session_ticket_functions.sh

fname=$(git log -1 --pretty=%B | head -n 1 | sed -e 's/:\s.*//' -e 's/\[//' -e 's/\]//')

prompt_to_stage_if_needed

echo "Committing to $fname"
echo "Message:"
read -p ">" msg

commit_ticket $fname "$msg"