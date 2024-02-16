issueToken() {
    TOKENS_TO_MINT=$(echo "scale=0; $CHAIN_SPECIFIC_TOKENS_TO_MINT*10^$NR_DECIMALS_CHAIN_SPECIFIC/1" | bc)

    mxpy --verbose contract call ${ESDT_SYSTEM_SC_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=60000000 \
        --value=${ESDT_ISSUE_COST} \
        --function="issue" \
        --arguments \
            str:${CHAIN_SPECIFIC_TOKEN_DISPLAY_NAME} \
            str:${CHAIN_SPECIFIC_TOKEN_TICKER} \
            ${TOKENS_TO_MINT} \
            ${NR_DECIMALS_CHAIN_SPECIFIC} \
            str:canAddSpecialRoles str:true \
        --recall-nonce \
        --wait-result \
        --send || return
}

depositTokenInSC() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

    mxpy --verbose contract call ${ESDT_SAFE_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=10000000 \
        --function="ESDTTransfer" \
        --arguments \
            str:${DEPOSIT_TOKEN_IDENTIFIER} \
            ${AMOUNT_TO_TRANSFER} \
            str:deposit \
        --recall-nonce \
        --wait-result \
        --send || return
}
