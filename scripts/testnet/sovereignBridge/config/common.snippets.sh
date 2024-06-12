deploySovereignWithCrossChainContracts() {
    deployMainChainContractsAndSetupObserver

    sovereignDeploy
}

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

    deployMultiSigVerifierContract

    setEsdtSafeAddress

    sovereignStart

    setMultiSigAddress

    setSovereignBridgeAddress

    getFundsInAddressSovereign

    setFeeMarketAddressSovereign

    disableFeeMarketContractSovereign

    unpauseEsdtSafeContractSovereign
}

sovereignStart() {
    updateAndStartBridgeService

    ../sovereignStart.sh

    deployObserver
}

stopSovereign() {
    ../stop.sh

    screen -S sovereignBridgeService -X kill

    stopObserver
}

stopAndCleanSovereign() {
    stopSovereign

    ../clean.sh
}
