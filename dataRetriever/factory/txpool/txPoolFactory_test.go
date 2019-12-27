package txpool

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/require"
)

func Test_CreateNewTxPool_ShardedData(t *testing.T) {
	config := storageUnit.CacheConfig{Type: storageUnit.FIFOShardedCache, Size: 100, Shards: 1}
	txPool, err := CreateTxPool(config)
	require.Nil(t, err)
	require.NotNil(t, txPool)
}

func Test_CreateNewTxPool_ShardedTxPool(t *testing.T) {
	config := storageUnit.CacheConfig{Size: 100, Shards: 1}
	txPool, err := CreateTxPool(config)
	require.Nil(t, err)
	require.NotNil(t, txPool)
}
