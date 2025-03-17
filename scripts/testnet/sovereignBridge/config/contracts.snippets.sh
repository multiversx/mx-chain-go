downloadCrossChainContracts() {
    echo "Downloading cross-chain contracts..."

    mkdir -p $(eval echo "${CONTRACTS_DIRECTORY}")
    version=$(basename `curl -s https://github.com/multiversx/mx-sovereign-sc/releases/latest -I | grep location | awk -F"https:/" '{print $2}' | tr -d "\r"`)
    wget -O $(eval echo ${MVX_ESDT_SAFE_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/mvx-esdt-safe.wasm
    wget -O $(eval echo ${SOV_ESDT_SAFE_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/sov-esdt-safe.wasm
    wget -O $(eval echo ${FEE_MARKET_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/fee-market.wasm
    wget -O $(eval echo ${HEADER_VERIFIER_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/header-verifier.wasm
}
