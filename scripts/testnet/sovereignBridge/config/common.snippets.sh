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
    copyContracts

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

    setSovereignBridgeAddress

    updateAndStartBridgeService

    getFundsInAddressSovereign

    setFeeMarketAddressSovereign

    disableFeeMarketContractSovereign

    unpauseEsdtSafeContractSovereign

    issueTokenSovereign
}

upgradeContractsAndStartSovereign() {
    copyContracts

    upgradeContracts

    sovereignInit
}

stopSovereign() {
    ../stop.sh

    ../clean.sh

    stopObserver
}