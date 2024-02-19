issueToken() {
    manualUpdateConfigFile #update config file

    TOKENS_TO_MINT=$(echo "scale=0; $INITIAL_SUPPLY*10^$NR_DECIMALS/1" | bc)

    mxpy --verbose contract call ${ESDT_SYSTEM_SC_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=60000000 \
        --value=${ESDT_ISSUE_COST} \
        --function="issue" \
        --arguments \
            str:${TOKEN_DISPLAY_NAME} \
            str:${TOKEN_TICKER} \
            ${TOKENS_TO_MINT} \
            ${NR_DECIMALS} \
            str:canAddSpecialRoles str:true \
        --outfile="${SCRIPT_PATH}/issue-token.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    HEX_TOKEN_IDENTIFIER=$(mxpy data parse --file="${SCRIPT_PATH}/issue-token.interaction.json"  --expression="data['transactionOnNetwork']['logs']['events'][2]['topics'][0]")
    TOKEN_IDENTIFIER=$(echo "$HEX_TOKEN_IDENTIFIER" | xxd -r -p)
    update-config DEPOSIT_TOKEN_IDENTIFIER $TOKEN_IDENTIFIER
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
