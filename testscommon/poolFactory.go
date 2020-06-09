package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/txpool"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

// CreateTxPool -
func CreateTxPool(numShards uint32, selfShard uint32) (dataRetriever.ShardedDataCacherNotifier, error) {
	return txpool.NewShardedTxPool(
		txpool.ArgShardedTxPool{
			Config: storageUnit.CacheConfig{
				Capacity:             100_000,
				SizePerSender:        1_000_000_000,
				SizeInBytes:          1_000_000_000,
				SizeInBytesPerSender: 33_554_432,
				Shards:               16,
			},
			MinGasPrice:    200000000000,
			NumberOfShards: numShards,
			SelfShardID:    selfShard,
		},
	)
}
