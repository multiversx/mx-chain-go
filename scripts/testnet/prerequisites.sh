#!/usr/bin/env bash

export ELRONDTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

export DISTRIBUTION=$(cat /etc/os-release | grep "^ID=" | sed 's/ID=//')


if [[ "$DISTRIBUTION" =~ ^(fedora|centos|rhel)$ ]]; then
  export PACKAGE_MANAGER="dnf"
  export REQUIRED_PACKAGES="git golang gcc lsof jq curl"
  echo "Using DNF to install required packages: $REQUIRED_PACKAGES"
fi

if [[ "$DISTRIBUTION" =~ ^(ubuntu|debian)$ ]]; then
  export PACKAGE_MANAGER="apt-get"
  export REQUIRED_PACKAGES="git gcc lsof jq curl"

  echo "Using APT to install required packages: $REQUIRED_PACKAGES"
fi

sudo $PACKAGE_MANAGER install -y $REQUIRED_PACKAGES

if [[ "$DISTRIBUTION" =~ ^(ubuntu|debian)$ ]]; then
  echo "Installing Go..."
  GO_LATEST=$(curl -sS https://golang.org/VERSION?m=text) 
  wget https://dl.google.com/go/$GO_LATEST.linux-amd64.tar.gz
  sudo tar -C /usr/local -xzf $GO_LATEST.linux-amd64.tar.gz
  rm $GO_LATEST.linux-amd64.tar.gz

  echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile
  mkdir -p "$HOME/go"
  echo "export GOPATH=$HOME/go" >> ~/.profile
  source ~/.profile 
fi


cd $(dirname $ELRONDDIR)
git clone https://github.com/ElrondNetwork/elrond-deploy-go.git
git clone https://github.com/ElrondNetwork/elrond-proxy-go.git

git clone https://github.com/ElrondNetwork/elrond-txgen-go.git
cd elrond-txgen-go
git checkout EN-5018/adapt-for-sc-arwen
