issueToken() {
    manualUpdateConfigFile #update config file

    local TOKENS_TO_MINT=$(echo "scale=0; $INITIAL_SUPPLY*10^$NR_DECIMALS/1" | bc)

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

    local HEX_TOKEN_IDENTIFIER=$(mxpy data parse --file="${SCRIPT_PATH}/issue-token.interaction.json"  --expression="data['transactionOnNetwork']['logs']['events'][2]['topics'][0]")
    local TOKEN_IDENTIFIER=$(echo "$HEX_TOKEN_IDENTIFIER" | xxd -r -p)
    update-config DEPOSIT_TOKEN_IDENTIFIER $TOKEN_IDENTIFIER
}

depositTokenInSC() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

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

nftDepositInSC() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

    mxpy --verbose contract call ${WALLET_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=20000000 \
        --function="MultiESDTNFTTransfer" \
        --arguments \
           ${ESDT_SAFE_ADDRESS} \
           1 \
           str:SNFT-34c04c \
           2 \
           1 \
           str:deposit \
           ${WALLET_ADDRESS} \
        --recall-nonce \
        --wait-result \
        --send || return
}

depositTokenInSCAdder() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

    mxpy --verbose contract call ${WALLET_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=30000000 \
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
           erd1qqqqqqqqqqqqqpgqp6k29tdnray9kzzetsv50fxgyzgkt5pqulmq0yp9c6 \
           0x0000000001312d00 \
           0x616464 \
           0x0000000401312d00 \
        --recall-nonce \
        --wait-result \
        --send || return
}

issueTokenSovereign() {
    manualUpdateConfigFile #update config file

    local TOKENS_TO_MINT=$(echo "scale=0; $INITIAL_SUPPLY*10^$NR_DECIMALS/1" | bc)

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

    local HEX_TOKEN_IDENTIFIER=$(mxpy data parse --file="${SCRIPT_PATH}/issue-sovereign-token.interaction.json"  --expression="data['transactionOnNetwork']['logs']['events'][2]['topics'][0]")
    local TOKEN_IDENTIFIER=$(echo "$HEX_TOKEN_IDENTIFIER" | xxd -r -p)
    update-config DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN $TOKEN_IDENTIFIER

    mxpy --verbose contract call ${ESDT_SAFE_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=100000000 \
        --function="registerToken" \
        --value=50000000000000000 \
        --arguments \
           str:${DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN} \
           0 \
           str:SovToken \
           str:SOVT \
           0x12 \
        --recall-nonce \
        --wait-result \
        --send || return
}

depositTokenInSCSovereign() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

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

depositTokenInSCSovereignFail() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

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
           erd1qqqqqqqqqqqqqpgqpzu5ezwlsm7zqnke62ylsxjtwta7qz50ulmqa9xpwa \
           0x0000000001312d00 \
           0x61646464 \
           0x0000000401312d00 \
        --recall-nonce \
        --wait-result \
        --send || return
}

depositTokenInSCSovereignBack() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

    mxpy --verbose contract call ${WALLET_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --gas-limit=20000000 \
        --function="MultiESDTNFTTransfer" \
        --arguments \
           ${ESDT_SAFE_ADDRESS_SOVEREIGN} \
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

depositTokenInSCSovereignWithArgs() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

    mxpy --verbose contract call ${WALLET_ADDRESS} \
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
           0x0000000001312d00 \
           0x616464 \
           0x0000000401312d00 \
        --recall-nonce \
        --wait-result \
        --send || return
}

getFundsInAddressSovereign() {
    mxpy tx new \
        --pem="~/MultiversX/testnet/node/config/walletKey.pem" \
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

#--------------------------------------------------------------------------------------
#--------------------------------------------------------------------------------------
#--------------------------------------------------------------------------------------
#--------------------------------------------------------------------------------------

getFundsInAddress2Sovereign() {
    mxpy tx new \
        --pem="~/MultiversX/testnet/node/config/walletKey.pem" \
        --pem-index 0 \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --receiver=erd1353zkd5yq8shpwk3djgz6r7qq4rkxtajfn2gkk79muvuhun0l5zqavu826 \
        --value=200000000000000000000 \
        --gas-limit=50000 \
        --recall-nonce \
        --send

    sleep 6
}

depositTokenInSC2() {
    manualUpdateConfigFile #update config file

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

    mxpy --verbose contract call erd1353zkd5yq8shpwk3djgz6r7qq4rkxtajfn2gkk79muvuhun0l5zqavu826 \
        --pem="/home/ubuntu/Wallets/wallet.pem" \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=20000000 \
        --function="MultiESDTNFTTransfer" \
        --arguments \
           ${ESDT_SAFE_ADDRESS} \
           1 \
           str:SVN-d9b66a \
           0 \
           ${AMOUNT_TO_TRANSFER} \
           str:deposit \
           erd1353zkd5yq8shpwk3djgz6r7qq4rkxtajfn2gkk79muvuhun0l5zqavu826 \
        --recall-nonce \
        --wait-result \
        --send || return
}

issueToken2Sovereign() {
    manualUpdateConfigFile #update config file

    local TOKENS_TO_MINT=$(echo "scale=0; $INITIAL_SUPPLY*10^$NR_DECIMALS/1" | bc)

    mxpy --verbose contract call ${ESDT_SYSTEM_SC_ADDRESS} \
        --pem="/home/ubuntu/Wallets/wallet.pem" \
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
        --outfile="${SCRIPT_PATH}/issue2-sovereign-token.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    local HEX_TOKEN_IDENTIFIER=$(mxpy data parse --file="${SCRIPT_PATH}/issue2-sovereign-token.interaction.json"  --expression="data['transactionOnNetwork']['logs']['events'][2]['topics'][0]")
    local TOKEN_IDENTIFIER=$(echo "$HEX_TOKEN_IDENTIFIER" | xxd -r -p)
    update-config DEPOSIT_TOKEN2_IDENTIFIER_SOVEREIGN $TOKEN_IDENTIFIER

    mxpy --verbose contract call ${ESDT_SAFE_ADDRESS} \
        --pem="/home/ubuntu/Wallets/wallet.pem" \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=100000000 \
        --function="registerToken" \
        --value=50000000000000000 \
        --arguments \
           str:${DEPOSIT_TOKEN2_IDENTIFIER_SOVEREIGN} \
           0 \
           str:SovToken \
           str:SOVT \
           0x12 \
        --recall-nonce \
        --wait-result \
        --send || return
}

depositTokenInSC2Sovereign() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER_SOVEREIGN DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

    mxpy --verbose contract call erd1353zkd5yq8shpwk3djgz6r7qq4rkxtajfn2gkk79muvuhun0l5zqavu826 \
        --pem="/home/ubuntu/Wallets/wallet.pem" \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --gas-limit=20000000 \
        --function="MultiESDTNFTTransfer" \
        --arguments \
           ${ESDT_SAFE_ADDRESS_SOVEREIGN} \
           1 \
           str:${DEPOSIT_TOKEN2_IDENTIFIER_SOVEREIGN} \
           0 \
           ${AMOUNT_TO_TRANSFER} \
           str:deposit \
           erd1353zkd5yq8shpwk3djgz6r7qq4rkxtajfn2gkk79muvuhun0l5zqavu826 \
        --recall-nonce \
        --wait-result \
        --send || return
}

dep() {
    manualUpdateConfigFile #update config file

    CHECK_VARIABLES DEPOSIT_TOKEN_IDENTIFIER DEPOSIT_TOKEN_NR_DECIMALS DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER || return

    local AMOUNT_TO_TRANSFER=$(echo "scale=0; $DEPOSIT_TOKEN_AMOUNT_TO_TRANSFER*10^$DEPOSIT_TOKEN_NR_DECIMALS/1" | bc)

    mxpy --verbose contract call ${WALLET_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=20000000 \
        --function="MultiESDTNFTTransfer" \
        --arguments \
           ${ESDT_SAFE_ADDRESS} \
           2 \
           str:SVN-c53da0 \
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

distributeFees() {
    mxpy --verbose contract call ${FEE_MARKET_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=20000000 \
        --function="distributeFees" \
        --arguments \
           erd1fp7lg768s4t7yk2e3yz6t69dt6n85hdxq477h7vq3dywszrng8fs3j4ku8 \
           0x2710 \
        --recall-nonce \
        --wait-result \
        --send || return
}
