package staking

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
)

type baseMetaProcessor struct {
	MetaBlockProcessor  process.BlockProcessor
	NodesCoordinator    nodesCoordinator.NodesCoordinator
	ValidatorStatistics process.ValidatorStatisticsProcessor
	EpochStartTrigger   integrationTests.TestEpochStartTrigger
	BlockChainHandler   data.ChainHandler
	NodesConfig         nodesConfig
	AccountsAdapter     state.AccountsAdapter
	Marshaller          marshal.Marshalizer

	currentRound uint64
}

func newBaseMetaProcessor(
	coreComponents factory.CoreComponentsHolder,
	dataComponents factory.DataComponentsHolder,
	bootstrapComponents factory.BootstrapComponentsHolder,
	statusComponents factory.StatusComponentsHolder,
	stateComponents factory.StateComponentsHandler,
	nc nodesCoordinator.NodesCoordinator,
	maxNodesConfig []config.MaxNodesChangeConfig,
	queue [][]byte,
) *baseMetaProcessor {
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

	return &baseMetaProcessor{
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
