package node

import (
	"io/ioutil"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/require"
)

func TestMemoryConfig(t *testing.T) {
	nodeConfig := config.Config{}

	tomlString, err := ioutil.ReadFile("../cmd/node/config/config.toml")
	require.Nil(t, err)
	require.True(t, len(tomlString) > 0)
	err = toml.Unmarshal([]byte(tomlString), &nodeConfig)
	require.Nil(t, err)

	numShards := 2

	plannedMemory := uint64(0)
	plannedMemory += nodeConfig.MiniBlocksStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.PeerBlockBodyStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.BlockHeaderStorage.Cache.SizeInBytes
	// todo: BootstrapStorage.Cache
	plannedMemory += nodeConfig.MetaBlockStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.TxStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.TxLogsStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.UnsignedTransactionStorage.Cache.SizeInBytes * uint64(2*numShards-1)
	plannedMemory += nodeConfig.RewardTxStorage.Cache.SizeInBytes
	// todo: StatusMetricsStorage.Cache
	plannedMemory += nodeConfig.ShardHdrNonceHashStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.MetaHdrNonceHashStorage.Cache.SizeInBytes
	plannedMemory += nodeConfig.AccountsTrieStorage.Cache.SizeInBytes
	// todo: EvictionWaitingList
	plannedMemory += nodeConfig.PeerAccountsTrieStorage.Cache.SizeInBytes
	// todo: HeadersPoolConfig
	plannedMemory += nodeConfig.BadBlocksCache.SizeInBytes
	plannedMemory += nodeConfig.PeerBlockBodyDataPool.SizeInBytes
	plannedMemory += nodeConfig.TxDataPool.SizeInBytes
	plannedMemory += nodeConfig.TrieNodesDataPool.SizeInBytes
	// todo: WhiteListPool
	// todo: WhiteListerVerifiedTxs
	plannedMemory += nodeConfig.UnsignedTransactionDataPool.SizeInBytes
	plannedMemory += nodeConfig.RewardTransactionDataPool.SizeInBytes
	// todo: PublicKeyShardId
	// todo: PeerIdShardId
	// todo: P2PMessageIDAdditionalCache
	// todo: antiflood cache
	// todo: heartbeat cache
	// todo: hardfork caches

	require.LessOrEqual(t, bToMb(plannedMemory), 2707)
}

func bToMb(b uint64) int {
	return int(b / 1024 / 1024)
}
