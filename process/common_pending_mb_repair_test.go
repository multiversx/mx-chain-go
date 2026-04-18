package process_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

type storedBlocks struct {
	metas  map[string][]byte
	shards map[string][]byte
}

func newStoredBlocks() *storedBlocks {
	return &storedBlocks{metas: map[string][]byte{}, shards: map[string][]byte{}}
}

func (s *storedBlocks) storageService() dataRetriever.StorageService {
	return &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			if unitType == dataRetriever.MetaBlockUnit {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						if v, ok := s.metas[string(key)]; ok {
							return v, nil
						}
						return nil, storage.ErrKeyNotFound
					},
				}, nil
			}
			return &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					if v, ok := s.shards[string(key)]; ok {
						return v, nil
					}
					return nil, storage.ErrKeyNotFound
				},
			}, nil
		},
	}
}

func putMeta(t *testing.T, s *storedBlocks, hash []byte, mb *block.MetaBlock) {
	t.Helper()
	marsh := &mock.MarshalizerMock{}
	buf, err := marsh.Marshal(mb)
	require.NoError(t, err)
	s.metas[string(hash)] = buf
}

func putShard(t *testing.T, s *storedBlocks, hash []byte, hdr *block.Header) {
	t.Helper()
	marsh := &mock.MarshalizerMock{}
	buf, err := marsh.Marshal(hdr)
	require.NoError(t, err)
	s.shards[string(hash)] = buf
}

func mbFinal(t *testing.T, hash []byte, sender, receiver uint32, final bool) block.MiniBlockHeader {
	t.Helper()
	mbh := block.MiniBlockHeader{Hash: hash, SenderShardID: sender, ReceiverShardID: receiver}
	if !final {
		err := mbh.SetConstructionState(int32(block.Proposed))
		require.NoError(t, err)
	}
	return mbh
}

// --- tests -----------------------------------------------------------------

func TestRepairPendingMiniBlocks_LastIsEpochStartShortcut(t *testing.T) {
	t.Parallel()

	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}
	tracker := &mock.BlockTrackerMock{}

	hash := []byte("ES")
	meta := &block.MetaBlock{
		Nonce: 100,
		Epoch: 7,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardID: 0,
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: []byte("mbA"), SenderShardID: 1, ReceiverShardID: 0},
						{Hash: []byte("mbSame"), SenderShardID: 0, ReceiverShardID: 0},
					},
				},
				{
					ShardID: 1,
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: []byte("mbB"), SenderShardID: core.MetachainShardId, ReceiverShardID: 1},
					},
				},
			},
		},
	}
	putMeta(t, s, hash, meta)

	out, err := process.RepairPendingMiniBlocks(hash, nil, 2, tracker, marsh, s.storageService())
	require.NoError(t, err)
	require.Len(t, out[0], 1)
	assert.Equal(t, []byte("mbA"), out[0][0])
	require.Len(t, out[1], 1)
	assert.Equal(t, []byte("mbB"), out[1][0])
}

func TestRepairPendingMiniBlocks_ShardsCaughtUp_UsePerShardAnchor(t *testing.T) {
	t.Parallel()

	// epochStart at nonce 50. Shard 0's anchor is meta 51 (in the new epoch).
	// Last committed meta = 52. Per-shard anchor is used (shard is past epochStart).
	// Walk covers only meta 52 (nonce 51 == anchor, stops).
	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	esHash := []byte("ES50")
	es := &block.MetaBlock{Nonce: 50, EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0}}}}
	putMeta(t, s, esHash, es)

	anchorHash := []byte("M51")
	anchorMeta := &block.MetaBlock{
		Nonce:    51,
		PrevHash: esHash,
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{Hash: []byte("mbPending"), SenderShardID: 1, ReceiverShardID: 0},
				},
			},
		},
	}
	putMeta(t, s, anchorHash, anchorMeta)

	lastHash := []byte("L52")
	last := &block.MetaBlock{
		Nonce:    52,
		PrevHash: anchorHash,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: []byte("mbForward"), SenderShardID: core.MetachainShardId, ReceiverShardID: 0},
		},
	}
	putMeta(t, s, lastHash, last)

	sh0Hash := []byte("SH0")
	sh0 := &block.Header{Nonce: 10, ShardID: 0, MetaBlockHashes: [][]byte{anchorHash}}
	putShard(t, s, sh0Hash, sh0)

	tracker := &mock.BlockTrackerMock{
		GetLastSelfNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID == 0 {
				return anchorMeta, anchorHash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
		GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID == 0 {
				return sh0, sh0Hash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
	}

	out, err := process.RepairPendingMiniBlocks(lastHash, esHash, 1, tracker, marsh, s.storageService())
	require.NoError(t, err)
	hashes := map[string]bool{}
	for _, h := range out[0] {
		hashes[string(h)] = true
	}
	assert.True(t, hashes["mbPending"])
	assert.True(t, hashes["mbForward"])
	assert.Equal(t, 2, len(out[0]))
}

func TestRepairPendingMiniBlocks_ShardLagsBehindEpochStart_UsesEpochStartSeed(t *testing.T) {
	t.Parallel()

	// epochStart at nonce 50 (with pending {mbSeeded}), lastMeta at nonce 51.
	// Shard 0's anchor points at nonce 49 (epoch 3, pre-epochStart).
	// Effective anchor = epochStart (nonce 50). Per-shard walk uses seed from
	// epochStart, then toggles forward just meta 51.
	// A pre-epochStart meta with an MB for shard 0 must not leak through.
	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	preEsHash := []byte("M49")
	preEs := &block.MetaBlock{
		Nonce: 49,
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{Hash: []byte("mbPreEs"), SenderShardID: 1, ReceiverShardID: 0},
				},
			},
		},
	}
	putMeta(t, s, preEsHash, preEs)

	esHash := []byte("ES50")
	es := &block.MetaBlock{
		Nonce:    50,
		PrevHash: preEsHash,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{
					ShardID: 0,
					PendingMiniBlockHeaders: []block.MiniBlockHeader{
						{Hash: []byte("mbSeeded"), SenderShardID: 1, ReceiverShardID: 0},
					},
				},
			},
		},
	}
	putMeta(t, s, esHash, es)

	lastHash := []byte("L51")
	last := &block.MetaBlock{
		Nonce:    51,
		PrevHash: esHash,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Hash: []byte("mbForward"), SenderShardID: core.MetachainShardId, ReceiverShardID: 0},
		},
	}
	putMeta(t, s, lastHash, last)

	// shard 0's anchor is nonce 49 (lagging behind epochStart nonce 50)
	tracker := &mock.BlockTrackerMock{
		GetLastSelfNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID == 0 {
				return preEs, preEsHash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
		GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			return nil, nil, process.ErrMissingHeader
		},
	}

	out, err := process.RepairPendingMiniBlocks(lastHash, esHash, 1, tracker, marsh, s.storageService())
	require.NoError(t, err)
	hashes := map[string]bool{}
	for _, h := range out[0] {
		hashes[string(h)] = true
	}
	assert.True(t, hashes["mbSeeded"], "epochStart seed entry must be present")
	assert.True(t, hashes["mbForward"], "post-epochStart MB must be added")
	assert.False(t, hashes["mbPreEs"], "pre-epochStart MB must NOT leak through (wiped at epochStart commit)")
	assert.Equal(t, 2, len(out[0]))
}

func TestRepairPendingMiniBlocks_PerShardAnchorSubtractsIsFinal(t *testing.T) {
	t.Parallel()

	// epochStart at 50, anchor at 51 (past epochStart). Anchor has 2 MBs for shard 0;
	// shard 0's shard header marks one as final. Only the other remains pending.
	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	esHash := []byte("ES50")
	es := &block.MetaBlock{Nonce: 50, EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0}}}}
	putMeta(t, s, esHash, es)

	anchorHash := []byte("M51")
	anchorMeta := &block.MetaBlock{
		Nonce:    51,
		PrevHash: esHash,
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{Hash: []byte("mbDone"), SenderShardID: 1, ReceiverShardID: 0},
					{Hash: []byte("mbLeft"), SenderShardID: 1, ReceiverShardID: 0},
				},
			},
		},
	}
	putMeta(t, s, anchorHash, anchorMeta)

	sh0Hash := []byte("SH0")
	sh0 := &block.Header{
		Nonce:           10,
		ShardID:         0,
		MetaBlockHashes: [][]byte{anchorHash},
		MiniBlockHeaders: []block.MiniBlockHeader{
			mbFinal(t, []byte("mbDone"), 1, 0, true),
			mbFinal(t, []byte("mbLeft"), 1, 0, false),
		},
	}
	putShard(t, s, sh0Hash, sh0)

	tracker := &mock.BlockTrackerMock{
		GetLastSelfNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID == 0 {
				return anchorMeta, anchorHash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
		GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID == 0 {
				return sh0, sh0Hash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
	}

	out, err := process.RepairPendingMiniBlocks(anchorHash, esHash, 1, tracker, marsh, s.storageService())
	require.NoError(t, err)
	require.Len(t, out[0], 1)
	assert.Equal(t, []byte("mbLeft"), out[0][0])
}

func TestRepairPendingMiniBlocks_MixedShardAnchors(t *testing.T) {
	t.Parallel()

	// Two shards with different anchor nonces relative to epochStart.
	//   shard 0: anchor = 48 (< epochStart=50) -> effective anchor = epochStart.
	//   shard 1: anchor = 52 (> epochStart) -> own anchor.
	// Walk range covers nonces (50, 54]. Shard 1 only picks up MBs in (52, 54].
	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	sh0AnchorHash := []byte("M48")
	sh0Anchor := &block.MetaBlock{Nonce: 48}
	putMeta(t, s, sh0AnchorHash, sh0Anchor)

	esHash := []byte("ES50")
	es := &block.MetaBlock{
		Nonce:    50,
		PrevHash: sh0AnchorHash,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{
				{ShardID: 0, PendingMiniBlockHeaders: []block.MiniBlockHeader{{Hash: []byte("sh0Seed"), SenderShardID: 1, ReceiverShardID: 0}}},
				{ShardID: 1},
			},
		},
	}
	putMeta(t, s, esHash, es)

	m51Hash := []byte("M51")
	m51 := &block.MetaBlock{
		Nonce:    51,
		PrevHash: esHash,
		ShardInfo: []block.ShardData{
			{
				ShardID: 2,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					// for shard 0 (past its anchor=epochStart=50): added
					{Hash: []byte("sh0After"), SenderShardID: 2, ReceiverShardID: 0},
					// for shard 1 (anchor=52, 51 <= 52): NOT added
					{Hash: []byte("sh1Pre"), SenderShardID: 2, ReceiverShardID: 1},
				},
			},
		},
	}
	putMeta(t, s, m51Hash, m51)

	sh1AnchorHash := []byte("M52")
	sh1Anchor := &block.MetaBlock{Nonce: 52, PrevHash: m51Hash}
	putMeta(t, s, sh1AnchorHash, sh1Anchor)

	m53Hash := []byte("M53")
	m53 := &block.MetaBlock{
		Nonce:    53,
		PrevHash: sh1AnchorHash,
		MiniBlockHeaders: []block.MiniBlockHeader{
			// for shard 1, 53 > anchor 52: added
			{Hash: []byte("sh1After"), SenderShardID: core.MetachainShardId, ReceiverShardID: 1},
		},
	}
	putMeta(t, s, m53Hash, m53)

	lastHash := []byte("L54")
	last := &block.MetaBlock{Nonce: 54, PrevHash: m53Hash}
	putMeta(t, s, lastHash, last)

	// shard 0's cross-notarized shard header: empty, referencing epochStart
	sh0Hash := []byte("SH0")
	sh0 := &block.Header{Nonce: 5, ShardID: 0, MetaBlockHashes: [][]byte{esHash}}
	putShard(t, s, sh0Hash, sh0)
	sh1Hash := []byte("SH1")
	sh1 := &block.Header{Nonce: 5, ShardID: 1, MetaBlockHashes: [][]byte{sh1AnchorHash}}
	putShard(t, s, sh1Hash, sh1)

	tracker := &mock.BlockTrackerMock{
		GetLastSelfNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			switch shardID {
			case 0:
				return sh0Anchor, sh0AnchorHash, nil
			case 1:
				return sh1Anchor, sh1AnchorHash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
		GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			switch shardID {
			case 0:
				return sh0, sh0Hash, nil
			case 1:
				return sh1, sh1Hash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
	}

	out, err := process.RepairPendingMiniBlocks(lastHash, esHash, 2, tracker, marsh, s.storageService())
	require.NoError(t, err)

	sh0Hashes := map[string]bool{}
	for _, h := range out[0] {
		sh0Hashes[string(h)] = true
	}
	assert.True(t, sh0Hashes["sh0Seed"])
	assert.True(t, sh0Hashes["sh0After"])
	assert.Equal(t, 2, len(out[0]))

	sh1Hashes := map[string]bool{}
	for _, h := range out[1] {
		sh1Hashes[string(h)] = true
	}
	assert.True(t, sh1Hashes["sh1After"])
	assert.False(t, sh1Hashes["sh1Pre"])
	assert.Equal(t, 1, len(out[1]))
}

func TestRepairPendingMiniBlocks_NoEpochStartHashReturnsEmpty(t *testing.T) {
	t.Parallel()

	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	lastHash := []byte("L")
	last := &block.MetaBlock{Nonce: 10}
	putMeta(t, s, lastHash, last)

	tracker := &mock.BlockTrackerMock{}
	out, err := process.RepairPendingMiniBlocks(lastHash, nil, 1, tracker, marsh, s.storageService())
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestRepairPendingMiniBlocks_NoAnchorsReturnsEmpty(t *testing.T) {
	t.Parallel()

	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	esHash := []byte("ES50")
	es := &block.MetaBlock{Nonce: 50, EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0}}}}
	putMeta(t, s, esHash, es)
	lastHash := []byte("L51")
	last := &block.MetaBlock{Nonce: 51, PrevHash: esHash}
	putMeta(t, s, lastHash, last)

	tracker := &mock.BlockTrackerMock{
		GetLastSelfNotarizedHeaderCalled: func(uint32) (data.HeaderHandler, []byte, error) {
			return nil, nil, process.ErrMissingHeader
		},
	}
	out, err := process.RepairPendingMiniBlocks(lastHash, esHash, 2, tracker, marsh, s.storageService())
	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestRepairPendingMiniBlocks_MissingLastMetaReturnsError(t *testing.T) {
	t.Parallel()

	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}
	tracker := &mock.BlockTrackerMock{}
	_, err := process.RepairPendingMiniBlocks([]byte("missing"), []byte("ES"), 1, tracker, marsh, s.storageService())
	assert.Error(t, err)
}

func TestRepairPendingMiniBlocks_MissingEpochStartReturnsError(t *testing.T) {
	t.Parallel()

	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	lastHash := []byte("L")
	last := &block.MetaBlock{Nonce: 10}
	putMeta(t, s, lastHash, last)

	tracker := &mock.BlockTrackerMock{}
	_, err := process.RepairPendingMiniBlocks(lastHash, []byte("missing"), 1, tracker, marsh, s.storageService())
	assert.Error(t, err)
}

func TestRepairPendingMiniBlocks_MissingIntermediateMetaInForwardWalkReturnsError(t *testing.T) {
	t.Parallel()

	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	esHash := []byte("ES50")
	es := &block.MetaBlock{Nonce: 50, EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0}}}}
	putMeta(t, s, esHash, es)

	// lastMeta points via PrevHash to a meta that is not stored.
	lastHash := []byte("L52")
	last := &block.MetaBlock{Nonce: 52, PrevHash: []byte("missingMid")}
	putMeta(t, s, lastHash, last)

	tracker := &mock.BlockTrackerMock{
		GetLastSelfNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID == 0 {
				return es, esHash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
	}

	_, err := process.RepairPendingMiniBlocks(lastHash, esHash, 1, tracker, marsh, s.storageService())
	assert.Error(t, err)
}

func TestRepairPendingMiniBlocks_FilterAllEdgeCasesInForwardWalk(t *testing.T) {
	t.Parallel()

	// The forward walk must drop: same-shard, shard->meta, meta->all. Only a
	// normal cross-shard entry survives.
	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	esHash := []byte("ES50")
	es := &block.MetaBlock{
		Nonce:      50,
		EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0}}},
	}
	putMeta(t, s, esHash, es)

	lastHash := []byte("L51")
	last := &block.MetaBlock{
		Nonce:    51,
		PrevHash: esHash,
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{Hash: []byte("mbSame"), SenderShardID: 0, ReceiverShardID: 0},
					{Hash: []byte("mbShardToMeta"), SenderShardID: 0, ReceiverShardID: core.MetachainShardId},
					{Hash: []byte("mbMetaToAll"), SenderShardID: core.MetachainShardId, ReceiverShardID: core.AllShardId},
					{Hash: []byte("mbKeep"), SenderShardID: 1, ReceiverShardID: 0},
				},
			},
		},
	}
	putMeta(t, s, lastHash, last)

	tracker := &mock.BlockTrackerMock{
		GetLastSelfNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID == 0 {
				return es, esHash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
	}

	out, err := process.RepairPendingMiniBlocks(lastHash, esHash, 1, tracker, marsh, s.storageService())
	require.NoError(t, err)
	require.Len(t, out[0], 1)
	assert.Equal(t, []byte("mbKeep"), out[0][0])
}

func TestRepairPendingMiniBlocks_NoAnchorForShardLeavesEntryAbsent(t *testing.T) {
	t.Parallel()

	// With two shards but only shard 0 having an anchor, the output map must
	// contain shard 0 only; shard 1 must be absent so the caller skips
	// overriding shard 1's saved record.
	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	esHash := []byte("ES50")
	es := &block.MetaBlock{
		Nonce: 50,
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0}, {ShardID: 1}},
		},
	}
	putMeta(t, s, esHash, es)
	lastHash := []byte("L51")
	last := &block.MetaBlock{Nonce: 51, PrevHash: esHash}
	putMeta(t, s, lastHash, last)

	tracker := &mock.BlockTrackerMock{
		GetLastSelfNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID == 0 {
				return es, esHash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
	}

	out, err := process.RepairPendingMiniBlocks(lastHash, esHash, 2, tracker, marsh, s.storageService())
	require.NoError(t, err)
	_, ok0 := out[0]
	assert.True(t, ok0, "shard 0 should have an entry")
	_, ok1 := out[1]
	assert.False(t, ok1, "shard 1 must not have an entry when no anchor was found")
}

func TestRepairPendingMiniBlocks_LastMetaIsOwnAnchorNoForwardWalk(t *testing.T) {
	t.Parallel()

	// Shard 0's anchor IS the last committed meta: no forward walk runs.
	// Pending is fully determined by step 1 at the anchor.
	s := newStoredBlocks()
	marsh := &mock.MarshalizerMock{}

	esHash := []byte("ES50")
	es := &block.MetaBlock{Nonce: 50, EpochStart: block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0}}}}
	putMeta(t, s, esHash, es)

	lastHash := []byte("L51")
	last := &block.MetaBlock{
		Nonce:    51,
		PrevHash: esHash,
		ShardInfo: []block.ShardData{
			{
				ShardID: 1,
				ShardMiniBlockHeaders: []block.MiniBlockHeader{
					{Hash: []byte("mbAtLast"), SenderShardID: 1, ReceiverShardID: 0},
				},
			},
		},
	}
	putMeta(t, s, lastHash, last)

	sh0Hash := []byte("SH0")
	sh0 := &block.Header{Nonce: 10, ShardID: 0, MetaBlockHashes: [][]byte{lastHash}}
	putShard(t, s, sh0Hash, sh0)

	tracker := &mock.BlockTrackerMock{
		GetLastSelfNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID == 0 {
				return last, lastHash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
		GetLastCrossNotarizedHeaderCalled: func(shardID uint32) (data.HeaderHandler, []byte, error) {
			if shardID == 0 {
				return sh0, sh0Hash, nil
			}
			return nil, nil, process.ErrMissingHeader
		},
	}

	out, err := process.RepairPendingMiniBlocks(lastHash, esHash, 1, tracker, marsh, s.storageService())
	require.NoError(t, err)
	require.Len(t, out[0], 1)
	assert.Equal(t, []byte("mbAtLast"), out[0][0])
}
