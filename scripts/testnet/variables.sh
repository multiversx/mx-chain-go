# These paths must be absolute

# Path to elrond-go. Determined automatically. Do not change.
export ELRONDDIR=$(dirname $(dirname $ELRONDTESTNETSCRIPTSDIR))

# Path where the testnet will be instantiated. This folder is assumed to not
# exist, but it doesn't matter if it already does. It will be created if not,
# anyway.
export TESTNETDIR="$HOME/Work/Elrond/testnet"

# Path to elrond-deploy-go, branch: master. Default: near elrond-go.
export CONFIGGENERATORDIR="$(dirname $ELRONDDIR)/elrond-deploy-go/cmd/filegen"
export CONFIGGENERATOR="$CONFIGGENERATORDIR/filegen"    # Leave unchanged.

# Path to the executable node. Leave unchanged unless well justified.
export NODEDIR="$ELRONDDIR/cmd/node"
export NODE="$NODEDIR/node"     # Leave unchanged

# Path to the executable seednode. Leave unchanged unless well justified.
export SEEDNODEDIR="$ELRONDDIR/cmd/seednode"
export SEEDNODE="$SEEDNODEDIR/seednode"   # Leave unchanged.

# Path to elrond-proxy-go, branch: master. Default: near elrond-go.
export PROXYDIR="$(dirname $ELRONDDIR)/elrond-proxy-go/cmd/proxy"
export PROXY=$PROXYDIR/proxy    # Leave unchanged.

# Path to elrond-txgen-go, branch: EN-5018/adapt-for-sc-arwen (will change eventually). Default: near elrond-go.
export TXGENDIR="$(dirname $ELRONDDIR)/elrond-txgen-go/cmd/txgen"
export TXGEN=$TXGENDIR/txgen    # Leave unchanged.

# Use tmux or not. If set to 1, only 2 terminal windows will be opened, and
# tmux will be used to display the running executables using split windows.
# Recommended. Tmux needs to be installed.
export USETMUX=1

# Start Nodes with TermUI or not. Looks good with TermUI, but if you want full
# info and saved logs, set this to 0. TermUI can't save logs.
export NODETERMUI=1

# Log level for the logger in the Node.
export LOGLEVEL="*:DEBUG"

# Delays after running executables.
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

# Leave unchanged.
let "total_observer_count = $SHARD_OBSERVERCOUNT * $SHARDCOUNT + $META_OBSERVERCOUNT"
export TOTAL_OBSERVERCOUNT=$total_observer_count

# Leave unchanged.
let "total_node_count = $SHARD_VALIDATORCOUNT * $SHARDCOUNT + $META_VALIDATORCOUNT + $TOTAL_OBSERVERCOUNT"
export TOTAL_NODECOUNT=$total_node_count

# Okay as defaults, change if needed.
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

# Number of accounts to be generated by txgen
export NUMACCOUNTS="100"

# Whether txgen should regenerate its accounts when starting, or not.
# Recommended value is 1, but 0 is useful to run the txgen a second time, to
# continue a testing session on the same accounts.
export REGENERATE_ACCOUNTS=1

if [ "$TESTNETMODE" == "debug" ]; then
  NODETERMUI=0
  USETMUX=1
  LOGLEVEL="*:DEBUG"
fi
