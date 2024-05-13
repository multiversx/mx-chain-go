deployAll() {
    deployEsdtSafeContract

    deployFeeMarketContract

    setFeeMarketAddress

    disableFeeMarketContract

    unpauseEsdtSafeContract

    setGenesisContract

    updateSovereignConfig

    prepareObserver
}

sovereignInit() {
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

stopSovereign() {
    ../stop.sh

    screen -S sovereignBridgeService -X kill

    ../clean.sh

    stopObserver
}