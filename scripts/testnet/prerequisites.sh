#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

# TODO adapt for Ubuntu (and systems other than Fedora)
sudo dnf install -y git golang gcc lsof jq

cd $(dirname $ELRONDDIR)
git clone https://github.com/ElrondNetwork/elrond-deploy-go.git
git clone https://github.com/ElrondNetwork/elrond-proxy-go.git

git clone https://github.com/ElrondNetwork/elrond-txgen-go.git
cd elrond-txgen-go
git checkout EN-5018/adapt-for-sc-arwen
