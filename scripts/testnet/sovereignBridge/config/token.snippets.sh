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

    mxpy --verbose contract call ${WALLET_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=20000000 \
        --function="MultiESDTNFTTransfer" \
        --arguments \
           ${ESDT_SAFE_ADDRESS} \
           2 \
           str:${DEPOSIT_TOKEN_IDENTIFIER} \
           0 \
           ${AMOUNT_TO_TRANSFER} \
           str:${DEPOSIT_TOKEN_IDENTIFIER} \
           0 \
           ${AMOUNT_TO_TRANSFER} \
           str:deposit \
           ${WALLET_ADDRESS} \
        --recall-nonce \
        --wait-result \
        --send || return
}

issueTokenSovereign() {
    manualUpdateConfigFile #update config file

    TOKENS_TO_MINT=$(echo "scale=0; $INITIAL_SUPPLY*10^$NR_DECIMALS/1" | bc)

    mxpy --verbose contract call ${ESDT_SYSTEM_SC_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --gas-limit=60000000 \
        --value=5000000000000000000 \
        --function="issue" \
        --arguments \
            str:${TOKEN_DISPLAY_NAME} \
            str:${TOKEN_TICKER} \
            ${TOKENS_TO_MINT} \
            ${NR_DECIMALS} \
            str:canAddSpecialRoles str:true \
        --outfile="${SCRIPT_PATH}/issue-sovereign-token.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    HEX_TOKEN_IDENTIFIER=$(mxpy data parse --file="${SCRIPT_PATH}/issue-sovereign-token.interaction.json"  --expression="data['transactionOnNetwork']['logs']['events'][2]['topics'][0]")
    TOKEN_IDENTIFIER=$(echo "$HEX_TOKEN_IDENTIFIER" | xxd -r -p)
    update-config DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN $TOKEN_IDENTIFIER
}

depositTokenInSCSovereign() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

    mxpy --verbose contract call ${WALLET_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --gas-limit=20000000 \
        --function="MultiESDTNFTTransfer" \
        --arguments \
           ${ESDT_SAFE_ADDRESS_SOVEREIGN} \
           2 \
           str:${DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN} \
           0 \
           ${AMOUNT_TO_TRANSFER} \
           str:${DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN} \
           0 \
           ${AMOUNT_TO_TRANSFER} \
           str:deposit \
           ${WALLET_ADDRESS} \
        --recall-nonce \
        --wait-result \
        --send || return
}

getFundsInAddressSovereign() {
    mxpy tx new \
        --pem="/home/ubuntu/MultiversX/testnet/node/config/walletKey.pem" \
        --pem-index 0 \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --receiver=${WALLET_ADDRESS} \
        --value=200000000000000000000 \
        --gas-limit=50000 \
        --recall-nonce \
        --send

    sleep 6
}
