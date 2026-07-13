package track_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/track"
)

func TestBaseBlockTrack_IsSettledCrossHeader(t *testing.T) {
	t.Parallel()

	headerHash := []byte("headerHash")
	header := &block.Header{Nonce: 5, Round: 10}

	t.Run("nil header or empty hash returns false", func(t *testing.T) {
		t.Parallel()

		args, _, _ := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close()

		require.False(t, sbt.IsSettledCrossHeader(nil, headerHash))
		require.False(t, sbt.IsSettledCrossHeader(header, nil))
	})

	t.Run("no child known returns false", func(t *testing.T) {
		t.Parallel()

		args, _, _ := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close()

		require.False(t, sbt.IsSettledCrossHeader(header, headerHash))
	})

	t.Run("tracked proofed child extending the header settles it", func(t *testing.T) {
		t.Parallel()

		args, _, _ := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close()

		child := &block.Header{Nonce: 6, Round: 11, PrevHash: headerHash}
		sbt.AddTrackedHeader(child, []byte("childHash"))

		// no proof for the child yet
		require.False(t, sbt.IsSettledCrossHeader(header, headerHash))

		_ = args.PoolsHolder.Proofs().AddProof(&block.HeaderProof{HeaderHash: []byte("childHash"), HeaderNonce: 6, HeaderRound: 11})
		require.True(t, sbt.IsSettledCrossHeader(header, headerHash))
	})

	t.Run("proofed child of a sibling does not settle the header", func(t *testing.T) {
		t.Parallel()

		args, _, _ := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close()

		siblingChild := &block.Header{Nonce: 6, Round: 11, PrevHash: []byte("siblingHash")}
		sbt.AddTrackedHeader(siblingChild, []byte("siblingChildHash"))
		_ = args.PoolsHolder.Proofs().AddProof(&block.HeaderProof{HeaderHash: []byte("siblingChildHash"), HeaderNonce: 6, HeaderRound: 11})

		require.False(t, sbt.IsSettledCrossHeader(header, headerHash))
	})

	t.Run("proofed child known only from the headers pool settles it", func(t *testing.T) {
		t.Parallel()

		args, _, _ := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close()

		child := &block.Header{Nonce: 6, Round: 11, PrevHash: headerHash}
		args.PoolsHolder.Headers().AddHeader([]byte("childHash"), child)
		_ = args.PoolsHolder.Proofs().AddProof(&block.HeaderProof{HeaderHash: []byte("childHash"), HeaderNonce: 6, HeaderRound: 11})

		require.True(t, sbt.IsSettledCrossHeader(header, headerHash))
	})
}
