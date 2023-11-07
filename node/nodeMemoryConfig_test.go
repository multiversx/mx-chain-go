package node

import (
	"os"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"
)

func TestMemoryConfig(t *testing.T) {
	nodeConfig := config.Config{}

	tomlBytes, err := os.ReadFile("../cmd/node/config/config.toml")
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
	plannedMemory += nodeConfig.LogsAndEvents.TxLogsStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.UnsignedTransactionStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.RewardTxStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.ShardHdrNonceHashStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.MetaHdrNonceHashStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.AccountsTrieStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.PeerAccountsTrieStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.ReceiptsStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.ScheduledSCRsStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.BadBlocksCache.SizeInBytes
	plannedMemory += nodeConfig.PeerBlockBodyDataPool.SizeInBytes
	plannedMemory += nodeConfig.TxDataPool.SizeInBytes
	plannedMemory += nodeConfig.TrieSyncStorage.SizeInBytes
	// One cache for each pair (shard, otherShard), including meta
	plannedMemory += nodeConfig.UnsignedTransactionDataPool.SizeInBytes * uint64(numShardsIncludingMeta*(numShardsIncludingMeta-1)) / 2
	// One cache for each pair (meta, shard)
	plannedMemory += nodeConfig.RewardTransactionDataPool.SizeInBytes * uint64(numShards)
	plannedMemory += nodeConfig.ValidatorInfoPool.SizeInBytes

	require.LessOrEqual(t, int(plannedMemory), 3000*core.MegabyteSize)
}
