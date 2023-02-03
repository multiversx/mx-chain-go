package processing

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/txsimulator"
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
	txSimulatorProcessorArgs *txsimulator.ArgsTxSimulator,
	wasmVMChangeLocker common.Locker,
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler,
	processedMiniBlocksTracker process.ProcessedMiniBlocksTracker,
	receiptsRepository factory.ReceiptsRepository,
) (process.BlockProcessor, process.VirtualMachinesContainerFactory, error) {
	blockProcessorComponents, err := pcf.newBlockProcessor(
		requestHandler,
		forkDetector,
		epochStartTrigger,
		bootStorer,
		validatorStatisticsProcessor,
		headerValidator,
		blockTracker,
		pendingMiniBlocksHandler,
		txSimulatorProcessorArgs,
		wasmVMChangeLocker,
		scheduledTxsExecutionHandler,
		processedMiniBlocksTracker,
		receiptsRepository,
	)
	if err != nil {
		return nil, nil, err
	}

	return blockProcessorComponents.blockProcessor, blockProcessorComponents.vmFactoryForTxSimulate, nil
}

// IndexGenesisBlocks -
func (pcf *processComponentsFactory) IndexGenesisBlocks(genesisBlocks map[uint32]data.HeaderHandler, indexingData map[uint32]*genesis.IndexingData) error {
	return pcf.indexGenesisBlocks(genesisBlocks, indexingData, map[string]*outport.AlteredAccount{})
}
