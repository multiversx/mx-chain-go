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
)

type pullRequest struct {
	shardID uint32
	nonce   uint64
}

type trackedHeaderAdder interface {
	AddTrackedHeader(header data.HeaderHandler, hash []byte)
}

func createProofPullTrackerScaffold(supernovaEnabled bool) (track.ArgShardTracker, *mock.RoundHandlerMock, *[]pullRequest) {
	args := CreateShardTrackerMockArguments()

	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return supernovaEnabled && flag == common.SupernovaFlag
		},
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return false
		},
	}

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	args.RoundHandler = roundHandler

	requests := &[]pullRequest{}
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestEquivalentProofByNonceCalled: func(headerShard uint32, headerNonce uint64) {
			*requests = append(*requests, pullRequest{shardID: headerShard, nonce: headerNonce})
		},
	}

	return args, roundHandler, requests
}

func addTipWithParent(sbt trackedHeaderAdder, tipRound uint64) {
	parentHash := []byte("parentHash")
	parent := &block.Header{Nonce: 1, Round: 1}
	tip := &block.Header{Nonce: 2, Round: tipRound, PrevHash: parentHash}

	sbt.AddTrackedHeader(parent, parentHash)
	sbt.AddTrackedHeader(tip, []byte("tipHash"))
}

func TestBaseBlockTrack_PullProofsForContendedTips(t *testing.T) {
	t.Parallel()

	t.Run("non-contended tip should not request", func(t *testing.T) {
		t.Parallel()

		args, _, requests := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close() // stop background loops; the test drives the pull explicitly

		addTipWithParent(sbt, 2) // round 2 = parent round + 1, no skipped round

		sbt.PullProofsForContendedTips()
		require.Empty(t, *requests)
	})

	t.Run("contended tip should request once per round with backoff until settled", func(t *testing.T) {
		t.Parallel()

		args, roundHandler, requests := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close() // stop background loops; the test drives the pull explicitly

		addTipWithParent(sbt, 5) // rounds 2-4 skipped after parent round 1

		sbt.PullProofsForContendedTips()
		require.Equal(t, []pullRequest{{shardID: 0, nonce: 2}}, *requests)

		// same round: no new request
		sbt.PullProofsForContendedTips()
		require.Equal(t, 1, len(*requests))

		// backoff 1: next round fires
		roundHandler.RoundIndex = 11
		sbt.PullProofsForContendedTips()
		require.Equal(t, 2, len(*requests))

		// backoff 2: round 12 skipped, round 13 fires
		roundHandler.RoundIndex = 12
		sbt.PullProofsForContendedTips()
		require.Equal(t, 2, len(*requests))
		roundHandler.RoundIndex = 13
		sbt.PullProofsForContendedTips()
		require.Equal(t, 3, len(*requests))

		// settled by a proofed child extending the tip: no more requests; the round is advanced
		// before AddProof, which dispatches the tracker's receivedProof subscriber on another goroutine
		roundHandler.RoundIndex = 17
		_ = args.PoolsHolder.Proofs().AddProof(&block.HeaderProof{
			HeaderHash:    []byte("childHash"),
			HeaderNonce:   3,
			HeaderRound:   6,
			HeaderShardId: 0,
		})
		// the child is known from the headers pool only, so the contended header stays the tracked tip
		child := &block.Header{Nonce: 3, Round: 6, PrevHash: []byte("tipHash")}
		args.PoolsHolder.Headers().AddHeader([]byte("childHash"), child)
		sbt.PullProofsForContendedTips()
		require.Equal(t, 3, len(*requests))
	})

	t.Run("supernova disabled should not request", func(t *testing.T) {
		t.Parallel()

		args, _, requests := createProofPullTrackerScaffold(false)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close() // stop background loops; the test drives the pull explicitly

		addTipWithParent(sbt, 5)

		sbt.PullProofsForContendedTips()
		require.Empty(t, *requests)
	})

	t.Run("contended tip with unknown parent should not request", func(t *testing.T) {
		t.Parallel()

		args, _, requests := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close() // stop background loops; the test drives the pull explicitly

		tip := &block.Header{Nonce: 2, Round: 5, PrevHash: []byte("unknownParent")}
		sbt.AddTrackedHeader(tip, []byte("tipHash"))

		sbt.PullProofsForContendedTips()
		require.Empty(t, *requests)
	})

	t.Run("new contended tip resets the backoff", func(t *testing.T) {
		t.Parallel()

		args, roundHandler, requests := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close() // stop background loops; the test drives the pull explicitly

		addTipWithParent(sbt, 5)

		sbt.PullProofsForContendedTips()
		roundHandler.RoundIndex = 11
		sbt.PullProofsForContendedTips()
		require.Equal(t, 2, len(*requests))

		// a new contended tip at the next nonce fires immediately, regardless of prior backoff
		newTip := &block.Header{Nonce: 3, Round: 9, PrevHash: []byte("tipHash")}
		sbt.AddTrackedHeader(newTip, []byte("newTipHash"))

		roundHandler.RoundIndex = 12
		sbt.PullProofsForContendedTips()
		require.Equal(t, 3, len(*requests))
		require.Equal(t, pullRequest{shardID: 0, nonce: 3}, (*requests)[2])
	})
}
