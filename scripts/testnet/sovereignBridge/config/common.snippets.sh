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

#    upgradeMultisigVerifierContract

    setGenesisContract
}

sovereignInit() {
    updateNotifierNotarizationRound

    ../config.sh

    ../sovereignStart.sh

    deployObserver

    deployMultisigVerifierContract

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