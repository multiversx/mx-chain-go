package shard

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/factory/containers"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type preProcessorsContainerFactory struct {
	shardCoordinator  sharding.Coordinator
	store             dataRetriever.StorageService
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	dataPool          dataRetriever.PoolsHolder
	addrConverter     state.AddressConverter
	txProcessor       process.TransactionProcessor
	scProcessor       process.SmartContractProcessor
	scResultProcessor process.SmartContractResultProcessor
	accounts          state.AccountsAdapter
	requestHandler    process.RequestHandler
}

// NewPreProcessorsContainerFactory is responsible for creating a new preProcessors factory object
func NewPreProcessorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	dataPool dataRetriever.PoolsHolder,
	addrConverter state.AddressConverter,
	accounts state.AccountsAdapter,
	requestHandler process.RequestHandler,
	txProcessor process.TransactionProcessor,
	scProcessor process.SmartContractProcessor,
	scResultProcessor process.SmartContractResultProcessor,
) (*preProcessorsContainerFactory, error) {

	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if store == nil {
		return nil, process.ErrNilStore
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if dataPool == nil {
		return nil, process.ErrNilDataPoolHolder
	}
	if addrConverter == nil {
		return nil, process.ErrNilAddressConverter
	}
	if txProcessor == nil {
		return nil, process.ErrNilTxProcessor
	}
	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}
	if scProcessor == nil {
		return nil, process.ErrNilSmartContractProcessor
	}
	if scResultProcessor == nil {
		return nil, process.ErrNilSmartContractResultProcessor
	}
	if requestHandler == nil {
		return nil, process.ErrNilRequestHandler
	}

	return &preProcessorsContainerFactory{
		shardCoordinator:  shardCoordinator,
		store:             store,
		marshalizer:       marshalizer,
		hasher:            hasher,
		dataPool:          dataPool,
		addrConverter:     addrConverter,
		txProcessor:       txProcessor,
		accounts:          accounts,
		scProcessor:       scProcessor,
		scResultProcessor: scResultProcessor,
		requestHandler:    requestHandler,
	}, nil
}

// Create returns a preprocessor container that will hold all preprocessors in the system
func (ppcm *preProcessorsContainerFactory) Create() (process.PreProcessorsContainer, error) {
	container := containers.NewPreProcessorsContainer()

	preproc, err := ppcm.createTxPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.TxBlock, preproc)
	if err != nil {
		return nil, err
	}

	preproc, err = ppcm.createSmartContractResultPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.SmartContractResultBlock, preproc)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ppcm *preProcessorsContainerFactory) createTxPreProcessor() (process.PreProcessor, error) {
	txPreprocessor, err := preprocess.NewTransactionPreprocessor(
		ppcm.dataPool.Transactions(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.txProcessor,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestTransaction,
	)

	return txPreprocessor, err
}

func (ppcm *preProcessorsContainerFactory) createSmartContractResultPreProcessor() (process.PreProcessor, error) {
	scrPreprocessor, err := preprocess.NewSmartContractResultPreprocessor(
		ppcm.dataPool.SmartContractResults(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.scResultProcessor,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestTransaction,
	)

	return scrPreprocessor, err
}
