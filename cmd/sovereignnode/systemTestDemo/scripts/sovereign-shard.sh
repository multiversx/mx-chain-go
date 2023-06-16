CURRENT_DIR=$(pwd)
WORKING_DIR="$CURRENT_DIR"
SOVEREIGN_DIR=$WORKING_DIR/sovereign
SOVEREIGN_OUTPUT_DIR=$SOVEREIGN_DIR/sovereign-local
SCRIPTS_DIR=mx-chain-go/scripts/testnet
VARIABLES_PATH=$SCRIPTS_DIR/variables.sh
OBSERVERS_PATH=$SCRIPTS_DIR/include/observers.sh
VALIDATORS_PATH=$SCRIPTS_DIR/include/validators.sh
ENABLE_EPOCH_DIR=$SOVEREIGN_DIR/mx-chain-go/cmd/node/config/enableEpochs.toml
CONFIG_DIR=$SOVEREIGN_DIR/mx-chain-go/cmd/node/config/config.toml
SYSTEM_SC_CONFIG_DIR=$SOVEREIGN_DIR/mx-chain-go/cmd/node/config/systemSmartContractsConfig.toml
SANDBOX_NAME=sandbox

cloneDependencies(){
  if [ -d "$SOVEREIGN_DIR" ]; then
    rm -rf $SOVEREIGN_DIR
  fi

  mkdir "$SOVEREIGN_DIR"

  git clone https://github.com/multiversx/mx-chain-go "$SOVEREIGN_DIR/mx-chain-go"
  cd $SOVEREIGN_DIR/mx-chain-go
  git checkout 954bae92b09c62317391c5f8af5831921ab2ff67
  cd ../..

  git clone https://github.com/multiversx/mx-chain-deploy-go "$SOVEREIGN_DIR/mx-chain-deploy-go"
  git clone https://github.com/multiversx/mx-chain-proxy-go "$SOVEREIGN_DIR/mx-chain-proxy-go"
}

sovereignRemove(){
  rm -rf "$SOVEREIGN_OUTPUT_DIR"
}

sovereignSetup(){
  mkdir "$SOVEREIGN_OUTPUT_DIR"
  cd "$SOVEREIGN_OUTPUT_DIR"
  ln -s "$SOVEREIGN_DIR"/mx-chain-go mx-chain-go
  ln -s "$SOVEREIGN_DIR"/mx-chain-deploy-go mx-chain-deploy-go
  ln -s "$SOVEREIGN_DIR"/mx-chain-proxy-go mx-chain-proxy-go
}

sovereignPrereq(){
  cd "$SOVEREIGN_OUTPUT_DIR" && \
    bash $SCRIPTS_DIR/prerequisites.sh && \
    echo -e "export TESTNETDIR=$SOVEREIGN_OUTPUT_DIR/$SANDBOX_NAME" > $SCRIPTS_DIR/local.sh && \
    bash $SCRIPTS_DIR/config.sh
}

sovereignUpdateVariables(){
  sed -i 's/export SHARDCOUNT=.*/export SHARDCOUNT=1/' $VARIABLES_PATH
  sed -i 's/SHARD_VALIDATORCOUNT=.*/SHARD_VALIDATORCOUNT=2/' $VARIABLES_PATH
  sed -i 's/SHARD_OBSERVERCOUNT=.*/SHARD_OBSERVERCOUNT=1/' $VARIABLES_PATH
  sed -i 's/SHARD_CONSENSUS_SIZE=.*/SHARD_CONSENSUS_SIZE=$SHARD_VALIDATORCOUNT/' $VARIABLES_PATH
  sed -i 's/export NODE_DELAY=.*/export NODE_DELAY=30/' $VARIABLES_PATH

  sed -i 's/export LOGLEVEL=.*/export LOGLEVEL="\*\:DEBUG"/' $VARIABLES_PATH
}

sovereignNew(){
  cloneDependencies
  sovereignRemove
  sovereignSetup
  sovereignUpdateVariables
  sovereignPrereq
}

sovereignStart(){
  cd "$SOVEREIGN_DIR" && \
    ./mx-chain-go/scripts/testnet/sovereignStart.sh
}

sovereignStop(){
  cd "$SOVEREIGN_DIR" && \
    ./mx-chain-go/scripts/testnet/stop.sh
}

sovereignReset(){
  cd "$SOVEREIGN_DIR" && \
    ./mx-chain-go/scripts/testnet/clean.sh && \
      ./mx-chain-go/scripts/testnet/config.sh
     ./mx-chain-go/scripts/testnet/sovereignStart.sh \

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
        sovereignNew ;;
      start)
        sovereignStart ;;
      stop)
        sovereignStop ;;
      reset)
        sovereignReset ;;
      *)
        echoOptions ;;
      esac
  else
    echoOptions
  fi
}

main "$@"
