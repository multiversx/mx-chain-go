# These paths must be absolute
export ELRONDDIR=$(dirname $(dirname $ELRONDTESTNETSCRIPTSDIR))
export TESTNETDIR="$HOME/Work/Elrond/testnet"
export CONFIGGENERATORDIR="$(dirname $ELRONDDIR)/elrond-deploy-go/cmd/filegen"
export CONFIGGENERATOR="$CONFIGGENERATORDIR/filegen"
export NODEDIR="$ELRONDDIR/cmd/node"
export NODE="$NODEDIR/node"
export SEEDNODEDIR="$ELRONDDIR/cmd/seednode"
export SEEDNODE="$SEEDNODEDIR/seednode"
export PROXYDIR="$(dirname $ELRONDDIR)/elrond-proxy-go/cmd/proxy"
export PROXY=$PROXYDIR/proxy
export TXGENDIR="$(dirname $ELRONDDIR)/elrond-txgen-go/cmd/txgen"
export TXGEN=$TXGENDIR/txgen

export USETMUX=1
export NODETERMUI=0

export SEEDNODE_DELAY=5
export NODE_DELAY=5
export PROXY_DELAY=60

# Shard structure
export SHARDCOUNT=3
export SHARD_VALIDATORCOUNT=2
export SHARD_OBSERVERCOUNT=1
export SHARD_CONSENSUS_SIZE=1

# Metashard structure
export META_VALIDATORCOUNT=1
export META_OBSERVERCOUNT=0
export META_CONSENSUS_SIZE=1

let "total_observer_count = $SHARD_OBSERVERCOUNT * $SHARDCOUNT + $META_OBSERVERCOUNT"
export TOTAL_OBSERVERCOUNT=$total_observer_count

let "total_node_count = $SHARD_VALIDATORCOUNT * $SHARDCOUNT + $META_VALIDATORCOUNT + $TOTAL_OBSERVERCOUNT"
export TOTAL_NODECOUNT=$total_node_count

export CONSENSUS_TYPE="bls"

export MINT_VALUE="1000000000000000000000000000"

# Ports used by the Nodes
export PORT_SEEDNODE="9999"
export PORT_ORIGIN_OBSERVER="21100"
export PORT_ORIGIN_OBSERVER_REST="10000"
export PORT_ORIGIN_VALIDATOR="21500"
export PORT_ORIGIN_VALIDATOR_REST="9500"
export PORT_PROXY="7950"
export PORT_TXGEN="7951"

export P2P_SEEDNODE_ADDRESS="/ip4/127.0.0.1/tcp/$PORT_SEEDNODE/p2p/16Uiu2HAmAzokH1ozUF52Vy3RKqRfCMr9ZdNDkUQFEkXRs9DqvmKf"

export NUMACCOUNTS="100"

export REGENERATE_ACCOUNTS=1
