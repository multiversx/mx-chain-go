ESDT_SAFE_ADDRESS=$(mxpy data load --use-global --partition=${CHAIN_ID} --key=address-esdt-safe-contract)
ESDT_SAFE_ADDRESS_SOVEREIGN=$(mxpy data load --use-global --partition=sovereign --key=address-esdt-safe-contract)

deployEsdtSafeContract() {
    echo "Deploying ESDT Safe contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/deploy-esdt-safe.interaction.json"
    mxpy contract deploy \
        --bytecode=$(eval echo ${MVX_ESDT_SAFE_WASM}) \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --arguments \
            ${HEADER_VERIFIER_ADDRESS} \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    local ADDRESS=$(mxpy data parse --file=${OUTFILE} --expression="data['contractAddress']")
    mxpy data store --use-global --partition=${CHAIN_ID} --key=address-esdt-safe-contract --value=${ADDRESS}
    ESDT_SAFE_ADDRESS=$(mxpy data load --use-global --partition=${CHAIN_ID} --key=address-esdt-safe-contract)
    echo -e "ESDT Safe contract: ${ADDRESS}"

    local SOVEREIGN_CONTRACT_ADDRESS=$(computeFirstSovereignContractAddress)
    mxpy data store --use-global --partition=sovereign --key=address-esdt-safe-contract --value=${SOVEREIGN_CONTRACT_ADDRESS}
    ESDT_SAFE_ADDRESS_SOVEREIGN=$(mxpy data load --use-global --partition=sovereign --key=address-esdt-safe-contract)
    echo -e "ESDT Safe sovereign contract: ${SOVEREIGN_CONTRACT_ADDRESS}\n"
}

upgradeEsdtSafeContract() {
    echo "Upgrading ESDT Safe contract on main chain..."
    checkVariables ESDT_SAFE_ADDRESS || return

    local OUTFILE="${OUTFILE_PATH}/upgrade-esdt-safe.interaction.json"
    upgradeEsdtSafeContractCall ${ESDT_SAFE_ADDRESS} ${PROXY} ${CHAIN_ID} ${OUTFILE}
}

upgradeEsdtSafeContractSovereign() {
    echo "Upgrading ESDT Safe contract on sovereign chain..."
    checkVariables ESDT_SAFE_ADDRESS_SOVEREIGN || return

    local OUTFILE="${OUTFILE_PATH}/upgrade-esdt-safe-sovereign.interaction.json"
    upgradeEsdtSafeContractCall ${ESDT_SAFE_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN} ${OUTFILE}
}

upgradeEsdtSafeContractCall() {
    if [ $# -lt 4 ]; then
        echo "Usage: ${FUNCNAME[0]} <arg1> <arg2> <arg3> <arg4>"
        return 1
    fi

    local ADDRESS=$1
    local URL=$2
    local CHAIN=$3
    local OUTFILE=$4

    mxpy contract upgrade ${ADDRESS} \
        --bytecode=$(eval echo ${MVX_ESDT_SAFE_WASM}) \
        --pem=${WALLET} \
        --proxy=${URL} \
        --chain=${CHAIN} \
        --gas-limit=200000000 \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}

pauseEsdtSafeContract() {
    echo "Pausing ESDT Safe contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/pause-esdt-safe.interaction.json"
    pauseEsdtSafeContractCall ${ESDT_SAFE_ADDRESS} ${PROXY} ${CHAIN_ID} ${OUTFILE}
}
pauseEsdtSafeContractSovereign() {
    echo "Pausing ESDT Safe contract on sovereign chain..."

    local OUTFILE="${OUTFILE_PATH}/pause-esdt-safe-sovereign.interaction.json"
    pauseEsdtSafeContractCall ${ESDT_SAFE_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN} ${OUTFILE}
}
pauseEsdtSafeContractCall() {
    if [ $# -lt 4 ]; then
        echo "Usage: ${FUNCNAME[0]} <arg1> <arg2> <arg3> <arg4>"
        return 1
    fi

    local ADDRESS=$1
    local URL=$2
    local CHAIN=$3
    local OUTFILE=$4

    mxpy contract call ${ADDRESS} \
        --pem=${WALLET} \
        --proxy=${URL} \
        --chain=${CHAIN} \
        --gas-limit=10000000 \
        --function="pause" \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}

unpauseEsdtSafeContract() {
    echo "Unpausing ESDT Safe contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/unpause-esdt-safe.interaction.json"
    unpauseEsdtSafeContractCall ${ESDT_SAFE_ADDRESS} ${PROXY} ${CHAIN_ID} ${OUTFILE}
}
unpauseEsdtSafeContractSovereign() {
    echo "Unpausing ESDT Safe contract on sovereign chain..."

    local OUTFILE="${OUTFILE_PATH}/unpause-esdt-safe-sovereign.interaction.json"
    unpauseEsdtSafeContractCall ${ESDT_SAFE_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN} ${OUTFILE}
}
unpauseEsdtSafeContractCall() {
    if [ $# -lt 4 ]; then
        echo "Usage: ${FUNCNAME[0]} <arg1> <arg2> <arg3> <arg4>"
        return 1
    fi

    local ADDRESS=$1
    local URL=$2
    local CHAIN=$3
    local OUTFILE=$4

    mxpy contract call ${ADDRESS} \
        --pem=${WALLET} \
        --proxy=${URL} \
        --chain=${CHAIN} \
        --gas-limit=10000000 \
        --function="unpause" \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}

setFeeMarketAddress() {
    echo "Setting Fee Market address in ESDT Safe contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/set-feemarket-address.interaction.json"
    setFeeMarketAddressCall ${ESDT_SAFE_ADDRESS} ${FEE_MARKET_ADDRESS} ${PROXY} ${CHAIN_ID} ${OUTFILE}
}
setFeeMarketAddressSovereign() {
    echo "Setting Fee Market address in ESDT Safe contract on sovereign chain..."

    local OUTFILE="${OUTFILE_PATH}/set-feemarket-address-sovereign.interaction.json"
    setFeeMarketAddressCall ${ESDT_SAFE_ADDRESS_SOVEREIGN} ${FEE_MARKET_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN} ${OUTFILE}
}
setFeeMarketAddressCall() {
    if [ $# -lt 5 ]; then
        echo "Usage: ${FUNCNAME[0]} <arg1> <arg2> <arg3> <arg4> <arg5>"
        return 1
    fi

    local ESDT_SAFE_CONTRACT_ADDRESS=$1
    local FEE_MARKET_CONTRACT_ADDRESS=$2
    local URL=$3
    local CHAIN=$4
    local OUTFILE=$5

    mxpy contract call ${ESDT_SAFE_CONTRACT_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${URL} \
        --chain=${CHAIN} \
        --gas-limit=10000000 \
        --function="setFeeMarketAddress" \
        --arguments ${FEE_MARKET_CONTRACT_ADDRESS} \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}
