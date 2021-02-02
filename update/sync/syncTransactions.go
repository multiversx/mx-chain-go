package sync

import (
	"context"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update"
)

var _ update.PendingTransactionsSyncHandler = (*pendingTransactions)(nil)

type pendingTransactions struct {
	mutPendingTx            sync.Mutex
	mapTransactions         map[string]data.TransactionHandler
	mapHashes               map[string]*block.MiniBlock
	txPools                 map[block.Type]dataRetriever.ShardedDataCacherNotifier
	storage                 map[block.Type]update.HistoryStorer
	chReceivedAll           chan bool
	requestHandler          process.RequestHandler
	marshalizer             marshal.Marshalizer
	epochToSync             uint32
	stopSync                bool
	syncedAll               bool
	waitTimeBetweenRequests time.Duration
}

// ArgsNewPendingTransactionsSyncer defines the arguments needed for a new transactions syncer
type ArgsNewPendingTransactionsSyncer struct {
	DataPools      dataRetriever.PoolsHolder
	Storages       dataRetriever.StorageService
	Marshalizer    marshal.Marshalizer
	RequestHandler process.RequestHandler
}

// NewPendingTransactionsSyncer creates a new transactions syncer
func NewPendingTransactionsSyncer(args ArgsNewPendingTransactionsSyncer) (*pendingTransactions, error) {
	if check.IfNil(args.Storages) {
		return nil, dataRetriever.ErrNilHeadersStorage
	}
	if check.IfNil(args.DataPools) {
		return nil, dataRetriever.ErrNilDataPoolHolder
	}
	if check.IfNil(args.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}

	p := &pendingTransactions{
		mutPendingTx:            sync.Mutex{},
		mapTransactions:         make(map[string]data.TransactionHandler),
		mapHashes:               make(map[string]*block.MiniBlock),
		chReceivedAll:           make(chan bool),
		requestHandler:          args.RequestHandler,
		marshalizer:             args.Marshalizer,
		stopSync:                true,
		syncedAll:               true,
		waitTimeBetweenRequests: args.RequestHandler.RequestInterval(),
	}

	p.txPools = make(map[block.Type]dataRetriever.ShardedDataCacherNotifier)
	p.txPools[block.TxBlock] = args.DataPools.Transactions()
	p.txPools[block.SmartContractResultBlock] = args.DataPools.UnsignedTransactions()
	p.txPools[block.RewardsBlock] = args.DataPools.RewardTransactions()

	p.storage = make(map[block.Type]update.HistoryStorer)
	p.storage[block.TxBlock] = args.Storages.GetStorer(dataRetriever.TransactionUnit)
	p.storage[block.SmartContractResultBlock] = args.Storages.GetStorer(dataRetriever.UnsignedTransactionUnit)
	p.storage[block.RewardsBlock] = args.Storages.GetStorer(dataRetriever.RewardTransactionUnit)

	for _, pool := range p.txPools {
		pool.RegisterOnAdded(p.receivedTransaction)
	}

	return p, nil
}

// SyncPendingTransactionsFor syncs pending transactions for a list of miniblocks
func (p *pendingTransactions) SyncPendingTransactionsFor(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error {
	_ = core.EmptyChannel(p.chReceivedAll)

	for {
		p.mutPendingTx.Lock()
		p.epochToSync = epoch
		p.syncedAll = false
		p.stopSync = false

		requestedTxs := 0
		for _, miniBlock := range miniBlocks {
			for _, txHash := range miniBlock.TxHashes {
				p.mapHashes[string(txHash)] = miniBlock
			}
			requestedTxs += p.requestTransactionsFor(miniBlock)
		}
		p.mutPendingTx.Unlock()

		if requestedTxs == 0 {
			p.mutPendingTx.Lock()
			p.stopSync = true
			p.syncedAll = true
			p.mutPendingTx.Unlock()
			return nil
		}

		select {
		case <-p.chReceivedAll:
			p.mutPendingTx.Lock()
			p.stopSync = true
			p.syncedAll = true
			p.mutPendingTx.Unlock()
			return nil
		case <-time.After(p.waitTimeBetweenRequests):
			continue
		case <-ctx.Done():
			p.mutPendingTx.Lock()
			p.stopSync = true
			p.mutPendingTx.Unlock()
			return update.ErrTimeIsOut
		}
	}
}

func (p *pendingTransactions) requestTransactionsFor(miniBlock *block.MiniBlock) int {
	missingTxs := make([][]byte, 0)
	for _, txHash := range miniBlock.TxHashes {
		if _, ok := p.mapTransactions[string(txHash)]; ok {
			continue
		}

		tx, ok := p.getTransactionFromPoolOrStorage(txHash)
		if ok {
			p.mapTransactions[string(txHash)] = tx
			continue
		}

		log.Debug("TODO REMOVE THIS - request transaction", "hash", txHash, "mb type", miniBlock.Type.String())
		missingTxs = append(missingTxs, txHash)
	}

	for _, txHash := range missingTxs {
		p.mapHashes[string(txHash)] = miniBlock
	}

	switch miniBlock.Type {
	case block.TxBlock:
		p.requestHandler.RequestTransaction(miniBlock.SenderShardID, missingTxs)
	case block.SmartContractResultBlock:
		p.requestHandler.RequestUnsignedTransactions(miniBlock.SenderShardID, missingTxs)
	case block.RewardsBlock:
		p.requestHandler.RequestRewardTransactions(miniBlock.SenderShardID, missingTxs)
	}

	return len(missingTxs)
}

// receivedMiniBlock is a callback function when a new transactions was received
func (p *pendingTransactions) receivedTransaction(txHash []byte, val interface{}) {
	p.mutPendingTx.Lock()
	if p.stopSync {
		p.mutPendingTx.Unlock()
		return
	}

	if _, ok := p.mapHashes[string(txHash)]; !ok {
		p.mutPendingTx.Unlock()
		return
	}
	if _, ok := p.mapTransactions[string(txHash)]; ok {
		p.mutPendingTx.Unlock()
		return
	}

	tx, ok := val.(data.TransactionHandler)
	if !ok {
		p.mutPendingTx.Unlock()
		return
	}

	p.mapTransactions[string(txHash)] = tx
	receivedAllMissing := len(p.mapHashes) == len(p.mapTransactions)
	p.mutPendingTx.Unlock()

	if receivedAllMissing {
		p.chReceivedAll <- true
	}
}

func (p *pendingTransactions) getTransactionFromPool(txHash []byte) (data.TransactionHandler, bool) {
	mb := p.mapHashes[string(txHash)]
	storeId := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	shardTxStore := p.txPools[mb.Type].ShardDataStore(storeId)
	if check.IfNil(shardTxStore) {
		return nil, false
	}

	val, ok := shardTxStore.Peek(txHash)
	if !ok {
		return nil, false
	}

	tx, ok := val.(data.TransactionHandler)
	if !ok {
		return nil, false
	}

	return tx, true
}

func (p *pendingTransactions) getTransactionFromPoolOrStorage(hash []byte) (data.TransactionHandler, bool) {
	txFromPool, ok := p.getTransactionFromPool(hash)
	if ok {
		return txFromPool, true
	}

	miniBlock, ok := p.mapHashes[string(hash)]
	if !ok {
		return nil, false
	}

	txData, err := GetDataFromStorage(hash, p.storage[miniBlock.Type])
	if err != nil {
		return nil, false
	}

	var tx data.TransactionHandler
	switch miniBlock.Type {
	case block.TxBlock:
		tx = &transaction.Transaction{}
	case block.SmartContractResultBlock:
		tx = &smartContractResult.SmartContractResult{}
	case block.RewardsBlock:
		tx = &rewardTx.RewardTx{}
	}

	err = p.marshalizer.Unmarshal(tx, txData)
	if err != nil {
		return nil, false
	}

	return tx, true
}

// GetTransactions returns the synced transactions
func (p *pendingTransactions) GetTransactions() (map[string]data.TransactionHandler, error) {
	p.mutPendingTx.Lock()
	defer p.mutPendingTx.Unlock()
	if !p.syncedAll {
		return nil, update.ErrNotSynced
	}

	return p.mapTransactions, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (p *pendingTransactions) IsInterfaceNil() bool {
	return p == nil
}
