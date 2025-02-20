#!/usr/bin/env bash

# Delete the entire testnet folder, which includes configuration, executables and logs.

export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"

if [ $USE_ELASTICSEARCH -eq 1 ]; then
  echo "Removing Elasticsearch Docker container and volume..."
  ES_CONTAINER_ID=$(cat $TESTNETDIR/es_container_id.txt)
  sudo docker remove $ES_CONTAINER_ID

  sudo docker volume rm $ELASTICSEARCH_VOLUME
fi

echo "Removing $TESTNETDIR..."
rm -rf $TESTNETDIR
