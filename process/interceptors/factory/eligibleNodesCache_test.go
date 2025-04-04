package factory

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func TestNewEligibleNodesCache(t *testing.T) {
	t.Parallel()

	t.Run("nil PeerShardMapper should error", func(t *testing.T) {
		t.Parallel()

		cache, err := newEligibleNodesCache(nil, nil)
		require.Equal(t, process.ErrNilPeerShardMapper, err)
		require.Nil(t, cache)
	})
	t.Run("nil NodesCoordinator should error", func(t *testing.T) {
		t.Parallel()

		cache, err := newEligibleNodesCache(&mock.PeerShardMapperStub{}, nil)
		require.Equal(t, process.ErrNilNodesCoordinator, err)
		require.Nil(t, cache)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cache, err := newEligibleNodesCache(&mock.PeerShardMapperStub{}, &shardingMocks.NodesCoordinatorMock{})
		require.NoError(t, err)
		require.NotNil(t, cache)
	})
}

func TestEligibleNodesCache_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var cache *eligibleNodesCache
	require.True(t, cache.IsInterfaceNil())

	cache, _ = newEligibleNodesCache(&mock.PeerShardMapperStub{}, &shardingMocks.NodesCoordinatorMock{})
	require.False(t, cache.IsInterfaceNil())
}

func TestEligibleNodesCache_IsPeerEligible(t *testing.T) {
	t.Parallel()

	t.Run("nodes coordinator error on first call should error", func(t *testing.T) {
		t.Parallel()

		nodesCoordinator := &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return nil, expectedErr
			},
		}
		cache, err := newEligibleNodesCache(&mock.PeerShardMapperStub{}, nodesCoordinator)
		require.NoError(t, err)

		require.False(t, cache.IsPeerEligible("pid", 0, 0))
	})
	t.Run("nodes coordinator error on second call should error", func(t *testing.T) {
		t.Parallel()

		providedPid := core.PeerID("providedPid")
		providedPK := []byte("providedPK")

		cnt := 0
		nodesCoordinator := &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				cnt++
				if cnt > 1 {
					return nil, expectedErr
				}
				return []string{string(providedPK)}, nil
			},
		}
		peerShardMapper := &mock.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				require.Equal(t, providedPid, pid)
				return core.P2PPeerInfo{
					PkBytes: providedPK,
				}
			},
		}
		cache, err := newEligibleNodesCache(peerShardMapper, nodesCoordinator)
		require.NoError(t, err)

		require.True(t, cache.IsPeerEligible(providedPid, 0, 0))

		// new epoch, force second call
		require.False(t, cache.IsPeerEligible(providedPid, 0, 1))
	})
	t.Run("node not eligible should error", func(t *testing.T) {
		t.Parallel()

		providedPid := core.PeerID("providedPid")
		providedPK := []byte("providedPK")

		peerShardMapper := &mock.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				require.Equal(t, providedPid, pid)
				return core.P2PPeerInfo{
					PkBytes: providedPK,
				}
			},
		}
		nodesCoordinator := &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return []string{"otherEligible1", "otherEligible2"}, nil
			},
		}
		cache, err := newEligibleNodesCache(peerShardMapper, nodesCoordinator)
		require.NoError(t, err)

		require.False(t, cache.IsPeerEligible(providedPid, 0, 0))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedPid := core.PeerID("providedPid")
		providedPK := []byte("providedPK")

		peerShardMapper := &mock.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				require.Equal(t, providedPid, pid)
				return core.P2PPeerInfo{
					PkBytes: providedPK,
				}
			},
		}
		cntGetAllEligibleValidatorsPublicKeysForShardCalled := 0
		nodesCoordinator := &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				cntGetAllEligibleValidatorsPublicKeysForShardCalled++
				return []string{string(providedPK)}, nil
			},
		}
		cache, err := newEligibleNodesCache(peerShardMapper, nodesCoordinator)
		require.NoError(t, err)

		// should not return from cache, new shard and epoch
		require.True(t, cache.IsPeerEligible(providedPid, 0, 0))

		// should return from cache, same epoch and shard
		require.True(t, cache.IsPeerEligible(providedPid, 0, 0))

		// should not return from cache, new shard
		require.True(t, cache.IsPeerEligible(providedPid, 1, 0))

		// should return from cache, same epoch and shard
		require.True(t, cache.IsPeerEligible(providedPid, 1, 0))

		// should not return from cache, new epoch in delta limits
		require.True(t, cache.IsPeerEligible(providedPid, 1, 1))

		// expecting calls to nodesCoordinator:
		//	- first call on the initial proof
		//	- second call on the third proof, which has a different shard
		// 	- third call on the last proof, which has a different epoch
		expectedCalls := 3
		require.Equal(t, expectedCalls, cntGetAllEligibleValidatorsPublicKeysForShardCalled)

		// new epoch for shard 1 was in delta limits, no reset
		eligibleMap := cache.eligibleNodesMap
		require.Equal(t, 2, len(eligibleMap)) // 2 shards
		eligibleListForShard0ByEpoch, ok := eligibleMap[0]
		require.True(t, ok) // must have shard 0
		eligibleListForShard1ByEpoch, ok := eligibleMap[1]
		require.True(t, ok) // must have shard 1

		require.Equal(t, 1, len(eligibleListForShard0ByEpoch)) // tests ran with epoch 0 only
		require.Equal(t, 2, len(eligibleListForShard1ByEpoch)) // new epoch in delta limits

		shard0EligibleNodes, ok := eligibleListForShard0ByEpoch[0]
		require.True(t, ok) // must have one eligible for epoch 0
		require.Equal(t, 1, len(shard0EligibleNodes))

		shardMetaEligibleNodesEpoch0, ok := eligibleListForShard1ByEpoch[0]
		require.True(t, ok) // must have one eligible for epoch 0
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch0))

		shardMetaEligibleNodesEpoch1, ok := eligibleListForShard1ByEpoch[1]
		require.True(t, ok) // must have one eligible for epoch 1
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch1))
	})
	t.Run("should work extreme scenarios", func(t *testing.T) {
		t.Parallel()

		providedPid := core.PeerID("providedPid")
		providedPK := []byte("providedPK")

		peerShardMapper := &mock.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				require.Equal(t, providedPid, pid)
				return core.P2PPeerInfo{
					PkBytes: providedPK,
				}
			},
		}
		nodesCoordinator := &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return []string{string(providedPK)}, nil
			},
		}
		cache, err := newEligibleNodesCache(peerShardMapper, nodesCoordinator)
		require.NoError(t, err)

		// =================================================================
		// testing higher epochs

		// 1st call: should not return from cache, new shard
		require.True(t, cache.IsPeerEligible(providedPid, 1, 30))

		// 2nd call: should return from cache, same epoch and shard
		require.True(t, cache.IsPeerEligible(providedPid, 1, 30))

		// 3rd call: should not return from cache, new epoch in delta limits
		require.True(t, cache.IsPeerEligible(providedPid, 1, 32))

		// 4th call: should not return from cache, new epoch out of delta limits
		require.True(t, cache.IsPeerEligible(providedPid, 1, 34))

		// =================================================================
		// checking results
		eligibleMap := cache.eligibleNodesMap
		require.Equal(t, 1, len(eligibleMap)) // 1 shard
		eligibleListForShard1ByEpoch, ok := eligibleMap[1]
		require.True(t, ok) // must have shard 1

		// 4th call was outside of delta limits, should keep it and the prev one
		require.Equal(t, 2, len(eligibleListForShard1ByEpoch))

		_, ok = eligibleListForShard1ByEpoch[30]
		require.False(t, ok) // epoch 30 should have been cleaned, too old

		shardMetaEligibleNodesEpoch32, ok := eligibleListForShard1ByEpoch[32]
		require.True(t, ok) // must have one eligible for epoch 32
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch32))

		shardMetaEligibleNodesEpoch34, ok := eligibleListForShard1ByEpoch[34]
		require.True(t, ok) // must have one eligible for epoch 34
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch34))
		// =================================================================

		// 5th call: should not return from cache, new epoch out of delta limits
		require.True(t, cache.IsPeerEligible(providedPid, 1, 50))

		// =================================================================
		// checking results
		require.Equal(t, 1, len(eligibleMap)) // 1 shard
		eligibleListForShard1ByEpoch, ok = eligibleMap[1]
		require.True(t, ok) // must have shard 1

		// 5th call was outside of delta limits by too much, should have deleted the others
		require.Equal(t, 1, len(eligibleListForShard1ByEpoch))

		shardMetaEligibleNodesEpoch50, ok := eligibleListForShard1ByEpoch[50]
		require.True(t, ok) // must have one eligible for epoch 50
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch50))
		// =================================================================

		// =================================================================
		// testing lower epochs

		// 6th call: should not return from cache, new epoch inside delta
		require.True(t, cache.IsPeerEligible(providedPid, 1, 49))

		// 7th call: should not return from cache, new epoch inside delta
		require.True(t, cache.IsPeerEligible(providedPid, 1, 48))

		// =================================================================
		// checking results
		require.Equal(t, 1, len(eligibleMap)) // 1 shard
		eligibleListForShard1ByEpoch, ok = eligibleMap[1]
		require.True(t, ok) // must have shard 1

		// all calls were inside delta
		require.Equal(t, 3, len(eligibleListForShard1ByEpoch))

		shardMetaEligibleNodesEpoch50, ok = eligibleListForShard1ByEpoch[50]
		require.True(t, ok) // must have one eligible for epoch 50
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch50))

		shardMetaEligibleNodesEpoch49, ok := eligibleListForShard1ByEpoch[49]
		require.True(t, ok) // must have one eligible for epoch 49
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch49))

		shardMetaEligibleNodesEpoch48, ok := eligibleListForShard1ByEpoch[48]
		require.True(t, ok) // must have one eligible for epoch 48
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch48))

		// =================================================================

		// 8th call: should not return from cache, new epoch outside delta, should delete epoch 50
		require.True(t, cache.IsPeerEligible(providedPid, 1, 47))

		// =================================================================
		// checking results
		require.Equal(t, 1, len(eligibleMap)) // 1 shard
		eligibleListForShard1ByEpoch, ok = eligibleMap[1]
		require.True(t, ok) // must have shard 1

		// deleted one, added one
		require.Equal(t, 3, len(eligibleListForShard1ByEpoch))

		_, ok = eligibleListForShard1ByEpoch[50]
		require.False(t, ok)

		shardMetaEligibleNodesEpoch49, ok = eligibleListForShard1ByEpoch[49]
		require.True(t, ok) // must have one eligible for epoch 49
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch49))

		shardMetaEligibleNodesEpoch48, ok = eligibleListForShard1ByEpoch[48]
		require.True(t, ok) // must have one eligible for epoch 48
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch48))

		shardMetaEligibleNodesEpoch47, ok := eligibleListForShard1ByEpoch[47]
		require.True(t, ok) // must have one eligible for epoch 47
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch47))

		// =================================================================

		// 9th call: should not return from cache, new epoch outside delta
		require.True(t, cache.IsPeerEligible(providedPid, 1, 40))

		// =================================================================
		// checking results
		require.Equal(t, 1, len(eligibleMap)) // 1 shard
		eligibleListForShard1ByEpoch, ok = eligibleMap[1]
		require.True(t, ok) // must have shard 1

		// 10th call was outside of delta limits by too much, should have deleted the others
		require.Equal(t, 1, len(eligibleListForShard1ByEpoch))

		shardMetaEligibleNodesEpoch40, ok := eligibleListForShard1ByEpoch[40]
		require.True(t, ok) // must have one eligible for epoch 40
		require.Equal(t, 1, len(shardMetaEligibleNodesEpoch40))
		// =================================================================
	})
	t.Run("should work, concurrent calls", func(t *testing.T) {
		t.Parallel()

		providedPidShard0 := core.PeerID("providedPid_0")
		providedPKShard0 := []byte("providedPK_0")
		providedPidShardMeta := core.PeerID("providedPid_Meta")
		providedPKShardMeta := []byte("providedPK_Meta")

		peerShardMapper := &mock.PeerShardMapperStub{
			GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
				switch pid {
				case providedPidShard0:
					return core.P2PPeerInfo{
						PkBytes: providedPKShard0,
					}
				case providedPidShardMeta:
					return core.P2PPeerInfo{
						PkBytes: providedPKShardMeta,
					}
				default:
					require.Fail(t, fmt.Sprintf("should have not been called for pid %s", pid))
				}

				return core.P2PPeerInfo{}
			},
		}
		nodesCoordinator := &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				switch shardID {
				case 0:
					return []string{string(providedPKShard0)}, nil
				case core.MetachainShardId:
					return []string{string(providedPKShardMeta)}, nil
				default:
					require.Fail(t, fmt.Sprintf("should have not been called for shard %d", shardID))
				}

				return nil, expectedErr
			},
		}
		cache, err := newEligibleNodesCache(peerShardMapper, nodesCoordinator)
		require.NoError(t, err)

		numCalls := 3000
		wg := sync.WaitGroup{}
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(idx int) {
				shard := uint32(0)
				pid := providedPidShard0
				// 3000 calls => 2000 proofs for shard 0, 1000 for meta
				if idx%3 == 0 {
					shard = core.MetachainShardId
					pid = providedPidShardMeta
				}

				// same epoch 0 each time
				require.True(t, cache.IsPeerEligible(pid, shard, 0))

				wg.Done()
			}(i)
		}

		wg.Wait()

		eligibleMap := cache.eligibleNodesMap
		require.Equal(t, 2, len(eligibleMap)) // 2 shards
		eligibleListForShard0ByEpoch, ok := eligibleMap[0]
		require.True(t, ok) // must have shard 0
		eligibleListForShardMetaByEpoch, ok := eligibleMap[core.MetachainShardId]
		require.True(t, ok) // must have shard meta

		require.Equal(t, 1, len(eligibleListForShard0ByEpoch))    // tests ran with epoch 0 only
		require.Equal(t, 1, len(eligibleListForShardMetaByEpoch)) // tests ran with epoch 0 only

		shard0EligibleNodes, ok := eligibleListForShard0ByEpoch[0]
		require.True(t, ok) // must have one eligible for epoch 0
		require.Equal(t, 1, len(shard0EligibleNodes))
		_, ok = shard0EligibleNodes[string(providedPKShard0)]
		require.True(t, ok) // must have only the provided shard 0 pk

		shardMetaEligibleNodes, ok := eligibleListForShardMetaByEpoch[0]
		require.True(t, ok) // must have one eligible for epoch 0
		require.Equal(t, 1, len(shardMetaEligibleNodes))
		_, ok = shardMetaEligibleNodes[string(providedPKShardMeta)]
		require.True(t, ok) // must have only the provided shard meta pk
	})
}
