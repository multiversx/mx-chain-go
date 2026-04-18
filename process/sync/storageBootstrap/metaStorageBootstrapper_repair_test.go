package storageBootstrap

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

func triggerStub(hash []byte) *testscommon.EpochStartTriggerStub {
	return &testscommon.EpochStartTriggerStub{
		EpochStartMetaHdrHashCalled: func() []byte { return hash },
	}
}

type shardCoordStub struct {
	sharding.Coordinator
	numShards uint32
}

func (s *shardCoordStub) NumberOfShards() uint32 { return s.numShards }

func buildStoreFor(t *testing.T, metaHash []byte, meta *block.MetaBlock) dataRetriever.StorageService {
	t.Helper()
	marsh := &mock.MarshalizerMock{}
	buf, err := marsh.Marshal(meta)
	require.NoError(t, err)
	return &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if string(key) == string(metaHash) {
						return buf, nil
					}
					return nil, storage.ErrKeyNotFound
				},
			}, nil
		},
	}
}

func TestMetaStorageBootstrapper_repairPendingMiniBlocks_NoOpWhenMatch(t *testing.T) {
	t.Parallel()

	hash := []byte("ES")
	mb := []byte("mb1")
	meta := &block.MetaBlock{
		Nonce: 10,
		Epoch: 5,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardID: 0,
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: mb, SenderShardID: 1, ReceiverShardID: 0},
					},
				},
			},
		},
	}

	replaceCalled := 0
	handler := &mock.PendingMiniBlocksHandlerStub{
		GetPendingMiniBlocksCalled: func(shardID uint32) [][]byte {
			if shardID == 0 {
				return [][]byte{mb}
			}
			return nil
		},
		ReplacePendingMiniBlocksForShardCalled: func(uint32, [][]byte) {
			replaceCalled++
		},
	}

	msb := &metaStorageBootstrapper{
		storageBootstrapper: &storageBootstrapper{
			marshalizer:       &mock.MarshalizerMock{},
			store:             buildStoreFor(t, hash, meta),
			shardCoordinator:  &shardCoordStub{numShards: 1},
			blockTracker:      &mock.BlockTrackerMock{},
			epochStartTrigger: triggerStub(hash),
		},
		pendingMiniBlocksHandler: handler,
	}

	err := msb.repairPendingMiniBlocks(hash)
	require.NoError(t, err)
	assert.Equal(t, 0, replaceCalled)
}

func TestMetaStorageBootstrapper_repairPendingMiniBlocks_OverridesWhenDiffer(t *testing.T) {
	t.Parallel()

	hash := []byte("ES")
	mbLoaded := []byte("stale")
	mbCorrect := []byte("correct")
	meta := &block.MetaBlock{
		Nonce: 10,
		Epoch: 5,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardID: 0,
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: mbCorrect, SenderShardID: 1, ReceiverShardID: 0},
					},
				},
			},
		},
	}

	captured := map[uint32][][]byte{}
	handler := &mock.PendingMiniBlocksHandlerStub{
		GetPendingMiniBlocksCalled: func(shardID uint32) [][]byte {
			if shardID == 0 {
				return [][]byte{mbLoaded}
			}
			return nil
		},
		ReplacePendingMiniBlocksForShardCalled: func(shardID uint32, hashes [][]byte) {
			captured[shardID] = hashes
		},
	}

	msb := &metaStorageBootstrapper{
		storageBootstrapper: &storageBootstrapper{
			marshalizer:       &mock.MarshalizerMock{},
			store:             buildStoreFor(t, hash, meta),
			shardCoordinator:  &shardCoordStub{numShards: 1},
			blockTracker:      &mock.BlockTrackerMock{},
			epochStartTrigger: triggerStub(hash),
		},
		pendingMiniBlocksHandler: handler,
	}

	err := msb.repairPendingMiniBlocks(hash)
	require.NoError(t, err)

	got, ok := captured[0]
	require.True(t, ok)
	require.Len(t, got, 1)
	assert.Equal(t, mbCorrect, got[0])
}

func TestMetaStorageBootstrapper_repairPendingMiniBlocks_OverridesWithEmptyWhenComputedEmpty(t *testing.T) {
	t.Parallel()

	// Loaded has entries, computed is empty: override with empty (clear).
	hash := []byte("ES")
	meta := &block.MetaBlock{
		Nonce: 10,
		Epoch: 5,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0}},
		},
	}

	captured := map[uint32]int{}
	handler := &mock.PendingMiniBlocksHandlerStub{
		GetPendingMiniBlocksCalled: func(shardID uint32) [][]byte {
			if shardID == 0 {
				return [][]byte{[]byte("ghost")}
			}
			return nil
		},
		ReplacePendingMiniBlocksForShardCalled: func(shardID uint32, hashes [][]byte) {
			captured[shardID] = len(hashes)
		},
	}

	msb := &metaStorageBootstrapper{
		storageBootstrapper: &storageBootstrapper{
			marshalizer:       &mock.MarshalizerMock{},
			store:             buildStoreFor(t, hash, meta),
			shardCoordinator:  &shardCoordStub{numShards: 1},
			blockTracker:      &mock.BlockTrackerMock{},
			epochStartTrigger: triggerStub(hash),
		},
		pendingMiniBlocksHandler: handler,
	}

	err := msb.repairPendingMiniBlocks(hash)
	require.NoError(t, err)

	got, ok := captured[0]
	require.True(t, ok)
	assert.Equal(t, 0, got)
}

func TestMetaStorageBootstrapper_repairPendingMiniBlocks_NoOverrideForShardWithoutAnchor(t *testing.T) {
	t.Parallel()

	// Last meta is an epoch-start whose EpochStart shard data only covers shard 0.
	// Shard 1 is absent from the computed map -> its loaded record must be
	// left untouched.
	hash := []byte("ES")
	meta := &block.MetaBlock{
		Nonce: 10,
		Epoch: 5,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardID: 0,
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: []byte("mb0"), SenderShardID: 1, ReceiverShardID: 0},
					},
				},
			},
		},
	}

	saved := map[uint32][][]byte{
		0: {[]byte("mb0")},
		1: {[]byte("mbSavedShard1")},
	}
	replaceCalls := map[uint32]int{}
	handler := &mock.PendingMiniBlocksHandlerStub{
		GetPendingMiniBlocksCalled: func(shardID uint32) [][]byte {
			return saved[shardID]
		},
		ReplacePendingMiniBlocksForShardCalled: func(shardID uint32, _ [][]byte) {
			replaceCalls[shardID]++
		},
	}

	msb := &metaStorageBootstrapper{
		storageBootstrapper: &storageBootstrapper{
			marshalizer:       &mock.MarshalizerMock{},
			store:             buildStoreFor(t, hash, meta),
			shardCoordinator:  &shardCoordStub{numShards: 2},
			blockTracker:      &mock.BlockTrackerMock{},
			epochStartTrigger: triggerStub(hash),
		},
		pendingMiniBlocksHandler: handler,
	}

	err := msb.repairPendingMiniBlocks(hash)
	require.NoError(t, err)
	assert.Equal(t, 0, replaceCalls[0], "shard 0 already matches, no override")
	assert.Equal(t, 0, replaceCalls[1], "shard 1 not computed, saved record must stay")
}

func TestMetaStorageBootstrapper_repairPendingMiniBlocks_PropagatesComputeError(t *testing.T) {
	t.Parallel()

	handler := &mock.PendingMiniBlocksHandlerStub{}
	msb := &metaStorageBootstrapper{
		storageBootstrapper: &storageBootstrapper{
			marshalizer: &mock.MarshalizerMock{},
			store: &storageStubs.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
					return &storageStubs.StorerStub{
						GetCalled: func(_ []byte) ([]byte, error) {
							return nil, storage.ErrKeyNotFound
						},
					}, nil
				},
			},
			shardCoordinator:  &shardCoordStub{numShards: 1},
			blockTracker:      &mock.BlockTrackerMock{},
			epochStartTrigger: triggerStub([]byte("anyHash")),
		},
		pendingMiniBlocksHandler: handler,
	}

	err := msb.repairPendingMiniBlocks([]byte("missing"))
	require.Error(t, err)
}

func TestPendingMiniBlocksEqual(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		a    [][]byte
		b    [][]byte
		want bool
	}{
		{"both nil", nil, nil, true},
		{"nil and empty", nil, [][]byte{}, true},
		{"different lengths", [][]byte{[]byte("x")}, nil, false},
		{"same order", [][]byte{[]byte("a"), []byte("b")}, [][]byte{[]byte("a"), []byte("b")}, true},
		{"different order", [][]byte{[]byte("a"), []byte("b")}, [][]byte{[]byte("b"), []byte("a")}, true},
		{"different contents", [][]byte{[]byte("a")}, [][]byte{[]byte("b")}, false},
		{"one side has extra", [][]byte{[]byte("a"), []byte("b")}, [][]byte{[]byte("a"), []byte("c")}, false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, pendingMiniBlocksEqual(tc.a, tc.b))
		})
	}
}
