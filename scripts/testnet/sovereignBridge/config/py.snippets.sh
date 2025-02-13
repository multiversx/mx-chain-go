computeFirstSovereignContractAddress() {
    echo $(python3 $SCRIPT_PATH/pyScripts/compute_contract_address.py $WALLET_ADDRESS 0)
}

computeSecondSovereignContractAddress() {
    echo $(python3 $SCRIPT_PATH/pyScripts/compute_contract_address.py $WALLET_ADDRESS 1)
}

getShardOfAddress() {
  echo $(python3 $SCRIPT_PATH/pyScripts/address_shard.py $WALLET_ADDRESS)
}

bech32ToHex() {
  echo $(python3 $SCRIPT_PATH/pyScripts/address_convert.py $1)
}

displayContracts() {
    echo "ESDT-SAFE: $ESDT_SAFE_ADDRESS"
    echo "ESDT-SAFE SOVEREIGN: $ESDT_SAFE_ADDRESS_SOVEREIGN"
    echo "FEE-MARKET: $FEE_MARKET_ADDRESS"
    echo "FEE-MARKET SOVEREIGN: $FEE_MARKET_ADDRESS_SOVEREIGN"
    echo "HEADER-VERIFIER: $HEADER_VERIFIER_ADDRESS"
}

updateAndStartBridgeService() {
    python3 $SCRIPT_PATH/pyScripts/bridge_service.py $WALLET $PROXY $ESDT_SAFE_ADDRESS $HEADER_VERIFIER_ADDRESS
}

updateNotifierNotarizationRound() {
    python3 $SCRIPT_PATH/pyScripts/notifier_round.py $PROXY $(getShardOfAddress)
}

setGenesisContract() {
    local ESDT_SAFE_INIT_PARAMS="01"
    local FEE_MARKET_INIT_PARAMS="$(bech32ToHex $ESDT_SAFE_ADDRESS_SOVEREIGN)@00"

    python3 $SCRIPT_PATH/pyScripts/genesis_contract.py $WALLET_ADDRESS $ESDT_SAFE_WASM $ESDT_SAFE_INIT_PARAMS $FEE_MARKET_WASM $FEE_MARKET_INIT_PARAMS
}

updateSovereignConfig() {
    if [ -z "$1" ]; then
        ESDT_PREFIX=$(generateRandomEsdtPrefix)
    else
        ESDT_PREFIX=$1
    fi

    python3 $SCRIPT_PATH/pyScripts/update_toml.py $ESDT_SAFE_ADDRESS $ESDT_SAFE_ADDRESS_SOVEREIGN $ESDT_PREFIX $MAIN_CHAIN_ELASTIC
}

generateRandomEsdtPrefix() {
  LEN=$(shuf -i 1-4 -n 1)
  RANDOM_PREFIX=$(cat /dev/urandom | tr -dc 'a-z0-9' | head -c $LEN)

  echo $RANDOM_PREFIX
}
