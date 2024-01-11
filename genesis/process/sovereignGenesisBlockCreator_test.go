//go:build !race

package process

import (
	"encoding/hex"
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	nodeMock "github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func createGenesisBlockCreator(t *testing.T) *genesisBlockCreator {
	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	gbc, _ := NewGenesisBlockCreator(arg)
	return gbc
}

func createSovereignGenesisBlockCreator(t *testing.T) GenesisBlockCreatorHandler {
	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.ShardCoordinator = sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	arg.DNSV2Addresses = []string{"00000000000000000500761b8c4a25d3979359223208b412285f635e71300102"}
	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreator(gbc)
	return sgbc
}

func TestNewSovereignGenesisBlockCreator(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		gbc := createGenesisBlockCreator(t)
		sgbc, err := NewSovereignGenesisBlockCreator(gbc)
		require.Nil(t, err)
		require.NotNil(t, sgbc)
	})

	t.Run("nil genesis block creator, should return error", func(t *testing.T) {
		t.Parallel()

		sgbc, err := NewSovereignGenesisBlockCreator(nil)
		require.Equal(t, errNilGenesisBlockCreator, err)
		require.Nil(t, sgbc)
	})
}

func TestSovereignGenesisBlockCreator_CreateGenesisBlocksEmptyBlocks(t *testing.T) {
	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.StartEpochNum = 1
	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreator(gbc)

	blocks, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)
	require.Equal(t, map[uint32]data.HeaderHandler{
		core.SovereignChainShardId: &block.SovereignChainHeader{
			Header: &block.Header{
				ShardID: core.SovereignChainShardId,
			},
		},
	}, blocks)
}

func TestSovereignGenesisBlockCreator_CreateGenesisBaseProcess(t *testing.T) {
	sgbc := createSovereignGenesisBlockCreator(t)

	blocks, err := sgbc.CreateGenesisBlocks()
	require.Nil(t, err)
	require.Len(t, blocks, 1)
	require.Contains(t, blocks, core.SovereignChainShardId)

	indexingData := sgbc.GetIndexingData()
	require.Len(t, indexingData, 1)

	numDNSTypeScTxs := 2 * 256 // there are 2 contracts in testdata/smartcontracts.json
	numDefaultTypeScTxs := 1
	reqNumDeployInitialScTxs := numDNSTypeScTxs + numDefaultTypeScTxs

	sovereignIdxData := indexingData[core.SovereignChainShardId]
	require.Len(t, sovereignIdxData.DeployInitialScTxs, reqNumDeployInitialScTxs)
	require.Len(t, sovereignIdxData.DeploySystemScTxs, 4)
	require.Len(t, sovereignIdxData.DelegationTxs, 3)
	require.Len(t, sovereignIdxData.StakingTxs, 0)
	require.Len(t, sovereignIdxData.ScrsTxs, getRequiredNumScrsTxs(indexingData, core.SovereignChainShardId))
}

func TestSovereignGenesisBlockCreator_setSovereignStakedData(t *testing.T) {
	t.Parallel()

	args := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))

	acc := &state.AccountWrapMock{
		Balance: big.NewInt(1),
	}
	acc.IncreaseNonce(4)
	args.Accounts = &state.AccountsStub{
		LoadAccountCalled: func(addr []byte) (vmcommon.AccountHandler, error) {
			return acc, nil
		},
	}
	initialNode := &sharding.InitialNode{
		Address: "addr",
	}
	nodesSpliter := &mock.NodesListSplitterStub{
		GetAllNodesCalled: func() []nodesCoordinator.GenesisNodeInfoHandler {
			return []nodesCoordinator.GenesisNodeInfoHandler{initialNode}
		},
	}
	expectedTx := &transaction.Transaction{
		Nonce:     acc.GetNonce(),
		Value:     new(big.Int).Set(args.GenesisNodePrice),
		RcvAddr:   vm.ValidatorSCAddress,
		SndAddr:   initialNode.AddressBytes(),
		GasPrice:  0,
		GasLimit:  math.MaxUint64,
		Data:      []byte("stake@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString(initialNode.PubKeyBytes()) + "@" + hex.EncodeToString([]byte("genesis"))),
		Signature: nil,
	}
	processors := &genesisProcessors{
		txProcessor: &testscommon.TxProcessorStub{
			ProcessTransactionCalled: func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
				require.Equal(t, expectedTx, transaction)

				return vmcommon.Ok, nil
			},
		},
		queryService: &nodeMock.SCQueryServiceStub{
			ExecuteQueryCalled: func(query *process.SCQuery) (*vmcommon.VMOutput, common.BlockInfo, error) {
				require.Equal(t, &process.SCQuery{
					ScAddress: vm.StakingSCAddress,
					FuncName:  "isStaked",
					Arguments: [][]byte{initialNode.PubKeyBytes()}}, query)

				return &vmcommon.VMOutput{ReturnCode: vmcommon.Ok}, holders.NewBlockInfo(nil, 0, nil), nil
			},
		},
	}

	txs, err := setSovereignStakedData(args, processors, nodesSpliter)
	require.Nil(t, err)
	require.Equal(t, []data.TransactionHandler{expectedTx}, txs)
}

func TestSovereignGenesisBlockCreator_InitSystemAccountCalled(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.ShardCoordinator = sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	arg.DNSV2Addresses = []string{"00000000000000000500761b8c4a25d3979359223208b412285f635e71300102"}

	loadAccountWasCalled := false
	saveAccountWasCalled := false
	accountsAdapter := &state.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			if !loadAccountWasCalled { // first loaded account
				loadAccountWasCalled = true
				require.Equal(t, core.SystemAccountAddress, address)
			}
			return state.NewAccountWrapMock(address), nil
		},
		SaveAccountCalled: func(account vmcommon.AccountHandler) error {
			if !saveAccountWasCalled { // first saved account
				saveAccountWasCalled = true
				require.Equal(t, core.SystemAccountAddress, account.AddressBytes())
			}
			return nil
		},
	}
	arg.Accounts = accountsAdapter

	gbc, _ := NewGenesisBlockCreator(arg)
	sgbc, _ := NewSovereignGenesisBlockCreator(gbc)
	_, err := sgbc.CreateGenesisBlocks()
	require.NotNil(t, err)

	require.NotNil(t, sgbc)
	require.True(t, loadAccountWasCalled)
	require.True(t, saveAccountWasCalled)
}
