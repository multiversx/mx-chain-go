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
			GetCachedEpochsCalled: func() map[uint32]struct{} {
				return map[uint32]struct{}{
					0: {},
				}
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

		// expecting calls to nodesCoordinator:
		//	- first call on the initial proof
		//	- second call on the third proof, which has a different shard
		expectedCalls := 2
		require.Equal(t, expectedCalls, cntGetAllEligibleValidatorsPublicKeysForShardCalled)

		// same epoch, 2 shards
		eligibleMap := cache.eligibleNodesMap
		require.Equal(t, 1, len(eligibleMap))
		require.Equal(t, 2, len(eligibleMap[0])) // 2 shards
		eligibleListForShard0ByEpoch, ok := eligibleMap[0][0]
		require.True(t, ok) // must have shard 0
		eligibleListForShard1ByEpoch, ok := eligibleMap[0][1]
		require.True(t, ok) // must have shard 1

		require.Equal(t, 1, len(eligibleListForShard0ByEpoch)) // tests ran with epoch 0 only
		require.Equal(t, 1, len(eligibleListForShard1ByEpoch)) // tests ran with epoch 0 only

		_, ok = eligibleListForShard0ByEpoch[string(providedPK)]
		require.True(t, ok) // must have the eligible node

		_, ok = eligibleListForShard1ByEpoch[string(providedPK)]
		require.True(t, ok) // must have  the eligible node
	})
	t.Run("should work with cache cleanup", func(t *testing.T) {
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
		cachedEpochs := make(map[uint32]struct{})
		nodesCoordinator := &shardingMocks.NodesCoordinatorMock{
			GetAllEligibleValidatorsPublicKeysForShardCalled: func(epoch uint32, shardID uint32) ([]string, error) {
				return []string{string(providedPK)}, nil
			},
			GetCachedEpochsCalled: func() map[uint32]struct{} {
				return cachedEpochs
			},
		}
		cache, err := newEligibleNodesCache(peerShardMapper, nodesCoordinator)
		require.NoError(t, err)

		// 1st call: should not return from cache, new shard
		require.True(t, cache.IsPeerEligible(providedPid, 1, 30))

		// 2nd call: should return from cache, same epoch and shard
		require.True(t, cache.IsPeerEligible(providedPid, 1, 30))

		cachedEpochs = map[uint32]struct{}{
			30: {}, // old calls
			32: {}, // keep nodes coordinator aligned with the upcoming call
		}

		// 3rd call: should not return from cache, new epoch
		require.True(t, cache.IsPeerEligible(providedPid, 1, 32))

		cachedEpochs = map[uint32]struct{}{
			// 30: {}, <- deleted epoch 30 from nodesCoordinator, should remove it from local cache too
			32: {}, // keep nodes coordinator aligned with the upcoming call
		}

		// 4th call: should not return from cache, same epoch but new shard
		require.True(t, cache.IsPeerEligible(providedPid, 2, 32))

		eligibleMap := cache.eligibleNodesMap
		require.Equal(t, 1, len(eligibleMap)) // one epoch, 32
		nodesPerShardInEpoch, ok := eligibleMap[32]
		require.True(t, ok)
		require.Equal(t, 2, len(nodesPerShardInEpoch)) // 2 shards on epoch 32
		for _, eligibleInShard := range nodesPerShardInEpoch {
			require.Equal(t, 1, len(eligibleInShard)) // one node in each shard
			_, ok := eligibleInShard[string(providedPK)]
			require.True(t, ok)
		}
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
			GetCachedEpochsCalled: func() map[uint32]struct{} {
				// this test will only run with 0 and Meta
				return map[uint32]struct{}{
					0:                     {},
					core.MetachainShardId: {},
				}
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
		require.Equal(t, 1, len(eligibleMap)) // 1 epoch
		eligibleListForEpoch0ByShard, ok := eligibleMap[0]
		require.True(t, ok) // must have epoch 0
		shard0EligibleNodes, ok := eligibleListForEpoch0ByShard[0]
		require.True(t, ok) // must have eligible for epoch 0 shard 0
		_, ok = shard0EligibleNodes[string(providedPKShard0)]
		require.True(t, ok) // must have only the provided shard 0 pk
		shardMetaEligibleNodes, ok := eligibleListForEpoch0ByShard[core.MetachainShardId]
		require.True(t, ok) // must have eligible for epoch 0 shard meta
		_, ok = shardMetaEligibleNodes[string(providedPKShardMeta)]
		require.True(t, ok) // must have only the provided shard meta pk
	})
}
