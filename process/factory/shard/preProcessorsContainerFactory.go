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
	shardCoordinator   sharding.Coordinator
	store              dataRetriever.StorageService
	marshalizer        marshal.Marshalizer
	hasher             hashing.Hasher
	dataPool           dataRetriever.PoolsHolder
	addrConverter      state.AddressConverter
	txProcessor        process.TransactionProcessor
	scProcessor        process.SmartContractProcessor
	scResultProcessor  process.SmartContractResultProcessor
	rewardsTxProcessor process.RewardTransactionProcessor
	accounts           state.AccountsAdapter
	requestHandler     process.RequestHandler
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
	rewardsTxProcessor process.RewardTransactionProcessor,
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
	if addrConverter == nil || addrConverter.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if txProcessor == nil || txProcessor.IsInterfaceNil() {
		return nil, process.ErrNilTxProcessor
	}
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if scProcessor == nil || scProcessor.IsInterfaceNil() {
		return nil, process.ErrNilSmartContractProcessor
	}
	if scResultProcessor == nil || scResultProcessor.IsInterfaceNil() {
		return nil, process.ErrNilSmartContractResultProcessor
	}
    if rewardsTxProcessor == nil || rewardsTxProcessor.IsInterfaceNil() {
        return nil, process.ErrNilRewardsTxProcessor
    }
	if requestHandler == nil || requestHandler.IsInterfaceNil() {
		return nil, process.ErrNilRequestHandler
	}

	return &preProcessorsContainerFactory{
		shardCoordinator:   shardCoordinator,
		store:              store,
		marshalizer:        marshalizer,
		hasher:             hasher,
		dataPool:           dataPool,
		addrConverter:      addrConverter,
		txProcessor:        txProcessor,
		accounts:           accounts,
		scProcessor:        scProcessor,
		scResultProcessor:  scResultProcessor,
		rewardsTxProcessor: rewardsTxProcessor,
		requestHandler:     requestHandler,
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

	preproc, err = ppcm.createRewardsTransactionPreProcessor()
	if err != nil {
		return nil, err
	}

	err = container.Add(block.RewardsBlock, preproc)
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
		ppcm.dataPool.UnsignedTransactions(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.scResultProcessor,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestUnsignedTransactions,
	)

	return scrPreprocessor, err
}

func (ppcm *preProcessorsContainerFactory) createRewardsTransactionPreProcessor() (process.PreProcessor, error) {
	rewardTxPreprocessor, err := preprocess.NewRewardTxPreprocessor(
		ppcm.dataPool.RewardTransactions(),
		ppcm.store,
		ppcm.hasher,
		ppcm.marshalizer,
		ppcm.rewardsTxProcessor,
		ppcm.shardCoordinator,
		ppcm.accounts,
		ppcm.requestHandler.RequestRewardTransactions,
	)

	return rewardTxPreprocessor, err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppcm *preProcessorsContainerFactory) IsInterfaceNil() bool {
    if ppcm == nil {
        return true
    }
    return false
}
