#!/bin/bash

echo "Installing software prerequisites..."
sudo apt update
sudo apt install -y python3-pip pipx screen ca-certificates curl wget
pip install multiversx-sdk
pipx install multiversx-sdk-cli --force
pipx ensurepath

source config/configs.cfg
source config/contracts.snippets.sh
downloadCrossChainContracts

# install docker
if ! command -v docker &> /dev/null
then
  echo "Installing Docker..."
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
  sudo usermod -aG docker $USER
  newgrp docker
fi
