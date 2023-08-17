CURRENT_DIRECTORY := $(shell pwd)
TESTS_TO_RUN := $(shell go list ./... | grep -v /integrationTests/ | grep -v /testscommon/ | grep -v mock | grep -v disabled | grep -v defaults)

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

test-v: clean-test
	go test -v ./...

test-serial: clean-test
	go test -p 1 ./...

test-short:
	go test -short -count=1 ./...

test-short-v:
	go test -short -v -count=1 ./...

test-race:
	go test -short -race -v ./...

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

test-wasm:
	go test -count=1 -v ./integrationTests/vm/wasm/...

test-coverage:
	@echo "Running unit tests"
	CURRENT_DIRECTORY=$(CURRENT_DIRECTORY) go test -short -cover -coverprofile=coverage.txt -covermode=atomic -v ${TESTS_TO_RUN}

test-multishard-sc:
	go test -count=1 -v ./integrationTests/multiShard/smartContract

benchmark-wasm:
	go test -v -count=1 -test.bench 'Benchmark_VmDeployWithFibbonacciAndExecute' -test.run='noruns' ./integrationTests/vm/wasm
	go test -v -count=1 -test.bench 'Benchmark_VmDeployWithCPUCalculateAndExecute' -test.run='noruns' ./integrationTests/vm/wasm
	go test -v -count=1 -test.bench 'Benchmark_VmDeployWithStringConcatAndExecute' -test.run='noruns' ./integrationTests/vm/wasm

cli-docs:
	cd ./cmd && bash ./CLI.md.sh

lint-install:
ifeq (,$(wildcard test -f bin/golangci-lint))
	@echo "Installing golint"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s
endif

run-lint:
	@echo "Running golint"
	bin/golangci-lint run --max-issues-per-linter 0 --max-same-issues 0 --timeout=2m

lint: lint-install run-lint

install-proto:
	bash ./install-proto.sh

check-cli-md:
	cd ./cmd/assessment && go build
	cd ./cmd/keygenerator && go build
	cd ./cmd/logviewer && go build
	cd ./cmd/node && go build
	cd ./cmd/seednode && go build
	cd ./cmd/termui && go build
	cd ./cmd && bash ./CLI.md.sh
	@status=$$(git status --porcelain | grep CLI); \
    	if [ ! -z "$${status}" ]; \
    	then \
    		echo "Error - please update all CLI.md files by running the 'cli-docs' or 'check-cli-md' from Makefile!"; \
    		exit 1; \
    	fi
