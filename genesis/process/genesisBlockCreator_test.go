package process

import (
	"math/big"
	"testing"

	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/config"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/genesis/parsing"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//TODO improve code coverage of this package
func createMockArgument() ArgsGenesisBlockCreator {
	memDBMock := mock.NewMemDbMock()
	storageManager := &mock.StorageManagerStub{DatabaseCalled: func() data.DBWriteCacher {
		return memDBMock
	}}

	trieStorageManagers := make(map[string]data.StorageManager)
	trieStorageManagers[factory.UserAccountTrie] = storageManager
	trieStorageManagers[factory.PeerAccountTrie] = storageManager

	arg := ArgsGenesisBlockCreator{
		GenesisTime:   0,
		StartEpochNum: 0,
		Core: &mock.CoreComponentsMock{
			IntMarsh:            &mock.MarshalizerMock{},
			Hash:                &mock.HasherMock{},
			UInt64ByteSliceConv: &mock.Uint64ByteSliceConverterMock{},
			AddrPubKeyConv:      mock.NewPubkeyConverterMock(32),
			Chain:               "chainID",
		},
		Data: &mock.DataComponentsMock{
			Storage: &mock.ChainStorerMock{
				GetStorerCalled: func(unitType dataRetriever.UnitType) storage.Storer {
					return mock.NewStorerMock()
				},
			},
			Blkc:     &mock.BlockChainStub{},
			DataPool: mock.NewPoolsHolderMock(),
		},
		InitialNodesSetup:    &mock.InitialNodesSetupHandlerStub{},
		TxLogsProcessor:      &mock.TxLogProcessorMock{},
		VirtualMachineConfig: config.VirtualMachineConfig{},
		HardForkConfig:       config.HardforkConfig{},
		SystemSCConfig: config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "5000000000000000000000",
				OwnerAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
			},
		},
		TrieStorageManagers: trieStorageManagers,
	}

	arg.ShardCoordinator = &mock.ShardCoordinatorMock{
		NumOfShards: 2,
		SelfShardId: 0,
	}

	arg.Accounts = &mock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return make([]byte, 0), nil
		},
		CommitCalled: func() ([]byte, error) {
			return make([]byte, 0), nil
		},
		SaveAccountCalled: func(account state.AccountHandler) error {
			return nil
		},
		LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return state.NewEmptyUserAccount(), nil
		},
	}

	arg.ValidatorAccounts = &mock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return make([]byte, 0), nil
		},
		CommitCalled: func() ([]byte, error) {
			return make([]byte, 0), nil
		},
		SaveAccountCalled: func(account state.AccountHandler) error {
			return nil
		},
		LoadAccountCalled: func(address []byte) (state.AccountHandler, error) {
			return state.NewEmptyPeerAccount(), nil
		},
	}

	arg.GasMap = arwenConfig.MakeGasMap(1)
	defaults.FillGasMapInternal(arg.GasMap, 1)

	ted := &economics.TestEconomicsData{
		EconomicsData: &economics.EconomicsData{},
	}
	ted.SetGenesisNodePrice(big.NewInt(100))
	ted.SetMinStep(big.NewInt(1))
	ted.SetTotalSupply(big.NewInt(10000))
	ted.SetUnJailPrice(big.NewInt(1))
	arg.Economics = ted.EconomicsData
	arg.AccountsParser, _ = parsing.NewAccountsParser(
		"testdata/genesis.json",
		arg.Economics.TotalSupply(),
		arg.Core.AddressPubKeyConverter(),
	)

	arg.SmartContractParser, _ = parsing.NewSmartContractsParser(
		"testdata/smartcontracts.json",
		arg.Core.AddressPubKeyConverter(),
	)

	return arg
}

func TestGenesisBlockCreator_CreateGenesisBlocksShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	blocks, err := gbc.CreateGenesisBlocks()

	assert.Nil(t, err)
	assert.Equal(t, 3, len(blocks))
}
