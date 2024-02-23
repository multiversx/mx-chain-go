initTransactions() {
    deployEsdtSafeContract

    deployFeeMarketContract

    setFeeMarketAddress

    disableFeeMarketContract

    unpauseEsdtSafeContract

    issueToken

    setGenesisContract

    updateSovereignConfig
}

initTransactionsSovereign() {
    getFundsInAddress

    setFeeMarketAddressSovereign

    disableFeeMarketContractSovereign

    unpauseEsdtSafeContractSovereign

    issueTokenSovereign
}

getFundsInAddress() {
    mxpy tx new \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --receiver=${WALLET_ADDRESS} \
        --value=200000000000000000000 \
        --recall-nonce \
        --wait-result \
        --send || return
}