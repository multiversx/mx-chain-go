firstSovereignContractAddress() {
    echo $(python3 $SCRIPT_PATH/pyScripts/next_contract.py $WALLET_ADDRESS 0)
}

secondSovereignContractAddress() {
    echo $(python3 $SCRIPT_PATH/pyScripts/next_contract.py $WALLET_ADDRESS 1)
}

getShardOfAddress() {
  echo $(python3 $SCRIPT_PATH/pyScripts/address_shard.py $WALLET_ADDRESS)
}

displayContracts() {
    echo $"ESDT-SAFE CHAIN[$CHAIN_ID]: $ESDT_SAFE_ADDRESS"
    echo $"ESDT-SAFE SOVEREIGN: $ESDT_SAFE_ADDRESS_SOVEREIGN"
    echo $"FEE-MARKET: $FEE_MARKET_ADDRESS"
    echo $"MULTISIG-VERIFIER: $MULTISIG_VERIFIER_ADDRESS"
}

setGenesisContract() {
    ESDT_SAFE_INIT_PARAMS="@@d08355c11965f045b765085f0e39849447d59efd7c5222f265f2b9b5f8d8f654"
    FEE_MARKET_INIT_PARAMS="@d08355c11965f045b765085f0e39849447d59efd7c5222f265f2b9b5f8d8f654@000000000000000005004c13819a7f26de997e7c6720a6efe2d4b85c0609c9ad"

    python3 $SCRIPT_PATH/pyScripts/genesis_contract.py $WALLET_ADDRESS $ESDT_SAFE_WASM $ESDT_SAFE_INIT_PARAMS $FEE_MARKET_WASM $FEE_MARKET_INIT_PARAMS
}

updateSovereignConfig() {
    python3 $SCRIPT_PATH/pyScripts/update_toml.py $ESDT_SAFE_ADDRESS $ESDT_SAFE_ADDRESS_SOVEREIGN
}
