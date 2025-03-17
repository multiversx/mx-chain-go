HEADER_VERIFIER_ADDRESS=$(mxpy data load --use-global --partition=${CHAIN_ID} --key=address-header-verifier-contract)

deployHeaderVerifierContract() {
    manualUpdateConfigFile #update config file

    echo "Deploying Header Verifier contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/deploy-header-verifier.interaction.json"
    mxpy contract deploy \
        --bytecode=$(eval echo ${HEADER_VERIFIER_WASM}) \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    local ADDRESS=$(mxpy data parse --file=${OUTFILE}  --expression="data['contractAddress']")
    mxpy data store --use-global --partition=${CHAIN_ID} --key=address-header-verifier-contract --value=${ADDRESS}
    HEADER_VERIFIER_ADDRESS=$(mxpy data load --use-global --partition=${CHAIN_ID} --key=address-header-verifier-contract)
    echo -e "Header Verifier contract: ${ADDRESS}\n"
}

upgradeHeaderVerifierContract() {
    manualUpdateConfigFile #update config file

    echo "Upgrading Header Verifier contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/upgrade-header-verifier.interaction.json"
    mxpy contract upgrade ${HEADER_VERIFIER_ADDRESS} \
        --bytecode=$(eval echo ${HEADER_VERIFIER_WASM}) \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}

registerBlsPubKeysInHeaderVerifier() {
    echo "Setting BLS pub keys in Header Verifier contract on main chain..."
    checkVariables HEADER_VERIFIER_ADDRESS || return

    BLS_PUB_KEYS=$(python3 $SCRIPT_PATH/pyScripts/read_bls_keys.py)

    local OUTFILE="${OUTFILE_PATH}/register-bls-keys.interaction.json"
    mxpy contract call ${HEADER_VERIFIER_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=10000000 \
        --function="registerBlsPubKeys" \
        --arguments ${BLS_PUB_KEYS} \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}

setEsdtSafeAddressInHeaderVerifier() {
    echo "Setting ESDT Safe address in Header Verifier contract on main chain..."
    checkVariables HEADER_VERIFIER_ADDRESS ESDT_SAFE_ADDRESS || return

    local OUTFILE="${OUTFILE_PATH}/set-esdtsafe-address.interaction.json"
    mxpy contract call ${HEADER_VERIFIER_ADDRESS} \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=10000000 \
        --function="setEsdtSafeAddress" \
        --arguments ${ESDT_SAFE_ADDRESS} \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE}
}
