package poolsCleaner

import (
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// TxPoolsCleaner represents a pools cleaner that check if a transaction should be in pool
type TxPoolsCleaner struct {
	accounts         state.AccountsAdapter
	shardCoordinator sharding.Coordinator
	dataPool         dataRetriever.PoolsHolder
	addrConverter    state.AddressConverter
	numRemovedTxs    uint64
	canDoClean       chan struct{}
	economicsFee     process.FeeHandler
}

// NewTxsPoolsCleaner will return a new transaction pools cleaner
func NewTxsPoolsCleaner(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	dataPool dataRetriever.PoolsHolder,
	addrConverter state.AddressConverter,
	economicsFee process.FeeHandler,
) (*TxPoolsCleaner, error) {
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(dataPool) {
		return nil, process.ErrNilDataPoolHolder
	}

	transactionPool := dataPool.Transactions()
	if check.IfNil(transactionPool) {
		return nil, process.ErrNilTransactionPool
	}
	if check.IfNil(addrConverter) {
		return nil, process.ErrNilAddressConverter
	}
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}

	canDoClean := make(chan struct{}, 1)

	return &TxPoolsCleaner{
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
		dataPool:         dataPool,
		addrConverter:    addrConverter,
		numRemovedTxs:    0,
		canDoClean:       canDoClean,
		economicsFee:     economicsFee,
	}, nil
}

// Clean removes the transactions with lower nonces than the senders' accounts.
func (tpc *TxPoolsCleaner) Clean(duration time.Duration) (bool, error) {
	if duration == 0 {
		return false, process.ErrZeroMaxCleanTime
	}

	select {
	case tpc.canDoClean <- struct{}{}:
		startTime := time.Now()
		haveTime := func() bool {
			return time.Since(startTime) < duration
		}

		tpc.cleanPools(haveTime)
		<-tpc.canDoClean

		return true, nil
	default:
		return false, nil
	}
}

// TODO, tx cache cleanup optimization:
// Getting all the keys of the cache (see below) can be pretty time consuming especially when the txs pool is full.
// We can redesign the cleanup for the new cache type so that we improve the processing time.
// One idea is that when cleaning executed tx hashes for a block, we can remove all the txs with lower nonce from the accounts-txs cache, for the respective account as well.
// https://github.com/ElrondNetwork/elrond-go/pull/863#discussion_r363641694
func (tpc *TxPoolsCleaner) cleanPools(haveTime func() bool) {
	atomic.StoreUint64(&tpc.numRemovedTxs, 0)

	shardId := tpc.shardCoordinator.SelfId()
	transactions := tpc.dataPool.Transactions()
	numOfShards := tpc.shardCoordinator.NumberOfShards()

	for destShardId := uint32(0); destShardId < numOfShards; destShardId++ {
		cacherId := process.ShardCacherIdentifier(shardId, destShardId)
		txsPool := transactions.ShardDataStore(cacherId)

		tpc.cleanDataStore(txsPool, haveTime)
	}
}

func (tpc *TxPoolsCleaner) cleanDataStore(txsPool storage.Cacher, haveTime func() bool) {
	for _, key := range txsPool.Keys() {
		if !haveTime() {
			return
		}

		obj, ok := txsPool.Peek(key)
		if !ok {
			continue
		}

		tx, ok := obj.(*transaction.Transaction)
		if !ok {
			atomic.AddUint64(&tpc.numRemovedTxs, 1)
			txsPool.Remove(key)
			continue
		}

		sndAddr := tx.GetSndAddr()
		addr, err := tpc.addrConverter.CreateAddressFromPublicKeyBytes(sndAddr)
		if err != nil {
			txsPool.Remove(key)
			atomic.AddUint64(&tpc.numRemovedTxs, 1)
			continue
		}

		accountHandler, err := tpc.accounts.GetExistingAccount(addr)
		if err != nil {
			txsPool.Remove(key)
			atomic.AddUint64(&tpc.numRemovedTxs, 1)
			continue
		}

		account, ok := accountHandler.(*state.Account)
		if !ok {
			txsPool.Remove(key)
			atomic.AddUint64(&tpc.numRemovedTxs, 1)
			continue
		}

		shouldClean := tpc.invalidTransaction(account, tx)
		if shouldClean {
			txsPool.Remove(key)
			atomic.AddUint64(&tpc.numRemovedTxs, 1)
		}

		//TODO maybe integrate here a TTL mechanism on each stored transactions as to not store transactions indefinitely
	}
}

func (tpc *TxPoolsCleaner) invalidTransaction(
	account *state.Account,
	tx *transaction.Transaction,
) bool {
	accountNonce := account.GetNonce()
	txNonce := tx.Nonce
	if txNonce < accountNonce {
		return true
	}

	txFee := tpc.economicsFee.ComputeFee(tx)
	return account.Balance.Cmp(txFee) < 0
}

// NumRemovedTxs will return the number of removed txs from pools
func (tpc *TxPoolsCleaner) NumRemovedTxs() uint64 {
	return atomic.LoadUint64(&tpc.numRemovedTxs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tpc *TxPoolsCleaner) IsInterfaceNil() bool {
	if tpc == nil {
		return true
	}
	return false
}
