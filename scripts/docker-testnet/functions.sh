#!/usr/bin/env bash

IP_BIT=3

cloneRepositories() {
  if [[ -n $CI_RUN ]]; then
    echo "Repositories have been cloned in the CI"
  else
    cd $(dirname $MULTIVERSXDIR)

    git clone git@github.com:multiversx/mx-chain-deploy-go.git || true
    git clone git@github.com:multiversx/mx-chain-proxy-go.git || true
  fi
}

startSeedNode() {
  docker run -d --name seednode -v ${TESTNETDIR}/seednode/config:/go/mx-chain-go/cmd/seednode/config seednode:dev \
  --rest-api-interface=0.0.0.0:10000
}

startObservers() {
 local observerIdx=0
 # Example for loop with injected variables in Bash
 for ((i = 0; i < SHARDCOUNT; i++)); do
   for ((j = 0; j < SHARD_OBSERVERCOUNT; j++)); do
     # Your commands or code to be executed in each iteration
     KEY_INDEX=$((TOTAL_NODECOUNT - observerIdx - 1))

     docker run -d --name "observer${observerIdx}-172.17.0.${IP_BIT}-10200-shard${i}" \
     -v $TESTNETDIR/node/config:/go/mx-chain-go/cmd/node/config \
     node:dev \
     --destination-shard-as-observer $i \
     --rest-api-interface=0.0.0.0:10200 \
     --config ./config/config_observer.toml \
     --sk-index=${KEY_INDEX} \
     $EXTRA_OBSERVERS_FLAGS


     (( IP_BIT++ ))
     ((observerIdx++)) || true
   done
 done

 for ((i = 0; i < META_OBSERVERCOUNT; i++)); do
    KEY_INDEX=$((TOTAL_NODECOUNT - observerIdx - 1))

    docker run -d --name "observer${observerIdx}-172.17.0.${IP_BIT}-10200-metachain" \
        -v $TESTNETDIR/node/config:/go/mx-chain-go/cmd/node/config \
        node:dev \
        --destination-shard-as-observer "metachain" \
        --rest-api-interface=0.0.0.0:10200 \
        --config ./config/config_observer.toml \
        --sk-index=${KEY_INDEX} \
        $EXTRA_OBSERVERS_FLAGS

    (( IP_BIT++ ))
    ((observerIdx++)) || true
 done
}

startValidators() {
 validatorIdx=0
 # Example for loop with injected variables in Bash
 for ((i = 0; i < SHARDCOUNT; i++)); do
   for ((j = 0; j < SHARD_VALIDATORCOUNT; j++)); do

     docker run -d --name "validator${validatorIdx}-172.17.0.${IP_BIT}-10200-shard${i}" \
     -v $TESTNETDIR/node/config:/go/mx-chain-go/cmd/node/config \
     node:dev \
     --rest-api-interface=0.0.0.0:10200 \
     --config ./config/config_validator.toml \
     --sk-index=${validatorIdx} \

     (( IP_BIT++ ))
     ((validatorIdx++)) || true
   done
 done

  for ((i = 0; i < META_VALIDATORCOUNT; i++)); do
     docker run -d --name "validator${validatorIdx}-172.17.0.${IP_BIT}-10200-metachain" \
         -v $TESTNETDIR/node/config:/go/mx-chain-go/cmd/node/config \
         node:dev \
         --rest-api-interface=0.0.0.0:10200 \
         --config ./config/config_observer.toml \
         --sk-index=${validatorIdx} \

     (( IP_BIT++ ))
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
  local ipBit=3
  OUTPUTFILE=$!


  for ((i = 0; i < SHARDCOUNT; i++)); do
     for ((j = 0; j < SHARD_OBSERVERCOUNT; j++)); do

        echo "[[Observers]]" >> config_edit.toml
        echo "   ShardId = $i" >> config_edit.toml
        echo "   Address = \"http://172.17.0.${ipBit}:10200\"" >> config_edit.toml
        echo ""$'\n' >> config_edit.toml

        (( ipBit++ )) || true
     done
  done

  for META_OBSERVER in $(seq $META_OBSERVERCOUNT); do
        echo "[[Observers]]" >> config_edit.toml
        echo "   ShardId = $METASHARD_ID" >> config_edit.toml
        echo "   Address = \"http://172.17.0.${ipBit}:10200\"" >> config_edit.toml
        echo ""$'\n' >> config_edit.toml

        (( ipBit++ )) || true
  done
}

generateProxyValidatorListDocker() {
  local ipBit=3
  OUTPUTFILE=$!


  for ((i = 0; i < SHARDCOUNT; i++)); do
     for ((j = 0; j < SHARD_VALIDATORCOUNT; j++)); do

        echo "[[Observers]]" >> config_edit.toml
        echo "   ShardId = $i" >> config_edit.toml
        echo "   Address = \"http://172.17.0.${ipBit}:10200\"" >> config_edit.toml
        echo "   Type = \"Validator\"" >> config_edit.toml
        echo ""$'\n' >> config_edit.toml

        (( ipBit++ )) || true
     done
  done

  for META_OBSERVER in $(seq $META_VALIDATORCOUNT); do
        echo "[[Observers]]" >> config_edit.toml
        echo "   ShardId = $METASHARD_ID" >> config_edit.toml
        echo "   Address = \"http://172.17.0.${ipBit}:10200\"" >> config_edit.toml
        echo "   Type = \"Validator\"" >> config_edit.toml
        echo ""$'\n' >> config_edit.toml

        (( ipBit++ )) || true
  done
}

startProxyDocker() {
  docker run -d --name "proxy" \
           -p $PORT_PROXY:8080 \
           -v $TESTNETDIR/proxy/config:/mx-chain-proxy-go/cmd/proxy/config \
           multiversx/chain-proxy:v1.1.45-sp4
}
