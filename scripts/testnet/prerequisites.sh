#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

export DISTRIBUTION=$(cat /etc/os-release | grep "^ID=" | sed 's/ID=//')

export REQUIRED_PACKAGES="git golang gcc lsof jq"

if [[ "$DISTRIBUTION" =~ ^(fedora|centos|rhel)$ ]]; then
  export PACKAGE_MANAGER="dnf"
  echo "Using DNF to install required packages: $REQUIRED_PACKAGES"
fi

if [[ "$DISTRIBUTION" =~ ^(ubuntu|debian)$ ]]; then
  export PACKAGE_MANAGER="apt-get"
  echo "Using APT to install required packages: $REQUIRED_PACKAGES"
fi

sudo $PACKAGE_MANAGER install -y $REQUIRED_PACKAGES

cd $(dirname $ELRONDDIR)
git clone https://github.com/ElrondNetwork/elrond-deploy-go.git
git clone https://github.com/ElrondNetwork/elrond-proxy-go.git

git clone https://github.com/ElrondNetwork/elrond-txgen-go.git
cd elrond-txgen-go
git checkout EN-5018/adapt-for-sc-arwen
