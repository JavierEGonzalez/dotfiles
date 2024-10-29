#!/bin/zsh

target_dir=$PWD
verbose='false'
prefix="[INFO]"
get_unique_files='false'
query=''

while getopts 'q:fd:v' flag; do
  case $flag in 
    q) query="${OPTARG}";;
    f) get_unique_files='true';;
    d) target_dir="${OPTARG}";;
    v) verbose='true';;
  esac
done
if [[ -z $query ]]; then
  echo "No query provided" 1>&2
  echo "Pass query with -q flag" 1>&2
  exit 1
fi

if [[ $verbose == 'true' ]]; then
  echo "$prefix VERBOSE MODE" 1>&2
  echo "$prefix unique files: $get_unique_files" 1>&2
  echo "$prefix target_dir: $target_dir" 1>&2
  echo 1>&2
fi

command="grep -Rn --color=auto --exclude-dir={node_modules,bin,dist,css,qa-build,build,nginx-env,shop/frags,coverage} '$query' $target_dir"


if [[ $verbose == 'true' ]]; then
  echo "$prefix searching for string '$query' in $target_dir by running" 1>&2
  echo "$prefix $command" 1>&2
  echo ''
fi

if [[ $get_unique_files == 'true' ]]; then
  $command | sed -e 's/:.*//' | uniq
else
  $command
fi
