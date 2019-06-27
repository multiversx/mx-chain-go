package shard

import (
	"github.com/ElrondNetwork/elrond-go/crypto"
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
	shardCoordinator    sharding.Coordinator
	messenger           process.TopicHandler
	store               dataRetriever.StorageService
	marshalizer         marshal.Marshalizer
	hasher              hashing.Hasher
	keyGen              crypto.KeyGenerator
	singleSigner        crypto.SingleSigner
	multiSigner         crypto.MultiSigner
	dataPool            dataRetriever.PoolsHolder
	addrConverter       state.AddressConverter
	chronologyValidator process.ChronologyValidator
	txProcessor         process.TransactionProcessor
	scProcessor         process.SmartContractProcessor
	scResultProcessor   process.SmartContractResultProcessor
	interProcessor      process.IntermediateTransactionHandler
	accounts            state.AccountsAdapter
	requestHandler      process.RequestHandler
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
	interProcessor process.IntermediateTransactionHandler,
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
	if interProcessor == nil {
		return nil, process.ErrNilIntermediateTransactionHandler
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
		interProcessor:    interProcessor,
		requestHandler:    requestHandler,
	}, nil
}

func (ppcm *preProcessorsContainerFactory) Create() (process.PreProcessorsContainer, error) {
	container := containers.NewPreProcessorsContainer()

	key, preproc, err := ppcm.createTxPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(key, preproc)
	if err != nil {
		return nil, err
	}

	key, preproc, err = ppcm.createSmartContractResultPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(key, preproc)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (ppcm *preProcessorsContainerFactory) createTxPreProcessor() (block.Type, process.PreProcessor, error) {
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

	return block.TxBlock, txPreprocessor, err
}

func (ppcm *preProcessorsContainerFactory) createSmartContractResultPreProcessor() (block.Type, process.PreProcessor, error) {
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

	return block.SmartContractResultBlock, scrPreprocessor, err
}
