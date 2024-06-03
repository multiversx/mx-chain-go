package api_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	factoryErrors "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/api"
	"github.com/multiversx/mx-chain-go/factory/mock"
	testsMocks "github.com/multiversx/mx-chain-go/integrationTests/mock"
	trieIteratorsFactory "github.com/multiversx/mx-chain-go/node/trieIterators/factory"
	"github.com/multiversx/mx-chain-go/process"
	vmFactory "github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/process/sync/disabled"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMocks "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"
)

const unreachableStep = 10000

type failingSteps struct {
	marshallerStepCounter int
	marshallerFailingStep int

	enableEpochsHandlerStepCounter int
	enableEpochsHandlerFailingStep int

	uint64ByteSliceConvStepCounter int
	uint64ByteSliceConvFailingStep int

	addressPublicKeyConverterStepCounter int
	addressPublicKeyConverterFailingStep int

	economicsDataStepCounter int
	economicsDataFailingStep int
}

func (fs *failingSteps) reset() {
	fs.marshallerStepCounter = 0
	fs.marshallerFailingStep = unreachableStep

	fs.enableEpochsHandlerStepCounter = 0
	fs.enableEpochsHandlerFailingStep = unreachableStep

	fs.uint64ByteSliceConvStepCounter = 0
	fs.uint64ByteSliceConvFailingStep = unreachableStep

	fs.addressPublicKeyConverterStepCounter = 0
	fs.addressPublicKeyConverterFailingStep = unreachableStep

	fs.economicsDataStepCounter = 0
	fs.economicsDataFailingStep = unreachableStep
}

func createMockArgs(t *testing.T) *api.ApiResolverArgs {
	cfg := testscommon.GetGeneralConfig()
	coreComponents := componentsMock.GetCoreComponents(cfg)
	statusCoreComponents := componentsMock.GetStatusCoreComponents(cfg, coreComponents)
	cryptoComponents := componentsMock.GetCryptoComponents(coreComponents)
	networkComponents := componentsMock.GetNetworkComponents(cryptoComponents)
	runTypeComponents := componentsMock.GetRunTypeComponents(coreComponents, cryptoComponents)
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	bootstrapComponents := componentsMock.GetBootstrapComponents(cfg, statusCoreComponents, coreComponents, cryptoComponents, networkComponents, runTypeComponents)
	componentsMock.SetShardCoordinator(t, bootstrapComponents, shardCoordinator)
	dataComponents := componentsMock.GetDataComponents(cfg, statusCoreComponents, coreComponents, bootstrapComponents, cryptoComponents, runTypeComponents)
	stateComponents := componentsMock.GetStateComponents(cfg, coreComponents, dataComponents, statusCoreComponents, runTypeComponents)
	statusComponents := componentsMock.GetStatusComponents(cfg, statusCoreComponents, coreComponents, networkComponents, bootstrapComponents, stateComponents, &shardingMocks.NodesCoordinatorMock{}, cryptoComponents)
	processComponents := componentsMock.GetProcessComponents(cfg, runTypeComponents, coreComponents, cryptoComponents, networkComponents, bootstrapComponents, stateComponents, dataComponents, statusComponents, statusCoreComponents)

	gasSchedule, _ := common.LoadGasScheduleConfig("../../cmd/node/config/gasSchedules/gasScheduleV1.toml")
	economicsConfig := testscommon.GetEconomicsConfig()

	return &api.ApiResolverArgs{
		Configs: &config.Configs{
			FlagsConfig: &config.ContextFlagsConfig{
				WorkingDir: "",
			},
			GeneralConfig:   &cfg,
			EpochConfig:     &config.EpochConfig{},
			EconomicsConfig: &economicsConfig,
		},
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		StateComponents:      stateComponents,
		BootstrapComponents:  bootstrapComponents,
		CryptoComponents:     cryptoComponents,
		ProcessComponents:    processComponents,
		StatusCoreComponents: statusCoreComponents,
		GasScheduleNotifier: &testscommon.GasScheduleNotifierMock{
			GasSchedule: gasSchedule,
		},
		Bootstrapper:       disabled.NewDisabledBootstrapper(),
		AllowVMQueriesChan: common.GetClosedUnbufferedChannel(),
		StatusComponents: &mainFactoryMocks.StatusComponentsStub{
			ManagedPeersMonitorField: &testscommon.ManagedPeersMonitorStub{},
		},
		RunTypeComponents:              runTypeComponents,
		DelegatedListFactoryHandler:    trieIteratorsFactory.NewDelegatedListProcessorFactory(),
		DirectStakedListFactoryHandler: trieIteratorsFactory.NewDirectStakedListProcessorFactory(),
		TotalStakedValueFactoryHandler: trieIteratorsFactory.NewTotalStakedListProcessorFactory(),
	}
}

func createFailingMockArgs(t *testing.T, failingSteps *failingSteps) *api.ApiResolverArgs {
	args := createMockArgs(t)
	coreCompStub := factory.NewCoreComponentsHolderStubFromRealComponent(args.CoreComponents)

	internalMarshaller := args.CoreComponents.InternalMarshalizer()
	coreCompStub.InternalMarshalizerCalled = func() marshal.Marshalizer {
		failingSteps.marshallerStepCounter++
		if failingSteps.marshallerStepCounter > failingSteps.marshallerFailingStep {
			return nil
		}
		return internalMarshaller
	}

	enableEpochsHandler := args.CoreComponents.EnableEpochsHandler()
	coreCompStub.EnableEpochsHandlerCalled = func() common.EnableEpochsHandler {
		failingSteps.enableEpochsHandlerStepCounter++
		if failingSteps.enableEpochsHandlerStepCounter > failingSteps.enableEpochsHandlerFailingStep {
			return nil
		}
		return enableEpochsHandler
	}

	byteSliceConv := args.CoreComponents.Uint64ByteSliceConverter()
	coreCompStub.Uint64ByteSliceConverterCalled = func() typeConverters.Uint64ByteSliceConverter {
		failingSteps.uint64ByteSliceConvStepCounter++
		if failingSteps.uint64ByteSliceConvStepCounter > failingSteps.uint64ByteSliceConvFailingStep {
			return nil
		}
		return byteSliceConv
	}

	pubKeyConv := args.CoreComponents.AddressPubKeyConverter()
	coreCompStub.AddressPubKeyConverterCalled = func() core.PubkeyConverter {
		failingSteps.addressPublicKeyConverterStepCounter++
		if failingSteps.addressPublicKeyConverterStepCounter > failingSteps.addressPublicKeyConverterFailingStep {
			return nil
		}
		return pubKeyConv
	}

	economicsData := args.CoreComponents.EconomicsData()
	coreCompStub.EconomicsDataCalled = func() process.EconomicsDataHandler {
		failingSteps.economicsDataStepCounter++
		if failingSteps.economicsDataStepCounter > failingSteps.economicsDataFailingStep {
			return nil
		}
		return economicsData
	}

	args.CoreComponents = coreCompStub
	return args
}

func TestCreateApiResolver(t *testing.T) {
	t.Parallel()

	t.Run("createScQueryService fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.Configs.GeneralConfig.VirtualMachine.Querying.NumConcurrentVMs = 0
		apiResolver, err := api.CreateApiResolver(args)
		require.True(t, strings.Contains(err.Error(), "VirtualMachine.Querying.NumConcurrentVms"))
		require.True(t, check.IfNil(apiResolver))
	})

	failingStepsInstance := &failingSteps{}
	failingArgs := createFailingMockArgs(t, failingStepsInstance)
	// do not run these tests in parallel as they all use the same args
	t.Run("DecodeAddresses fails causing createScQueryElement error should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.addressPublicKeyConverterFailingStep = 0
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "public key converter"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewESDTTransferParser fails causing createScQueryElement error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 4
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("DecodeAddresses fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.addressPublicKeyConverterFailingStep = 3
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "nil address converter"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createBuiltinFuncs fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 7
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewESDTTransferParser fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 8
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewTxTypeHandler fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.enableEpochsHandlerFailingStep = 6
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "enable epochs handler"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewFeeComputer fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.economicsDataFailingStep = 3
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "economics data"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createLogsFacade fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 10
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshaller"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewOperationDataFieldParser fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 11
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewAPITransactionProcessor fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 12
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createAPIBlockProcessor fails because createAPIBlockProcessorArgs fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.uint64ByteSliceConvFailingStep = 3
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "uint64"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createAPIInternalBlockProcessor fails because createAPIBlockProcessorArgs fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.uint64ByteSliceConvFailingStep = 5
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "uint64"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createAPIBlockProcessorArgs fails because createLogsFacade fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 13
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createAPIBlockProcessorArgs fails because NewAlteredAccountsProvider fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.addressPublicKeyConverterFailingStep = 10
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "pubkey converter"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("should work", func(t *testing.T) {
		failingStepsInstance.reset() // no failure
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.Nil(t, err)
		require.False(t, check.IfNil(apiResolver))
	})
	t.Run("DelegatedListFactoryHandler nil should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.DelegatedListFactoryHandler = nil
		apiResolver, err := api.CreateApiResolver(args)
		require.Equal(t, factoryErrors.ErrNilDelegatedListFactory, err)
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("DirectStakedListFactoryHandler nil should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.DirectStakedListFactoryHandler = nil
		apiResolver, err := api.CreateApiResolver(args)
		require.Equal(t, factoryErrors.ErrNilDirectStakedListFactory, err)
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("TotalStakedValueFactoryHandler nil should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs(t)
		args.TotalStakedValueFactoryHandler = nil
		apiResolver, err := api.CreateApiResolver(args)
		require.Equal(t, factoryErrors.ErrNilTotalStakedValueFactory, err)
		require.True(t, check.IfNil(apiResolver))
	})
}

func createMockSCQueryElementArgs(shardId uint32) api.SCQueryElementArgs {
	coreComp := componentsMock.GetCoreComponents(testscommon.GetGeneralConfig())
	cryptoComp := componentsMock.GetCryptoComponents(coreComp)
	return api.SCQueryElementArgs{
		GeneralConfig: &config.Config{
			BuiltInFunctions: config.BuiltInFunctionsConfig{
				MaxNumAddressesInTransferRole: 1,
				AutomaticCrawlerAddresses:     []string{"addr1"},
			},
			SmartContractDataPool: config.CacheConfig{
				Type:     "LRU",
				Capacity: 100,
			},
			EvictionWaitingList: config.EvictionWaitingListConfig{
				RootHashesSize: 100,
				HashesSize:     10000,
			},
			TrieStorageManagerConfig: config.TrieStorageManagerConfig{
				SnapshotsGoroutineNum: 1,
			},
			StateTriesConfig: config.StateTriesConfig{
				MaxStateTrieLevelInMemory: 5,
			},
			VirtualMachine: config.VirtualMachineServicesConfig{
				Querying: config.QueryVirtualMachineConfig{
					VirtualMachineConfig: config.VirtualMachineConfig{
						WasmVMVersions: []config.WasmVMVersionByEpoch{
							{StartEpoch: 0, Version: "*"},
						},
						TransferAndExecuteByUserAddresses: []string{"erd1qqqqqqqqqqqqqpgqr46jrxr6r2unaqh75ugd308dwx5vgnhwh47qtvepe3"},
					},
				},
			},
		},
		EpochConfig: &config.EpochConfig{},
		CoreComponents: &mock.CoreComponentsMock{
			AddrPubKeyConv: &testscommon.PubkeyConverterStub{
				DecodeCalled: func(humanReadable string) ([]byte, error) {
					return []byte(humanReadable), nil
				},
			},
			IntMarsh:                     &marshallerMock.MarshalizerStub{},
			EpochChangeNotifier:          &epochNotifierMock.EpochNotifierStub{},
			EnableEpochsHandlerField:     &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			UInt64ByteSliceConv:          &testsMocks.Uint64ByteSliceConverterMock{},
			EconomicsHandler:             &economicsmocks.EconomicsHandlerStub{},
			NodesConfig:                  &genesisMocks.NodesSetupStub{},
			Hash:                         &testscommon.HasherStub{},
			RatingHandler:                &testscommon.RaterMock{},
			WasmVMChangeLockerInternal:   &sync.RWMutex{},
			PathHdl:                      &testscommon.PathManagerStub{},
			ProcessStatusHandlerInternal: &testscommon.ProcessStatusHandlerStub{},
		},
		StateComponents: &mock.StateComponentsHolderStub{
			AccountsAdapterAPICalled: func() state.AccountsAdapter {
				return &stateMocks.AccountsStub{}
			},
			PeerAccountsCalled: func() state.AccountsAdapter {
				return &stateMocks.AccountsStub{}
			},
		},
		StatusCoreComponents: &factory.StatusCoreComponentsStub{
			AppStatusHandlerCalled: func() core.AppStatusHandler {
				return &statusHandler.AppStatusHandlerStub{}
			},
			StateStatsHandlerField: &testscommon.StateStatisticsHandlerStub{},
		},
		DataComponents: &mock.DataComponentsMock{
			Storage:  genericMocks.NewChainStorerMock(0),
			Blkc:     &testscommon.ChainHandlerMock{},
			DataPool: &dataRetriever.PoolsHolderMock{},
		},
		ProcessComponents: &mock.ProcessComponentsMock{
			ShardCoord: &testscommon.ShardsCoordinatorMock{
				CurrentShard: shardId,
			},
			NodesCoord: &shardingMocks.NodesCoordinatorStub{},
		},
		GasScheduleNotifier: &testscommon.GasScheduleNotifierMock{
			LatestGasScheduleCalled: func() map[string]map[string]uint64 {
				gasSchedule, _ := common.LoadGasScheduleConfig("../../cmd/node/config/gasSchedules/gasScheduleV1.toml")
				return gasSchedule
			},
		},
		MessageSigVerifier: &testscommon.MessageSignVerifierMock{},
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "1000",
				OwnerAddress:    "erd1fpkcgel4gcmh8zqqdt043yfcn5tyx8373kg6q2qmkxzu4dqamc0swts65c",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				V1: config.GovernanceSystemSCConfigV1{
					ProposalCost:     "500",
					NumNodes:         100,
					MinQuorum:        50,
					MinPassThreshold: 50,
					MinVetoThreshold: 50,
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
				},
				OwnerAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				MinStakeValue:                        "1",
				UnJailValue:                          "1",
				MinStepValue:                         "1",
				UnBondPeriod:                         0,
				NumRoundsWithoutBleed:                0,
				MaximumPercentageToBleed:             0,
				BleedPercentagePerRound:              0,
				MaxNumberOfNodesForStake:             10,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
				NodeLimitPercentage:                  100.0,
				StakeLimitPercentage:                 100.0,
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "100",
				MinStakeAmount:      "100",
				ConfigChangeAddress: "erd1vxy22x0fj4zv6hktmydg8vpfh6euv02cz4yg0aaws6rrad5a5awqgqky80",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 100,
			},
		},
		Bootstrapper:          testsMocks.NewTestBootstrapperMock(),
		AllowVMQueriesChan:    make(chan struct{}, 1),
		WorkingDir:            "",
		Index:                 0,
		GuardedAccountHandler: &guardianMocks.GuardedAccountHandlerStub{},
		RunTypeComponents:     componentsMock.GetRunTypeComponents(coreComp, cryptoComp),
	}
}

func TestCreateApiResolver_createScQueryElement(t *testing.T) {
	t.Parallel()

	t.Run("nil guardian handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		args.GuardedAccountHandler = nil
		scQueryService, storageManager, err := api.CreateScQueryElement(args)
		require.Equal(t, process.ErrNilGuardedAccountHandler, err)
		require.Nil(t, scQueryService)
		require.Nil(t, storageManager)
	})
	t.Run("DecodeAddresses fails", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		args.CoreComponents = &mock.CoreComponentsMock{
			AddrPubKeyConv: nil,
		}
		scQueryService, storageManager, err := api.CreateScQueryElement(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "public key converter"))
		require.Nil(t, scQueryService)
		require.Nil(t, storageManager)
	})
	t.Run("createBuiltinFuncs fails", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		coreCompMock := args.CoreComponents.(*mock.CoreComponentsMock)
		coreCompMock.IntMarsh = nil
		scQueryService, storageManager, err := api.CreateScQueryElement(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.Nil(t, scQueryService)
		require.Nil(t, storageManager)
	})
	t.Run("NewCache fails", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		args.GeneralConfig.SmartContractDataPool = config.CacheConfig{
			Type:        "LRU",
			SizeInBytes: 1,
		}
		scQueryService, storageManager, err := api.CreateScQueryElement(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "lru"))
		require.Nil(t, scQueryService)
		require.Nil(t, storageManager)
	})
	t.Run("metachain - createNewAccountsAdapterApi fails", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		args.ProcessComponents = &mock.ProcessComponentsMock{
			ShardCoord: &testscommon.ShardsCoordinatorMock{
				SelfIDCalled: func() uint32 {
					return common.MetachainShardId
				},
			},
		}
		coreCompMock := args.CoreComponents.(*mock.CoreComponentsMock)
		coreCompMock.Hash = nil
		scQueryService, storageManager, err := api.CreateScQueryElement(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "hasher"))
		require.Nil(t, scQueryService)
		require.Nil(t, storageManager)
	})
	t.Run("shard - createNewAccountsAdapterApi fails", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		coreCompMock := args.CoreComponents.(*mock.CoreComponentsMock)
		coreCompMock.Hash = nil
		scQueryService, storageManager, err := api.CreateScQueryElement(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "hasher"))
		require.Nil(t, scQueryService)
		require.Nil(t, storageManager)
	})
}

func TestCreateApiResolver_createBlockchainForScQuery(t *testing.T) {
	t.Parallel()

	t.Run("for metachain", func(t *testing.T) {
		t.Parallel()

		apiBlockchain, err := api.CreateBlockchainForScQuery(core.MetachainShardId)
		require.NoError(t, err)
		require.Equal(t, "*blockchain.metaChain", fmt.Sprintf("%T", apiBlockchain))
	})

	t.Run("for shard", func(t *testing.T) {
		t.Parallel()

		apiBlockchain, err := api.CreateBlockchainForScQuery(0)
		require.NoError(t, err)
		require.Equal(t, "*blockchain.blockChain", fmt.Sprintf("%T", apiBlockchain))
	})
}

func TestCreateApiResolver_createArgsSCQueryService(t *testing.T) {
	t.Parallel()

	t.Run("sovereign chain should add systemVM", func(t *testing.T) {
		t.Parallel()

		coreComp := componentsMock.GetSovereignCoreComponents(testscommon.GetGeneralConfig())
		cryptoComp := componentsMock.GetCryptoComponents(coreComp)
		args := createMockSCQueryElementArgs(0)
		args.RunTypeComponents = componentsMock.GetSovereignRunTypeComponents(coreComp, cryptoComp)

		argsScQueryService, err := api.CreateArgsSCQueryService(args)
		require.Nil(t, err)
		require.NotNil(t, argsScQueryService.VmContainer)

		require.Equal(t, 2, argsScQueryService.VmContainer.Len())

		svm, err := argsScQueryService.VmContainer.Get(vmFactory.SystemVirtualMachine)
		require.Nil(t, err)
		require.NotNil(t, svm)
		require.Equal(t, "*process.systemVM", fmt.Sprintf("%T", svm))

		wasmvm, err := argsScQueryService.VmContainer.Get(vmFactory.WasmVirtualMachine)
		require.Nil(t, err)
		require.NotNil(t, wasmvm)
		require.Equal(t, "*hostCore.vmHost", fmt.Sprintf("%T", wasmvm))
	})
	t.Run("regular chain for shards should only add wasm vm", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)

		argsScQueryService, err := api.CreateArgsSCQueryService(args)
		require.Nil(t, err)
		require.NotNil(t, argsScQueryService.VmContainer)

		require.Equal(t, 1, argsScQueryService.VmContainer.Len())

		svm, err := argsScQueryService.VmContainer.Get(vmFactory.SystemVirtualMachine)
		require.NotNil(t, err)
		require.Nil(t, svm)

		wasmvm, err := argsScQueryService.VmContainer.Get(vmFactory.WasmVirtualMachine)
		require.Nil(t, err)
		require.NotNil(t, wasmvm)
		require.Equal(t, "*hostCore.vmHost", fmt.Sprintf("%T", wasmvm))
	})
	t.Run("regular chain for meta should only add systemVM", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(common.MetachainShardId)

		argsScQueryService, err := api.CreateArgsSCQueryService(args)
		require.Nil(t, err)
		require.NotNil(t, argsScQueryService.VmContainer)

		require.Equal(t, 1, argsScQueryService.VmContainer.Len())

		svm, err := argsScQueryService.VmContainer.Get(vmFactory.SystemVirtualMachine)
		require.Nil(t, err)
		require.NotNil(t, svm)
	})
}
