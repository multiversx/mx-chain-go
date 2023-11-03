package sharding

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

func TestSovereignShardCoordinator_ComputeId(t *testing.T) {
	shardCoordinator := NewSovereignShardCoordinator(core.SovereignChainShardId)
	require.Equal(t, core.SovereignChainShardId, shardCoordinator.ComputeId([]byte("address")))

	require.Equal(t, uint32(1), shardCoordinator.NumberOfShards())
}

func TestNewSovereignShardCoordinator_SameShard(t *testing.T) {
	shardCoordinator := NewSovereignShardCoordinator(core.SovereignChainShardId)
	metaShardAddress := vm.ESDTSCAddress
	addr1 := getAddressFromUint32(uint32(1))
	addr2 := getAddressFromUint32(uint32(2))

	require.True(t, shardCoordinator.SameShard(metaShardAddress, addr1))
	require.True(t, shardCoordinator.SameShard(addr1, addr2))
	require.True(t, shardCoordinator.SameShard(addr2, metaShardAddress))
}
