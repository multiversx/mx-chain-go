package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

var testOkCfg = storage.CacheConfig{
	Size: 10,
	Type: storage.LRUCache,
}

//------- NewTransientDataPool

func TestNewTransientDataPool_NilTransactionsShouldErr(t *testing.T) {
	transactions := data.ShardedDataCacherNotifier(nil)

	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headerNonces, err := dataPool.NewNonceToHashCacher(testOkCfg)
	assert.Nil(t, err)

	txBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	peersBlock, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	stateBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	_, err = dataPool.NewTransientDataPool(
		transactions,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilTxDataPool, err)
}

func TestNewTransientDataPool_NilHeadersShouldErr(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers := data.ShardedDataCacherNotifier(nil)

	headerNonces, err := dataPool.NewNonceToHashCacher(testOkCfg)
	assert.Nil(t, err)

	txBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	peersBlock, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	stateBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	_, err = dataPool.NewTransientDataPool(
		transactions,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilHeadersDataPool, err)
}

func TestNewTransientDataPool_NilHeaderNoncesShouldErr(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headerNonces := data.Uint64Cacher(nil)

	txBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	peersBlock, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	stateBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	_, err = dataPool.NewTransientDataPool(
		transactions,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilHeadersNoncesDataPool, err)
}

func TestNewTransientDataPool_NilTxBlocksShouldErr(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headerNonces, err := dataPool.NewNonceToHashCacher(testOkCfg)
	assert.Nil(t, err)

	txBlocks := storage.Cacher(nil)

	peersBlock, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	stateBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	_, err = dataPool.NewTransientDataPool(
		transactions,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilTxBlockDataPool, err)
}

func TestNewTransientDataPool_NilPeerBlocksShouldErr(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headerNonces, err := dataPool.NewNonceToHashCacher(testOkCfg)
	assert.Nil(t, err)

	txBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	peersBlock := storage.Cacher(nil)

	stateBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	_, err = dataPool.NewTransientDataPool(
		transactions,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilPeerChangeBlockDataPool, err)
}

func TestNewTransientDataPool_NilStateBlocksShouldErr(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headerNonces, err := dataPool.NewNonceToHashCacher(testOkCfg)
	assert.Nil(t, err)

	txBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	peersBlock, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	stateBlocks := storage.Cacher(nil)

	_, err = dataPool.NewTransientDataPool(
		transactions,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilStateBlockDataPool, err)
}

func TestNewTransientDataPool_OkValsShouldWork(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headerNonces, err := dataPool.NewNonceToHashCacher(testOkCfg)
	assert.Nil(t, err)

	txBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	peersBlock, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	stateBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	tdp, err := dataPool.NewTransientDataPool(
		transactions,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
		stateBlocks,
	)

	assert.Nil(t, err)
	//pointer checking
	assert.True(t, transactions == tdp.Transactions())
	assert.True(t, headers == tdp.Headers())
	assert.True(t, headerNonces == tdp.HeadersNonces())
	assert.True(t, txBlocks == tdp.TxBlocks())
	assert.True(t, peersBlock == tdp.PeerChangesBlocks())
	assert.True(t, stateBlocks == tdp.StateBlocks())
}
