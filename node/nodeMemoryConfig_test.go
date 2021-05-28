package node

import (
	"io/ioutil"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"
)

func TestMemoryConfig(t *testing.T) {
	nodeConfig := config.Config{}

	tomlBytes, err := ioutil.ReadFile("../cmd/node/config/config.toml")
	require.Nil(t, err)
	err = toml.Unmarshal(tomlBytes, &nodeConfig)
	require.Nil(t, err)

	numShards := 2
	numShardsIncludingMeta := 3

	plannedMemory := uint64(0)
	plannedMemory += nodeConfig.MiniBlocksStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.PeerBlockBodyStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.BlockHeaderStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.MetaBlockStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.TxStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.TxLogsStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.UnsignedTransactionStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.RewardTxStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.ShardHdrNonceHashStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.MetaHdrNonceHashStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.AccountsTrieStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.PeerAccountsTrieStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.ReceiptsStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.BadBlocksCache.SizeInBytes
	plannedMemory += nodeConfig.PeerBlockBodyDataPool.SizeInBytes
	plannedMemory += nodeConfig.TxDataPool.SizeInBytes
	plannedMemory += nodeConfig.TrieSyncStorage.SizeInBytes
	// One cache for each pair (shard, otherShard), including meta
	plannedMemory += nodeConfig.UnsignedTransactionDataPool.SizeInBytes * uint64(numShardsIncludingMeta*(numShardsIncludingMeta-1)) / 2
	// One cache for each pair (meta, shard)
	plannedMemory += nodeConfig.RewardTransactionDataPool.SizeInBytes * uint64(numShards)

	require.LessOrEqual(t, int(plannedMemory), 3000*core.MegabyteSize)
}
