deploySovereignWithCrossChainContracts() {
    deployMainChainContractsAndSetupObserver || return

    sovereignDeploy || return
}

deployMainChainContractsAndSetupObserver() {
    checkWalletBalance || return

    deployEsdtSafeContract || return

    deployFeeMarketContract || return

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

    deployMultiSigVerifierContract || return

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
