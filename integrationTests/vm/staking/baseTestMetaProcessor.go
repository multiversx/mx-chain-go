package staking

import (
	arwenConfig "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/config"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
)

func newTestMetaProcessor(
	coreComponents factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	bootstrapComponents factory.BootstrapComponentsHolder,
	statusComponents factory.StatusComponentsHolder,
	stateComponents factory.StateComponentsHandler,
	nc nodesCoordinator.NodesCoordinator,
	maxNodesConfig []config.MaxNodesChangeConfig,
	queue [][]byte,
) *TestMetaProcessor {
	gasScheduleNotifier := createGasScheduleNotifier()
	blockChainHook := createBlockChainHook(
		dataComponents, coreComponents,
		stateComponents.AccountsAdapter(),
		bootstrapComponents.ShardCoordinator(),
		gasScheduleNotifier,
	)

	metaVmFactory := createVMContainerFactory(
		coreComponents,
		gasScheduleNotifier,
		blockChainHook,
		stateComponents.PeerAccounts(),
		bootstrapComponents.ShardCoordinator(),
		nc,
		maxNodesConfig[0].MaxNumNodes,
	)
	vmContainer, _ := metaVmFactory.Create()

	validatorStatisticsProcessor := createValidatorStatisticsProcessor(
		dataComponents,
		coreComponents,
		nc,
		bootstrapComponents.ShardCoordinator(),
		stateComponents.PeerAccounts(),
	)
	scp := createSystemSCProcessor(
		nc,
		coreComponents,
		stateComponents,
		bootstrapComponents.ShardCoordinator(),
		maxNodesConfig,
		validatorStatisticsProcessor,
		vmContainer,
	)

	epochStartTrigger := createEpochStartTrigger(coreComponents, dataComponents.StorageService())

	eligible, _ := nc.GetAllEligibleValidatorsPublicKeys(0)
	waiting, _ := nc.GetAllWaitingValidatorsPublicKeys(0)
	shuffledOut, _ := nc.GetAllShuffledOutValidatorsPublicKeys(0)

	return &TestMetaProcessor{
		AccountsAdapter: stateComponents.AccountsAdapter(),
		Marshaller:      coreComponents.InternalMarshalizer(),
		NodesConfig: nodesConfig{
			eligible:    eligible,
			waiting:     waiting,
			shuffledOut: shuffledOut,
			queue:       queue,
			auction:     make([][]byte, 0),
		},
		MetaBlockProcessor: createMetaBlockProcessor(
			nc,
			scp,
			coreComponents,
			dataComponents,
			bootstrapComponents,
			statusComponents,
			stateComponents,
			validatorStatisticsProcessor,
			blockChainHook,
			metaVmFactory,
			epochStartTrigger,
			vmContainer,
		),
		currentRound:        1,
		NodesCoordinator:    nc,
		ValidatorStatistics: validatorStatisticsProcessor,
		EpochStartTrigger:   epochStartTrigger,
		BlockChainHandler:   dataComponents.Blockchain(),
	}
}

func createGasScheduleNotifier() core.GasScheduleNotifier {
	gasSchedule := arwenConfig.MakeGasMapForTests()
	defaults.FillGasMapInternal(gasSchedule, 1)
	return mock.NewGasScheduleNotifierMock(gasSchedule)
}

func createEpochStartTrigger(
	coreComponents factory.CoreComponentsHolder,
	storageService dataRetriever.StorageService,
) integrationTests.TestEpochStartTrigger {
	argsEpochStart := &metachain.ArgsNewMetaEpochStartTrigger{
		Settings: &config.EpochStartConfig{
			MinRoundsBetweenEpochs: 10,
			RoundsPerEpoch:         10,
		},
		Epoch:              0,
		EpochStartNotifier: coreComponents.EpochStartNotifierWithConfirm(),
		Storage:            storageService,
		Marshalizer:        coreComponents.InternalMarshalizer(),
		Hasher:             coreComponents.Hasher(),
		AppStatusHandler:   coreComponents.StatusHandler(),
	}

	epochStartTrigger, _ := metachain.NewEpochStartTrigger(argsEpochStart)
	testTrigger := &metachain.TestTrigger{}
	testTrigger.SetTrigger(epochStartTrigger)

	return testTrigger
}
