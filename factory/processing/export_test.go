package processing

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/cutoff"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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

// CreateAPITransactionEvaluator -
func (pcf *processComponentsFactory) CreateAPITransactionEvaluator() (factory.TransactionEvaluator, process.VirtualMachinesContainerFactory, error) {
	return pcf.createAPITransactionEvaluator()
}

// SetChainRunType -
func (pcf *processComponentsFactory) SetChainRunType(chainRunType common.ChainRunType) {
	pcf.chainRunType = chainRunType
}

// AddSystemVMToContainer -
func (pcf *processComponentsFactory) AddSystemVMToContainerIfNeeded(vmContainer process.VirtualMachinesContainer, builtInFuncFactory vmcommon.BuiltInFunctionFactory) error {
	return pcf.addSystemVMToContainerIfNeeded(vmContainer, builtInFuncFactory)
}
