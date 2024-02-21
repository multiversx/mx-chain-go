#!/bin/bash

# Current location
SCRIPT_PATH=$(dirname "$(realpath "$BASH_SOURCE")")

# Source all scripts
source $SCRIPT_PATH/config/configs.cfg
source $SCRIPT_PATH/config/helper.cfg
source $SCRIPT_PATH/config/esdt-safe.snippets.sh
source $SCRIPT_PATH/config/fee-market.snippets.sh
source $SCRIPT_PATH/config/multisigverifier.snippets.sh
source $SCRIPT_PATH/config/token.snippets.sh
source $SCRIPT_PATH/config/py.snippets.sh
source $SCRIPT_PATH/observer/deployObserver.sh
