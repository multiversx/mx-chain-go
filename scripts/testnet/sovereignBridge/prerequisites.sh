#!/bin/bash

runPrerequisites() {
   installSoftwareDependenciesl

    downloadCrossChainContracts
}

installSoftwareDependencies() {
    apt update
    apt install -y python3-pip pipx screen tmux ca-certificates wget
    pip install multiversx-sdk
    pipx install multiversx-sdk-cli --force
    pipx ensurepath
    export PATH="$PATH:/root/.local/bin"

    # install docker-ce
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    chmod a+r /etc/apt/keyrings/docker.asc
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" > /etc/apt/sources.list.d/docker.list
    apt update
    apt install -y docker-ce
}

downloadCrossChainContracts() {
    mkdir -p $(eval echo "${CONTRACTS_DIRECTORY}")
    version=$(basename `curl -s https://github.com/multiversx/mx-sovereign-sc/releases/latest -I | grep location | awk -F"https:/" '{print $2}' | tr -d "\r"`)
    wget -O $(eval echo ${ESDT_SAFE_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/esdt-safe.wasm
    wget -O $(eval echo ${FEE_MARKET_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/fee-market.wasm
    wget -O $(eval echo ${MULTISIG_VERIFIER_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/multisigverifier.wasm
}