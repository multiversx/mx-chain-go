# This function will deploy full sovereign setup:
# - deploy all main chain contracts and update sovereign configs
# - deploy sovereign nodes with all services
deploySovereignWithCrossChainContracts() {
    deployMainChainContractsAndSetupObserver $1 || return

    sovereignDeploy
}

# This function will deploy main chain services:
# - deploy all main chain contracts
# - update sovereign configs
# - prepare a main chain observer for sovereign nodes
deployMainChainContractsAndSetupObserver() {
    deployEsdtSafeContract || return

    deployFeeMarketContract || return

    setFeeMarketAddress

    unpauseEsdtSafeContract

    setGenesisContract

    updateSovereignConfig $1

    prepareObserver
}

# This function will deploy sovereign:
# - update some parameter in notifier
# - run the sovereign nodes config
# - deploy header verifier contract on main chain
# - start the bridge service, nodes and the observer
# - do other transactions in sovereign contracts
sovereignDeploy() {
    updateNotifierNotarizationRound

    $TESTNET_DIR/config.sh

    deployHeaderVerifierContract || return

    setEsdtSafeAddressInHeaderVerifier

    setHeaderVerifierAddressInEsdtSafe

    createObserver

    sovereignStart

    mxpy config set default_address_hrp vibe

    getFundsInAddressSovereign

    setFeeMarketAddressSovereign

    unpauseEsdtSafeContractSovereign

    mxpy config set default_address_hrp erd
}

# This function will start sovereign:
# - update and start bridge service
# - start sovereign nodes
# - deploy main chain observer
sovereignStart() {
    deployObserver

    updateAndStartBridgeService

    $TESTNET_DIR/sovereignStart.sh

    /home/ubuntu/mx-services/redeploy.sh
}

# This function will reset sovereign:
# - stop sovereign nodes and services
# - deploy sovereign nodes with all services
sovereignReset() {
    stopAndCleanSovereign

    sovereignDeploy
}

# This function will upgrade and reset sovereign:
# - stop sovereign and clean nodes
# - pull the latest changes for all the repositories
# - download the new version of the contracts and update them on main chain
# - update sovereign configs
# - deploy sovereign nodes with all services
sovereignUpgradeAndReset() {
    stopAndCleanSovereign

    gitPullAllChanges || return

    downloadCrossChainContracts

    upgradeEsdtSafeContract

    upgradeFeeMarketContract

    setGenesisContract

    updateSovereignConfig

    prepareObserver

    sovereignDeploy
}

# This function will stop sovereign:
# - stop sovereign nodes
# - stop the bridge service
# - stop the main chain observer
stopSovereign() {
    $TESTNET_DIR/stop.sh

    screen -S sovereignBridgeService -X kill

    stopObserver

    /home/ubuntu/mx-services/stop.sh
}

# This function will stop and clean sovereign:
# - stop sovereign nodes and services
# - clean the sovereign configuration and observer
stopAndCleanSovereign() {
    stopSovereign

    $TESTNET_DIR/clean.sh

    cleanObserver
}
