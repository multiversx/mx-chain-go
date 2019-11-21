source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

prepareFolders() {
  [ -d $TESTNETDIR ] || mkdir -p $TESTNETDIR
  cd $TESTNETDIR
  [ -d node_working_dirs ] || mkdir -p node_working_dirs
  [ -d node ] || mkdir -p node
  [ -d node/config ] || mkdir -p node/config
  [ -d seednode ] || mkdir -p seednode
  [ -d seednode/config ] || mkdir -p seednode/config
  [ -d proxy ] || mkdir -p proxy
  [ -d ./proxy/config ] || mkdir -p ./proxy/config
  [ -d txgen ] || mkdir -p txgen
  [ -d ./txgen/config ] || mkdir -p ./txgen/config
}


buildNode() {
  echo "Building Node executable..."
  cd $NODEDIR
  go build .

  cd $TESTNETDIR
  cp $NODE ./node/
  echo "Node executable built."
}

buildSeednode() {
  echo "Building Seednode executable..."
  cd $SEEDNODEDIR
  go build .

  cd $TESTNETDIR
  cp $SEEDNODE ./seednode/
  echo "Seednode executable built."
}

buildProxy() {
  echo "Building Proxy executable..."
  cd $PROXYDIR
  go build .

  cd $TESTNETDIR
  cp $PROXY ./proxy/
  echo "Proxy executable built."
}

buildTxGen() {
  echo "Building TxGen executable..."
  cd $TXGENDIR
  go build .

  cd $TESTNETDIR
  cp $TXGEN ./txgen/
  echo "TxGen executable built."
}
