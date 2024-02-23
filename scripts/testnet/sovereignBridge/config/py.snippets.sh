firstSovereignContractAddress() {
    echo $(python3 $SCRIPT_PATH/pyScripts/next_contract.py $WALLET_ADDRESS 0)
}

secondSovereignContractAddress() {
    echo $(python3 $SCRIPT_PATH/pyScripts/next_contract.py $WALLET_ADDRESS 1)
}

getShardOfAddress() {
  echo $(python3 $SCRIPT_PATH/pyScripts/address_shard.py $WALLET_ADDRESS)
}

bech32ToHex() {
  echo $(python3 $SCRIPT_PATH/pyScripts/address_convert.py $1)
}

displayContracts() {
    echo $"ESDT-SAFE CHAIN[$CHAIN_ID]: $ESDT_SAFE_ADDRESS"
    echo $"ESDT-SAFE SOVEREIGN: $ESDT_SAFE_ADDRESS_SOVEREIGN"
    echo $"FEE-MARKET: $FEE_MARKET_ADDRESS"
    echo $"MULTISIG-VERIFIER: $MULTISIG_VERIFIER_ADDRESS"
}

setGenesisContract() {
    setGenesisContractOperation ${ESDT_SAFE_ADDRESS}
}

setGenesisContractSovereign() {
    setGenesisContractOperation ${ESDT_SAFE_ADDRESS_SOVEREIGN}
}

setGenesisContractOperation() {
    if [ $# -eq 0 ]; then
        echo "No arguments provided"
        return
    fi

    if [ "$MIN_VALID_SIGNERS" = "0" ]; then
        SIGNERS=""
    else
        SIGNERS=$MIN_VALID_SIGNERS
    fi
    ESDT_SAFE_INIT_PARAMS="${SIGNERS}@$(bech32ToHex $INITIATOR_ADDRESS)"
    FEE_MARKET_INIT_PARAMS="$(bech32ToHex $1)@$(bech32ToHex $PRICE_AGGREGATOR_ADDRESS)"

    python3 $SCRIPT_PATH/pyScripts/genesis_contract.py $WALLET_ADDRESS $ESDT_SAFE_WASM $ESDT_SAFE_INIT_PARAMS $FEE_MARKET_WASM $FEE_MARKET_INIT_PARAMS
}

updateSovereignConfig() {
    python3 $SCRIPT_PATH/pyScripts/update_toml.py $ESDT_SAFE_ADDRESS $ESDT_SAFE_ADDRESS_SOVEREIGN
}
