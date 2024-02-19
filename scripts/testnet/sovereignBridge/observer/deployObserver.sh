prepareObserver() {
    manualUpdateConfigFile #update config file

    if [ -n "$1" ]; then
        DOCKER_IMAGE=$1
    else
        case $CHAIN_ID in
            "1")
                TAG=$(curl -s https://registry.hub.docker.com/v2/repositories/multiversx/chain-mainnet/tags | jq -r '.results[0].name')
                DOCKER_IMAGE="multiversx/chain-mainnet:$TAG"
                ;;
            "D")
                TAG=$(curl -s https://registry.hub.docker.com/v2/repositories/multiversx/chain-devnet/tags | jq -r '.results[0].name')
                DOCKER_IMAGE="multiversx/chain-devnet:$TAG"
                ;;
            "T")
                TAG=$(curl -s https://registry.hub.docker.com/v2/repositories/multiversx/chain-testnet/tags | jq -r '.results[0].name')
                DOCKER_IMAGE="multiversx/chain-testnet:$TAG"
                ;;
        esac
    fi

    LINE="FROM $DOCKER_IMAGE"
    sed -i "1s,.*,${LINE}," "$SCRIPT_PATH/observer/shard-observer" # replace first line with the docker image

    docker image build . -t multiversx/sov-observer -f $SCRIPT_PATH/observer/shard-observer
}

deployObserver() {
    SHARD=$(getShardOfAddress)

    docker run -d -p 8083:8080 -p 22111 multiversx/sov-observer --destination-shard-as-observer $SHARD
}