package dataPool_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
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
	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	cache, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)
	headerNonces, err := dataPool.NewNonceToHashCacher(cache, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	txBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	peersBlock, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	stateBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	tdp, err := dataPool.NewTransientDataPool(
		nil,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilTxDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewTransientDataPool_NilHeadersShouldErr(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	cache, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)
	headerNonces, err := dataPool.NewNonceToHashCacher(cache, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	txBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	peersBlock, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	stateBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	tdp, err := dataPool.NewTransientDataPool(
		transactions,
		nil,
		headerNonces,
		txBlocks,
		peersBlock,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilHeadersDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewTransientDataPool_NilHeaderNoncesShouldErr(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers, err := shardedData.NewShardedData(testOkCfg)
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
		nil,
		txBlocks,
		peersBlock,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilHeadersNoncesDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewTransientDataPool_NilTxBlocksShouldErr(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	cache, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)
	headerNonces, err := dataPool.NewNonceToHashCacher(cache, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	peersBlock, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	stateBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	tdp, err := dataPool.NewTransientDataPool(
		transactions,
		headers,
		headerNonces,
		nil,
		peersBlock,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilTxBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewTransientDataPool_NilPeerBlocksShouldErr(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	cache, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)
	headerNonces, err := dataPool.NewNonceToHashCacher(cache, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	txBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	stateBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	tdp, err := dataPool.NewTransientDataPool(
		transactions,
		headers,
		headerNonces,
		txBlocks,
		nil,
		stateBlocks,
	)

	assert.Equal(t, data.ErrNilPeerChangeBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewTransientDataPool_NilStateBlocksShouldErr(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	cache, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)
	headerNonces, err := dataPool.NewNonceToHashCacher(cache, mock.NewNonceHashConverterMock())
	assert.Nil(t, err)

	txBlocks, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	peersBlock, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)

	tdp, err := dataPool.NewTransientDataPool(
		transactions,
		headers,
		headerNonces,
		txBlocks,
		peersBlock,
		nil,
	)

	assert.Equal(t, data.ErrNilStateBlockDataPool, err)
	assert.Nil(t, tdp)
}

func TestNewTransientDataPool_OkValsShouldWork(t *testing.T) {
	transactions, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	headers, err := shardedData.NewShardedData(testOkCfg)
	assert.Nil(t, err)

	cache, err := storage.NewCache(testOkCfg.Type, testOkCfg.Size)
	assert.Nil(t, err)
	headerNonces, err := dataPool.NewNonceToHashCacher(cache, mock.NewNonceHashConverterMock())
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
