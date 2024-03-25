FEE_MARKET_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-fee-market-contract)
FEE_MARKET_ADDRESS_SOVEREIGN=$(mxpy data load --partition=sovereign --key=address-fee-market-contract)

deployFeeMarketContract() {
    CHECK_VARIABLES ESDT_SAFE_ADDRESS || return

    mxpy --verbose contract deploy \
        --bytecode=$(eval echo ${FEE_MARKET_WASM}) \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --arguments \
            ${ESDT_SAFE_ADDRESS} \
            ${PRICE_AGGREGATOR_ADDRESS} \
        --outfile="${SCRIPT_PATH}/deploy-fee-market.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    local TX_STATUS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-fee-market.interaction.json"  --expression="data['transactionOnNetwork']['status']")
    if [ "$TX_STATUS" != "success" ]; then
        echo "Transaction was not successful"
        return
    fi

    local ADDRESS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-fee-market.interaction.json" --expression="data['contractAddress']")
    mxpy data store --partition=${CHAIN_ID} --key=address-fee-market-contract --value=${ADDRESS}
    FEE_MARKET_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-fee-market-contract)
    echo -e "Fee Market contract: ${ADDRESS}"

    local SOVEREIGN_CONTRACT_ADDRESS=$(secondSovereignContractAddress)
    mxpy data store --partition=sovereign --key=address-fee-market-contract --value=${SOVEREIGN_CONTRACT_ADDRESS}
    FEE_MARKET_ADDRESS_SOVEREIGN=$(mxpy data load --partition=sovereign --key=address-fee-market-contract)
    echo -e "Fee Market sovereign contract: ${SOVEREIGN_CONTRACT_ADDRESS}"
}

upgradeFeeMarketContract() {
    CHECK_VARIABLES ESDT_SAFE_ADDRESS || return

    mxpy --verbose contract upgrade ${FEE_MARKET_ADDRESS} \
        --bytecode=$(eval echo ${FEE_MARKET_WASM}) \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --outfile="${SCRIPT_PATH}/upgrade-fee-market.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    local TX_STATUS=$(mxpy data parse --file="${SCRIPT_PATH}/upgrade-fee-market.interaction.json"  --expression="data['transactionOnNetwork']['status']")
    if [ "$TX_STATUS" != "success" ]; then
        echo "Transaction was not successful"
        return
    fi
}

enableFeeMarketContract() {
    enableFeeMarketContractCall ${FEE_MARKET_ADDRESS} ${PROXY} ${CHAIN_ID}
}
enableFeeMarketContractSovereign() {
    enableFeeMarketContractCall ${FEE_MARKET_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN}
}
enableFeeMarketContractCall() {
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
        --function="enableFee" \
        --recall-nonce \
        --wait-result \
        --send || return
}

disableFeeMarketContract() {
    disableFeeMarketContractCall ${FEE_MARKET_ADDRESS} ${PROXY} ${CHAIN_ID}
}
disableFeeMarketContractSovereign() {
    disableFeeMarketContractCall ${FEE_MARKET_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN}
}
disableFeeMarketContractCall() {
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
        --function="disableFee" \
        --recall-nonce \
        --wait-result \
        --send || return
}

setFixedFeeMarketContract() {
    setFixedFeeMarketContractCall ${FEE_MARKET_ADDRESS} ${PROXY} ${CHAIN_ID}
}
setFixedFeeMarketContractSovereign() {
    setFixedFeeMarketContractCall ${FEE_MARKET_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN}
}
setFixedFeeMarketContractCall() {
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
            --function="addFee" \
            --arguments \
                str:SVN-c53da0 \
                0x010000000a53564e2d6335336461300000000901314fb370629800000000000901c9f78d2893e40000 \
            --recall-nonce \
            --wait-result \
            --send || return
}

setAnyTokenFeeMarketContract() {
    setAnyTokenFeeMarketContractCall ${FEE_MARKET_ADDRESS} ${PROXY} ${CHAIN_ID}
}
setAnyTokenFeeMarketContractSovereign() {
    setAnyTokenFeeMarketContractCall ${FEE_MARKET_ADDRESS_SOVEREIGN} ${PROXY_SOVEREIGN} ${CHAIN_ID_SOVEREIGN}
}
setAnyTokenFeeMarketContractCall() {
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
            --function="addFee" \
            --arguments \
                str:SVN-cb685a \
                0x020000000c5745474c442d64643834373100000008de0b6b3a7640000000000008ebec21ee1da40000 \
            --recall-nonce \
            --wait-result \
            --send || return
}