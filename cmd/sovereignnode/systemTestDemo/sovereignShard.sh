CURRENT_DIR=$(pwd)
WORKING_DIR="$CURRENT_DIR"
TESTNET_DIR=$WORKING_DIR/sovereign
TESTNET_OUTPUT_DIR=$TESTNET_DIR/sovereign-local
SCRIPTS_DIR=mx-chain-go/scripts/testnet
VARIABLES_PATH=$SCRIPTS_DIR/variables.sh
OBSERVERS_PATH=$SCRIPTS_DIR/include/observers.sh
VALIDATORS_PATH=$SCRIPTS_DIR/include/validators.sh
ENABLE_EPOCH_DIR=$TESTNET_DIR/mx-chain-go/cmd/node/config/enableEpochs.toml
CONFIG_DIR=$TESTNET_DIR/mx-chain-go/cmd/node/config/config.toml
SYSTEM_SC_CONFIG_DIR=$TESTNET_DIR/mx-chain-go/cmd/node/config/systemSmartContractsConfig.toml
SANDBOX_NAME=sandbox

cloneDependencies(){
  if [ -d "$TESTNET_DIR" ]; then
    rm -rf $TESTNET_DIR
  fi

  mkdir "$TESTNET_DIR"

  git clone https://github.com/multiversx/mx-chain-go "$TESTNET_DIR/mx-chain-go"
  cd $TESTNET_DIR/mx-chain-go
  git checkout f59d7b307351ea2d00e616ca6d02e7c16411deb7
  cd ../..

  git clone https://github.com/multiversx/mx-chain-deploy-go "$TESTNET_DIR/mx-chain-deploy-go"
  git clone https://github.com/multiversx/mx-chain-proxy-go "$TESTNET_DIR/mx-chain-proxy-go"
}

testnetRemove(){
  rm -rf "$TESTNET_OUTPUT_DIR"
}

testnetSetup(){
  mkdir "$TESTNET_OUTPUT_DIR"
  cd "$TESTNET_OUTPUT_DIR"
  ln -s "$TESTNET_DIR"/mx-chain-go mx-chain-go
  ln -s "$TESTNET_DIR"/mx-chain-deploy-go mx-chain-deploy-go
  ln -s "$TESTNET_DIR"/mx-chain-proxy-go mx-chain-proxy-go
}

testnetPrereq(){
  cd "$TESTNET_OUTPUT_DIR" && \
    bash $SCRIPTS_DIR/prerequisites.sh && \
    echo -e "export TESTNETDIR=$TESTNET_OUTPUT_DIR/$SANDBOX_NAME" > $SCRIPTS_DIR/local.sh && \
    bash $SCRIPTS_DIR/config.sh
}

testnetUpdateVariables(){
  sed -i 's/export SHARDCOUNT=.*/export SHARDCOUNT=1/' $VARIABLES_PATH
  sed -i 's/SHARD_VALIDATORCOUNT=.*/SHARD_VALIDATORCOUNT=2/' $VARIABLES_PATH
  sed -i 's/SHARD_OBSERVERCOUNT=.*/SHARD_OBSERVERCOUNT=1/' $VARIABLES_PATH
  sed -i 's/SHARD_CONSENSUS_SIZE=.*/SHARD_CONSENSUS_SIZE=$SHARD_VALIDATORCOUNT/' $VARIABLES_PATH
  sed -i 's/export NODE_DELAY=.*/export NODE_DELAY=30/' $VARIABLES_PATH

  sed -i 's/export LOGLEVEL=.*/export LOGLEVEL="\*\:DEBUG"/' $VARIABLES_PATH
}

testnetNew(){
  cloneDependencies
  testnetRemove
  testnetSetup
  testnetUpdateVariables
  testnetPrereq
}

testnetStart(){
  cd "$TESTNET_DIR" && \
    ./mx-chain-go/scripts/testnet/sovereignStart.sh
}

testnetReset(){
  cd "$TESTNET_DIR" && \
    ./mx-chain-go/scripts/testnet/reset.sh
}

testnetStop(){
  cd "$TESTNET_DIR" && \
    ./mx-chain-go/scripts/testnet/stop.sh
}

echoOptions(){
  echo "ERROR!!! Please choose one of the following parameters:
  - new to create a new sovereign shard
  - start to start the sovereign shard
  - reset to reset the sovereign shard
  - stop to stop the sovereign shard"
}

main(){
  if [ $# -eq 1 ]; then
    case "$1" in
      new)
        testnetNew ;;
      start)
        testnetStart ;;
      reset)
        testnetReset ;;
      stop)
        testnetStop ;;
      *)
        echoOptions ;;
      esac
  else
    echoOptions
  fi
}

main "$@"