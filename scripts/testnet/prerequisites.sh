#!/usr/bin/env bash

export MULTIVERSXTESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
source "$MULTIVERSXTESTNETSCRIPTSDIR/variables.sh"

export DISTRIBUTION=$(cat /etc/os-release | grep "^ID=" | sed 's/ID=//')


if [[ "$DISTRIBUTION" =~ ^(fedora|centos|rhel)$ ]]; then
  export PACKAGE_MANAGER="dnf"
  export REQUIRED_PACKAGES="git golang gcc lsof jq curl"
  export INSTALL_PACKAGES_COMMAND="sudo $PACKAGE_MANAGER install -y $REQUIRED_PACKAGES"
elif [[ "$DISTRIBUTION" =~ ^(ubuntu|debian)$ ]]; then
  export PACKAGE_MANAGER="apt-get"
  export REQUIRED_PACKAGES="git gcc lsof jq curl"
  export INSTALL_PACKAGES_COMMAND="sudo $PACKAGE_MANAGER install -y $REQUIRED_PACKAGES"
elif [[ "$DISTRIBUTION" =~ ^(arch)$ ]]; then
  export PACKAGE_MANAGER="pacman"
  export REQUIRED_PACKAGES="git gcc lsof jq curl"
  export INSTALL_PACKAGES_COMMAND="sudo $PACKAGE_MANAGER -S $REQUIRED_PACKAGES"
fi

if [ -z "$INSTALL_PACKAGES_COMMAND" ]; then
  echo "Your operating system was not identified. The required packages were not installed, however they could still be part of your system."
  read -r -p "Would you like to try to continue anyway? [Y/n] " response
  if [[ "$response" =~ ^[Nn][Oo]?$ ]]; then
    exit
  fi
else
  echo "Using $PACKAGE_MANAGER to install required packages: $REQUIRED_PACKAGES"
  $INSTALL_PACKAGES_COMMAND
fi

if [[ "$DISTRIBUTION" =~ ^(ubuntu|debian)$ ]]; then

  if ! [ -x "$(command -v go)" ]; then
    echo "Installing Go..."
    GO_LATEST=$(curl -sS https://golang.org/VERSION?m=text) 
    wget https://dl.google.com/go/$GO_LATEST.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf $GO_LATEST.linux-amd64.tar.gz
    rm $GO_LATEST.linux-amd64.tar.gz

    export GOROOT="/usr/local/go"
    export GOBIN="$HOME/go/bin"
    export PATH=$PATH:$GOROOT/bin:$GOBIN
    mkdir -p $GOBIN

    echo "export GOROOT=/usr/local/go" >> ~/.profile
    echo "export GOBIN=$HOME/go/bin" >> ~/.profile
    echo "export PATH=$PATH:$GOROOT/bin:$GOBIN" >> ~/.profile
    source ~/.profile 
  fi
fi


cd $(dirname $MULTIVERSXDIR)
git clone git@github.com:multiversx/mx-chain-deploy-go.git
git clone git@github.com:multiversx/mx-chain-proxy-go.git

if [ "$SOVEREIGN_DEPLOY" = true ]; then
    git clone https://github.com/multiversx/mx-chain-sovereign-bridge-go.git
    cd mx-chain-sovereign-bridge-go
    git checkout eb39c56a1539
    cd cert/cmd/cert
    go build
    ./cert
    cd ../../../../
fi

if [[ $PRIVATE_REPOS -eq 1 ]]; then
  git clone git@github.com:multiversx/mx-chain-txgen-go.git
  cd mx-chain-txgen-go
  git checkout master
fi
