source "$ELRONDTESTNETSCRIPTSDIR/variables.sh"

prepareFolders() {
  [ -d $TESTNETDIR ] || mkdir -p $TESTNETDIR
  cd $TESTNETDIR
  [ -d filegen ] || mkdir -p filegen
  [ -d node ] || mkdir -p node
  [ -d node/config ] || mkdir -p node/config
  [ -d seednode ] || mkdir -p seednode
  [ -d seednode/config ] || mkdir -p seednode/config
  [ -d node_working_dirs ] || mkdir -p node_working_dirs
}

prepareFolders_PrivateRepos() {
  [ -d $TESTNETDIR ] || mkdir -p $TESTNETDIR
  cd $TESTNETDIR
  [ -d proxy ] || mkdir -p proxy
  [ -d ./proxy/config ] || mkdir -p ./proxy/config
  [ -d txgen ] || mkdir -p txgen
  [ -d ./txgen/config ] || mkdir -p ./txgen/config
}

buildConfigGenerator() {
  echo "Building Configuration Generator..."
  pushd $CONFIGGENERATORDIR
  go build .
  popd

  pushd $TESTNETDIR
  mv $CONFIGGENERATOR ./filegen/
  echo "Configuration Generator built..."
  popd
}


buildNode() {
  echo "Building Node executable..."
  pushd $NODEDIR
  go build -gcflags="all=-N -l" .
  popd

  pushd $TESTNETDIR
  mv $NODE ./node/
  echo "Node executable built."
  popd
}

buildSeednode() {
  echo "Building Seednode executable..."
  pushd $SEEDNODEDIR
  go build .
  popd

  pushd $TESTNETDIR
  mv $SEEDNODE ./seednode/
  echo "Seednode executable built."
  popd
}

buildProxy() {
  echo "Building Proxy executable..."
  pushd $PROXYDIR
  go build .
  popd

  pushd $TESTNETDIR
  mv $PROXY ./proxy/
  echo "Proxy executable built."
  popd
}

buildTxGen() {
  echo "Building TxGen executable..."
  pushd $TXGENDIR
  go build .
  popd

  pushd $TESTNETDIR
  mv $TXGEN ./txgen/
  echo "TxGen executable built."
  popd
}
