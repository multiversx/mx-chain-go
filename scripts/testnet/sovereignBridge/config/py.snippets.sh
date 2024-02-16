firstSovereignContractAddress() {
    echo $(python3 next_contract.py $WALLET_ADDRESS 0)
}

displayContracts() {
    echo $"ESDT-SAFE CHAIN[$CHAIN_ID]: $ESDT_SAFE_ADDRESS"
    echo $"ESDT-SAFE SOVEREIGN: $ESDT_SAFE_ADDRESS_SOVEREIGN"
    echo $"FEE-MARKET: $FEE_MARKET_ADDRESS"
}

setGenesisContract() {
    ESDT_SAFE_PATH="${ROOT}/${ESDT_SAFE_WASM}"

    python3 genesis_contract.py $ESDT_SAFE_ADDRESS_SOVEREIGN $WALLET_ADDRESS
}

updateSovereignConfig() {
    python3 update_toml.py $ESDT_SAFE_ADDRESS $ESDT_SAFE_ADDRESS_SOVEREIGN
}
