ESDT_SAFE_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-esdt-safe-contract)
ESDT_SAFE_ADDRESS_SOVEREIGN=$(mxpy data load --partition=sovereign --key=address-esdt-safe-contract)

deployEsdtSafeContract() {
    mxpy --verbose contract deploy \
        --bytecode=$(eval echo ${ESDT_SAFE_WASM}) \
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

    local TX_STATUS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-esdt-safe.interaction.json"  --expression="data['transactionOnNetwork']['status']")
    if [ "$TX_STATUS" != "success" ]; then
        echo "Transaction was not successful"
        return
    fi

    local ADDRESS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-esdt-safe.interaction.json"  --expression="data['contractAddress']")
    mxpy data store --partition=${CHAIN_ID} --key=address-esdt-safe-contract --value=${ADDRESS}
    ESDT_SAFE_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-esdt-safe-contract)
    echo -e "ESDT Safe contract: ${ADDRESS}"

    local SOVEREIGN_CONTRACT_ADDRESS=$(firstSovereignContractAddress)
    mxpy data store --partition=sovereign --key=address-esdt-safe-contract --value=${SOVEREIGN_CONTRACT_ADDRESS}
    ESDT_SAFE_ADDRESS_SOVEREIGN=$(mxpy data load --partition=sovereign --key=address-esdt-safe-contract)
    echo -e "ESDT Safe sovereign contract: ${SOVEREIGN_CONTRACT_ADDRESS}"
}

upgradeEsdtSafeContract() {
    mxpy --verbose contract upgrade ${ESDT_SAFE_ADDRESS} \
        --bytecode=$(eval echo ${ESDT_SAFE_WASM}) \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --outfile="${SCRIPT_PATH}/upgrade-esdt-safe.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    local TX_STATUS=$(mxpy data parse --file="${SCRIPT_PATH}/upgrade-esdt-safe.interaction.json"  --expression="data['transactionOnNetwork']['status']")
    if [ "$TX_STATUS" != "success" ]; then
        echo "Transaction was not successful"
        return
    fi
}

pauseEsdtSafeContract() {
    pauseEsdtSafeContractCall ${ESDT_SAFE_ADDRESS} ${PROXY} ${CHAIN_ID}
}
pauseEsdtSafeContractSovereign() {
    pauseEsdtSafeContractCall ${ESDT_SAFE_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN}
}
pauseEsdtSafeContractCall() {
    if [ $# -lt 3 ]; then
        echo "Usage: $0 <arg1> <arg2> <arg3>"
        exit 1
    fi

    local ADDRESS=$1
    local URL=$2
    local CHAIN=$3

    mxpy --verbose contract call ${ADDRESS} \
        --pem=${WALLET} \
        --proxy=${URL} \
        --chain=${CHAIN} \
        --gas-limit=10000000 \
        --function="pause" \
        --recall-nonce \
        --wait-result \
        --send || return
}

unpauseEsdtSafeContract() {
    unpauseEsdtSafeContractCall ${ESDT_SAFE_ADDRESS} ${PROXY} ${CHAIN_ID}
}
unpauseEsdtSafeContractSovereign() {
    unpauseEsdtSafeContractCall ${ESDT_SAFE_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN}
}
unpauseEsdtSafeContractCall() {
    if [ $# -lt 3 ]; then
        echo "Usage: $0 <arg1> <arg2> <arg3>"
        exit 1
    fi

    local ADDRESS=$1
    local URL=$2
    local CHAIN=$3

    mxpy --verbose contract call ${ADDRESS} \
        --pem=${WALLET} \
        --proxy=${URL} \
        --chain=${CHAIN} \
        --gas-limit=10000000 \
        --function="unpause" \
        --recall-nonce \
        --wait-result \
        --send || return
}

setFeeMarketAddress() {
    setFeeMarketAddressCall ${ESDT_SAFE_ADDRESS} ${FEE_MARKET_ADDRESS} ${PROXY} ${CHAIN_ID}
}
setFeeMarketAddressSovereign() {
    setFeeMarketAddressCall ${ESDT_SAFE_ADDRESS_SOVEREIGN} ${FEE_MARKET_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN}
}
setFeeMarketAddressCall() {
    if [ $# -lt 3 ]; then
            echo "Usage: $0 <arg1> <arg2> <arg3>"
            exit 1
        fi

    local ESDT_SAFE_CONTRACT_ADDRESS=$1
    local FEE_MARKET_CONTRACT_ADDRESS=$2
    local URL=$3
    local CHAIN=$4

    mxpy --verbose contract call ESDT_SAFE_CONTRACT_ADDRESS \
        --pem=${WALLET} \
        --proxy=${URL} \
        --chain=${CHAIN} \
        --gas-limit=10000000 \
        --function="setFeeMarketAddress" \
        --arguments $FEE_MARKET_CONTRACT_ADDRESS \
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
