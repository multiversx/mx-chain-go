package processing

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/txsimulator"
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
	arwenChangeLocker common.Locker,
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
		arwenChangeLocker,
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

// DecodeAddresses -
func DecodeAddresses(pkConverter core.PubkeyConverter, automaticCrawlerAddressesStrings []string) ([][]byte, error) {
	return factory.DecodeAddresses(pkConverter, automaticCrawlerAddressesStrings)
}
