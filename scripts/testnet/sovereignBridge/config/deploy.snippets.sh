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

    deployEsdtSafeContract

    deployFeeMarketContract

    setFeeMarketAddress

    disableFeeMarketContract

    unpauseEsdtSafeContract

    setGenesisContract

    updateSovereignConfig

    prepareObserver
}

# This function will:
# - update some parameter in notifier
# - run the sovereign nodes config
# - deploy multisig contract on main chain
# - start the bridge service, nodes and the observer
# - do other transactions in sovereign contracts
sovereignDeploy() {
    checkWalletBalanceOnMainChain || return

    updateNotifierNotarizationRound

    ../config.sh

    deployMultiSigVerifierContract

    setEsdtSafeAddressInMultiSigVerifier

    sovereignStart

    setMultiSigVerifierAddressInEsdtSafe

    setSovereignBridgeAddressInEsdtSafe

    getFundsInAddressSovereign

    setFeeMarketAddressSovereign

    disableFeeMarketContractSovereign

    unpauseEsdtSafeContractSovereign
}

# This function will:
# - update and start bridge service
# - start sovereign nodes
# - deploy the main chain observer
sovereignStart() {
    updateAndStartBridgeService

    ../sovereignStart.sh

    deployObserver
}

# This function will:
# - stop sovereign nodes and services
# - start again the sovereign nodes and services
sovereignRestart() {
    stopSovereign

    sovereignStart
}

# This function will:
# - stop sovereign nodes
# - stop the bridge service
# - stop the main chain observer
stopSovereign() {
    ../stop.sh

    screen -S sovereignBridgeService -X kill

    stopObserver
}

# This function will:
# - stop sovereign nodes and services
# - clean the sovereign configuration
stopAndCleanSovereign() {
    stopSovereign

    ../clean.sh
}
