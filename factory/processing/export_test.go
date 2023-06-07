package processing

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/cutoff"
)

// NewBlockProcessor calls the unexported method with the same name in order to use it in tests
func (pcf *processComponentsFactory) NewBlockProcessor(
	requestHandler process.RequestHandler,
	forkDetector process.ForkDetector,
	epochStartTrigger epochStart.TriggerHandler,
	bootStorer process.BootStorer,
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor,
	headerValidator process.HeaderConstructionValidator,
	blockTracker process.BlockTracker,
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler,
	wasmVMChangeLocker common.Locker,
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler,
	processedMiniBlocksTracker process.ProcessedMiniBlocksTracker,
	receiptsRepository factory.ReceiptsRepository,
	blockProcessingCutoff cutoff.BlockProcessingCutoffHandler,
	missingTrieNodesNotifier common.MissingTrieNodesNotifier,
) (process.BlockProcessor, error) {
	blockProcessorComponents, err := pcf.newBlockProcessor(
		requestHandler,
		forkDetector,
		epochStartTrigger,
		bootStorer,
		validatorStatisticsProcessor,
		headerValidator,
		blockTracker,
		pendingMiniBlocksHandler,
		wasmVMChangeLocker,
		scheduledTxsExecutionHandler,
		processedMiniBlocksTracker,
		receiptsRepository,
		blockProcessingCutoff,
		missingTrieNodesNotifier,
	)
	if err != nil {
		return nil, err
	}

	return blockProcessorComponents.blockProcessor, nil
}

// CreateTxSimulatorProcessor -
func (pcf *processComponentsFactory) CreateTxSimulatorProcessor() (factory.TransactionSimulatorProcessor, process.VirtualMachinesContainerFactory, error) {
	return pcf.createTxSimulatorProcessor()
}

// SetChainRunType -
func (pcf *processComponentsFactory) SetChainRunType(chainRunType common.ChainRunType) {
	pcf.chainRunType = chainRunType
}
