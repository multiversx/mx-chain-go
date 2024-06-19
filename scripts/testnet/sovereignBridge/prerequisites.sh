#!/bin/bash

sudo apt update
sudo apt install -y python3-pip pipx screen ca-certificates curl wget
pip install multiversx-sdk
pipx install multiversx-sdk-cli --force
pipx ensurepath

# install docker
if ! command -v docker &> /dev/null
then
  sudo apt-get update
  sudo apt-get install ca-certificates curl
  sudo install -m 0755 -d /etc/apt/keyrings
  sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  sudo chmod a+r /etc/apt/keyrings/docker.asc
  # Add the repository to Apt sources:
  echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
    $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
    sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
  sudo apt-get update
  sudo apt install -y docker-ce
  # docker post install
  sudo groupadd docker
  sudo usermod -aG docker $USER
  newgrp docker
fi

# load config paths
source config/configs.cfg

mkdir -p $(eval echo "${CONTRACTS_DIRECTORY}")
version=$(basename `curl -s https://github.com/multiversx/mx-sovereign-sc/releases/latest -I | grep location | awk -F"https:/" '{print $2}' | tr -d "\r"`)
wget -O $(eval echo ${ESDT_SAFE_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/esdt-safe.wasm
wget -O $(eval echo ${FEE_MARKET_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/fee-market.wasm
wget -O $(eval echo ${HEADER_VERIFIER_WASM}) https://github.com/multiversx/mx-sovereign-sc/releases/download/${version}/header-verifier.wasm
