#/usr/local/bin/bash
source ~/.scratch/scripts/session_ticket_functions.sh

fname=$(git log | head -n 5 | tail -n 1 | sed -e 's/^[[:space:]]*//; s/:.*//')

check_committed_files

echo "Committing to $fname"
echo "Message:"
read -p ">" msg

commit_ticket $fname $msg
