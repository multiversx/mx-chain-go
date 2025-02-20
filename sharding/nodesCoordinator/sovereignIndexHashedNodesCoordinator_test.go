package nodesCoordinator

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignIndexHashedNodesCoordinator(t *testing.T) {
	t.Parallel()

	t.Run("invalid consensus size, should return error", func(t *testing.T) {
		args := createSovereignArgs()
		args.ShardConsensusGroupSize = 0

		ihnc, err := NewSovereignIndexHashedNodesCoordinator(args)
		require.Nil(t, ihnc)
		require.Equal(t, ErrInvalidConsensusGroupSize, err)
	})

	t.Run("invalid number of shards, should return error", func(t *testing.T) {
		args := createSovereignArgs()
		args.NbShards = 0

		ihnc, err := NewSovereignIndexHashedNodesCoordinator(args)
		require.Nil(t, ihnc)
		require.Equal(t, ErrInvalidNumberOfShards, err)
	})

	t.Run("small eligible list, should return error", func(t *testing.T) {
		args := createSovereignArgs()
		args.ShardConsensusGroupSize = 9999

		ihnc, err := NewSovereignIndexHashedNodesCoordinator(args)
		require.Nil(t, ihnc)
		require.Equal(t, ErrSmallShardEligibleListSize, err)
	})

	t.Run("should work", func(t *testing.T) {
		args := createSovereignArgs()
		ihnc, err := NewSovereignIndexHashedNodesCoordinator(args)
		require.Nil(t, err)
		require.False(t, ihnc.IsInterfaceNil())
	})
}

func TestSovereignIndexHashedNodesCoordinator_ComputeValidatorsGroup(t *testing.T) {
	t.Parallel()

	list := []Validator{
		newValidatorMock([]byte("pk0"), 1, defaultSelectionChances),
	}
	nodesMap := map[uint32][]Validator{
		core.SovereignChainShardId: list,
	}
	arguments := createArguments()
	arguments.EligibleNodes = nodesMap
	ihnc, _ := NewSovereignIndexHashedNodesCoordinator(arguments)

	t.Run("nil randomness, should return error", func(t *testing.T) {
		t.Parallel()

		list2, err := ihnc.ComputeConsensusGroup(nil, 0, core.SovereignChainShardId, 0)
		require.Empty(t, list2)
		require.Equal(t, ErrNilRandomness, err)
	})

	t.Run("invalid shard id, should return error", func(t *testing.T) {
		t.Parallel()

		list2, err := ihnc.ComputeConsensusGroup([]byte("randomness"), 0, core.MetachainShardId, 0)
		require.Empty(t, list2)
		require.Equal(t, ErrInvalidShardId, err)
	})

	t.Run("config not found for requested epoch, should return error", func(t *testing.T) {
		t.Parallel()

		list2, err := ihnc.ComputeConsensusGroup([]byte("randomness"), 0, core.SovereignChainShardId, 99999)
		require.Empty(t, list2)
		require.True(t, strings.Contains(err.Error(), ErrEpochNodesConfigDoesNotExist.Error()))
		require.True(t, strings.Contains(err.Error(), "99999"))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		list2, err := ihnc.ComputeConsensusGroup([]byte("randomness"), 0, core.SovereignChainShardId, 0)
		require.Equal(t, list, list2)
		require.Nil(t, err)
	})
}

func TestSovereignIndexHashedNodesCoordinator_GetConsensusValidatorsPublicKeys(t *testing.T) {
	t.Parallel()

	list := []Validator{
		newValidatorMock([]byte("pk0"), 1, defaultSelectionChances),
	}
	nodesMap := map[uint32][]Validator{
		core.SovereignChainShardId: list,
	}
	arguments := createArguments()
	arguments.EligibleNodes = nodesMap
	ihnc, _ := NewSovereignIndexHashedNodesCoordinator(arguments)

	t.Run("nil randomness, cannot compute consensus group, should return error", func(t *testing.T) {
		t.Parallel()

		list2, err := ihnc.GetConsensusValidatorsPublicKeys(nil, 0, core.SovereignChainShardId, 0)
		require.Empty(t, list2)
		require.Equal(t, ErrNilRandomness, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pubKeys, err := ihnc.GetConsensusValidatorsPublicKeys([]byte("randomness"), 0, core.SovereignChainShardId, 0)
		require.Equal(t, []string{string(list[0].PubKey())}, pubKeys)
		require.Nil(t, err)
	})
}
