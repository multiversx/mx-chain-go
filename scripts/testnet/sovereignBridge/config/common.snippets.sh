checkWalletBalanceOnMainChain() {
    local BALANCE=$(mxpy account get --address ${WALLET_ADDRESS} --proxy ${PROXY} --balance)
    if [ "$BALANCE" == "0" ]; then
        echo -e "Your wallet balance is zero on main chain"
        return 1
    fi
    return 0
}

fund() {
    echo "Getting funds in wallet on sovereign chain..."

    local OUTFILE="${OUTFILE_PATH}/fund.interaction.json"
    mxpy tx new \
       --pem="~/MultiversX/testnet/node/config/walletKey.pem" \
       --pem-index 0 \
       --proxy=${PROXY_SOVEREIGN} \
       --chain=${CHAIN_ID_SOVEREIGN} \
       --receiver=$1 \
       --value=10000000000000000000000 \
       --gas-limit=50000 \
       --outfile=${OUTFILE} \
       --recall-nonce \
       --send

    sleep 6
}

getFundsInAddressSovereign() {
    echo "Getting funds in wallet on sovereign chain..."

    local OUTFILE="${OUTFILE_PATH}/get-funds-sovereign.interaction.json"
    mxpy tx new \
        --pem="~/MultiversX/testnet/node/config/walletKey.pem" \
        --pem-index 0 \
        --proxy=${PROXY_SOVEREIGN} \
        --chain=${CHAIN_ID_SOVEREIGN} \
        --receiver=${WALLET_ADDRESS} \
        --value=100000000000000000000000 \
        --gas-limit=50000 \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --send

    sleep 6
}

gitPullAllChanges()
{
    pushd .

    # Traverse up to the parent directory of "mx-chain-go"
    while [[ ! -d "mx-chain-go" && $(pwd) != "/" ]]; do
      cd ..
    done

    # Check if we found the directory
    if [[ ! -d "mx-chain-go" ]]; then
      echo "mx-chain-go directory not found"
      popd
      return 1
    fi

    echo -e "Pulling changes for mx-chain-go..."
    cd mx-chain-go
    git pull
    cd ..

    echo -e "Pulling changes for mx-chain-deploy-go..."
    cd mx-chain-deploy-go
    git pull
    cd ..

    echo -e "Pulling changes for mx-chain-proxy-go..."
    cd mx-chain-proxy-go
    git pull
    cd ..

    echo -e "Pulling changes for mx-chain-sovereign-bridge-go..."
    cd mx-chain-sovereign-bridge-go
    git pull
    cd ..

    popd
}
