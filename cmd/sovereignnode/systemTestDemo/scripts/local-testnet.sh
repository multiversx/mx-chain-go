CURRENT_DIR=$(pwd)
WORKING_DIR="$CURRENT_DIR"
TESTNET_DIR=$WORKING_DIR/testnet
TESTNET_OUTPUT_DIR=$TESTNET_DIR/testnet-local
SCRIPTS_DIR=mx-chain-go/scripts/testnet
VARIABLES_PATH=$SCRIPTS_DIR/variables.sh
OBSERVERS_PATH=$SCRIPTS_DIR/include/observers.sh
VALIDATORS_PATH=$SCRIPTS_DIR/include/validators.sh
ENABLE_EPOCH_DIR=$TESTNET_DIR/mx-chain-go/cmd/node/config/enableEpochs.toml
EXTERNAL_CONFIG_DIR=$TESTNET_DIR/mx-chain-go/cmd/node/config/external.toml
CONFIG_DIR=$TESTNET_DIR/mx-chain-go/cmd/node/config/config.toml
SYSTEM_SC_CONFIG_DIR=$TESTNET_DIR/mx-chain-go/cmd/node/config/systemSmartContractsConfig.toml

# Possible values: "server"/"client"
OBSERVER_MODE="client"

SANDBOX_NAME=sandbox

cloneDependencies(){
  if [ -d "$TESTNET_DIR" ]; then
    rm -rf $TESTNET_DIR
  fi

  mkdir "$TESTNET_DIR"

  git clone https://github.com/multiversx/mx-chain-go "$TESTNET_DIR/mx-chain-go"
  checkoutStableVersion mx-chain-go 0bcc42220f436b40db3f15cb611e2713d43c04fa

  git clone https://github.com/multiversx/mx-chain-deploy-go "$TESTNET_DIR/mx-chain-deploy-go"
  git clone https://github.com/multiversx/mx-chain-proxy-go "$TESTNET_DIR/mx-chain-proxy-go"
}

checkoutStableVersion(){
    cd $TESTNET_DIR/$1
    git checkout $2
    cd ../..
}

testnetRemove(){
  rm -rf "$TESTNET_OUTPUT_DIR"
}

testnetSetup(){
  sed -i 's/TransactionSignedWithTxHashEnableEpoch =.*/TransactionSignedWithTxHashEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/BuiltInFunctionsEnableEpoch =.*/BuiltInFunctionsEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/ESDTEnableEpoch =.*/ESDTEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/ESDTMultiTransferEnableEpoch =.*/ESDTMultiTransferEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/ESDTTransferRoleEnableEpoch =.*/ESDTTransferRoleEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/MetaESDTSetEnableEpoch =.*/MetaESDTSetEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/ESDTRegisterAndSetAllRolesEnableEpoch =.*/ESDTRegisterAndSetAllRolesEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"
  sed -i 's/ESDTRegisterAndSetAllRolesEnableEpoch =.*/ESDTRegisterAndSetAllRolesEnableEpoch = 0/' "$ENABLE_EPOCH_DIR"

  sed -i 's/ScheduledMiniBlocksEnableEpoch =.*/ScheduledMiniBlocksEnableEpoch = 1/' "$ENABLE_EPOCH_DIR"
  sed -i 's/{ StartEpoch = 0, Version = "\*\" },/# Removed headerV1 version/' "$CONFIG_DIR"
  sed -i 's/{ StartEpoch = 1, Version = "2" },/{ StartEpoch = 0, Version = "2" },/' "$CONFIG_DIR"

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
  sed -i 's/SHARD_VALIDATORCOUNT=.*/SHARD_VALIDATORCOUNT=1/' $VARIABLES_PATH
  sed -i 's/SHARD_OBSERVERCOUNT=.*/SHARD_OBSERVERCOUNT=1/' $VARIABLES_PATH
  sed -i 's/SHARD_CONSENSUS_SIZE=.*/SHARD_CONSENSUS_SIZE=$SHARD_VALIDATORCOUNT/' $VARIABLES_PATH
  sed -i 's/META_VALIDATORCOUNT=.*/META_VALIDATORCOUNT=1/' $VARIABLES_PATH
  sed -i 's/META_OBSERVERCOUNT=.*/META_OBSERVERCOUNT=1/' $VARIABLES_PATH
  sed -i 's/META_CONSENSUS_SIZE=.*/META_CONSENSUS_SIZE=$META_VALIDATORCOUNT/' $VARIABLES_PATH
  sed -i 's/export NODE_DELAY=.*/export NODE_DELAY=30/' $VARIABLES_PATH

  sed -i 's/export PORT_SEEDNODE=.*/export PORT_SEEDNODE="9998"/' "$VARIABLES_PATH"
  sed -i 's/export PORT_ORIGIN_OBSERVER=.*/export PORT_ORIGIN_OBSERVER="21110"/' "$VARIABLES_PATH"
  sed -i 's/export PORT_ORIGIN_OBSERVER_REST=.*/export PORT_ORIGIN_OBSERVER_REST="10010"/' "$VARIABLES_PATH"
  sed -i 's/export PORT_ORIGIN_VALIDATOR=.*/export PORT_ORIGIN_VALIDATOR="21510"/' "$VARIABLES_PATH"
  sed -i 's/export PORT_ORIGIN_VALIDATOR_REST=.*/export PORT_ORIGIN_VALIDATOR_REST="9510"/' "$VARIABLES_PATH"
  sed -i 's/export PORT_PROXY=.*/export PORT_PROXY="7960"/' "$VARIABLES_PATH"
  sed -i 's/export PORT_TXGEN=.*/export PORT_TXGEN="7961"/' "$VARIABLES_PATH"

  sed -i 's/export USETMUX=.*/export USETMUX=0/' "$VARIABLES_PATH"
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
    ./mx-chain-go/scripts/testnet/start.sh
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
  - new to create a new testnet
  - start to start the testnet
  - reset to reset the testnet
  - stop to stop the testnet"
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
