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

func createSchedulingArgs(recorder *requestRecorder) headerForBlock.ArgHeadersForBlock {
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

func TestScheduleMiniBlocksRequest_PreAndromedaRequestsImmediately(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createSchedulingArgs(recorder)
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	hfb.ScheduleMiniBlocksRequestIfNeeded(header, hash)

	requireEventuallyNumRequests(t, recorder, 1)
	require.Zero(t, hfb.NumPendingMbRequests())
}

func TestScheduleMiniBlocksRequest_ProofAlreadyPresentRequestsImmediately(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createSchedulingArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash, 1)))

	hfb.ScheduleMiniBlocksRequestIfNeeded(header, hash)

	requireEventuallyNumRequests(t, recorder, 1)
	require.Zero(t, hfb.NumPendingMbRequests())
}

func TestScheduleMiniBlocksRequest_WaitsForProofThenRequestsOnce(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createSchedulingArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	hfb.ScheduleMiniBlocksRequestIfNeeded(header, hash)

	time.Sleep(100 * time.Millisecond)
	require.Zero(t, recorder.count())
	require.Equal(t, 1, hfb.NumPendingMbRequests())

	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash, 1)))

	requireEventuallyNumRequests(t, recorder, 1)
	require.Zero(t, hfb.NumPendingMbRequests())

	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 1, recorder.count())
}

func TestScheduleMiniBlocksRequest_EndToEndThroughHeadersPool(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createSchedulingArgs(recorder)
	_, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	args.DataPool.Headers().AddHeader(hash, header)

	time.Sleep(100 * time.Millisecond)
	require.Zero(t, recorder.count())

	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash, 1)))

	requireEventuallyNumRequests(t, recorder, 1)
}

func TestScheduleMiniBlocksRequest_NodeBehindRequestsImmediately(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createSchedulingArgs(recorder)
	args.RoundHandler = &testscommon.RoundHandlerMock{
		IndexForCurrentTimeCalled: func() int64 {
			return 10
		},
		IndexCalled: func() int64 {
			return 1 // stale stored index; the arithmetic index must be used instead
		},
	}
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	hfb.ScheduleMiniBlocksRequestIfNeeded(header, hash)

	requireEventuallyNumRequests(t, recorder, 1)
	require.Zero(t, hfb.NumPendingMbRequests())
}

func TestScheduleMiniBlocksRequest_FallbackDispatchesStaleEntries(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createSchedulingArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)
	hfb.SetPendingMbRequestFallbackDelay(50 * time.Millisecond)

	header1, hash1 := createTestMetaHeader(1)
	hfb.ScheduleMiniBlocksRequestIfNeeded(header1, hash1)

	time.Sleep(100 * time.Millisecond)
	require.Zero(t, recorder.count()) // no events yet, sweep not triggered

	header2, hash2 := createTestMetaHeader(2)
	hfb.ScheduleMiniBlocksRequestIfNeeded(header2, hash2)

	require.Eventually(t, func() bool {
		return recorder.has(1)
	}, time.Second, 5*time.Millisecond)
	require.False(t, recorder.has(2))
	require.Equal(t, 1, hfb.NumPendingMbRequests())
}

func TestScheduleMiniBlocksRequest_DuplicateScheduleSingleRequest(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createSchedulingArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	header, hash := createTestMetaHeader(1)
	hfb.ScheduleMiniBlocksRequestIfNeeded(header, hash)
	hfb.ScheduleMiniBlocksRequestIfNeeded(header, hash)
	require.Equal(t, 1, hfb.NumPendingMbRequests())

	require.True(t, args.DataPool.Proofs().AddProof(createTestProof(hash, 1)))

	requireEventuallyNumRequests(t, recorder, 1)
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 1, recorder.count())
}

func TestScheduleMiniBlocksRequest_CapEvictionDispatchesOldest(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createSchedulingArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)
	hfb.SetMaxPendingMbRequests(2)

	for nonce := uint64(1); nonce <= 3; nonce++ {
		header, hash := createTestMetaHeader(nonce)
		hfb.ScheduleMiniBlocksRequestIfNeeded(header, hash)
	}

	require.Eventually(t, func() bool {
		return recorder.has(1)
	}, time.Second, 5*time.Millisecond)
	require.Equal(t, 1, recorder.count())
	require.Equal(t, 2, hfb.NumPendingMbRequests())
}

func TestScheduleMiniBlocksRequest_ConcurrentScheduleAndProof(t *testing.T) {
	t.Parallel()

	recorder := &requestRecorder{}
	args := createSchedulingArgs(recorder)
	hfb, err := headerForBlock.NewHeadersForBlock(args)
	require.NoError(t, err)

	numHeaders := 15 // stays within MaxHeadersToRequestInAdvance of the tracker stub's nonce 0
	wg := sync.WaitGroup{}
	for i := 1; i <= numHeaders; i++ {
		header, hash := createTestMetaHeader(uint64(i))
		proof := createTestProof(hash, uint64(i))
		wg.Add(2)
		go func() {
			hfb.ScheduleMiniBlocksRequestIfNeeded(header, hash)
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
