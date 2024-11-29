issueToken() {
    manualUpdateConfigFile #update config file

    echo "Issuing fungible token on main chain..."

    local TOKENS_TO_MINT=$(echo "scale=0; $INITIAL_SUPPLY*10^$NR_DECIMALS/1" | bc)
    local OUTFILE="${OUTFILE_PATH}/issue-token.interaction.json"
    mxpy contract call ${ESDT_SYSTEM_SC_ADDRESS} \
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
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    local HEX_TOKEN_IDENTIFIER=$(mxpy data parse --file=${OUTFILE}  --expression="data['transactionOnNetwork']['logs']['events'][2]['topics'][0]")
    local TOKEN_IDENTIFIER=$(echo "$HEX_TOKEN_IDENTIFIER" | xxd -r -p)
    updateConfig DEPOSIT_TOKEN_IDENTIFIER $TOKEN_IDENTIFIER
    echo "Issued Token identifier: ${DEPOSIT_TOKEN_IDENTIFIER}"
}

depositTokenInSC() {
    manualUpdateConfigFile #update config file

    echo "Depositing token in ESDT Safe contract on main chain..."
    checkVariables DEPOSIT_TOKEN_IDENTIFIER DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)
    local OUTFILE="${OUTFILE_PATH}/deposit-token.interaction.json"
    mxpy contract call ${WALLET_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=20000000 \
        --function="MultiESDTNFTTransfer" \
        --arguments \
           ${ESDT_SAFE_ADDRESS} \
           1 \
           str:${DEPOSIT_TOKEN_IDENTIFIER} \
           0 \
           ${AMOUNT_TO_TRANSFER} \
           str:deposit \
           ${WALLET_ADDRESS} \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}

issueTokenSovereign() {
    manualUpdateConfigFile #update config file

    echo "Issuing fungible token on sovereign chain..."

    local TOKENS_TO_MINT=$(echo "scale=0; $INITIAL_SUPPLY_SOVEREIGN*10^$NR_DECIMALS_SOVEREIGN/1" | bc)
    local OUTFILE="${OUTFILE_PATH}/issue-token-sovereign.interaction.json"
    mxpy contract call ${ESDT_SYSTEM_SC_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --gas-limit=60000000 \
        --value=${ESDT_ISSUE_COST} \
        --function="issue" \
        --arguments \
            str:${TOKEN_DISPLAY_NAME_SOVEREIGN} \
            str:${TOKEN_TICKER_SOVEREIGN} \
            ${TOKENS_TO_MINT} \
            ${NR_DECIMALS_SOVEREIGN} \
            str:canAddSpecialRoles str:true \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    local HEX_TOKEN_IDENTIFIER=$(mxpy data parse --file=${OUTFILE}  --expression="data['transactionOnNetwork']['logs']['events'][2]['topics'][0]")
    local TOKEN_IDENTIFIER=$(echo "$HEX_TOKEN_IDENTIFIER" | xxd -r -p)
    updateConfig DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN $TOKEN_IDENTIFIER
    echo "Issued Sovereign Token identifier: ${DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN}"
}

registerSovereignToken() {
    echo "Registering Sovereign token in ESDT Safe contract on main chain..."
    checkVariables DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN || return

    local OUTFILE="${OUTFILE_PATH}/register-token.interaction.json"
    mxpy contract call ${ESDT_SAFE_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=100000000 \
        --function="registerToken" \
        --value=${ESDT_ISSUE_COST} \
        --arguments \
            str:${DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN} \
            0 \
            str:${TOKEN_DISPLAY_NAME_SOVEREIGN} \
            str:${TOKEN_TICKER_SOVEREIGN} \
            ${NR_DECIMALS_SOVEREIGN} \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}

setLocalBurnRoleSovereign() {
    echo "Setting local burn role for ESDT Safe contract on sovereign chain..."
    checkVariables DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN ESDT_SAFE_ADDRESS_SOVEREIGN || return

    local OUTFILE="${OUTFILE_PATH}/set-localburn-role-sovereign.interaction.json"
    mxpy contract call ${ESDT_SYSTEM_SC_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --gas-limit=60000000 \
        --function="setSpecialRole" \
        --arguments \
            str:${DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN} \
            ${ESDT_SAFE_ADDRESS_SOVEREIGN} \
            str:ESDTRoleLocalBurn \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}

depositTokenInSCSovereign() {
    echo "Depositing token in ESDT Safe contract on sovereign chain..."
    checkVariables DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN DEPOSIT_TOKEN_NR_DECIMALS_SOVEREIGN DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER_SOVEREIGN || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER_SOVEREIGN*10^$DEPOSIT_TOKEN_NR_DECIMALS_SOVEREIGN/1" | bc)
    local OUTFILE="${OUTFILE_PATH}/deposit-token-sovereign.interaction.json"
    mxpy contract call ${WALLET_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --gas-limit=20000000 \
        --function="MultiESDTNFTTransfer" \
        --arguments \
           ${ESDT_SAFE_ADDRESS_SOVEREIGN} \
           1 \
           str:${DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN} \
           0 \
           ${AMOUNT_TO_TRANSFER} \
           str:deposit \
           ${WALLET_ADDRESS} \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}
