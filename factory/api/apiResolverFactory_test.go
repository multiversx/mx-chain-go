package api_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	factoryErrors "github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/api"
	"github.com/multiversx/mx-chain-go/factory/bootstrap"
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
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMocks "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
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
}

func createMockArgs(t *testing.T) *api.ApiResolverArgs {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	coreComponents := componentsMock.GetCoreComponents()
	cryptoComponents := componentsMock.GetCryptoComponents(coreComponents)
	networkComponents := componentsMock.GetNetworkComponents(cryptoComponents)
	dataComponents := componentsMock.GetDataComponents(coreComponents, shardCoordinator)
	stateComponents := componentsMock.GetStateComponents(coreComponents)
	processComponents := componentsMock.GetProcessComponents(shardCoordinator, coreComponents, networkComponents, dataComponents, cryptoComponents, stateComponents)
	argsB := componentsMock.GetBootStrapFactoryArgs()

	bcf, _ := bootstrap.NewBootstrapComponentsFactory(argsB)
	mbc, err := bootstrap.NewManagedBootstrapComponents(bcf)
	require.Nil(t, err)
	err = mbc.Create()
	require.Nil(t, err)

	gasSchedule, _ := common.LoadGasScheduleConfig("../../cmd/node/config/gasSchedules/gasScheduleV1.toml")
	economicsConfig := testscommon.GetEconomicsConfig()
	cfg := componentsMock.GetGeneralConfig()

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
		BootstrapComponents:  mbc,
		CryptoComponents:     cryptoComponents,
		ProcessComponents:    processComponents,
		StatusCoreComponents: componentsMock.GetStatusCoreComponents(),
		GasScheduleNotifier: &testscommon.GasScheduleNotifierMock{
			GasSchedule: gasSchedule,
		},
		Bootstrapper:       disabled.NewDisabledBootstrapper(),
		AllowVMQueriesChan: common.GetClosedUnbufferedChannel(),
		StatusComponents: &mainFactoryMocks.StatusComponentsStub{
			ManagedPeersMonitorField: &testscommon.ManagedPeersMonitorStub{},
		},
		ChainRunType:                   common.ChainRunTypeRegular,
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
	t.Run("DecodeAddresses fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.addressPublicKeyConverterFailingStep = 3
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "public key converter"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createBuiltinFuncs fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 3
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewESDTTransferParser fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 5
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		println(err.Error())
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshaller"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewTxTypeHandler fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.enableEpochsHandlerFailingStep = 4
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "enable epochs handler"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewTransactionCostEstimator fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.enableEpochsHandlerFailingStep = 5
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "enable epochs handler"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createLogsFacade fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 9
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewOperationDataFieldParser fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 10
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("NewAPITransactionProcessor fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 11
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createAPIBlockProcessor fails because createAPIBlockProcessorArgs fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.uint64ByteSliceConvFailingStep = 2
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "uint64"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createAPIInternalBlockProcessor fails because createAPIBlockProcessorArgs fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.uint64ByteSliceConvFailingStep = 4
		apiResolver, err := api.CreateApiResolver(failingArgs)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "uint64"))
		require.True(t, check.IfNil(apiResolver))
	})
	t.Run("createAPIBlockProcessorArgs fails because createLogsFacade fails should error", func(t *testing.T) {
		failingStepsInstance.reset()
		failingStepsInstance.marshallerFailingStep = 12
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
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "public key converter"))
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
			NodesConfig:                  &testscommon.NodesSetupStub{},
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
				GenesisNodePrice:                     "2500000000000000000000",
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
		ChainRunType:          common.ChainRunTypeRegular,
	}
}

func TestCreateApiResolver_createScQueryElement(t *testing.T) {
	t.Parallel()

	t.Run("nil guardian handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		args.GuardedAccountHandler = nil
		scQueryService, err := api.CreateScQueryElement(args)
		require.Equal(t, process.ErrNilGuardedAccountHandler, err)
		require.Nil(t, scQueryService)
	})
	t.Run("DecodeAddresses fails", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		args.CoreComponents = &mock.CoreComponentsMock{
			AddrPubKeyConv: nil,
		}
		scQueryService, err := api.CreateScQueryElement(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "public key converter"))
		require.Nil(t, scQueryService)
	})
	t.Run("createBuiltinFuncs fails", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		coreCompMock := args.CoreComponents.(*mock.CoreComponentsMock)
		coreCompMock.IntMarsh = nil
		scQueryService, err := api.CreateScQueryElement(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "marshalizer"))
		require.Nil(t, scQueryService)
	})
	t.Run("NewCache fails", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		args.GeneralConfig.SmartContractDataPool = config.CacheConfig{
			Type:        "LRU",
			SizeInBytes: 1,
		}
		scQueryService, err := api.CreateScQueryElement(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "lru"))
		require.Nil(t, scQueryService)
	})
	t.Run("metachain - NewVMContainerFactory fails", func(t *testing.T) {
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
		scQueryService, err := api.CreateScQueryElement(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "hasher"))
		require.Nil(t, scQueryService)
	})
	t.Run("shard - NewVMContainerFactory fails", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		coreCompMock := args.CoreComponents.(*mock.CoreComponentsMock)
		coreCompMock.Hash = nil
		scQueryService, err := api.CreateScQueryElement(args)
		require.NotNil(t, err)
		require.True(t, strings.Contains(strings.ToLower(err.Error()), "hasher"))
		require.Nil(t, scQueryService)
	})

}

func TestCreateApiResolver_createArgsSCQueryService(t *testing.T) {
	t.Parallel()

	t.Run("sovereign chain should add systemVM", func(t *testing.T) {
		t.Parallel()

		args := createMockSCQueryElementArgs(0)
		args.ChainRunType = common.ChainRunTypeSovereign

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
		args.ChainRunType = common.ChainRunTypeRegular

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
		args.ChainRunType = common.MetachainShardName

		argsScQueryService, err := api.CreateArgsSCQueryService(args)
		require.Nil(t, err)
		require.NotNil(t, argsScQueryService.VmContainer)

		require.Equal(t, 1, argsScQueryService.VmContainer.Len())

		svm, err := argsScQueryService.VmContainer.Get(vmFactory.SystemVirtualMachine)
		require.Nil(t, err)
		require.NotNil(t, svm)
	})
}
