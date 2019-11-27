#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# TODO adapt for Ubuntu (and systems other than Fedora)
sudo dnf install -y git golang gcc

cd $ELRONDDIR
git clone https://github.com/ElrondNetwork/elrond-txgen-go.git
git clone https://github.com/ElrondNetwork/elrond-proxy-go.git
