package txpool

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/require"
)

func Test_CreateNewTxPool_ShardedData(t *testing.T) {
	config := storageUnit.CacheConfig{Type: storageUnit.FIFOShardedCache, Size: 100, SizeInBytes: 40960, Shards: 1}
	txPool, err := CreateTxPool(config, mock.NewEconomicsStub(100000000000000))
	require.Nil(t, err)
	require.NotNil(t, txPool)

	config = storageUnit.CacheConfig{Type: storageUnit.LRUCache, Size: 100, SizeInBytes: 40960, Shards: 1}
	txPool, err = CreateTxPool(config, mock.NewEconomicsStub(100000000000000))
	require.Nil(t, err)
	require.NotNil(t, txPool)
}

func Test_CreateNewTxPool_ShardedTxPool(t *testing.T) {
	config := storageUnit.CacheConfig{Size: 100, SizeInBytes: 40960, Shards: 1}
	txPool, err := CreateTxPool(config, mock.NewEconomicsStub(100000000000000))
	require.Nil(t, err)
	require.NotNil(t, txPool)
}
