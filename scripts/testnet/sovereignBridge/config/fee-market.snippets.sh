FEE_MARKET_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-fee-market-contract)
FEE_MARKET_ADDRESS_SOVEREIGN=$(mxpy data load --partition=sovereign --key=address-fee-market-contract)

deployFeeMarketContract() {
    CHECK_VARIABLES ESDT_SAFE_ADDRESS || return

    mxpy --verbose contract deploy \
        --bytecode="${FEE_MARKET_WASM}" \
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

    TX_STATUS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-fee-market.interaction.json"  --expression="data['transactionOnNetwork']['status']")
    if [ "$TX_STATUS" != "success" ]; then
        echo "Transaction was not successful"
        return
    fi

    ADDRESS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-fee-market.interaction.json" --expression="data['contractAddress']")
    mxpy data store --partition=${CHAIN_ID} --key=address-fee-market-contract --value=${ADDRESS}
    FEE_MARKET_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-fee-market-contract)
    echo -e "Fee Market contract: ${ADDRESS}"

    SOVEREIGN_CONTRACT_ADDRESS=$(secondSovereignContractAddress)
    mxpy data store --partition=sovereign --key=address-fee-market-contract --value=${SOVEREIGN_CONTRACT_ADDRESS}
    FEE_MARKET_ADDRESS_SOVEREIGN=$(mxpy data load --partition=sovereign --key=address-fee-market-contract)
    echo -e "Fee Market sovereign contract: ${SOVEREIGN_CONTRACT_ADDRESS}"
}

upgradeFeeMarketContract() {
    CHECK_VARIABLES ESDT_SAFE_ADDRESS || return

    mxpy --verbose contract upgrade ${FEE_MARKET_ADDRESS} \
        --bytecode="${FEE_MARKET_WASM}" \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --arguments \
            ${ESDT_SAFE_ADDRESS} \
            ${PRICE_AGGREGATOR_ADDRESS} \
        --outfile="${SCRIPT_PATH}/upgrade-fee-market.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    TX_STATUS=$(mxpy data parse --file="${SCRIPT_PATH}/upgrade-fee-market.interaction.json"  --expression="data['transactionOnNetwork']['status']")
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

    ADDRESS=$1
    URL=$2
    CHAIN=$3

    mxpy --verbose contract call $ADDRESS \
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

    ADDRESS=$1
    URL=$2
    CHAIN=$3

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
