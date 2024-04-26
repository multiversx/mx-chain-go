//go:build !race

package process

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	commonMocks "github.com/multiversx/mx-chain-go/testscommon/common"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	storageCommon "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/update"
	updateMock "github.com/multiversx/mx-chain-go/update/mock"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var nodePrice = big.NewInt(5000)

// TODO improve code coverage of this package
func createMockArgument(
	t *testing.T,
	genesisFilename string,
	initialNodes genesis.InitialNodesHandler,
	entireSupply *big.Int,
) ArgsGenesisBlockCreator {
	trieStorageManagers := createTrieStorageManagers()
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
				GenesisNodePrice:                     nodePrice.Text(10),
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
				StakeLimitPercentage:                 100.0,
				NodeLimitPercentage:                  100.0,
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
			SoftAuctionConfig: config.SoftAuctionConfig{
				TopUpStep:             "10",
				MinTopUp:              "1",
				MaxTopUp:              "32000000",
				MaxNumberOfIterations: 100000,
			},
		},
		TrieStorageManagers: trieStorageManagers,
		BlockSignKeyGen:     &mock.KeyGenMock{},
		GenesisNodePrice:    nodePrice,
		EpochConfig: config.EpochConfig{
			EnableEpochs: config.EnableEpochs{
				SCDeployEnableEpoch:               unreachableEpoch,
				CleanUpInformativeSCRsEnableEpoch: unreachableEpoch,
				SCProcessorV2EnableEpoch:          unreachableEpoch,
				StakeLimitsEnableEpoch:            10,
			},
		},
		RoundConfig:             testscommon.GetDefaultRoundsConfig(),
		HeaderVersionConfigs:    testscommon.GetDefaultHeaderVersionConfig(),
		HistoryRepository:       &dblookupext.HistoryRepositoryStub{},
		TxExecutionOrderHandler: &commonMocks.TxExecutionOrderHandlerStub{},
		versionedHeaderFactory: &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32) data.HeaderHandler {
				return &block.Header{}
			},
		},
		RunTypeComponents: genesisMocks.NewRunTypeComponentsStub(),
	}

	arg.ShardCoordinator = &mock.ShardCoordinatorMock{
		NumOfShards: 2,
		SelfShardId: 0,
	}

	var err error
	arg.Accounts, err = createAccountAdapter(
		&mock.MarshalizerMock{},
		&hashingMocks.HasherMock{},
		arg.RunTypeComponents.AccountsCreator(),
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
			return accounts.NewPeerAccount(address)
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
			return math.MaxInt64
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

func createTrieStorageManagers() map[string]common.StorageManager {
	storageManagerArgs := storageCommon.GetStorageManagerArgs()
	storageManager, _ := trie.CreateTrieStorageManager(storageManagerArgs, storageCommon.GetStorageManagerOptions())

	trieStorageManagers := make(map[string]common.StorageManager)
	trieStorageManagers[dataRetriever.UserAccountsUnit.String()] = storageManager
	trieStorageManagers[dataRetriever.PeerAccountsUnit.String()] = storageManager

	return trieStorageManagers
}

func TestNewGenesisBlockCreator(t *testing.T) {
	t.Parallel()

	t.Run("nil Accounts should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.Accounts = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilAccountsAdapter)
		require.Nil(t, gbc)
	})
	t.Run("nil Core should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.Core = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilCoreComponentsHolder)
		require.Nil(t, gbc)
	})
	t.Run("nil Data should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.Data = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilDataComponentsHolder)
		require.Nil(t, gbc)
	})
	t.Run("nil AddressPubKeyConverter should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.Core = &mock.CoreComponentsMock{
			AddrPubKeyConv: nil,
		}

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilPubkeyConverter)
		require.Nil(t, gbc)
	})
	t.Run("nil InitialNodesSetup should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.InitialNodesSetup = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilNodesSetup)
		require.Nil(t, gbc)
	})
	t.Run("nil Economics should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.Economics = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilEconomicsData)
		require.Nil(t, gbc)
	})
	t.Run("nil ShardCoordinator should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.ShardCoordinator = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilShardCoordinator)
		require.Nil(t, gbc)
	})
	t.Run("nil StorageService should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.Data = &mock.DataComponentsMock{
			Storage: nil,
		}

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilStore)
		require.Nil(t, gbc)
	})
	t.Run("nil InternalMarshalizer should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.Core = &mock.CoreComponentsMock{
			AddrPubKeyConv: testscommon.NewPubkeyConverterMock(32),
			IntMarsh:       nil,
		}

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilMarshalizer)
		require.Nil(t, gbc)
	})
	t.Run("nil Hasher should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.Core = &mock.CoreComponentsMock{
			AddrPubKeyConv: testscommon.NewPubkeyConverterMock(32),
			IntMarsh:       &mock.MarshalizerMock{},
			Hash:           nil,
		}

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilHasher)
		require.Nil(t, gbc)
	})
	t.Run("nil DataPool should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.Data = &mock.DataComponentsMock{
			Storage:  &storageCommon.ChainStorerStub{},
			DataPool: nil,
		}

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilPoolsHolder)
		require.Nil(t, gbc)
	})
	t.Run("nil AccountsParser should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.AccountsParser = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, genesis.ErrNilAccountsParser)
		require.Nil(t, gbc)
	})
	t.Run("nil GasSchedule should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.GasSchedule = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, process.ErrNilGasSchedule)
		require.Nil(t, gbc)
	})
	t.Run("nil SmartContractParser should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.SmartContractParser = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, genesis.ErrNilSmartContractParser)
		require.Nil(t, gbc)
	})
	t.Run("nil RunTypeComponents should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.RunTypeComponents = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, errorsMx.ErrNilRunTypeComponents)
		require.Nil(t, gbc)
	})
	t.Run("nil BlockchainHookHandlerCreator should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		rtComponents := genesisMocks.NewRunTypeComponentsStub()
		rtComponents.BlockChainHookHandlerFactory = nil
		arg.RunTypeComponents = rtComponents

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, errorsMx.ErrNilBlockChainHookHandlerCreator)
		require.Nil(t, gbc)
	})
	t.Run("nil SCResultsPreProcessorCreator should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		rtComponents := genesisMocks.NewRunTypeComponentsStub()
		rtComponents.SCResultsPreProcessorFactory = nil
		arg.RunTypeComponents = rtComponents

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, errorsMx.ErrNilSCResultsPreProcessorCreator)
		require.Nil(t, gbc)
	})
	t.Run("nil SCProcessorCreator should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		rtComponents := genesisMocks.NewRunTypeComponentsStub()
		rtComponents.SCProcessorFactory = nil
		arg.RunTypeComponents = rtComponents

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, errorsMx.ErrNilSCProcessorCreator)
		require.Nil(t, gbc)
	})
	t.Run("nil TransactionCoordinatorCreator should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		rtComponents := genesisMocks.NewRunTypeComponentsStub()
		rtComponents.TransactionCoordinatorFactory = nil
		arg.RunTypeComponents = rtComponents

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, errorsMx.ErrNilTransactionCoordinatorCreator)
		require.Nil(t, gbc)
	})
	t.Run("nil AccountCreator should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		rtComponents := genesisMocks.NewRunTypeComponentsStub()
		rtComponents.AccountCreator = nil
		arg.RunTypeComponents = rtComponents

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, state.ErrNilAccountFactory)
		require.Nil(t, gbc)
	})
	t.Run("nil ShardCoordinatorFactory should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		rtComponents := genesisMocks.NewRunTypeComponentsStub()
		rtComponents.ShardCoordinatorFactory = nil
		arg.RunTypeComponents = rtComponents

		gbc, err := NewGenesisBlockCreator(arg)
		require.True(t, errors.Is(err, errorsMx.ErrNilShardCoordinatorFactory))
		require.Nil(t, gbc)
	})
	t.Run("nil TxPreProcessorFactory should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		rtComponents := genesisMocks.NewRunTypeComponentsStub()
		rtComponents.TxPreProcessorFactory = nil
		arg.RunTypeComponents = rtComponents

		gbc, err := NewGenesisBlockCreator(arg)
		require.True(t, errors.Is(err, errorsMx.ErrNilTxPreProcessorCreator))
		require.Nil(t, gbc)
	})
	t.Run("nil TrieStorageManagers should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.TrieStorageManagers = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, genesis.ErrNilTrieStorageManager)
		require.Nil(t, gbc)
	})
	t.Run("invalid GenesisNodePrice should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.SystemSCConfig.StakingSystemSCConfig.GenesisNodePrice = "0"

		gbc, err := NewGenesisBlockCreator(arg)
		require.ErrorIs(t, err, genesis.ErrInvalidInitialNodePrice)
		require.Nil(t, gbc)
	})
	t.Run("nil HistoryRepository should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		arg.HistoryRepository = nil

		gbc, err := NewGenesisBlockCreator(arg)
		require.True(t, errors.Is(err, process.ErrNilHistoryRepository))
		require.Nil(t, gbc)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
		gbc, err := NewGenesisBlockCreator(arg)
		require.NoError(t, err)
		require.NotNil(t, gbc)
	})
}

func TestGenesisBlockCreator_CreateGenesisBlockAfterHardForkShouldCreateSCResultingAddresses(t *testing.T) {
	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
				},
			}, make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	blocks, err := gbc.CreateGenesisBlocks()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(blocks))

	mapAddressesWithDeploy, err := arg.SmartContractParser.GetDeployedSCAddresses(genesis.DNSType)
	assert.Nil(t, err)
	assert.Equal(t, len(mapAddressesWithDeploy), core.MaxNumShards)

	newArgs := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)
	hardForkGbc, err := NewGenesisBlockCreator(newArgs)
	assert.Nil(t, err)
	err = hardForkGbc.computeInitialDNSAddresses(gbc.arg.EpochConfig.EnableEpochs)
	assert.Nil(t, err)

	mapAfterHardForkAddresses, err := newArgs.SmartContractParser.GetDeployedSCAddresses(genesis.DNSType)
	assert.Nil(t, err)
	assert.Equal(t, len(mapAfterHardForkAddresses), core.MaxNumShards)
	for address := range mapAddressesWithDeploy {
		_, ok := mapAfterHardForkAddresses[address]
		assert.True(t, ok)
	}
}

func TestGenesisBlockCreator_CreateGenesisBlocksJustDelegationShouldWorkAndDNS(t *testing.T) {
	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	stakedAddr, _ := hex.DecodeString("b00102030405060708090001020304050607080900010203040506070809000b")
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: stakedAddr,
						PubKeyBytesValue:  bytes.Repeat([]byte{2}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
				},
			}, make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)

	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	blocks, err := gbc.CreateGenesisBlocks()

	assert.Nil(t, err)
	assert.Equal(t, 3, len(blocks))
}

func TestGenesisBlockCreator_CreateGenesisBlocksStakingAndDelegationShouldWorkAndDNS(t *testing.T) {
	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	stakedAddr, _ := hex.DecodeString("b00102030405060708090001020304050607080900010203040506070809000b")
	stakedAddr2, _ := hex.DecodeString("d00102030405060708090001020304050607080900010203040506070809000d")
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: stakedAddr,
						PubKeyBytesValue:  bytes.Repeat([]byte{2}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: stakedAddr2,
						PubKeyBytesValue:  bytes.Repeat([]byte{8}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{4}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{5}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: stakedAddr2,
						PubKeyBytesValue:  bytes.Repeat([]byte{6}, 96),
					},
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: stakedAddr2,
						PubKeyBytesValue:  bytes.Repeat([]byte{7}, 96),
					},
				},
			}, make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest2.json",
		initialNodesSetup,
		big.NewInt(47000),
	)
	arg.ShardCoordinator, _ = sharding.NewMultiShardCoordinator(2, 1)
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	blocks, err := gbc.CreateGenesisBlocks()

	assert.Nil(t, err)
	assert.Equal(t, 3, len(blocks))

	_, err = arg.Accounts.Commit()
	require.Nil(t, err)

	t.Run("backwards compatibility on nonces: for a shard != 0, all accounts not having a delegation value would "+
		"have caused an artificial increase in their accounts nonce", func(t *testing.T) {
		accnt, errGet := arg.Accounts.GetExistingAccount(stakedAddr)
		require.Nil(t, errGet)
		assert.Equal(t, uint64(2), accnt.GetNonce())
	})
}

func TestGenesisBlockCreator_GetIndexingDataShouldWork(t *testing.T) {
	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	stakedAddr, _ := hex.DecodeString("b00102030405060708090001020304050607080900010203040506070809000b")
	stakedAddr2, _ := hex.DecodeString("d00102030405060708090001020304050607080900010203040506070809000d")
	initialGenesisNodes := map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
		0: {
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: scAddressBytes,
				PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: stakedAddr,
				PubKeyBytesValue:  bytes.Repeat([]byte{2}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: scAddressBytes,
				PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: stakedAddr2,
				PubKeyBytesValue:  bytes.Repeat([]byte{8}, 96),
			},
		},
		1: {
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: scAddressBytes,
				PubKeyBytesValue:  bytes.Repeat([]byte{4}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: scAddressBytes,
				PubKeyBytesValue:  bytes.Repeat([]byte{5}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: stakedAddr2,
				PubKeyBytesValue:  bytes.Repeat([]byte{6}, 96),
			},
			&mock.GenesisNodeInfoHandlerMock{
				AddressBytesValue: stakedAddr2,
				PubKeyBytesValue:  bytes.Repeat([]byte{7}, 96),
			},
		},
	}
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return initialGenesisNodes, make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest2.json",
		initialNodesSetup,
		big.NewInt(47000),
	)
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	blocks, err := gbc.CreateGenesisBlocks()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(blocks))

	indexingData := gbc.GetIndexingData()

	numDNSTypeScTxs := 256
	numDefaultTypeScTxs := 1
	numSystemSC := 4

	numInitialNodes := 0
	for k := range initialGenesisNodes {
		numInitialNodes += len(initialGenesisNodes[k])
	}

	reqNumDeployInitialScTxs := numDNSTypeScTxs + numDefaultTypeScTxs
	reqNumScrs := getRequiredNumScrsTxs(indexingData, 0)
	reqNumDelegationTxs := 4
	assert.Equal(t, reqNumDeployInitialScTxs, len(indexingData[0].DeployInitialScTxs))
	assert.Equal(t, 0, len(indexingData[0].DeploySystemScTxs))
	assert.Equal(t, reqNumDelegationTxs, len(indexingData[0].DelegationTxs))
	assert.Equal(t, 0, len(indexingData[0].StakingTxs))
	assert.Equal(t, reqNumScrs, len(indexingData[0].ScrsTxs))

	reqNumDeployInitialScTxs = numDNSTypeScTxs
	reqNumScrs = getRequiredNumScrsTxs(indexingData, 1)
	assert.Equal(t, reqNumDeployInitialScTxs, len(indexingData[1].DeployInitialScTxs))
	assert.Equal(t, 0, len(indexingData[1].DeploySystemScTxs))
	assert.Equal(t, 0, len(indexingData[1].DelegationTxs))
	assert.Equal(t, 0, len(indexingData[1].StakingTxs))
	assert.Equal(t, reqNumScrs, len(indexingData[1].ScrsTxs))

	reqNumScrs = getRequiredNumScrsTxs(indexingData, core.MetachainShardId)
	assert.Equal(t, 0, len(indexingData[core.MetachainShardId].DeployInitialScTxs))
	assert.Equal(t, numSystemSC, len(indexingData[core.MetachainShardId].DeploySystemScTxs))
	assert.Equal(t, 0, len(indexingData[core.MetachainShardId].DelegationTxs))
	assert.Equal(t, numInitialNodes, len(indexingData[core.MetachainShardId].StakingTxs))
	assert.Equal(t, reqNumScrs, len(indexingData[core.MetachainShardId].ScrsTxs))
}

func getRequiredNumScrsTxs(idata map[uint32]*genesis.IndexingData, shardId uint32) int {
	n := 2 * (len(idata[shardId].DeployInitialScTxs) + len(idata[shardId].DeploySystemScTxs) + len(idata[shardId].DelegationTxs))
	n += 3 * len(idata[shardId].StakingTxs)
	return n
}

func TestCreateArgsGenesisBlockCreator_ShouldErrWhenGetNewArgForShardFails(t *testing.T) {
	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	shardIDs := []uint32{0, 1}
	mapArgsGenesisBlockCreator := make(map[uint32]ArgsGenesisBlockCreator)
	initialNodesSetup := createDummyNodesHandler(scAddressBytes)
	arg := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)

	arg.ShardCoordinator = &mock.ShardCoordinatorMock{SelfShardId: 1}
	arg.TrieStorageManagers = make(map[string]common.StorageManager)
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	err = gbc.createArgsGenesisBlockCreator(shardIDs, mapArgsGenesisBlockCreator)
	assert.True(t, errors.Is(err, trie.ErrNilTrieStorage))
}

func TestCreateArgsGenesisBlockCreator_ShouldWork(t *testing.T) {
	shardIDs := []uint32{0, 1}
	mapArgsGenesisBlockCreator := make(map[uint32]ArgsGenesisBlockCreator)
	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
				},
			}, make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	err = gbc.createArgsGenesisBlockCreator(shardIDs, mapArgsGenesisBlockCreator)
	assert.Nil(t, err)
	require.Equal(t, 2, len(mapArgsGenesisBlockCreator))
	assert.Equal(t, uint32(0), mapArgsGenesisBlockCreator[0].ShardCoordinator.SelfId())
	assert.Equal(t, uint32(1), mapArgsGenesisBlockCreator[1].ShardCoordinator.SelfId())
}

func TestCreateHardForkBlockProcessors_ShouldWork(t *testing.T) {
	selfShardID := uint32(0)
	shardIDs := []uint32{1, core.MetachainShardId}
	mapArgsGenesisBlockCreator := make(map[uint32]ArgsGenesisBlockCreator)
	mapHardForkBlockProcessor := make(map[uint32]update.HardForkBlockProcessor)
	scAddressBytes, _ := hex.DecodeString("00000000000000000500761b8c4a25d3979359223208b412285f635e71300102")
	initialNodesSetup := &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
				},
			}, make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
	arg := createMockArgument(
		t,
		"testdata/genesisTest1.json",
		initialNodesSetup,
		big.NewInt(22000),
	)
	arg.importHandler = &updateMock.ImportHandlerStub{
		GetAccountsDBForShardCalled: func(shardID uint32) state.AccountsAdapter {
			return &stateMock.AccountsStub{}
		},
	}
	gbc, err := NewGenesisBlockCreator(arg)
	require.Nil(t, err)

	_ = gbc.createArgsGenesisBlockCreator(shardIDs, mapArgsGenesisBlockCreator)

	err = createHardForkBlockProcessors(selfShardID, shardIDs, mapArgsGenesisBlockCreator, mapHardForkBlockProcessor)
	assert.Nil(t, err)
	require.Equal(t, 2, len(mapHardForkBlockProcessor))
}

func createDummyNodesHandler(scAddressBytes []byte) genesis.InitialNodesHandler {
	return &mock.InitialNodesHandlerStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
				0: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{1}, 96),
					},
				},
				1: {
					&mock.GenesisNodeInfoHandlerMock{
						AddressBytesValue: scAddressBytes,
						PubKeyBytesValue:  bytes.Repeat([]byte{3}, 96),
					},
				},
			}, make(map[uint32][]nodesCoordinator.GenesisNodeInfoHandler)
		},
		MinNumberOfNodesCalled: func() uint32 {
			return 1
		},
	}
}

func TestCreateArgsGenesisBlockCreator_ShouldWorkAndCreateEmpty(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(t, "testdata/genesisTest1.json", &mock.InitialNodesHandlerStub{}, big.NewInt(22000))
	arg.StartEpochNum = 1
	gbc, err := NewGenesisBlockCreator(arg)
	require.NoError(t, err)
	require.NotNil(t, gbc)

	blocks, err := gbc.CreateGenesisBlocks()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(blocks))
	for _, blockInstance := range blocks {
		assert.Zero(t, blockInstance.GetNonce())
		assert.Zero(t, blockInstance.GetRound())
		assert.Zero(t, blockInstance.GetEpoch())
	}
}
