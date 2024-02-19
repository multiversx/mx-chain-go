firstSovereignContractAddress() {
    echo $(python3 $SCRIPT_PATH/pyScripts/next_contract.py $WALLET_ADDRESS 0)
}

getShardOfAddress() {
  echo $(python3 $SCRIPT_PATH/pyScripts/address_shard.py $WALLET_ADDRESS)
}

displayContracts() {
    echo $"ESDT-SAFE CHAIN[$CHAIN_ID]: $ESDT_SAFE_ADDRESS"
    echo $"ESDT-SAFE SOVEREIGN: $ESDT_SAFE_ADDRESS_SOVEREIGN"
    echo $"FEE-MARKET: $FEE_MARKET_ADDRESS"
}

setGenesisContract() {
    ESDT_SAFE_PATH="${ROOT}/${ESDT_SAFE_WASM}"

    python3 $SCRIPT_PATH/pyScripts/genesis_contract.py $ESDT_SAFE_PATH $WALLET_ADDRESS
}

updateSovereignConfig() {
    python3 $SCRIPT_PATH/pyScripts/update_toml.py $ESDT_SAFE_ADDRESS $ESDT_SAFE_ADDRESS_SOVEREIGN
}
