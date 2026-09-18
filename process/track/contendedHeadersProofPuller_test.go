package track_test

import (
	"sync"
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

type pullRequestRecorder struct {
	mut      sync.Mutex
	requests []pullRequest
}

func (recorder *pullRequestRecorder) add(request pullRequest) {
	recorder.mut.Lock()
	recorder.requests = append(recorder.requests, request)
	recorder.mut.Unlock()
}

func (recorder *pullRequestRecorder) snapshot() []pullRequest {
	recorder.mut.Lock()
	defer recorder.mut.Unlock()

	return append([]pullRequest(nil), recorder.requests...)
}

type trackedHeaderAdder interface {
	AddTrackedHeader(header data.HeaderHandler, hash []byte)
}

func createProofPullTrackerScaffold(supernovaEnabled bool) (track.ArgShardTracker, *mock.RoundHandlerMock, *pullRequestRecorder) {
	args := CreateShardTrackerMockArguments()

	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledCalled: func(flag core.EnableEpochFlag) bool {
			return supernovaEnabled && flag == common.SupernovaFlag
		},
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
			return supernovaEnabled && flag == common.SupernovaFlag
		},
	}
	args.EnableRoundsHandler = testscommon.NewEnableRoundsHandlerStub(common.SupernovaRoundFlag)

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	args.RoundHandler = roundHandler

	requests := &pullRequestRecorder{}
	args.RequestHandler = &testscommon.RequestHandlerStub{
		RequestEquivalentProofByNonceCalled: func(headerShard uint32, headerNonce uint64) {
			requests.add(pullRequest{shardID: headerShard, nonce: headerNonce})
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

func TestBaseBlockTrack_PullProofsForContendedNonces(t *testing.T) {
	t.Parallel()

	t.Run("non-contended tip should not request", func(t *testing.T) {
		t.Parallel()

		args, _, requests := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close() // stop background loops; the test drives the pull explicitly

		addTipWithParent(sbt, 2) // round 2 = parent round + 1, no skipped round

		sbt.PullProofsForContendedNonces()
		require.Empty(t, requests.snapshot())
	})

	t.Run("contended tip should request once per round with backoff until settled", func(t *testing.T) {
		t.Parallel()

		args, roundHandler, requests := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close() // stop background loops; the test drives the pull explicitly

		addTipWithParent(sbt, 5) // rounds 2-4 skipped after parent round 1

		sbt.PullProofsForContendedNonces()
		require.Equal(t, []pullRequest{{shardID: 0, nonce: 2}}, requests.snapshot())

		// same round: no new request
		sbt.PullProofsForContendedNonces()
		require.Len(t, requests.snapshot(), 1)

		// backoff 1: next round fires
		roundHandler.RoundIndex = 11
		sbt.PullProofsForContendedNonces()
		require.Len(t, requests.snapshot(), 2)

		// backoff 2: round 12 skipped, round 13 fires
		roundHandler.RoundIndex = 12
		sbt.PullProofsForContendedNonces()
		require.Len(t, requests.snapshot(), 2)
		roundHandler.RoundIndex = 13
		sbt.PullProofsForContendedNonces()
		require.Len(t, requests.snapshot(), 3)

		// backoff 4: round 17 fires
		roundHandler.RoundIndex = 17
		sbt.PullProofsForContendedNonces()
		require.Len(t, requests.snapshot(), 4)

		// a shard child does not settle the contended header: pulling must continue at its nonce
		child := &block.Header{Nonce: 3, Round: 6, PrevHash: []byte("headHash")}
		sbt.AddTrackedHeader(child, []byte("childHash"))
		roundHandler.RoundIndex = 25
		sbt.PullProofsForContendedNonces()
		requestSnapshot := requests.snapshot()
		require.Len(t, requestSnapshot, 5)
		require.Equal(t, pullRequest{shardID: 0, nonce: 2}, requestSnapshot[4])

		// notarization passing the nonce concludes the arbitration and drops the state
		sbt.AddSelfNotarizedHeader(core.MetachainShardId, &block.Header{Nonce: 2}, []byte("notarizedHash"))
		roundHandler.RoundIndex = 40
		sbt.PullProofsForContendedNonces()
		require.Len(t, requests.snapshot(), 5)
	})

	t.Run("a child arriving before the first pull does not hide its contended parent", func(t *testing.T) {
		t.Parallel()

		args, _, requests := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close()

		// both land between pulls, e.g. batched delivery: the tip is the ordinary child
		addTipWithParent(sbt, 5) // contended at nonce 2
		child := &block.Header{Nonce: 3, Round: 6, PrevHash: []byte("tipHash")}
		sbt.AddTrackedHeader(child, []byte("childHash"))

		sbt.PullProofsForContendedNonces()
		require.Equal(t, []pullRequest{{shardID: 0, nonce: 2}}, requests.snapshot())
	})

	t.Run("two unresolved contended nonces pull independently", func(t *testing.T) {
		t.Parallel()

		args, roundHandler, requests := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close()

		addTipWithParent(sbt, 5) // contended at nonce 2

		sbt.PullProofsForContendedNonces()
		require.Equal(t, []pullRequest{{shardID: 0, nonce: 2}}, requests.snapshot())

		// the child skips rounds as well: a second contended nonce while the first is unresolved
		contendedChild := &block.Header{Nonce: 3, Round: 9, PrevHash: []byte("tipHash")}
		sbt.AddTrackedHeader(contendedChild, []byte("contendedChildHash"))

		roundHandler.RoundIndex = 20
		sbt.PullProofsForContendedNonces()
		requestSnapshot := requests.snapshot()
		require.Contains(t, requestSnapshot, pullRequest{shardID: 0, nonce: 2})
		require.Contains(t, requestSnapshot, pullRequest{shardID: 0, nonce: 3})
	})

	t.Run("supernova disabled should not request", func(t *testing.T) {
		t.Parallel()

		args, _, requests := createProofPullTrackerScaffold(false)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close() // stop background loops; the test drives the pull explicitly

		addTipWithParent(sbt, 5)

		sbt.PullProofsForContendedNonces()
		require.Empty(t, requests.snapshot())
	})

	t.Run("Supernova epoch active but round not active should not request", func(t *testing.T) {
		t.Parallel()

		args, _, requests := createProofPullTrackerScaffold(true)
		args.EnableRoundsHandler = &testscommon.EnableRoundsHandlerStub{
			IsFlagEnabledCalled: func(_ common.EnableRoundFlag) bool {
				return false
			},
			IsFlagEnabledInRoundCalled: func(_ common.EnableRoundFlag, _ uint64) bool {
				require.Fail(t, "header round should not be checked while Supernova round is inactive")
				return true
			},
		}
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close()

		addTipWithParent(sbt, 5)

		sbt.PullProofsForContendedNonces()
		require.Empty(t, requests.snapshot())
	})

	t.Run("contended tip with unknown parent should not request", func(t *testing.T) {
		t.Parallel()

		args, _, requests := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close() // stop background loops; the test drives the pull explicitly

		tip := &block.Header{Nonce: 2, Round: 5, PrevHash: []byte("unknownParent")}
		sbt.AddTrackedHeader(tip, []byte("tipHash"))

		sbt.PullProofsForContendedNonces()
		require.Empty(t, requests.snapshot())
	})

	t.Run("new contended tip resets the backoff", func(t *testing.T) {
		t.Parallel()

		args, roundHandler, requests := createProofPullTrackerScaffold(true)
		sbt, err := track.NewShardBlockTrack(args)
		require.Nil(t, err)
		_ = sbt.Close() // stop background loops; the test drives the pull explicitly

		addTipWithParent(sbt, 5)

		sbt.PullProofsForContendedNonces()
		roundHandler.RoundIndex = 11
		sbt.PullProofsForContendedNonces()
		require.Len(t, requests.snapshot(), 2)

		// a new contended tip at the next nonce fires immediately, regardless of prior backoff
		newTip := &block.Header{Nonce: 3, Round: 9, PrevHash: []byte("tipHash")}
		sbt.AddTrackedHeader(newTip, []byte("newTipHash"))

		roundHandler.RoundIndex = 12
		sbt.PullProofsForContendedNonces()
		requestSnapshot := requests.snapshot()
		require.Len(t, requestSnapshot, 3)
		require.Equal(t, pullRequest{shardID: 0, nonce: 3}, requestSnapshot[2])
	})
}
