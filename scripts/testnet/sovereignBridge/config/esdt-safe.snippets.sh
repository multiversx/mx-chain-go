ESDT_SAFE_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-esdt-safe-contract)
ESDT_SAFE_ADDRESS_SOVEREIGN=$(mxpy data load --partition=sovereign --key=address-esdt-safe-contract)

deployEsdtSafeContract() {
    mxpy --verbose contract deploy \
        --bytecode="${ESDT_SAFE_WASM}" \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --arguments \
            ${MIN_VALID_SIGNERS} \
            ${INITIATOR_ADDRESS} \
            ${SIGNERS} \
        --outfile="${SCRIPT_PATH}/deploy-esdt-safe.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    TX_STATUS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-esdt-safe.interaction.json"  --expression="data['transactionOnNetwork']['status']")
    if [ "$TX_STATUS" != "success" ]; then
        echo "Transaction was not successful"
        return
    fi

    ADDRESS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-esdt-safe.interaction.json"  --expression="data['contractAddress']")
    mxpy data store --partition=${CHAIN_ID} --key=address-esdt-safe-contract --value=${ADDRESS}
    ESDT_SAFE_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-esdt-safe-contract)
    echo -e "\nESDT Safe contract: ${ADDRESS}"

    SOVEREIGN_CONTRACT_ADDRESS=$(firstSovereignContractAddress)
    mxpy data store --partition=sovereign --key=address-esdt-safe-contract --value=${SOVEREIGN_CONTRACT_ADDRESS}
    ESDT_SAFE_ADDRESS_SOVEREIGN=$(mxpy data load --partition=sovereign --key=address-esdt-safe-contract)
}

pauseEsdtSafeContract() {
    pauseEsdtSafe ${ESDT_SAFE_ADDRESS}
}
pauseEsdtSafeContractSovereign() {
    pauseEsdtSafe ${ESDT_SAFE_ADDRESS_SOVEREIGN}
}
pauseEsdtSafe() {
    if [ $# -eq 0 ]; then
        echo "No arguments provided"
        return
    fi

    mxpy --verbose contract call $1\
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=10000000 \
        --function="pause" \
        --recall-nonce \
        --wait-result \
        --send || return
}

unpauseEsdtSafeContract() {
    unpauseEsdtSafe ${ESDT_SAFE_ADDRESS}
}
unpauseEsdtSafeContractSovereign() {
    unpauseEsdtSafe ${ESDT_SAFE_ADDRESS_SOVEREIGN}
}
unpauseEsdtSafe() {
    if [ $# -eq 0 ]; then
        echo "No arguments provided"
        return
    fi

    mxpy --verbose contract call $1 \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=10000000 \
        --function="unpause" \
        --recall-nonce \
        --wait-result \
        --send || return
}

setFeeMarketAddress() {
    setFeeMarket ${ESDT_SAFE_ADDRESS} ${FEE_MARKET_ADDRESS}
}
setFeeMarketAddressSovereign() {
    setFeeMarket ${ESDT_SAFE_ADDRESS_SOVEREIGN} ${FEE_MARKET_ADDRESS_SOVEREIGN}
}
setFeeMarket() {
    if [ $# -eq 0 ]; then
        echo "No arguments provided"
        return
    fi

    mxpy --verbose contract call $1 \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=10000000 \
        --function="setFeeMarketAddress" \
        --arguments $2 \
        --recall-nonce \
        --wait-result \
        --send || return
}

changeEsdtSafeContractOwnerToMultisig() {
    CHECK_VARIABLES ESDT_SAFE_ADDRESS MULTISIG_VERIFIER_ADDRESS || return

    mxpy --verbose contract call ${ESDT_SAFE_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=10000000 \
        --function="ChangeOwnerAddress" \
        --arguments ${MULTISIG_VERIFIER_ADDRESS} \
        --recall-nonce \
        --wait-result \
        --send || return
}
