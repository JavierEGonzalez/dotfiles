#/usr/local/bin/bash
if [[ -z $1 ]] ; then
  yarn run dev
else
  echo "Running settings for $1"
  $("yarn run dev --dotenv ~/.scratch/.env.$1")
fi
