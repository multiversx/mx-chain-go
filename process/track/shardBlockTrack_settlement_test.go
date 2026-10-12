package track_test

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-go/common"
	requesterMock "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	storageRequesters "github.com/multiversx/mx-chain-go/dataRetriever/storageRequesters"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	processSync "github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	storageMocks "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestShardBlockTrack_SettlementRetiresObsoleteSource(t *testing.T) {
	for _, importMode := range []bool{false, true} {
		t.Run(fmt.Sprintf("import=%v", importMode), func(t *testing.T) {
			args := CreateShardTrackerMockArguments()
			args.IsImportDBMode = importMode
			args.ProofsPool = args.PoolsHolder.Proofs()
			var requests atomic.Int64
			var missingStorageReads atomic.Int64
			var enableMissingLookup atomic.Bool
			stop := make(chan endProcess.ArgEndProcess, 1)
			missingStorage := &storageMocks.StorerStub{
				SearchFirstCalled: func(_ []byte) ([]byte, error) {
					missingStorageReads.Add(1)
					return nil, fmt.Errorf("historical header outside active persisters")
				},
			}
			storageRequester, err := storageRequesters.NewHeaderRequester(storageRequesters.ArgHeaderRequester{
				Messenger:      &requesterMock.MessageHandlerStub{},
				NonceConverter: &requesterMock.Uint64ByteSliceConverterMock{},
				HdrStorage:     missingStorage, HeadersNoncesStorage: missingStorage,
				ManualEpochStartNotifier: &requesterMock.ManualEpochStartNotifierStub{
					CurrentEpochCalled: func() uint32 { return 2246 },
				},
				ChanGracefullyClose: stop,
			})
			require.NoError(t, err)
			require.NoError(t, storageRequester.SetEpochHandler(&requesterMock.EpochHandlerStub{
				MetaEpochCalled: func() uint32 { return 2244 },
			}))
			oldHash := []byte("old-source-2242")
			requestHandler := &requestHandlerWithIntervalHook{interval: time.Nanosecond}
			requestHandler.RequestMetaHeaderForEpochCalled = func(hash []byte, epoch uint32) {
				if bytes.Equal(hash, oldHash) && epoch == 2242 {
					requests.Add(1)
					if importMode && enableMissingLookup.Load() {
						_ = storageRequester.RequestDataFromHash(hash, epoch)
					}
				}
			}
			args.RequestHandler = requestHandler
			tracker, err := track.NewShardBlockTrack(args)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, tracker.Close()) })
			flags := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
				IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
					return flag == common.AndromedaFlag || flag == common.SupernovaFlag
				},
			}
			detector, err := processSync.NewShardForkDetector(
				&mock.RoundHandlerMock{RoundIndex: 1000}, &testscommon.TimeCacheStub{}, tracker,
				0, 0, flags, testscommon.NewEnableRoundsHandlerStub(common.SupernovaRoundFlag),
				args.ProofsPool, &chainParameters.ChainParametersHandlerStub{},
				testscommon.GetDefaultProcessConfigsHandler(), 0,
			)
			require.NoError(t, err)
			headers := args.PoolsHolder.Headers()
			proofs := args.PoolsHolder.Proofs()
			hash1, hash2 := []byte("shard-1"), []byte("shard-2")
			header1 := &block.HeaderV3{ShardID: 0, Nonce: 1, Round: 1, Epoch: 2242}
			header2 := &block.HeaderV3{ShardID: 0, Nonce: 2, Round: 2, Epoch: 2244, PrevHash: hash1}
			addProof := func(header data.HeaderHandler, hash []byte) {
				proofs.AddProof(&block.HeaderProof{
					HeaderHash: hash, HeaderNonce: header.GetNonce(), HeaderRound: header.GetRound(),
					HeaderEpoch: header.GetEpoch(), HeaderShardId: header.GetShardID(),
				})
			}
			headers.AddHeader(hash1, header1)
			headers.AddHeader(hash2, header2)
			addProof(header1, hash1)
			addProof(header2, hash2)
			oldSource := &block.MetaBlockV3{Nonce: 7, Round: 7, Epoch: 2242,
				ShardInfoProposal: []block.ShardDataProposal{{ShardID: 0, HeaderHash: hash1, Nonce: 1, Epoch: 2242}},
			}
			headers.AddHeader(oldHash, oldSource)
			addProof(oldSource, oldHash)
			oldRefs := tracker.GetSelfHeadersWithSource(oldSource, oldHash)
			require.Len(t, oldRefs, 1)
			tracker.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, oldRefs)
			require.Equal(t, int64(1), tracker.NumPendingSelfHeaders())
			require.NoError(t, detector.AddHeader(header1, hash1, process.BHProcessed, nil, nil))
			require.NoError(t, detector.AddHeader(header2, hash2, process.BHProcessed, nil, nil))
			require.Equal(t, uint64(2), detector.GetHighestFinalBlockNonce())
			nonce, _ := detector.GetHighestSettledBlockInfo()
			require.Zero(t, nonce)
			tracker.CleanupHeadersBehindNonce(0, 2, 0)
			require.Equal(t, int64(1), tracker.NumPendingSelfHeaders())
			sourceHash, childHash := []byte("current-source"), []byte("current-child")
			source := &block.MetaBlockV3{Nonce: 100, Round: 100, Epoch: 2244,
				ShardInfoProposal: []block.ShardDataProposal{{ShardID: 0, HeaderHash: hash2, Nonce: 2, Epoch: 2244}},
			}
			child := &block.MetaBlockV3{Nonce: 101, Round: 101, Epoch: 2244, PrevHash: sourceHash}
			headers.AddHeader(sourceHash, source)
			headers.AddHeader(childHash, child)
			addProof(source, sourceHash)
			addProof(child, childHash)
			require.True(t, tracker.IsSettledCrossHeader(source, sourceHash))
			tracker.AddCrossNotarizedHeader(core.MetachainShardId, source, sourceHash)
			require.Eventually(t, func() bool { n, _ := detector.GetHighestSettledBlockInfo(); return n == 2 }, time.Second, time.Millisecond)
			tracker.CleanupHeadersBehindNonce(0, 2, 0)
			tracker.CleanupHeadersBehindNonce(core.MetachainShardId, 0, 100)
			headers.RemoveHeaderByHash(oldHash)
			require.Zero(t, tracker.NumPendingSelfHeaders())
			tracker.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, oldRefs)
			require.Empty(t, tracker.GetSelfHeadersWithSource(oldSource, oldHash))
			require.Zero(t, tracker.NumPendingSelfHeaders())
			before := requests.Load()
			enableMissingLookup.Store(true)
			nextHash := []byte("current-grandchild")
			next := &block.MetaBlockV3{Nonce: 102, Round: 102, Epoch: 2244, PrevHash: childHash}
			headers.AddHeader(nextHash, next)
			addProof(next, nextHash)
			tracker.AddCrossNotarizedHeader(core.MetachainShardId, child, childHash)
			tracker.CleanupHeadersBehindNonce(0, 2, 0)
			require.Equal(t, before, requests.Load())
			require.Zero(t, tracker.NumPendingSelfHeaders())
			require.Zero(t, missingStorageReads.Load())
			require.Empty(t, stop)
			beforeNonce, beforeHash := detector.GetHighestSettledBlockInfo()
			detector.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, []data.HeaderHandler{header1}, [][]byte{hash1})
			afterNonce, afterHash := detector.GetHighestSettledBlockInfo()
			require.Equal(t, beforeNonce, afterNonce)
			require.Equal(t, beforeHash, afterHash)
			detector.ReceivedSelfNotarizedFromCrossHeaders(core.MetachainShardId, []data.HeaderHandler{header2}, [][]byte{[]byte("conflicting-settled-hash")})
			afterNonce, afterHash = detector.GetHighestSettledBlockInfo()
			require.Equal(t, beforeNonce, afterNonce)
			require.Equal(t, beforeHash, afterHash)
			if importMode {
				require.Error(t, storageRequester.RequestDataFromHash(oldHash, 2242))
				select {
				case signal := <-stop:
					require.Equal(t, common.ImportComplete, signal.Reason)
					require.Contains(t, signal.Description, "2245 or 2246")
				case <-time.After(time.Second):
					t.Fatal("control lookup did not emit the existing import-completion signal")
				}
			}
			tracker.RestoreToGenesis()
			tracker.CleanupHeadersBehindNonce(0, 0, 0)
			detector.RestoreToGenesis()
			args.PoolsHolder.Headers().AddHeader(hash1, header1)
			require.Len(t, tracker.GetSelfHeadersWithSource(oldSource, oldHash), 1)
		})
	}
}

func TestShardBlockTrack_SettlementPreservesNeededRecovery(t *testing.T) {
	t.Parallel()

	args := CreateShardTrackerMockArguments()
	var requests atomic.Int64
	args.RequestHandler = &requestHandlerWithIntervalHook{
		interval: time.Nanosecond,
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestShardHeaderForEpochCalled: func(_ uint32, _ []byte, _ uint32) { requests.Add(1) },
		},
	}
	tracker, err := track.NewShardBlockTrack(args)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, tracker.Close()) })
	var settled atomic.Uint64
	tracker.SetSettledNonceProvider(settled.Load)
	tracker.SetMetaFinalityView(&testscommon.MetaFinalityViewStub{
		IsMetaHeaderSettlementReadyCalled: func(_ data.HeaderHandler, _ []byte) bool { return false },
	})
	source := &block.MetaBlockV3{Nonce: 7, Round: 7, Epoch: 2242}
	for _, nonce := range []uint64{3, 4, 5, 6} {
		hash := []byte(fmt.Sprint(nonce))
		source.ShardInfoProposal = append(source.ShardInfoProposal, block.ShardDataProposal{
			ShardID: 0, HeaderHash: hash, Nonce: nonce, Epoch: 2242,
		})
		if nonce == 3 || nonce == 6 {
			args.PoolsHolder.Headers().AddHeader(hash, &block.HeaderV3{ShardID: 0, Nonce: nonce, Epoch: 2242})
		}
	}
	refs := tracker.GetSelfHeadersWithSource(source, []byte("source"))
	require.Len(t, refs, 2)
	tracker.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, refs)
	require.Equal(t, int64(4), tracker.NumPendingSelfHeaders())
	settled.Store(5)
	tracker.CleanupHeadersBehindNonce(0, 100, 0)
	require.Equal(t, int64(2), tracker.NumPendingSelfHeaders())
	require.Zero(t, tracker.NumPendingSources([]byte("3")))
	require.Zero(t, tracker.NumPendingSources([]byte("4")))
	require.Equal(t, 1, tracker.NumPendingSources([]byte("5")))
	require.Equal(t, 1, tracker.NumPendingSources([]byte("6")))
	before := requests.Load()
	tracker.GetSelfHeadersWithSource(source, []byte("source"))
	require.Greater(t, requests.Load(), before)
	tracker.PublishSelfNotarizedFromCrossHeaders(core.MetachainShardId, refs)
	require.Equal(t, int64(2), tracker.NumPendingSelfHeaders())
	tracker.RemoveLastNotarizedHeaders()
	tracker.CleanupHeadersBehindNonce(0, 100, 0)
	require.Equal(t, int64(2), tracker.NumPendingSelfHeaders())
}
