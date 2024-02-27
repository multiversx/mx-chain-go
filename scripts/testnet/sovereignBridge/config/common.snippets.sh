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

sovereignInit() {
    ../config.sh

    ../sovereignStart.sh

    deployObserver

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