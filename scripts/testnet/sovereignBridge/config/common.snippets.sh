deployMainChainContractsAndSetupObserver() {
    deployEsdtSafeContract

    deployFeeMarketContract

    setFeeMarketAddress

    disableFeeMarketContract

    unpauseEsdtSafeContract

    setGenesisContract

    updateSovereignConfig

    prepareObserver
}

sovereignDeploy() {
    updateNotifierNotarizationRound

    ../config.sh

    deployMultisigVerifierContract

    setEsdtSafeAddress

    updateAndStartBridgeService

    ../sovereignStart.sh

    deployObserver

    setMultisigAddress

    setSovereignBridgeAddress

    getFundsInAddressSovereign

    setFeeMarketAddressSovereign

    disableFeeMarketContractSovereign

    unpauseEsdtSafeContractSovereign
}

stopAndCleanSovereign() {
    ../stop.sh

    screen -S sovereignBridgeService -X kill

    ../clean.sh

    stopObserver
}
