#!/bin/bash

# Current location
SCRIPT_PATH=$(dirname "$(realpath "$BASH_SOURCE")")

# Source node variables
TESTNET_DIR=$(dirname $SCRIPT_PATH)
source $TESTNET_DIR/variables.sh

# Source all scripts
source $SCRIPT_PATH/config/configs.cfg
source $SCRIPT_PATH/config/helper.cfg
source $SCRIPT_PATH/config/esdt-safe.snippets.sh
source $SCRIPT_PATH/config/fee-market.snippets.sh
source $SCRIPT_PATH/config/header-verifier.snippets.sh
source $SCRIPT_PATH/config/common.snippets.sh
source $SCRIPT_PATH/config/token.snippets.sh
source $SCRIPT_PATH/config/py.snippets.sh
source $SCRIPT_PATH/config/contracts.snippets.sh
source $SCRIPT_PATH/config/deploy.snippets.sh
source $SCRIPT_PATH/observer/deployObserver.sh

# Create necessary directories
mkdir -p $(eval echo "${SOVEREIGN_DIRECTORY}")
mkdir -p $(eval echo "${TXS_OUTFILE_DIRECTORY}")
mkdir -p $(eval echo "${CONTRACTS_DIRECTORY}")

# Define other variables
WALLET_ADDRESS=$(echo "$(head -n 1 $(eval echo ${WALLET}))" | sed -n 's/.* for \([^-]*\)-----.*/\1/p')
OUTFILE_PATH=$(eval echo "${TXS_OUTFILE_DIRECTORY}")
