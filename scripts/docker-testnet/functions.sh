#!/usr/bin/env bash

# Starts from 3, if the DOCKER_NETWORK_SUBNET ends with a 0. The first IP address is reserved for the gateway and the
# second one is allocated to the seednode. Therefore the counting starts from 3. If you modify the DOCKER_NETWORK_SUBNET
# variable, make sure to change this one accordingly too.
IP_HOST_BYTE=3

cloneRepositories() {
  cd $(dirname $MULTIVERSXDIR)

  git clone git@github.com:multiversx/mx-chain-deploy-go.git || true
  git clone git@github.com:multiversx/mx-chain-proxy-go.git || true
}

buildNodeImages() {
  cd $MULTIVERSXDIR

  docker build -f docker/seednode/Dockerfile . -t seednode:dev

  docker build -f docker/node/Dockerfile . -t node:dev
}

createDockerNetwork() {
  docker network create -d bridge --subnet=${DOCKER_NETWORK_SUBNET} ${DOCKER_NETWORK_NAME}

  # this variable is used to keep track of the allocated IP addresses in the network, by removing the last byte
  # of the DOCKER_NETWORK_SUBNET. One can consider this the host network address without the last byte at the end.
  export NETWORK_ADDRESS=$(echo "$DOCKER_NETWORK_SUBNET" | rev | cut -d. -f2- | rev)
}

startSeedNode() {
  docker run -d --name seednode \
      -v ${TESTNETDIR}/seednode/config:/go/mx-chain-go/cmd/seednode/config \
      --network ${DOCKER_NETWORK_NAME} \
      seednode:dev \
      --rest-api-interface=0.0.0.0:10000
}

startObservers() {
 local observerIdx=0
 # Example for loop with injected variables in Bash
 for ((i = 0; i < SHARDCOUNT; i++)); do
   for ((j = 0; j < SHARD_OBSERVERCOUNT; j++)); do
     # Your commands or code to be executed in each iteration
     KEY_INDEX=$((TOTAL_NODECOUNT - observerIdx - 1))

     docker run -d --name "observer${observerIdx}-${NETWORK_ADDRESS}.${IP_HOST_BYTE}-10200-shard${i}" \
         -v $TESTNETDIR/node/config:/go/mx-chain-go/cmd/node/config \
         --network ${DOCKER_NETWORK_NAME} \
         node:dev \
         --destination-shard-as-observer $i \
         --rest-api-interface=0.0.0.0:10200 \
         --config ./config/config_observer.toml \
         --sk-index=${KEY_INDEX} \
         $EXTRA_OBSERVERS_FLAGS


     (( IP_HOST_BYTE++ ))
     ((observerIdx++)) || true
   done
 done

 for ((i = 0; i < META_OBSERVERCOUNT; i++)); do
    KEY_INDEX=$((TOTAL_NODECOUNT - observerIdx - 1))

    docker run -d --name "observer${observerIdx}-${NETWORK_ADDRESS}.${IP_HOST_BYTE}-10200-metachain" \
        -v $TESTNETDIR/node/config:/go/mx-chain-go/cmd/node/config \
        --network ${DOCKER_NETWORK_NAME} \
        node:dev \
        --destination-shard-as-observer "metachain" \
        --rest-api-interface=0.0.0.0:10200 \
        --config ./config/config_observer.toml \
        --sk-index=${KEY_INDEX} \
        $EXTRA_OBSERVERS_FLAGS

    (( IP_HOST_BYTE++ ))
    ((observerIdx++)) || true
 done
}

startValidators() {
 validatorIdx=0
 # Example for loop with injected variables in Bash
 for ((i = 0; i < SHARDCOUNT; i++)); do
   for ((j = 0; j < SHARD_VALIDATORCOUNT; j++)); do

     docker run -d --name "validator${validatorIdx}-${NETWORK_ADDRESS}.${IP_HOST_BYTE}-10200-shard${i}" \
         -v $TESTNETDIR/node/config:/go/mx-chain-go/cmd/node/config \
         --network ${DOCKER_NETWORK_NAME} \
         node:dev \
         --rest-api-interface=0.0.0.0:10200 \
         --config ./config/config_validator.toml \
         --sk-index=${validatorIdx} \

     (( IP_HOST_BYTE++ ))
     ((validatorIdx++)) || true
   done
 done

  for ((i = 0; i < META_VALIDATORCOUNT; i++)); do
     docker run -d --name "validator${validatorIdx}-${NETWORK_ADDRESS}.${IP_HOST_BYTE}-10200-metachain" \
         -v $TESTNETDIR/node/config:/go/mx-chain-go/cmd/node/config \
         --network ${DOCKER_NETWORK_NAME} \
         node:dev \
         --rest-api-interface=0.0.0.0:10200 \
         --config ./config/config_observer.toml \
         --sk-index=${validatorIdx} \

     (( IP_HOST_BYTE++ ))
     ((validatorIdx++)) || true
  done
}

updateProxyConfigDocker() {
    pushd $TESTNETDIR/proxy/config
    cp config.toml config_edit.toml

    # Truncate config.toml before the [[Observers]] list
    sed -i -n '/\[\[Observers\]\]/q;p' config_edit.toml

    if [ "$SHARD_OBSERVERCOUNT" -le 0 ]; then
        generateProxyValidatorListDocker config_edit.toml
    else
        generateProxyObserverListDocker config_edit.toml
    fi

    mv config_edit.toml config.toml

    echo "Updated configuration for the Proxy."
    popd
}

generateProxyObserverListDocker() {
  local ipByte=3

  for ((i = 0; i < SHARDCOUNT; i++)); do
     for ((j = 0; j < SHARD_OBSERVERCOUNT; j++)); do

        echo "[[Observers]]" >> config_edit.toml
        echo "   ShardId = $i" >> config_edit.toml
        echo "   Address = \"http://${NETWORK_ADDRESS}.${ipByte}:10200\"" >> config_edit.toml
        echo ""$'\n' >> config_edit.toml

        (( ipByte++ )) || true
     done
  done

  for META_OBSERVER in $(seq $META_OBSERVERCOUNT); do
        echo "[[Observers]]" >> config_edit.toml
        echo "   ShardId = $METASHARD_ID" >> config_edit.toml
        echo "   Address = \"http://${NETWORK_ADDRESS}.${ipByte}:10200\"" >> config_edit.toml
        echo ""$'\n' >> config_edit.toml

        (( ipByte++ )) || true
  done
}

generateProxyValidatorListDocker() {
  local ipByte=3

  for ((i = 0; i < SHARDCOUNT; i++)); do
     for ((j = 0; j < SHARD_VALIDATORCOUNT; j++)); do

        echo "[[Observers]]" >> config_edit.toml
        echo "   ShardId = $i" >> config_edit.toml
        echo "   Address = \"http://${NETWORK_ADDRESS}.${ipByte}:10200\"" >> config_edit.toml
        echo "   Type = \"Validator\"" >> config_edit.toml
        echo ""$'\n' >> config_edit.toml

        (( ipByte++ )) || true
     done
  done

  for META_OBSERVER in $(seq $META_VALIDATORCOUNT); do
        echo "[[Observers]]" >> config_edit.toml
        echo "   ShardId = $METASHARD_ID" >> config_edit.toml
        echo "   Address = \"http://${NETWORK_ADDRESS}.${ipByte}:10200\"" >> config_edit.toml
        echo "   Type = \"Validator\"" >> config_edit.toml
        echo ""$'\n' >> config_edit.toml

        (( ipByte++ )) || true
  done
}

buildProxyImage() {
  pushd ${PROXYDIR}
  cd ../..
  docker build -f docker/Dockerfile . -t proxy:dev
}

startProxyDocker() {
  docker run -d --name "proxy" \
      -v $TESTNETDIR/proxy/config:/mx-chain-proxy-go/cmd/proxy/config \
      --network ${DOCKER_NETWORK_NAME} \
      -p $PORT_PROXY:8080 \
      proxy:dev
}
