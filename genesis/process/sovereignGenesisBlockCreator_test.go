package process

import (
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	factoryState "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageCommon "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/stretchr/testify/require"
)

func createSovereignMockArgument(
	t *testing.T,
	genesisFilename string,
	initialNodes genesis.InitialNodesHandler,
	entireSupply *big.Int,
) ArgsGenesisBlockCreator {

	storageManagerArgs := storageCommon.GetStorageManagerArgs()
	storageManager, _ := trie.CreateTrieStorageManager(storageManagerArgs, storageCommon.GetStorageManagerOptions())

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[dataRetriever.UserAccountsUnit.String()] = storageManager
	trieStorageManagers[dataRetriever.PeerAccountsUnit.String()] = storageManager

	arg := ArgsGenesisBlockCreator{
		GenesisTime:   0,
		StartEpochNum: 0,
		Core: &mock.CoreComponentsMock{
			IntMarsh:                 &mock.MarshalizerMock{},
			TxMarsh:                  &mock.MarshalizerMock{},
			Hash:                     &hashingMocks.HasherMock{},
			UInt64ByteSliceConv:      &mock.Uint64ByteSliceConverterMock{},
			AddrPubKeyConv:           testscommon.NewPubkeyConverterMock(32),
			Chain:                    "chainID",
			TxVersionCheck:           &testscommon.TxVersionCheckerStub{},
			MinTxVersion:             1,
			EnableEpochsHandlerField: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		},
		Data: &mock.DataComponentsMock{
			Storage: &storageCommon.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
					return genericMocks.NewStorerMock(), nil
				},
			},
			Blkc:     &testscommon.ChainHandlerStub{},
			DataPool: dataRetrieverMock.NewPoolsHolderMock(),
		},
		InitialNodesSetup: &mock.InitialNodesSetupHandlerStub{},
		TxLogsProcessor:   &mock.TxLogProcessorMock{},
		VirtualMachineConfig: config.VirtualMachineConfig{
			WasmVMVersions: []config.WasmVMVersionByEpoch{
				{StartEpoch: 0, Version: "*"},
			},
		},
		HardForkConfig: config.HardforkConfig{
			ImportKeysStorageConfig: config.StorageConfig{
				Cache: config.CacheConfig{
					Type:     "LRU",
					Capacity: 1000,
					Shards:   1,
				},
				DB: config.DBConfig{
					Type:              "MemoryDB",
					BatchDelaySeconds: 1,
					MaxBatchSize:      1,
					MaxOpenFiles:      10,
				},
			},
			ImportStateStorageConfig: config.StorageConfig{
				Cache: config.CacheConfig{
					Type:     "LRU",
					Capacity: 1000,
					Shards:   1,
				},
				DB: config.DBConfig{
					Type:              "MemoryDB",
					BatchDelaySeconds: 1,
					MaxBatchSize:      1,
					MaxOpenFiles:      10,
				},
			},
		},
		SystemSCConfig: config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "5000000000000000000000",
				OwnerAddress:    "erd1932eft30w753xyvme8d49qejgkjc09n5e49w4mwdjtm0neld797su0dlxp",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				V1: config.GovernanceSystemSCConfigV1{
					ProposalCost: "500",
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
					LostProposalFee:  "1",
				},
				OwnerAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     big.NewInt(5000).Text(10),
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             10,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "100",
				MinStakeAmount:      "100",
				ConfigChangeAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 100,
			},
		},
		TrieStorageManagers: trieStorageManagers,
		BlockSignKeyGen:     &mock.KeyGenMock{},
		GenesisNodePrice:    big.NewInt(5000),
		EpochConfig: &config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				BuiltInFunctionsEnableEpoch:    0,
				SCDeployEnableEpoch:            0,
				RelayedTransactionsEnableEpoch: 0,
				PenalizedTooMuchGasEnableEpoch: 0,
			},
		},
		RoundConfig: &config.RoundConfig{
			RoundActivations: map[string]config.ActivationRoundByName{
				"DisableAsyncCallV1": {
					Round: "18446744073709551615",
				},
			},
		},
		ChainRunType:            common.ChainRunTypeRegular,
		ShardCoordinatorFactory: sharding.NewMultiShardCoordinatorFactory(),
	}

	arg.ShardCoordinator = &mock.ShardCoordinatorMock{
		NumOfShards: 2,
		SelfShardId: 0,
	}

	argsAccCreator := state.ArgsAccountCreation{
		Hasher:              &hashingMocks.HasherMock{},
		Marshaller:          &mock.MarshalizerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	accCreator, err := factoryState.NewAccountCreator(argsAccCreator)
	require.Nil(t, err)

	arg.Accounts, err = createAccountAdapter(
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		accCreator,
		trieStorageManagers[dataRetriever.UserAccountsUnit.String()],
		&testscommon.PubkeyConverterMock{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)
	require.Nil(t, err)

	arg.ValidatorAccounts = &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return make([]byte, 0), nil
		},
		CommitCalled: func() ([]byte, error) {
			return make([]byte, 0), nil
		},
		SaveAccountCalled: func(account vmcommon.AccountHandler) error {
			return nil
		},
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return state.NewEmptyPeerAccount(), nil
		},
	}

	gasMap := wasmConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasMap, 1)
	arg.GasSchedule = testscommon.NewGasScheduleNotifierMock(gasMap)
	ted := &economicsmocks.EconomicsHandlerStub{
		GenesisTotalSupplyCalled: func() *big.Int {
			return entireSupply
		},
		MaxGasLimitPerBlockCalled: func(shardID uint32) uint64 {
			return math.MaxUint64
		},
	}
	arg.Economics = ted

	args := genesis.AccountsParserArgs{
		GenesisFilePath: genesisFilename,
		EntireSupply:    arg.Economics.GenesisTotalSupply(),
		MinterAddress:   "",
		PubkeyConverter: arg.Core.AddressPubKeyConverter(),
		KeyGenerator:    &mock.KeyGeneratorStub{},
		Hasher:          &hashingMocks.HasherMock{},
		Marshalizer:     &mock.MarshalizerMock{},
	}

	arg.AccountsParser, err = parsing.NewAccountsParser(args)
	require.Nil(t, err)

	arg.SmartContractParser, err = parsing.NewSmartContractsParser(
		"testdata/smartcontracts.json",
		arg.Core.AddressPubKeyConverter(),
		&mock.KeyGeneratorStub{},
	)
	require.Nil(t, err)

	arg.InitialNodesSetup = initialNodes

	return arg
}

func createGenesisBlockCreator(t *testing.T) *genesisBlockCreator {
	arg := createSovereignMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	gbc, _ := NewGenesisBlockCreator(arg)
	return gbc
}

func createSovereignGenesisBlockCreator(t *testing.T) GenesisBlockCreatorHandler {
	arg := createSovereignMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.ShardCoordinator = sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
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
	arg := createSovereignMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
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
}
