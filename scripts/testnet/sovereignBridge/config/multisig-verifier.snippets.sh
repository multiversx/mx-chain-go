MULTISIG_VERIFIER_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-multisig-verifier-contract)

deployMultiSigVerifierContract() {
    manualUpdateConfigFile #update config file

    echo "Deploying MultiSig Verifier contract on main chain..."

    BLS_PUB_KEYS=$(python3 $SCRIPT_PATH/pyScripts/read_bls_keys.py)

    local OUTFILE="${OUTFILE_PATH}/deploy-multisig-verifier.interaction.json"
    mxpy contract deploy \
        --bytecode=$(eval echo ${MULTISIG_VERIFIER_WASM}) \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --arguments ${BLS_PUB_KEYS} \
        --outfile=${OUTFILE} \
        --recall-nonce \
        --wait-result \
        --send || return

    printTxStatus ${OUTFILE} || return

    local ADDRESS=$(mxpy data parse --file=${OUTFILE}  --expression="data['contractAddress']")
    mxpy data store --partition=${CHAIN_ID} --key=address-multisig-verifier-contract --value=${ADDRESS}
    MULTISIG_VERIFIER_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-multisig-verifier-contract)
    echo -e "MultiSig Verifier contract: ${ADDRESS}\n"
}

upgradeMultiSigVerifierContract() {
    manualUpdateConfigFile #update config file

    echo "Upgrading MultiSig Verifier contract on main chain..."

    local OUTFILE="${OUTFILE_PATH}/upgrade-multisig-verifier.interaction.json"
    mxpy contract upgrade ${MULTISIG_VERIFIER_ADDRESS} \
        --bytecode=$(eval echo ${MULTISIG_VERIFIER_WASM}) \
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

setEsdtSafeAddressInMultiSigVerifier() {
    echo "Setting ESDT Safe address in MultiSig Verifier contract on main chain..."
    checkVariables ESDT_SAFE_ADDRESS ESDT_SAFE_ADDRESS_SOVEREIGN || return

    mxpy contract call ${MULTISIG_VERIFIER_ADDRESS} \
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
