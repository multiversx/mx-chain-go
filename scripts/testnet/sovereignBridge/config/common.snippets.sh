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
}

sovereignInit() {
    ../config.sh

    ../sovereignStart.sh

    deployObserver

    deployMultisigVerifierContract

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