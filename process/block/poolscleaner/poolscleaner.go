package poolscleaner

import (
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// TxPoolsCleaner represents a pools cleaner that check if a transaction should be in pool
type TxPoolsCleaner struct {
	accounts         state.AccountsAdapter
	shardCoordinator sharding.Coordinator
	dataPool         dataRetriever.PoolsHolder
	addrConverter    state.AddressConverter
	numRemovedTxs    uint64
	canDoClean       chan struct{}
}

// NewTxsPoolsCleaner will return a new transaction pools cleaner
func NewTxsPoolsCleaner(
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	dataPool dataRetriever.PoolsHolder,
	addrConverter state.AddressConverter,
) (*TxPoolsCleaner, error) {
	if accounts == nil || accounts.IsInterfaceNil() {
		return nil, process.ErrNilAccountsAdapter
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}
	if dataPool == nil || dataPool.IsInterfaceNil() {
		return nil, process.ErrNilDataPoolHolder
	}
	transactionPool := dataPool.Transactions()
	if transactionPool == nil {
		return nil, process.ErrNilTransactionPool
	}
	if addrConverter == nil || addrConverter.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}

	canDoClean := make(chan struct{}, 1)

	return &TxPoolsCleaner{
		accounts:         accounts,
		shardCoordinator: shardCoordinator,
		dataPool:         dataPool,
		addrConverter:    addrConverter,
		numRemovedTxs:    0,
		canDoClean:       canDoClean,
	}, nil
}

// Clean will check if in pools exits transactions with nonce low that transaction sender account nonce
// and if tx have low nonce will be removed from pools
func (tpc *TxPoolsCleaner) Clean(duration time.Duration) (bool, error) {
	if duration == 0 {
		return false, process.ErrZeroCleaningTime
	}

	select {
	case tpc.canDoClean <- struct{}{}:
		startTime := time.Now()
		haveTime := func() bool {
			return time.Now().Sub(startTime) < duration
		}

		tpc.cleanPools(haveTime)
		<-tpc.canDoClean

		return true, nil
	default:
		return false, nil
	}
}

func (tpc *TxPoolsCleaner) cleanPools(haveTime func() bool) {
	shardId := tpc.shardCoordinator.SelfId()
	transactions := tpc.dataPool.Transactions()
	numOfShards := tpc.shardCoordinator.NumberOfShards()

	for destShardId := uint32(0); destShardId < numOfShards; destShardId++ {
		cacherId := process.ShardCacherIdentifier(shardId, destShardId)
		txsPool := transactions.ShardDataStore(cacherId)

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

			sndAddr := tx.GetSndAddress()
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

			accountNonce := accountHandler.GetNonce()
			txNonce := tx.Nonce
			lowerNonceInTx := txNonce < accountNonce
			if lowerNonceInTx {
				txsPool.Remove(key)
				atomic.AddUint64(&tpc.numRemovedTxs, 1)
			}
		}
	}
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
