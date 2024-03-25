deployAll() {
    deployEsdtSafeContract

    deployFeeMarketContract

    setFeeMarketAddress

    disableFeeMarketContract

    unpauseEsdtSafeContract

#    issueToken

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

    deployMultisigVerifierContract

    updateAndStartBridgeService

    ../sovereignStart.sh

    deployObserver

    setMultisigAddress

    setSovereignBridgeAddress

    getFundsInAddressSovereign

    setFeeMarketAddressSovereign

    disableFeeMarketContractSovereign

    unpauseEsdtSafeContractSovereign

    issueTokenSovereign
}

upgradeContractsAndStartSovereign() {
    upgradeContracts

    sovereignInit
}

stopSovereign() {
    ../stop.sh

    screen -S sovereignBridgeService -X kill

    ../clean.sh

    stopObserver
}