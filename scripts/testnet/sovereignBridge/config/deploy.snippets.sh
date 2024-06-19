# This function will:
# - deploy all main chain contracts and update sovereign configs
# - deploy sovereign nodes with all services
deploySovereignWithCrossChainContracts() {
    deployMainChainContractsAndSetupObserver || return

    sovereignDeploy
}

# This function will:
# - deploy all main chain contracts
# - update sovereign configs
# - prepare a main chain observer for sovereign nodes
deployMainChainContractsAndSetupObserver() {
    checkWalletBalanceOnMainChain || return

    deployEsdtSafeContract || return

    deployFeeMarketContract || return

    setFeeMarketAddress

    disableFeeInFeeMarketContract

    unpauseEsdtSafeContract

    setGenesisContract

    updateSovereignConfig

    prepareObserver
}

# This function will:
# - update some parameter in notifier
# - run the sovereign nodes config
# - deploy header verifier contract on main chain
# - start the bridge service, nodes and the observer
# - do other transactions in sovereign contracts
sovereignDeploy() {
    checkWalletBalanceOnMainChain || return

    updateNotifierNotarizationRound

    $TESTNET_DIR/config.sh

    deployHeaderVerifierContract || return

    setEsdtSafeAddressInHeaderVerifier

    sovereignStart

    setHeaderVerifierAddressInEsdtSafe

    getFundsInAddressSovereign

    setFeeMarketAddressSovereign

    disableFeeInFeeMarketContractSovereign

    unpauseEsdtSafeContractSovereign
}

# This function will:
# - update and start bridge service
# - start sovereign nodes
# - deploy the main chain observer
sovereignStart() {
    updateAndStartBridgeService

    $TESTNET_DIR/sovereignStart.sh

    deployObserver
}

# This function will:
# - stop sovereign nodes and services
# - deploy sovereign nodes with all services
sovereignRestart() {
    stopAndCleanSovereign

    sovereignDeploy
}

# This function will:
# - stop sovereign and clean nodes
# - pull the latest changes for all the repositories
# - download the new version of the contracts and update them on main chain
# - deploy sovereign nodes with all services
sovereignUpgradeAndRestart() {
    stopAndCleanSovereign

    gitPullAllChanges || return

    downloadCrossChainContracts

    upgradeEsdtSafeContract

    upgradeFeeMarketContract

    sovereignDeploy
}

# This function will:
# - stop sovereign nodes
# - stop the bridge service
# - stop the main chain observer
stopSovereign() {
    $TESTNET_DIR/stop.sh

    screen -S sovereignBridgeService -X kill

    stopObserver
}

# This function will:
# - stop sovereign nodes and services
# - clean the sovereign configuration
stopAndCleanSovereign() {
    stopSovereign

    $TESTNET_DIR/clean.sh
}
