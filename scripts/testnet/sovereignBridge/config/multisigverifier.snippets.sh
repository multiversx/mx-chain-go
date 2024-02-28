MULTISIG_VERIFIER_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-multisig-verifier-contract)

deployMultisigVerifierContract() {
    manualUpdateConfigFile #update config file

    BLS_PUB_KEYS=$(python3 $SCRIPT_PATH/pyScripts/read_bls_keys.py)

    mxpy --verbose contract deploy \
        --bytecode=$(eval echo ${MULTISIG_VERIFIER_WASM}) \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --arguments ${BLS_PUB_KEYS} \
        --outfile="${SCRIPT_PATH}/deploy-multisig-verifier.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    local TX_STATUS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-multisig-verifier.interaction.json"  --expression="data['transactionOnNetwork']['status']")
    if [ "$TX_STATUS" != "success" ]; then
        echo "Transaction was not successful"
        return
    fi

    local ADDRESS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-multisig-verifier.interaction.json"  --expression="data['contractAddress']")
    mxpy data store --partition=${CHAIN_ID} --key=address-multisig-verifier-contract --value=${ADDRESS}
    MULTISIG_VERIFIER_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-multisig-verifier-contract)
    echo -e "Multisig Verifier contract: ${ADDRESS}"
}

upgradeMultisigVerifierContract() {
    manualUpdateConfigFile #update config file

    local BLS_PUB_KEYS=$(python3 $SCRIPT_PATH/pyScripts/read_bls_keys.py)

    mxpy --verbose contract upgrade ${MULTISIG_VERIFIER_ADDRESS} \
        --bytecode=$(eval echo ${MULTISIG_VERIFIER_WASM}) \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --outfile="${SCRIPT_PATH}/upgrade-multisig-verifier.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    local TX_STATUS=$(mxpy data parse --file="${SCRIPT_PATH}/upgrade-multisig-verifier.interaction.json"  --expression="data['transactionOnNetwork']['status']")
    if [ "$TX_STATUS" != "success" ]; then
        echo "Transaction was not successful"
        return
    fi
}
