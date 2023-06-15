#!/usr/bin/env bash

CURRENT_DIR=$(pwd)
SANDBOX_PATH=$CURRENT_DIR/testnet/testnet-local/sandbox
KEY_GENERATOR_PATH=$CURRENT_DIR/testnet/mx-chain-go/cmd/keygenerator
EXTERNAL_CONFIG_DIR=$CURRENT_DIR/sovereignNotifier/config/external.toml

# Possible values: "server"/"client"
OBSERVER_MODE="server"

createObserverKey(){
  pushd $CURRENT_DIR

  cd testnet/mx-chain-go/cmd/keygenerator
  go build
  ./keygenerator

  popd
}

resetWorkingDir(){
  rm -rf sovereignNotifier
  mkdir "sovereignNotifier/"
  cd "sovereignNotifier/"
  mkdir "config/"
}

setupSovereignNotifier(){
  SOVEREIGN_OBSERVER_PATH=$(pwd)
  cp $SANDBOX_PATH/node/node $SOVEREIGN_OBSERVER_PATH
  cp -R $SANDBOX_PATH/node/config $SOVEREIGN_OBSERVER_PATH
  mv config/config_observer.toml config/config.toml
  mv $KEY_GENERATOR_PATH/validatorKey.pem config/

  sed -i 's/DestinationShardAsObserver =.*/DestinationShardAsObserver = "0"/' $SOVEREIGN_OBSERVER_PATH/config/prefs.toml

  sed -i '/HostDriverConfig\]/!b;n;n;c\    Enabled = true' "$EXTERNAL_CONFIG_DIR"
  sed -i "s@Mode =.*@Mode = \"$OBSERVER_MODE\"@" "$EXTERNAL_CONFIG_DIR"
  sed -i 's/MarshallerType =.*/MarshallerType = "json"/' "$EXTERNAL_CONFIG_DIR"
  sed -i 's/BlockingAckOnError =.*/BlockingAckOnError = false/' "$EXTERNAL_CONFIG_DIR"

  ./node --log-level *:DEBUG
}

createObserverKey
resetWorkingDir
setupSovereignNotifier
