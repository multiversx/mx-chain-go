package metachain

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/postprocess"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/multiversx/mx-chain-go/sharding"
)

type intermediateProcessorsContainerFactory struct {
	shardCoordinator      sharding.Coordinator
	marshalizer           marshal.Marshalizer
	hasher                hashing.Hasher
	pubkeyConverter       core.PubkeyConverter
	store                 dataRetriever.StorageService
	poolsHolder           dataRetriever.PoolsHolder
	economicsFee          process.FeeHandler
	enableEpochsHandler   common.EnableEpochsHandler
	executionOrderHandler common.TxExecutionOrderHandler
}

// ArgsNewIntermediateProcessorsContainerFactory defines the argument list to create a new container factory
type ArgsNewIntermediateProcessorsContainerFactory struct {
	ShardCoordinator        sharding.Coordinator
	Marshalizer             marshal.Marshalizer
	Hasher                  hashing.Hasher
	PubkeyConverter         core.PubkeyConverter
	Store                   dataRetriever.StorageService
	PoolsHolder             dataRetriever.PoolsHolder
	EconomicsFee            process.FeeHandler
	EnableEpochsHandler     common.EnableEpochsHandler
	TxExecutionOrderHandler common.TxExecutionOrderHandler
}

// NewIntermediateProcessorsContainerFactory is responsible for creating a new intermediate processors factory object
func NewIntermediateProcessorsContainerFactory(
	args ArgsNewIntermediateProcessorsContainerFactory,
) (*intermediateProcessorsContainerFactory, error) {

	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.PubkeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Store) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(args.PoolsHolder) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.TxExecutionOrderHandler) {
		return nil, process.ErrNilTxExecutionOrderHandler
	}

	return &intermediateProcessorsContainerFactory{
		shardCoordinator:      args.ShardCoordinator,
		marshalizer:           args.Marshalizer,
		hasher:                args.Hasher,
		pubkeyConverter:       args.PubkeyConverter,
		store:                 args.Store,
		poolsHolder:           args.PoolsHolder,
		economicsFee:          args.EconomicsFee,
		enableEpochsHandler:   args.EnableEpochsHandler,
		executionOrderHandler: args.TxExecutionOrderHandler,
	}, nil
}

// Create returns a preprocessor container that will hold all preprocessors in the system
func (ppcm *intermediateProcessorsContainerFactory) Create() (process.IntermediateProcessorContainer, error) {
	container := containers.NewIntermediateTransactionHandlersContainer()

	interproc, err := ppcm.createSmartContractResultsIntermediateProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.SmartContractResultBlock, interproc)
	if err != nil {
		return nil, err
	}

	interproc, err = ppcm.createBadTransactionsIntermediateProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.InvalidBlock, interproc)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ppcm *intermediateProcessorsContainerFactory) createSmartContractResultsIntermediateProcessor() (process.IntermediateTransactionHandler, error) {
	args := postprocess.ArgsNewIntermediateResultsProcessor{
		Hasher:                  ppcm.hasher,
		Marshalizer:             ppcm.marshalizer,
		Coordinator:             ppcm.shardCoordinator,
		PubkeyConv:              ppcm.pubkeyConverter,
		Store:                   ppcm.store,
		BlockType:               block.SmartContractResultBlock,
		CurrTxs:                 ppcm.poolsHolder.CurrentBlockTxs(),
		EconomicsFee:            ppcm.economicsFee,
		EnableEpochsHandler:     ppcm.enableEpochsHandler,
		TxExecutionOrderHandler: ppcm.executionOrderHandler,
	}
	irp, err := postprocess.NewIntermediateResultsProcessor(args)
	return irp, err
}

func (ppcm *intermediateProcessorsContainerFactory) createBadTransactionsIntermediateProcessor() (process.IntermediateTransactionHandler, error) {
	irp, err := postprocess.NewOneMiniBlockPostProcessor(
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.shardCoordinator,
		ppcm.store,
		block.InvalidBlock,
		dataRetriever.TransactionUnit,
		ppcm.economicsFee,
	)

	return irp, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *intermediateProcessorsContainerFactory) IsInterfaceNil() bool {
	return ppcm == nil
}
