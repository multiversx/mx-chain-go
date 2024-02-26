MULTISIG_VERIFIER_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-multisig-verifier-contract)

deployMultisigVerifierContract() {
    manualUpdateConfigFile #update config file

    BLS_PUB_KEYS=$(python3 $SCRIPT_PATH/pyScripts/read_bls_keys.py)

    mxpy --verbose contract deploy \
        --bytecode="${MULTISIG_VERIFIER_WASM}" \
        --pem=${WALLET} \
        --proxy=${PROXY} \
        --chain=${CHAIN_ID} \
        --gas-limit=200000000 \
        --arguments ${BLS_PUB_KEYS} \
        --outfile="${SCRIPT_PATH}/deploy-multisig-verifier.interaction.json" \
        --recall-nonce \
        --wait-result \
        --send || return

    TX_STATUS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-multisig-verifier.interaction.json"  --expression="data['transactionOnNetwork']['status']")
    if [ "$TX_STATUS" != "success" ]; then
        echo "Transaction was not successful"
        return
    fi

    ADDRESS=$(mxpy data parse --file="${SCRIPT_PATH}/deploy-multisig-verifier.interaction.json"  --expression="data['contractAddress']")
    mxpy data store --partition=${CHAIN_ID} --key=address-multisig-verifier-contract --value=${ADDRESS}
    MULTISIG_VERIFIER_ADDRESS=$(mxpy data load --partition=${CHAIN_ID} --key=address-multisig-verifier-contract)
    echo -e "Multisig Verifier contract: ${ADDRESS}"
}