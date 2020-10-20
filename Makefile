build:
	go build ./...

build-cmd:
	(cd cmd/node && go build)

clean-test:
	go clean -testcache ./...

clean: clean-test
	go clean -cache ./...
	go clean ./...

test: clean-test
	go test ./...

test-serial: clean-test
	go test -p 1 ./...

test-short:
	go test -short -count=1 ./...

test-short-v:
	go test -short -v -count=1 ./...

test-race:
	go test -short -race -count=1 ./...

test-memp2p-v:
	go test -v -count=1 ./p2p/memp2p

test-consensusBLS-memp2p:
	go test -count=1 --run=TestConsensusBLSFullTest ./integrationTests/consensus_memp2p/

test-consensusBLS-memp2p-v:
	go test -v -count=1 --run=TestConsensusBLSFullTest ./integrationTests/consensus_memp2p/

test-consensusBLS-v:
	go test -v -count=1 --run=TestConsensusBLSFullTest ./integrationTests/consensus/

test-consensusBLS:
	go test -count=1 --run=TestConsensusBLSFullTest ./integrationTests/consensus/

test-miniblocks-memp2p-v:
	go test -count=1 -v --run=TestShouldProcessBlocksInMultiShardArchitecture_withMemP2P ./integrationTests/multiShard/block_memp2p/

test-miniblocks-v:
	go test -count=1 -v --run=TestShouldProcessBlocksInMultiShardArchitecture ./integrationTests/multiShard/block/

test-agario-join-reward:
	go test -count=1 -v --run=TestShouldProcessBlocksWithScTxsJoinAndReward ./integrationTests/singleShard/block/

test-miniblocks-sc-v:
	go test -count=1 -v ./integrationTests/multiShard/block/executingMiniblocksSc_test.go

test-arwen:
	go test -count=1 -v ./integrationTests/vm/arwen/...

test-multishard-sc:
	go test -count=1 -v ./integrationTests/multiShard/smartContract

benchmark-arwen:
	go test -v -count=1 -test.bench 'Benchmark_VmDeployWithFibbonacciAndExecute' -test.run='noruns' ./integrationTests/vm/arwen
	go test -v -count=1 -test.bench 'Benchmark_VmDeployWithCPUCalculateAndExecute' -test.run='noruns' ./integrationTests/vm/arwen
	go test -v -count=1 -test.bench 'Benchmark_VmDeployWithStringConcatAndExecute' -test.run='noruns' ./integrationTests/vm/arwen

arwen:
ifndef ARWEN_PATH
	$(error ARWEN_PATH is undefined)
endif
	# When referencing a non-release version, add the commit hash, like this:
	go get github.com/ElrondNetwork/arwen-wasm-vm/cmd/arwen@f5d1cc592d3c1171e7952833670e892a629aa35a
	# When referencing a released version, use this instead:
	#go get github.com/ElrondNetwork/arwen-wasm-vm/cmd/arwen@$(shell cat go.mod | grep arwen-wasm-vm | sed 's/.* //')
	go build -o ${ARWEN_PATH} github.com/ElrondNetwork/arwen-wasm-vm/cmd/arwen
	stat ${ARWEN_PATH}

cli-docs:
	cd ./cmd && bash ./CLI.md.sh
