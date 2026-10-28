package headerForBlock_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process/block/headerForBlock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
)

type requestRecorder struct {
	mut    sync.Mutex
	nonces []uint64
}

func (rr *requestRecorder) record(header data.HeaderHandler) {
	rr.mut.Lock()
	rr.nonces = append(rr.nonces, header.GetNonce())
	rr.mut.Unlock()
}

func (rr *requestRecorder) count() int {
	rr.mut.Lock()
	defer rr.mut.Unlock()
	return len(rr.nonces)
}

func (rr *requestRecorder) has(nonce uint64) bool {
	rr.mut.Lock()
	defer rr.mut.Unlock()
	for _, n := range rr.nonces {
		if n == nonce {
			return true
		}
	}
	return false
}

func createRequestOnProofArgs(recorder *requestRecorder) headerForBlock.ArgHeadersForBlock {
	args := createMockArgs()
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
			return flag == common.AndromedaFlag
		},
	}
	args.BlockTracker = &mock.BlockTrackerStub{
		GetLastCrossNotarizedHeaderCalled: func(_ uint32) (data.HeaderHandler, []byte, error) {
			return &block.MetaBlock{}, nil, nil
		},
	}
	args.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		RequestMiniBlocksAndTransactionsCalled: recorder.record,
	}
	args.RoundHandler = &testscommon.RoundHandlerMock{
		IndexForCurrentTimeCalled: func() int64 {
			return 1
		},
	}
	return args
}

func createTestMetaHeader(nonce uint64) (*block.MetaBlock, []byte) {
	return &block.MetaBlock{Nonce: nonce, Round: 1, Epoch: 2}, []byte(fmt.Sprintf("meta hash %d", nonce))
}

func createTestProof(headerHash []byte, nonce uint64) *block.HeaderProof {
	return &block.HeaderProof{
		HeaderHash:    headerHash,
		HeaderShardId: core.MetachainShardId,
		HeaderNonce:   nonce,
		HeaderEpoch:   2,
		HeaderRound:   1,
	}
}

func requireEventuallyNumRequests(t *testing.T, recorder *requestRecorder, expected int) {
	require.Eventually(t, func() bool {
		return recorder.count() == expected
	}, time.Second, 5*time.Millisecond)
}

func TestRequestMiniBlocksOnProof_PreAndromedaRequestsImmediately(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createRequestOnProofArgs(recorder)
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	hfb.RequestMiniBlocksOnProofIfNeeded(header, hash)

	requireEventuallyNumRequests(t, recorder, 1)
	require.Zero(t, hfb.NumPendingMbRequests())
}

func TestRequestMiniBlocksOnProof_PatienceIsTakenForTheHeaderRound(t *testing.T) {
	t.Parallel()

	providedRound := uint64(1)
	chRound := make(chan uint64, 1)

	recorder := &requestRecorder{}
	args := createRequestOnProofArgs(recorder)
	args.ProcessConfigsHandler = &testscommon.ProcessConfigsHandlerStub{
		GetExtraDelayForRequestBlockInfoCalled: func(round uint64) time.Duration {
			select {
			case chRound <- round:
			default:
			}
			return 0
		},
	}
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	require.Equal(t, providedRound, header.GetRound())
	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash, 1)))

	hfb.RequestMiniBlocksOnProofIfNeeded(header, hash)

	requireEventuallyNumRequests(t, recorder, 1)

	select {
	case round := <-chRound:
		require.Equal(t, providedRound, round)
	case <-time.After(time.Second):
		require.Fail(t, "the request patience was not looked up for the header round")
	}
}

func TestRequestMiniBlocksOnProof_ProofAlreadyPresentRequestsImmediately(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createRequestOnProofArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash, 1)))

	hfb.RequestMiniBlocksOnProofIfNeeded(header, hash)

	requireEventuallyNumRequests(t, recorder, 1)
	require.Zero(t, hfb.NumPendingMbRequests())
}

func TestRequestMiniBlocksOnProof_WaitsForProofThenRequestsOnce(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createRequestOnProofArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	hfb.RequestMiniBlocksOnProofIfNeeded(header, hash)

	time.Sleep(100 * time.Millisecond)
	require.Zero(t, recorder.count())
	require.Equal(t, 1, hfb.NumPendingMbRequests())

	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash, 1)))

	requireEventuallyNumRequests(t, recorder, 1)
	require.Zero(t, hfb.NumPendingMbRequests())

	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 1, recorder.count())
}

func TestRequestMiniBlocksOnProof_EndToEndThroughHeadersPool(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createRequestOnProofArgs(recorder)
	_, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	args.DataPool.Headers().AddHeader(hash, header)

	time.Sleep(100 * time.Millisecond)
	require.Zero(t, recorder.count())

	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash, 1)))

	requireEventuallyNumRequests(t, recorder, 1)
}

func TestRequestMiniBlocksOnProof_BehindNodeStillRequiresProof(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createRequestOnProofArgs(recorder)
	args.RoundHandler = &testscommon.RoundHandlerMock{
		IndexForCurrentTimeCalled: func() int64 {
			return 10
		},
		IndexCalled: func() int64 {
			return 1 // stale stored index; the arithmetic index must be used for the wait bypass
		},
	}
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	hfb.RequestMiniBlocksOnProofIfNeeded(header, hash)

	time.Sleep(100 * time.Millisecond)
	require.Zero(t, recorder.count()) // behind or not, no proof means no request
	require.Equal(t, 1, hfb.NumPendingMbRequests())

	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash, 1)))

	requireEventuallyNumRequests(t, recorder, 1)
	require.Zero(t, hfb.NumPendingMbRequests())
}

func TestRequestMiniBlocksOnProof_StaleEntriesDroppedNotRequested(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createRequestOnProofArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)
	hfb.SetPendingMbRequestMaxAge(50 * time.Millisecond)

	header1, hash1 := createTestMetaHeader(1)
	hfb.RequestMiniBlocksOnProofIfNeeded(header1, hash1)

	time.Sleep(100 * time.Millisecond)

	header2, hash2 := createTestMetaHeader(2)
	hfb.RequestMiniBlocksOnProofIfNeeded(header2, hash2)

	require.Equal(t, 1, hfb.NumPendingMbRequests()) // header1 dropped by cleanup, header2 pending

	// a late proof for the dropped entry must not trigger a request
	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash1, 1)))
	time.Sleep(100 * time.Millisecond)
	require.Zero(t, recorder.count())
}

func TestRequestMiniBlocksOnProof_DuplicateScheduleSingleRequest(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createRequestOnProofArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	hfb.RequestMiniBlocksOnProofIfNeeded(header, hash)
	hfb.RequestMiniBlocksOnProofIfNeeded(header, hash)
	require.Equal(t, 1, hfb.NumPendingMbRequests())

	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash, 1)))

	requireEventuallyNumRequests(t, recorder, 1)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 1, recorder.count())
}

func TestRequestMiniBlocksOnProof_CapEvictionDropsOldestWithoutRequest(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createRequestOnProofArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)
	hfb.SetMaxPendingMbRequests(2)

	for nonce := uint64(1); nonce <= 3; nonce++ {
		header, hash := createTestMetaHeader(nonce)
		hfb.RequestMiniBlocksOnProofIfNeeded(header, hash)
	}

	time.Sleep(100 * time.Millisecond)
	require.Zero(t, recorder.count()) // evicted oldest is dropped, never requested
	require.Equal(t, 2, hfb.NumPendingMbRequests())

	// remaining pending entries still dispatch on their proofs
	_, hash2 := createTestMetaHeader(2)
	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash2, 2)))
	require.Eventually(t, func() bool {
		return recorder.has(2)
	}, time.Second, 5*time.Millisecond)
	require.Equal(t, 1, recorder.count())
}

func TestRequestMiniBlocksOnProof_ConcurrentScheduleAndProof(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createRequestOnProofArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	numHeaders := 15 // stays within MaxHeadersToRequestInAdvance of the tracker stub's nonce 0
	wg := sync.WaitGroup{}
	for i := 1; i <= numHeaders; i++ {
		header, hash := createTestMetaHeader(uint64(i))
		proof := createTestProof(hash, uint64(i))
		wg.Add(2)
		go func() {
			hfb.RequestMiniBlocksOnProofIfNeeded(header, hash)
			wg.Done()
		}()
		go func() {
			args.DataPool.Proofs().AddProof(proof)
			wg.Done()
		}()
	}
	wg.Wait()

	requireEventuallyNumRequests(t, recorder, numHeaders)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, numHeaders, recorder.count())
	require.Zero(t, hfb.NumPendingMbRequests())
}
