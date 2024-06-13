deploySovereignWithCrossChainContracts() {
    deployMainChainContractsAndSetupObserver || return

    sovereignDeploy
}

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

sovereignDeploy() {
    checkWalletBalanceOnMainChain || return

    updateNotifierNotarizationRound

    ../config.sh

    deployMultiSigVerifierContract

    setEsdtSafeAddressInMultiSigVerifier

    sovereignStart

    setMultiSigVerifierAddressInEsdtSafe

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
