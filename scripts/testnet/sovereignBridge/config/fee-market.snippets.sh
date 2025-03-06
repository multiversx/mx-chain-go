FEE_MARKET_ADDRESS=$(mxpy data load --use-global --partition=${CHAIN_ID} --key=address-fee-market-contract)
FEE_MARKET_ADDRESS_SOVEREIGN=$(mxpy data load --use-global --partition=sovereign --key=address-fee-market-contract)

deployFeeMarketContract() {
    checkVariables ESDT_SAFE_ADDRESS || return

    echo "Deploying Fee Market contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/deploy-fee-market.interaction.json"
    mxpy contract deploy \
        --bytecode=$(eval echo ${FEE_MARKET_WASM}) \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --arguments \
            ${ESDT_SAFE_ADDRESS} \
            00 \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    local ADDRESS=$(mxpy data parse --file=${OUTFILE} --expression="data['contractAddress']")
    mxpy data store --use-global --partition=${CHAIN_ID} --key=address-fee-market-contract --value=${ADDRESS}
    FEE_MARKET_ADDRESS=$(mxpy data load --use-global --partition=${CHAIN_ID} --key=address-fee-market-contract)
    echo -e "Fee Market contract: ${ADDRESS}"

    local SOVEREIGN_CONTRACT_ADDRESS=$(computeSecondSovereignContractAddress)
    mxpy data store --use-global --partition=sovereign --key=address-fee-market-contract --value=${SOVEREIGN_CONTRACT_ADDRESS}
    FEE_MARKET_ADDRESS_SOVEREIGN=$(mxpy data load --use-global --partition=sovereign --key=address-fee-market-contract)
    echo -e "Fee Market sovereign contract: ${SOVEREIGN_CONTRACT_ADDRESS}\n"
}

upgradeFeeMarketContract() {
    echo "Upgrading Fee Market contract on main chain..."
    checkVariables FEE_MARKET_ADDRESS || return

    local OUTFILE="${OUTFILE_PATH}/upgrade-fee-market.interaction.json"
    upgradeEsdtSafeContractCall ${ESDT_SAFE_ADDRESS} ${PROXY} ${CHAIN_ID} ${OUTFILE}
}

upgradeFeeMarketContractSovereign() {
    echo "Upgrading Fee Market contract on sovereign chain..."
    checkVariables FEE_MARKET_ADDRESS_SOVEREIGN || return

    local OUTFILE="${OUTFILE_PATH}/upgrade-fee-market-sovereign.interaction.json"
    upgradeEsdtSafeContractCall ${ESDT_SAFE_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN} ${OUTFILE}
}

upgradeFeeMarketContractCall() {
    if [ $# -lt 4 ]; then
        echo "Usage: ${FUNCNAME[0]} <arg1> <arg2> <arg3> <arg4>"
        return 1
    fi

    local ADDRESS=$1
    local URL=$2
    local CHAIN=$3
    local OUTFILE=$4

    mxpy contract upgrade ${ADDRESS} \
        --bytecode=$(eval echo ${FEE_MARKET_WASM}) \
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

removeFeeInFeeMarketContract() {
    echo "Removing fee in Fee Market contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/remove-fee-fee-market-contract.interaction.json"
    removeFeeInFeeMarketContractCall ${FEE_MARKET_ADDRESS} ${PROXY} ${CHAIN_ID} ${OUTFILE}
}
removeFeeInFeeMarketContractSovereign() {
    echo "Removing fee in Fee Market contract on sovereign chain..."

    local OUTFILE="${OUTFILE_PATH}/remove-fee-fee-market-contract-sovereign.interaction.json"
    removeFeeInFeeMarketContractCall ${FEE_MARKET_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN} ${OUTFILE}
}
removeFeeInFeeMarketContractCall() {
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
        --function="removeFee" \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}

setFixedFeeMarketContract() {
    echo "Setting fixed fee in Fee Market contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/set-fixed-fee-fee-market-contract.interaction.json"
    setFixedFeeMarketContractCall ${FEE_MARKET_ADDRESS} ${PROXY} ${CHAIN_ID} ${OUTFILE}
}
setFixedFeeMarketContractSovereign() {
    echo "Setting fixed fee in Fee Market contract on sovereign chain..."

    local OUTFILE="${OUTFILE_PATH}/set-fixed-fee-fee-market-contract-sovereign.interaction.json"
    setFixedFeeMarketContractCall ${FEE_MARKET_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN} ${OUTFILE}
}
setFixedFeeMarketContractCall() {
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
        --function="setFee" \
        --arguments \
            0x010000000a53564e2d6335336461300000000901314fb370629800000000000901c9f78d2893e40000 \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}

setAnyTokenFeeMarketContract() {
    echo "Setting any token fee in Fee Market contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/set-anytoken-fee-fee-market-contract.interaction.json"
    setAnyTokenFeeMarketContractCall ${FEE_MARKET_ADDRESS} ${PROXY} ${CHAIN_ID} ${OUTFILE}
}
setAnyTokenFeeMarketContractSovereign() {
    echo "Setting any token fee in Fee Market contract on sovereign chain..."

    local OUTFILE="${OUTFILE_PATH}/set-anytoken-fee-fee-market-contract-sovereign.interaction.json"
    setAnyTokenFeeMarketContractCall ${FEE_MARKET_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN} ${OUTFILE}
}
setAnyTokenFeeMarketContractCall() {
    if [ $# -lt 4 ]; then
        echo "Usage: ${FUNCNAME[0]} <arg1> <arg2> <arg3> <arg4>"
        return 1
    fi

    local ADDRESS=$1
    local URL=$2
    local CHAIN=$3
    local OUTFILE=$4

    echo "NOT IMPLEMENTED YET"
}

# distribute all the fees to wallet
distributeFees() {
    echo "Distributing 100% fees to owner wallet on main chain..."

    local OUTFILE="${OUTFILE_PATH}/distribute-fees.interaction.json"
    mxpy contract call ${FEE_MARKET_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=20000000 \
        --function="distributeFees" \
        --arguments \
           ${WALLET_ADDRESS} \
           0x2710 \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}
