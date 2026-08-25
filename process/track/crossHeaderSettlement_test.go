package track_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
)

type settlementTracker interface {
	trackedHeaderAdder
	IsSettledCrossHeader(header data.HeaderHandler, headerHash []byte) bool
}

func TestBaseBlockTrack_IsSettledCrossHeader(t *testing.T) {
	t.Parallel()

	headerHash := []byte("headerHash")
	shardHeader := &block.Header{Nonce: 5, Round: 10}
	metaHeader := &block.MetaBlock{Nonce: 5, Round: 10}

	newTracker := func(t *testing.T) (track.ArgShardTracker, settlementTracker) {
		args, _, _ := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close()

		return args, sbt
	}

	t.Run("nil header or empty hash returns false", func(t *testing.T) {
		t.Parallel()

		_, sbt := newTracker(t)

		require.False(t, sbt.IsSettledCrossHeader(nil, headerHash))
		require.False(t, sbt.IsSettledCrossHeader(metaHeader, nil))
	})

	t.Run("meta header with no child known returns false", func(t *testing.T) {
		t.Parallel()

		_, sbt := newTracker(t)

		require.False(t, sbt.IsSettledCrossHeader(metaHeader, headerHash))
	})

	t.Run("tracked proofed child settles a meta header", func(t *testing.T) {
		t.Parallel()

		// the depth-1 meta rule that checkNotContendedUnsettled relies on
		args, sbt := newTracker(t)

		child := &block.MetaBlock{Nonce: 6, Round: 11, PrevHash: headerHash}
		sbt.AddTrackedHeader(child, []byte("childHash"))

		// no proof for the child yet
		require.False(t, sbt.IsSettledCrossHeader(metaHeader, headerHash))

		_ = args.PoolsHolder.Proofs().AddProof(&block.HeaderProof{
			HeaderHash:    []byte("childHash"),
			HeaderNonce:   6,
			HeaderRound:   11,
			HeaderShardId: core.MetachainShardId,
		})
		require.True(t, sbt.IsSettledCrossHeader(metaHeader, headerHash))
	})

	t.Run("tracked proofed child does not settle a shard header", func(t *testing.T) {
		t.Parallel()

		// a proofed shard child does not exclude a lower-round sibling
		args, sbt := newTracker(t)

		child := &block.Header{Nonce: 6, Round: 11, PrevHash: headerHash}
		sbt.AddTrackedHeader(child, []byte("childHash"))
		_ = args.PoolsHolder.Proofs().AddProof(&block.HeaderProof{
			HeaderHash:  []byte("childHash"),
			HeaderNonce: 6,
			HeaderRound: 11,
		})

		require.False(t, sbt.IsSettledCrossHeader(shardHeader, headerHash))
	})

	t.Run("proofed child of a sibling does not settle the meta header", func(t *testing.T) {
		t.Parallel()

		args, sbt := newTracker(t)

		siblingChild := &block.MetaBlock{Nonce: 6, Round: 11, PrevHash: []byte("siblingHash")}
		sbt.AddTrackedHeader(siblingChild, []byte("siblingChildHash"))
		_ = args.PoolsHolder.Proofs().AddProof(&block.HeaderProof{
			HeaderHash:    []byte("siblingChildHash"),
			HeaderNonce:   6,
			HeaderRound:   11,
			HeaderShardId: core.MetachainShardId,
		})

		require.False(t, sbt.IsSettledCrossHeader(metaHeader, headerHash))
	})

	t.Run("proofed child known only from the headers pool settles a meta header", func(t *testing.T) {
		t.Parallel()

		args, sbt := newTracker(t)

		child := &block.MetaBlock{Nonce: 6, Round: 11, PrevHash: headerHash}
		args.PoolsHolder.Headers().AddHeader([]byte("childHash"), child)
		_ = args.PoolsHolder.Proofs().AddProof(&block.HeaderProof{
			HeaderHash:    []byte("childHash"),
			HeaderNonce:   6,
			HeaderRound:   11,
			HeaderShardId: core.MetachainShardId,
		})

		require.True(t, sbt.IsSettledCrossHeader(metaHeader, headerHash))
	})

	t.Run("proofed child known only from the headers pool does not settle a shard header", func(t *testing.T) {
		t.Parallel()

		args, sbt := newTracker(t)

		child := &block.Header{Nonce: 6, Round: 11, PrevHash: headerHash}
		args.PoolsHolder.Headers().AddHeader([]byte("childHash"), child)
		_ = args.PoolsHolder.Proofs().AddProof(&block.HeaderProof{
			HeaderHash:  []byte("childHash"),
			HeaderNonce: 6,
			HeaderRound: 11,
		})

		require.False(t, sbt.IsSettledCrossHeader(shardHeader, headerHash))
	})
}

// a contended shard header holding a proofed child must not enter the ordinary longest chain, so
// the shard adds no header and meta routes it to arbitration instead
func TestMetaBlockTrack_ComputeLongestChain_ContendedShardHeader(t *testing.T) {
	t.Parallel()

	hasher := &hashingMocks.HasherMock{}
	marshaller := &mock.MarshalizerMock{}

	hashOf := func(header data.HeaderHandler) []byte {
		headerBytes, _ := marshaller.Marshal(header)

		return hasher.Compute(string(headerBytes))
	}

	parent := &block.Header{ShardID: 0, Nonce: 5, Round: 10}
	parentHash := hashOf(parent)

	newTracker := func(t *testing.T, headers []data.HeaderHandler) interface {
		ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte)
	} {
		args := CreateMetaTrackerMockArguments()
		args.Hasher = hasher
		args.Marshalizer = marshaller
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
				return flag == common.SupernovaFlag
			},
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
				return flag == common.AndromedaFlag || flag == common.SupernovaFlag
			},
		}
		args.EnableRoundsHandler = testscommon.NewEnableRoundsHandlerStub(common.SupernovaRoundFlag)
		// the tracker settles from PoolsHolder, the block processor attests finality from
		// ProofsPool; production wires both to the same pool
		args.ProofsPool = args.PoolsHolder.Proofs()

		mbt, err := track.NewMetaBlockTrack(args)
		require.Nil(t, err)
		_ = mbt.Close()

		for _, header := range headers {
			headerHash := hashOf(header)
			added := args.PoolsHolder.Proofs().AddProof(&block.HeaderProof{
				HeaderHash:    headerHash,
				HeaderNonce:   header.GetNonce(),
				HeaderRound:   header.GetRound(),
				HeaderShardId: header.GetShardID(),
			})
			require.True(t, added)

			mbt.AddTrackedHeader(header, headerHash)
		}

		return mbt
	}

	t.Run("contended shard header with a proofed child is not selected", func(t *testing.T) {
		t.Parallel()

		siblingLowRound := &block.Header{ShardID: 0, Nonce: 6, Round: 12, PrevHash: parentHash}
		// higher-round sibling, settled by its own proofed child under the old proofed-child rule
		contended := &block.Header{ShardID: 0, Nonce: 6, Round: 14, PrevHash: parentHash}
		child := &block.Header{ShardID: 0, Nonce: 7, Round: 15, PrevHash: hashOf(contended)}

		mbt := newTracker(t, []data.HeaderHandler{siblingLowRound, contended, child})

		headers, hashes := mbt.ComputeLongestChain(0, parent)

		require.Empty(t, headers)
		require.Empty(t, hashes)
	})

	t.Run("non-contended shard header is selected", func(t *testing.T) {
		t.Parallel()

		clean := &block.Header{ShardID: 0, Nonce: 6, Round: 11, PrevHash: parentHash}
		cleanHash := hashOf(clean)

		mbt := newTracker(t, []data.HeaderHandler{clean})

		headers, hashes := mbt.ComputeLongestChain(0, parent)

		require.Equal(t, []data.HeaderHandler{clean}, headers)
		require.Equal(t, [][]byte{cleanHash}, hashes)
	})
}
