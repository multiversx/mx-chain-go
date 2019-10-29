package metachain

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
	shardCoordinator sharding.Coordinator
	store            dataRetriever.StorageService
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	dataPool         dataRetriever.MetaPoolsHolder
	txProcessor      process.TransactionProcessor
	accounts         state.AccountsAdapter
	requestHandler   process.RequestHandler
	economicsFee     process.FeeHandler
}

// NewPreProcessorsContainerFactory is responsible for creating a new preProcessors factory object
func NewPreProcessorsContainerFactory(
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	dataPool dataRetriever.MetaPoolsHolder,
	accounts state.AccountsAdapter,
	requestHandler process.RequestHandler,
	txProcessor process.TransactionProcessor,
	economicsFee process.FeeHandler,
) (*preProcessorsContainerFactory, error) {

	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if store == nil || store.IsInterfaceNil() {
		return nil, process.ErrNilStore
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if dataPool == nil || dataPool.IsInterfaceNil() {
		return nil, process.ErrNilDataPoolHolder
	}
	if txProcessor == nil || txProcessor.IsInterfaceNil() {
		return nil, process.ErrNilTxProcessor
	}
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if requestHandler == nil || requestHandler.IsInterfaceNil() {
		return nil, process.ErrNilRequestHandler
	}
	if economicsFee == nil || economicsFee.IsInterfaceNil() {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	return &preProcessorsContainerFactory{
		shardCoordinator: shardCoordinator,
		store:            store,
		marshalizer:      marshalizer,
		hasher:           hasher,
		dataPool:         dataPool,
		txProcessor:      txProcessor,
		accounts:         accounts,
		requestHandler:   requestHandler,
		economicsFee:     economicsFee,
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
		ppcm.economicsFee,
	)

	return txPreprocessor, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *preProcessorsContainerFactory) IsInterfaceNil() bool {
	if ppcm == nil {
		return true
	}
	return false
}
