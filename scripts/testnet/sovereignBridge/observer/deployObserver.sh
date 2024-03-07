IMAGE_NAME="multiversx/sov-observer"

prepareObserver() {
    manualUpdateConfigFile #update config file

    local DOCKER_IMAGE=""

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

    local LINE="FROM $DOCKER_IMAGE"
    sed -i "1s,.*,${LINE}," "$SCRIPT_PATH/observer/shard-observer" # replace first line with the docker image

    docker image build . -t $IMAGE_NAME -f $SCRIPT_PATH/observer/shard-observer
}

deployObserver() {
    local SHARD=$(getShardOfAddress)

    docker run -d -p 8083:8080 -p 22111:22111 $IMAGE_NAME --destination-shard-as-observer=$SHARD
}

stopObserver() {
    local CONTAINER_IDS=$(docker ps -q --filter "ancestor=${IMAGE_NAME}")

    if [ -z "$CONTAINER_IDS" ]; then
        echo "No containers running based on image ${IMAGE_NAME}"
        exit 1
    fi

    for CONTAINER_ID in $CONTAINER_IDS; do
        docker stop $CONTAINER_ID
    done
}
