checkWalletBalanceOnMainChain() {
    local BALANCE=$(mxpy account get --address ${WALLET_ADDRESS} --proxy ${PROXY} --balance)
    if [ "$BALANCE" == "0" ]; then
        echo -e "Your wallet balance is zero on main chain"
        return 1
    fi
    return 0
}

getFundsInAddressSovereign() {
    echo "Getting funds in wallet on sovereign chain..."

    local OUTFILE="${OUTFILE_PATH}/get-funds-sovereign.interaction.json"
    mxpy tx new \
        --pem="~/MultiversX/testnet/node/config/walletKey.pem" \
        --pem-index 0 \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --receiver=${WALLET_ADDRESS} \
        --value=200000000000000000000 \
        --gas-limit=50000 \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --send

    sleep 6
}