deployAll() {
    deployEsdtSafeContract

    deployFeeMarketContract

    setFeeMarketAddress

    disableFeeMarketContract

    unpauseEsdtSafeContract

    issueToken

    setGenesisContract

    updateSovereignConfig

    prepareObserver
}

upgradeContracts() {
    upgradeEsdtSafeContract

    upgradeFeeMarketContract

    setGenesisContract
}

sovereignInit() {
    updateNotifierNotarizationRound

    ../config.sh

    ../sovereignStart.sh

    deployObserver

    deployMultisigVerifierContract

    setMultisigAddress

#    updateAndStartBridgeService

    getFundsInAddressSovereign

    setFeeMarketAddressSovereign

    disableFeeMarketContractSovereign

    unpauseEsdtSafeContractSovereign

    issueTokenSovereign
}

stopSovereign() {
    ../stop.sh

    ../clean.sh

    stopObserver
}