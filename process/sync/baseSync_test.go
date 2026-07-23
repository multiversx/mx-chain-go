package sync

import (
	"bytes"
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	testscommonDataRetriever "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

func TestBaseBootstrap_SyncBlocksShouldNotCallSyncIfNotConnectedToTheNetwork(t *testing.T) {
	t.Parallel()

	var numCalls uint32
	boot := &baseBootstrap{
		chStopSync: make(chan bool),
		syncStarter: &mock.SyncStarterStub{
			SyncBlockCalled: func() error {
				atomic.AddUint32(&numCalls, 1)
				return nil
			},
		},
		networkWatcher: &mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return false
			},
		},
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	go boot.syncBlocks(ctx)

	// make sure go routine started and waited a few cycles of boot.syncBlocks
	time.Sleep(time.Second + sleepTime*10)
	cancelFunc()

	assert.Equal(t, uint32(0), atomic.LoadUint32(&numCalls))
}

func TestBaseBootstrap_SyncBlocksShouldCallSyncIfConnectedToTheNetwork(t *testing.T) {
	t.Parallel()

	var numCalls uint32
	boot := &baseBootstrap{
		chStopSync: make(chan bool),
		syncStarter: &mock.SyncStarterStub{
			SyncBlockCalled: func() error {
				atomic.AddUint32(&numCalls, 1)
				return nil
			},
		},
		networkWatcher: &mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		roundHandler: &mock.RoundHandlerMock{
			BeforeGenesisCalled: func() bool {
				return false
			},
		},
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	go boot.syncBlocks(ctx)

	// make sure go routine started and waited a few cycles of boot.syncBlocks
	time.Sleep(time.Second + sleepTime*10)
	cancelFunc()

	assert.True(t, atomic.LoadUint32(&numCalls) > 0)
}

func TestBaseBootstrap_GetOrderedMiniBlocksShouldErrMissingBody(t *testing.T) {
	t.Parallel()

	hashes := [][]byte{[]byte("hash1")}
	orderedMiniBlocks, err := getOrderedMiniBlocks(hashes, nil)

	assert.Nil(t, orderedMiniBlocks)
	assert.Equal(t, process.ErrMissingBody, err)
}

func TestBaseBootstrap_GetOrderedMiniBlocksShouldWork(t *testing.T) {
	t.Parallel()

	hash1 := []byte("hash1")
	miniBlock1 := &block.MiniBlock{SenderShardID: 0}

	hash2 := []byte("hash2")
	miniBlock2 := &block.MiniBlock{SenderShardID: 1}

	hash3 := []byte("hash3")
	miniBlock3 := &block.MiniBlock{SenderShardID: 2}

	hashes := [][]byte{hash1, hash2, hash3}
	miniBlocksAndHashes := []*block.MiniblockAndHash{
		{
			Hash:      hash1,
			Miniblock: miniBlock1,
		},
		{
			Hash:      hash2,
			Miniblock: miniBlock2,
		},
		{
			Hash:      hash3,
			Miniblock: miniBlock3,
		},
	}

	orderedMiniBlocks, err := getOrderedMiniBlocks(hashes, miniBlocksAndHashes)

	assert.Nil(t, err)
	require.Equal(t, 3, len(orderedMiniBlocks))
	assert.Equal(t, uint32(0), orderedMiniBlocks[0].SenderShardID)
	assert.Equal(t, uint32(1), orderedMiniBlocks[1].SenderShardID)
	assert.Equal(t, uint32(2), orderedMiniBlocks[2].SenderShardID)
}

func TestBaseBootstrap_GetNodeState(t *testing.T) {
	t.Parallel()

	boot := &baseBootstrap{
		isInImportMode:        true,
		isNodeStateCalculated: true,
		roundHandler:          &mock.RoundHandlerMock{},
		chainHandler:          getMockChainHandler(),
		currentEpochProvider:  &testscommon.CurrentEpochProviderStub{},
	}
	assert.Equal(t, common.NsNotSynchronized, boot.GetNodeState())

	boot = &baseBootstrap{
		isInImportMode:        false,
		isNodeStateCalculated: true,
		roundHandler:          &mock.RoundHandlerMock{},
		chainHandler:          getMockChainHandler(),
		currentEpochProvider:  &testscommon.CurrentEpochProviderStub{},
	}
	assert.Equal(t, common.NsNotSynchronized, boot.GetNodeState())

	boot = &baseBootstrap{
		roundIndex:            1,
		isInImportMode:        false,
		isNodeStateCalculated: true,
		roundHandler:          &mock.RoundHandlerMock{},
		chainHandler:          getMockChainHandler(),
		currentEpochProvider:  &testscommon.CurrentEpochProviderStub{},
	}
	assert.Equal(t, common.NsNotCalculated, boot.GetNodeState())

	boot = &baseBootstrap{
		roundIndex:            1,
		isInImportMode:        false,
		isNodeStateCalculated: true,
		roundHandler:          &mock.RoundHandlerMock{},
		chainHandler:          getMockChainHandler(),
		currentEpochProvider: &testscommon.CurrentEpochProviderStub{
			EpochIsActiveInNetworkCalled: func(epoch uint32) bool {
				return false
			},
		},
	}
	assert.Equal(t, common.NsNotSynchronized, boot.GetNodeState())
}

// createBootForRollBackOneBlockForcedTest builds a baseBootstrap whose current block (the one that
// will be rolled back, nonce currNonce, epoch currEpoch) sits on top of prevHeader (nonce
// currNonce-1, epoch prevEpoch). The proof gate keys off the rolled-back block, so currEpoch is the
// epoch that matters for the request decision.
func createBootForRollBackOneBlockForcedTest(
	selfID uint32,
	currNonce uint64,
	currEpoch uint32,
	prevEpoch uint32,
	enableEpochsHandler common.EnableEpochsHandler,
	requestProof func(nonce uint64),
) *baseBootstrap {
	currentHeader := data.HeaderHandler(&block.Header{
		Nonce:    currNonce,
		Epoch:    currEpoch,
		PrevHash: []byte("previous hash"),
		RootHash: []byte("current root hash"),
	})
	prevHeader := data.HeaderHandler(&block.Header{
		Nonce:    currNonce - 1,
		Epoch:    prevEpoch,
		RootHash: []byte("previous root hash"),
	})
	currentHash := []byte("current hash")

	return &baseBootstrap{
		chainHandler: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{}
			},
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return currentHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return currentHash
			},
			SetCurrentBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, rootHash []byte) error {
				currentHeader = header
				return nil
			},
			SetCurrentBlockHeaderHashCalled: func(hash []byte) {
				currentHash = hash
			},
			SetLastExecutedBlockHeaderAndRootHashCalled: func(header data.HeaderHandler, headerHash []byte, rootHash []byte) {
			},
		},
		blockBootstrapper: &blockBootstrapperStub{
			getCurrHeaderCalled: func() (data.HeaderHandler, error) {
				return currentHeader, nil
			},
			getPrevHeaderCalled: func(data.HeaderHandler, storage.Storer) (data.HeaderHandler, error) {
				return prevHeader, nil
			},
			getBlockBodyCalled: func(data.HeaderHandler) (data.BodyHandler, error) {
				return &block.Body{}, nil
			},
			requestProofByNonceCalled: requestProof,
		},
		blockProcessor: &testscommon.BlockProcessorStub{
			NonceOfFirstCommittedBlockCalled: func() core.OptionalUint64 {
				return core.OptionalUint64{
					Value:    1,
					HasValue: true,
				}
			},
		},
		forkDetector: &mock.ForkDetectorMock{
			// keep the highest final nonce below the current one so shouldAllowRollback permits the rollback
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 0
			},
		},
		headers: &mock.HeadersCacherStub{
			NoncesCalled: func(shardId uint32) []uint64 {
				return nil
			},
		},
		headerNonceHashStore:         &storageStubs.StorerStub{},
		historyRepo:                  &dblookupext.HistoryRepositoryStub{},
		marshalizer:                  &marshal.GogoProtoMarshalizer{},
		hasher:                       &hashingMocks.HasherMock{},
		outportHandler:               &outport.OutportStub{},
		scheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		shardCoordinator: &mock.ShardCoordinatorStub{
			SelfIdCalled: func() uint32 {
				return selfID
			},
		},
		requestHandler:      &testscommon.RequestHandlerStub{},
		store:               &storageStubs.ChainStorerStub{},
		bootStorer:          &mock.BoostrapStorerMock{},
		uint64Converter:     &mock.Uint64ByteSliceConverterMock{},
		statusHandler:       &statusHandlerMock.AppStatusHandlerStub{},
		forkInfo:            &process.ForkInfo{},
		enableEpochsHandler: enableEpochsHandler,
	}
}

func TestBaseBootstrap_RollBackOneBlockForcedShouldRequestEquivalentProofForNextNonce(t *testing.T) {
	t.Parallel()

	selfID := uint32(1)

	// andromedaFromEpoch returns a handler where the Andromeda flag activates at the given epoch.
	andromedaFromEpoch := func(activationEpoch uint32) *enableEpochsHandlerMock.EnableEpochsHandlerStub {
		return &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag && epoch >= activationEpoch
			},
		}
	}

	t.Run("proofs flag enabled should request proof for rolled-back nonce", func(t *testing.T) {
		t.Parallel()

		currNonce := uint64(10)
		requestCalled := false
		requestedNonce := uint64(0)
		boot := createBootForRollBackOneBlockForcedTest(selfID, currNonce, 6, 6, andromedaFromEpoch(5), func(nonce uint64) {
			requestCalled = true
			requestedNonce = nonce
		})

		boot.rollBackOneBlockForced()

		assert.True(t, requestCalled)
		assert.Equal(t, currNonce, requestedNonce)
	})

	t.Run("proofs flag disabled before activation should not request proof", func(t *testing.T) {
		t.Parallel()

		currNonce := uint64(10)
		requestCalled := false
		boot := createBootForRollBackOneBlockForcedTest(selfID, currNonce, 2, 2, andromedaFromEpoch(5), func(nonce uint64) {
			requestCalled = true
		})

		boot.rollBackOneBlockForced()

		// the rollback must actually have succeeded (current block dropped to currNonce-1), so that the
		// missing request is attributable to the disabled flag and not to a silently failed rollback
		require.Equal(t, currNonce-1, boot.getCurrentBlock().GetNonce())
		assert.False(t, requestCalled)
	})

	t.Run("rollback of first block with Andromeda active from epoch 0 should still request proof", func(t *testing.T) {
		t.Parallel()

		// the rolled-back block is nonce 1 (epoch 0); its proof must still be requested. Rolling it
		// back leaves genesis as the current block, but the gate keys off the captured nonce-1 header,
		// so IsProofsFlagEnabledForHeader (which excludes nonce 0) still returns true.
		currNonce := uint64(1)
		requestCalled := false
		requestedNonce := uint64(0)
		boot := createBootForRollBackOneBlockForcedTest(selfID, currNonce, 0, 0, andromedaFromEpoch(0), func(nonce uint64) {
			requestCalled = true
			requestedNonce = nonce
		})

		boot.rollBackOneBlockForced()

		assert.True(t, requestCalled)
		assert.Equal(t, currNonce, requestedNonce)
	})

	t.Run("activation boundary should request proof using the rolled-back block epoch", func(t *testing.T) {
		t.Parallel()

		// the rolled-back block is the activation-epoch block (epoch 5, proofs enabled) sitting on top
		// of a pre-activation parent (epoch 4); the gate must key off the rolled-back block, not the
		// parent, so the proof is requested.
		currNonce := uint64(10)
		requestCalled := false
		requestedNonce := uint64(0)
		boot := createBootForRollBackOneBlockForcedTest(selfID, currNonce, 5, 4, andromedaFromEpoch(5), func(nonce uint64) {
			requestCalled = true
			requestedNonce = nonce
		})

		boot.rollBackOneBlockForced()

		assert.True(t, requestCalled)
		assert.Equal(t, currNonce, requestedNonce)
	})

	t.Run("failed rollback should not request proof", func(t *testing.T) {
		t.Parallel()

		currNonce := uint64(10)
		requestCalled := false
		boot := createBootForRollBackOneBlockForcedTest(selfID, currNonce, 6, 6, andromedaFromEpoch(5), func(nonce uint64) {
			requestCalled = true
		})
		// make rollBack(false) fail so no block is actually rolled back
		getPrevHeaderCalled := false
		boot.blockBootstrapper.(*blockBootstrapperStub).getPrevHeaderCalled = func(data.HeaderHandler, storage.Storer) (data.HeaderHandler, error) {
			getPrevHeaderCalled = true
			return nil, errors.New("rollback error")
		}

		boot.rollBackOneBlockForced()

		// the error path must actually have been reached (rollBack attempted) and left the current block
		// unchanged, so the missing request is attributable to the rollback error and not to skipping the gate
		require.True(t, getPrevHeaderCalled)
		require.Equal(t, currNonce, boot.getCurrentBlock().GetNonce())
		assert.False(t, requestCalled)
	})
}

func TestBaseSync_getEpochOfCurrentBlockGenesis(t *testing.T) {
	t.Parallel()

	genesisEpoch := uint32(1123)
	boot := &baseBootstrap{
		chainHandler: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Epoch: genesisEpoch,
				}
			},
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		},
	}

	epoch := boot.getEpochOfCurrentBlock()
	assert.Equal(t, genesisEpoch, epoch)
}

func TestBaseSync_getEpochOfCurrentBlockHeader(t *testing.T) {
	t.Parallel()

	genesisEpoch := uint32(1123)
	headerEpoch := uint32(97493)
	boot := &baseBootstrap{
		chainHandler: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Epoch: genesisEpoch,
				}
			},
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Epoch: headerEpoch,
				}
			},
		},
	}

	epoch := boot.getEpochOfCurrentBlock()
	assert.Equal(t, headerEpoch, epoch)
}

func TestBaseBootstrap_confirmHeaderReceivedByHashShouldRequestMissingProof(t *testing.T) {
	t.Parallel()

	headerHash := []byte("requested-hash")
	expectedEpoch := uint32(7)
	expectedShardID := uint32(1)
	expectedNonce := uint64(42)

	var requestedEpoch uint32
	var requestedShardID uint32
	var requestedHash []byte

	requestHandler := &requestHandlerWithSetEpochStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestEquivalentProofByHashCalled: func(headerShard uint32, hash []byte) {
				requestedShardID = headerShard
				requestedHash = append([]byte(nil), hash...)
			},
		},
		SetEpochCalled: func(epoch uint32) {
			requestedEpoch = epoch
		},
	}

	boot := &baseBootstrap{
		requestHandler: requestHandler,
		proofs: &testscommonDataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, hash []byte) bool {
				return false
			},
		},
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		},
	}
	boot.setRequestedHeaderHash(headerHash)

	header := &block.Header{
		ShardID: expectedShardID,
		Epoch:   expectedEpoch,
		Nonce:   expectedNonce,
	}

	boot.confirmHeaderReceivedByHash(header, headerHash)

	require.Equal(t, expectedEpoch, requestedEpoch)
	require.Equal(t, expectedShardID, requestedShardID)
	require.True(t, bytes.Equal(headerHash, requestedHash))
}

func TestBaseBootstrap_confirmHeaderReceivedByNonceShouldRequestMissingProof(t *testing.T) {
	t.Parallel()

	headerHash := []byte("requested-hash")
	expectedEpoch := uint32(9)
	expectedShardID := uint32(2)
	expectedNonce := uint64(64)

	var requestedEpoch uint32
	var requestedShardID uint32
	var requestedHash []byte

	requestHandler := &requestHandlerWithSetEpochStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestEquivalentProofByHashCalled: func(headerShard uint32, hash []byte) {
				requestedShardID = headerShard
				requestedHash = append([]byte(nil), hash...)
			},
		},
		SetEpochCalled: func(epoch uint32) {
			requestedEpoch = epoch
		},
	}

	boot := &baseBootstrap{
		requestHandler: requestHandler,
		proofs: &testscommonDataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, hash []byte) bool {
				return false
			},
		},
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		},
	}
	boot.setRequestedHeaderNonce(&expectedNonce)

	header := &block.Header{
		ShardID: expectedShardID,
		Epoch:   expectedEpoch,
		Nonce:   expectedNonce,
	}

	boot.confirmHeaderReceivedByNonce(header, headerHash)

	require.Equal(t, expectedEpoch, requestedEpoch)
	require.Equal(t, expectedShardID, requestedShardID)
	require.True(t, bytes.Equal(headerHash, requestedHash))
}

func TestBaseSync_shouldAllowRollback(t *testing.T) {
	t.Parallel()

	finalBlockHash := []byte("final block hash")
	notFinalBlockHash := []byte("not final block hash")
	firstBlockNonce := &core.OptionalUint64{
		HasValue: true,
		Value:    2,
	}
	boot := &baseBootstrap{
		forkDetector: &mock.ForkDetectorMock{
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return 10
			},
			GetHighestFinalBlockHashCalled: func() []byte {
				return finalBlockHash
			},
		},
		blockProcessor: &testscommon.BlockProcessorStub{
			NonceOfFirstCommittedBlockCalled: func() core.OptionalUint64 {
				return *firstBlockNonce
			},
		},
		executionManager: &processMocks.ExecutionManagerMock{},
	}

	t.Run("should allow rollback nonces above final", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 11
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return false
			},
		}
		require.True(t, boot.shouldAllowRollback(header, finalBlockHash))
		require.True(t, boot.shouldAllowRollback(header, notFinalBlockHash))

		header.HasScheduledMiniBlocksCalled = func() bool {
			return true
		}
		require.True(t, boot.shouldAllowRollback(header, finalBlockHash))
		require.True(t, boot.shouldAllowRollback(header, notFinalBlockHash))
	})

	t.Run("should not allow rollback of a final header with the same final hash if it doesn't have scheduled miniBlocks", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 10
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return false
			},
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
	})

	t.Run("should allow rollback of a final header without the same final hash", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 10
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return false
			},
		}
		require.True(t, boot.shouldAllowRollback(header, notFinalBlockHash))
	})

	t.Run("should allow rollback of a final header if it holds scheduled miniBlocks", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 10
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return true
			},
		}
		require.True(t, boot.shouldAllowRollback(header, finalBlockHash))
	})

	t.Run("should not allow rollback of a final header if it holds scheduled miniBlocks but no commit was done", func(t *testing.T) {
		firstBlockNonce.HasValue = false
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 10
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return true
			},
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
		firstBlockNonce.HasValue = true
	})

	t.Run("should not allow rollback of a final header if it holds scheduled miniBlocks but first committed nonce is higher", func(t *testing.T) {
		firstBlockNonce.Value = 11
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 10
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return true
			},
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
		firstBlockNonce.Value = 2
	})

	t.Run("should not allow any rollBack of a header if nonce is behind final", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 9
			},
			HasScheduledMiniBlocksCalled: func() bool {
				return true
			},
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
		require.False(t, boot.shouldAllowRollback(header, notFinalBlockHash))

		header.HasScheduledMiniBlocksCalled = func() bool {
			return false
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
		require.False(t, boot.shouldAllowRollback(header, notFinalBlockHash))
	})

	t.Run("should allow rollback of a header v3 only above the final nonce", func(t *testing.T) {
		header := &testscommon.HeaderHandlerStub{
			GetNonceCalled: func() uint64 {
				return 11
			},
			IsHeaderV3Called: func() bool {
				return true
			},
		}
		require.True(t, boot.shouldAllowRollback(header, finalBlockHash))

		header.GetNonceCalled = func() uint64 {
			return 10
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
		require.False(t, boot.shouldAllowRollback(header, notFinalBlockHash))

		header.GetNonceCalled = func() uint64 {
			return 9
		}
		require.False(t, boot.shouldAllowRollback(header, finalBlockHash))
	})
}

func TestBaseBootstrap_PrepareForSyncAtBootstrapIfNeeded(t *testing.T) {
	t.Parallel()

	t.Run("should run only once", func(t *testing.T) {
		t.Parallel()

		lastExecHeaderHash := []byte("lastExecHeaderHash")

		lastHeader := &block.HeaderV3{
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  lastExecHeaderHash,
					HeaderNonce: 9,
				},
			},
			Nonce: 10,
		}

		numCalls := 0
		boot := &baseBootstrap{
			chainHandler: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					numCalls++
					return lastHeader
				},
			},
			preparedForSync: true, // prepared for sync already called, we are not testing here
			// the behaviour of preparedForSyncIfNeeded
		}

		err := boot.PrepareForSyncAtBoostrapIfNeeded()
		require.Nil(t, err)

		require.Equal(t, 1, numCalls)

		err = boot.PrepareForSyncAtBoostrapIfNeeded()
		require.Nil(t, err)

		require.Equal(t, 1, numCalls) // still 1 call
	})

	t.Run("should not trigger for non header v3", func(t *testing.T) {
		t.Parallel()

		lastHeader := &block.Header{
			Nonce: 10,
		}

		numCalls := 0
		boot := &baseBootstrap{
			chainHandler: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					numCalls++
					return lastHeader
				},
			},
			preparedForSync: false,
		}

		err := boot.PrepareForSyncAtBoostrapIfNeeded()
		require.Nil(t, err)

		require.Equal(t, 1, numCalls)

		err = boot.PrepareForSyncAtBoostrapIfNeeded()
		require.Nil(t, err)

		require.Equal(t, 1, numCalls) // still 1 call
	})
}

func TestBaseBootstrap_SaveProposedTxsToPool(t *testing.T) {
	t.Parallel()

	marshaller := &marshal.GogoProtoMarshalizer{}

	txCalls := 0
	scCalls := 0
	rwCalls := 0
	peerCalls := 0

	boot := &baseBootstrap{
		marshalizer: marshaller,
		dataPool: &testscommonDataRetriever.PoolsHolderStub{
			TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
					AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
						txCalls++
					},
				}
			},
			UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
					AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
						scCalls++
					},
				}
			},
			RewardTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
					AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
						rwCalls++
					},
				}
			},
			ValidatorsInfoCalled: func() dataRetriever.ShardedDataCacherNotifier {
				return &testscommon.ShardedDataStub{
					AddDataCalled: func(key []byte, data interface{}, sizeInBytes int, cacheID string) {
						peerCalls++
					},
				}
			},
		},
		store: &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					GetCalled: func(key []byte) ([]byte, error) {
						switch string(key) {
						case "txHash1":
							tx := &transaction.Transaction{
								Nonce: 1,
							}
							txBytes, _ := marshaller.Marshal(tx)
							return txBytes, nil
						case "txHash2":
							tx := &transaction.Transaction{
								Nonce: 2,
							}
							txBytes, _ := marshaller.Marshal(tx)
							return txBytes, nil
						case "txHash3":
							tx := &smartContractResult.SmartContractResult{
								Nonce:        3,
								CodeMetadata: []byte("codeMetadata"),
							}
							txBytes, _ := marshaller.Marshal(tx)
							return txBytes, nil
						case "txHash4":
							tx := &rewardTx.RewardTx{
								Round: 1,
							}
							txBytes, _ := marshaller.Marshal(tx)
							return txBytes, nil
						case "txHash5":
							tx := &state.ShardValidatorInfo{
								PublicKey: []byte("pubKey"),
							}
							txBytes, _ := marshaller.Marshal(tx)
							return txBytes, nil
						default:
							return nil, errors.New("err")
						}
					},
				}, nil
			},
		},
	}

	header := &block.HeaderV3{}
	body := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				TxHashes: [][]byte{[]byte("txHash1")},
				Type:     block.TxBlock,
			},
			{
				TxHashes: [][]byte{[]byte("txHash2")},
				Type:     block.InvalidBlock,
			},
			{
				TxHashes: [][]byte{[]byte("txHash3")},
				Type:     block.SmartContractResultBlock,
			},
			{
				TxHashes: [][]byte{[]byte("txHash4")},
				Type:     block.RewardsBlock,
			},
			{
				TxHashes: [][]byte{[]byte("txHash5")},
				Type:     block.PeerBlock,
			},
		},
	}

	err := boot.SaveProposedTxsToPool(header, body)
	require.Nil(t, err)

	require.Equal(t, 2, txCalls)
	require.Equal(t, 1, scCalls)
	require.Equal(t, 1, rwCalls)
	require.Equal(t, 1, peerCalls)
}

func TestBaseBootstrap_SyncBlocksWakesUpOnSignal(t *testing.T) {
	t.Parallel()

	signalChan := make(chan uint64, 1)
	var numCalls uint32
	syncError := errors.New("sync error to trigger wait")

	boot := &baseBootstrap{
		chStopSync:                  make(chan bool),
		signalProcessCompletionChan: signalChan,
		syncStarter: &mock.SyncStarterStub{
			SyncBlockCalled: func() error {
				atomic.AddUint32(&numCalls, 1)
				return syncError
			},
		},
		networkWatcher: &mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool {
				return true
			},
		},
		roundHandler: &mock.RoundHandlerMock{
			BeforeGenesisCalled: func() bool {
				return false
			},
		},
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	go boot.syncBlocks(ctx)

	// Wait for first sync call
	time.Sleep(50 * time.Millisecond)
	initialCalls := atomic.LoadUint32(&numCalls)
	require.GreaterOrEqual(t, initialCalls, uint32(1))

	// Signal the channel - this should wake up the loop immediately
	signalChan <- 42

	// Wait a short time - much less than sleepTimeOnFail (400ms)
	time.Sleep(50 * time.Millisecond)

	// Should have made another call due to signal wakeup
	finalCalls := atomic.LoadUint32(&numCalls)
	require.Greater(t, finalCalls, initialCalls)

	cancelFunc()
}

func TestBaseBootstrap_CleanChannelsDrainsSignalChannel(t *testing.T) {
	t.Parallel()

	signalChan := make(chan uint64, 5)
	signalChan <- 1
	signalChan <- 2
	signalChan <- 3

	boot := &baseBootstrap{
		chRcvHdrNonce:               make(chan bool, 1),
		chRcvHdrHash:                make(chan bool, 1),
		chRcvMiniBlocks:             make(chan bool, 1),
		signalProcessCompletionChan: signalChan,
	}

	boot.cleanChannels()

	assert.Equal(t, 0, len(signalChan))
}

func TestBaseBootstrap_ReconcileEquivocation(t *testing.T) {
	t.Parallel()

	finalNonce := uint64(10)
	localHash, competitorHash := []byte("localHash"), []byte("competitorHash")
	localHead := &block.HeaderV3{Nonce: finalNonce, Round: 12}

	competitorProof := &block.HeaderProof{
		HeaderHash:    competitorHash,
		HeaderNonce:   finalNonce,
		HeaderRound:   11,
		HeaderShardId: 0,
	}

	type reconcileCalls struct {
		reconciledNonce uint64
		rollBackNonce   uint64
		blacklisted     []string
	}

	buildBootstrapperWithChecker := func(childrenOf []byte, calls *reconcileCalls, checker settlementChecker, roundHandler *mock.RoundHandlerMock) *baseBootstrap {
		childHash := []byte("childHash")
		child := &block.HeaderV3{Nonce: finalNonce + 1, Round: 13, PrevHash: childrenOf}

		return &baseBootstrap{
			settlementChecker: checker,
			roundHandler:      roundHandler,
			chainHandler: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
					return localHead
				},
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return localHash
				},
			},
			forkDetector: &mock.ForkDetectorMock{
				GetHighestFinalBlockNonceCalled: func() uint64 {
					return finalNonce
				},
				ReconcileFinalCheckpointCalled: func(nonce uint64) {
					calls.reconciledNonce = nonce
				},
				SetRollBackNonceCalled: func(nonce uint64) {
					calls.rollBackNonce = nonce
				},
			},
			headers: &mock.HeadersCacherStub{
				GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
					if len(childrenOf) == 0 {
						return nil, nil, errors.New("no headers")
					}
					return []data.HeaderHandler{child}, [][]byte{childHash}, nil
				},
			},
			proofs: &testscommonDataRetriever.ProofsPoolMock{
				HasProofCalled: func(shardID uint32, headerHash []byte) bool {
					return string(headerHash) == string(competitorHash) || string(headerHash) == string(childHash)
				},
			},
			shardCoordinator: mock.NewOneShardCoordinatorMock(),
			blackListHandler: &testscommon.TimeCacheStub{
				AddCalled: func(key string) error {
					calls.blacklisted = append(calls.blacklisted, key)
					return nil
				},
			},
			statusHandler: &statusHandlerMock.AppStatusHandlerStub{},
		}
	}

	settlesOnly := func(hashes ...[]byte) *settlementCheckerStub {
		return &settlementCheckerStub{
			isSettledCalled: func(_ uint64, headerHash []byte) bool {
				for _, hash := range hashes {
					if bytes.Equal(hash, headerHash) {
						return true
					}
				}
				return false
			},
		}
	}

	buildBootstrapper := func(childrenOf []byte, calls *reconcileCalls) *baseBootstrap {
		checker := &settlementCheckerStub{
			isSettledCalled: func(nonce uint64, headerHash []byte) bool {
				return len(childrenOf) > 0 && bytes.Equal(childrenOf, headerHash)
			},
		}

		return buildBootstrapperWithChecker(childrenOf, calls, checker, &mock.RoundHandlerMock{})
	}

	t.Run("fires when the final head is childless and the competitor has a proofed child", func(t *testing.T) {
		t.Parallel()

		calls := &reconcileCalls{}
		boot := buildBootstrapper(competitorHash, calls)

		boot.onEquivocationEvidence(competitorProof, nil)
		require.NotNil(t, boot.pendingReconcile)

		require.True(t, boot.tryReconcileEquivocation())
		require.Equal(t, finalNonce, calls.reconciledNonce)
		require.Equal(t, finalNonce, calls.rollBackNonce)
		require.Equal(t, []string{string(localHash)}, calls.blacklisted)
		require.Nil(t, boot.pendingReconcile)
	})

	t.Run("never fires against a block with a settled descendant", func(t *testing.T) {
		t.Parallel()

		calls := &reconcileCalls{}
		boot := buildBootstrapper(localHash, calls)

		boot.onEquivocationEvidence(competitorProof, nil)
		require.False(t, boot.tryReconcileEquivocation())
		require.Equal(t, uint64(0), calls.reconciledNonce)
		require.Empty(t, calls.blacklisted)
	})

	t.Run("keeps the evidence armed while the competitor is unsettled", func(t *testing.T) {
		t.Parallel()

		calls := &reconcileCalls{}
		boot := buildBootstrapper(nil, calls)

		boot.onEquivocationEvidence(competitorProof, nil)
		require.False(t, boot.tryReconcileEquivocation())
		require.Equal(t, uint64(0), calls.reconciledNonce)
		// the settling child may still arrive: the evidence must survive the failed attempt
		require.NotNil(t, boot.pendingReconcile)
	})

	t.Run("ignores evidence away from the final head nonce", func(t *testing.T) {
		t.Parallel()

		calls := &reconcileCalls{}
		boot := buildBootstrapper(competitorHash, calls)

		otherProof := &block.HeaderProof{HeaderHash: competitorHash, HeaderNonce: finalNonce + 3, HeaderShardId: 0}
		boot.onEquivocationEvidence(otherProof, nil)
		require.Nil(t, boot.pendingReconcile)
		require.False(t, boot.tryReconcileEquivocation())
	})

	// signing locks per round, not per nonce, so a stranded loser can also hold a proofed child;
	// that child must no longer protect it from the authority's verdict
	t.Run("switches away from a local block that has its own proofed child", func(t *testing.T) {
		t.Parallel()

		calls := &reconcileCalls{}
		boot := buildBootstrapperWithChecker(localHash, calls, settlesOnly(competitorHash), &mock.RoundHandlerMock{})

		boot.onEquivocationEvidence(competitorProof, nil)
		require.True(t, boot.tryReconcileEquivocation())
		require.Equal(t, finalNonce, calls.reconciledNonce)
		require.Equal(t, []string{string(localHash)}, calls.blacklisted)
	})

	// the authority's verdict on the local hash beats any competitor evidence
	t.Run("never switches when the authority settled the local block", func(t *testing.T) {
		t.Parallel()

		calls := &reconcileCalls{}
		boot := buildBootstrapperWithChecker(competitorHash, calls, settlesOnly(localHash, competitorHash), &mock.RoundHandlerMock{})

		boot.onEquivocationEvidence(competitorProof, nil)
		require.False(t, boot.tryReconcileEquivocation())
		require.Equal(t, uint64(0), calls.reconciledNonce)
		require.Empty(t, calls.blacklisted)
		require.Nil(t, boot.pendingReconcile)
	})

	// the arbitration outcome: meta arbitrates the lowest-round sibling, which has no child at all
	t.Run("switches onto a childless competitor the authority notarized", func(t *testing.T) {
		t.Parallel()

		calls := &reconcileCalls{}
		boot := buildBootstrapperWithChecker(nil, calls, settlesOnly(competitorHash), &mock.RoundHandlerMock{})

		boot.onEquivocationEvidence(competitorProof, nil)
		require.True(t, boot.tryReconcileEquivocation())
		require.Equal(t, finalNonce, calls.reconciledNonce)
	})

	t.Run("meta node: depth-1 double extension keeps the evidence armed, a settled child fires", func(t *testing.T) {
		t.Parallel()

		localChildHash, competitorChildHash := []byte("localChildHash"), []byte("competitorChildHash")
		grandChildHash := []byte("grandChildHash")

		// both siblings hold a proofed child (per-round signing allows it); the competitor gains depth-2 later
		childrenByNonce := map[uint64][]pooledHeader{
			finalNonce + 1: {
				{&block.MetaBlock{Nonce: finalNonce + 1, PrevHash: localHash}, localChildHash},
				{&block.MetaBlock{Nonce: finalNonce + 1, PrevHash: competitorHash}, competitorChildHash},
			},
		}
		checker := newMetaCheckerWithPools(childrenByNonce, localChildHash, competitorChildHash, grandChildHash)

		calls := &reconcileCalls{}
		roundHandler := &mock.RoundHandlerMock{RoundIndex: 1}
		boot := buildBootstrapperWithChecker(nil, calls, checker, roundHandler)

		boot.onEquivocationEvidence(competitorProof, nil)

		require.False(t, boot.tryReconcileEquivocation())
		require.NotNil(t, boot.pendingReconcile)
		require.Equal(t, uint64(0), calls.reconciledNonce)

		// depth-2 on the competitor is the authority-grade evidence; depth-2 on BOTH siblings is
		// the accepted residual boundary, where the disarm keeps the node in place
		childrenByNonce[finalNonce+2] = []pooledHeader{
			{&block.MetaBlock{Nonce: finalNonce + 2, PrevHash: competitorChildHash}, grandChildHash},
		}
		roundHandler.RoundIndex = 2

		require.True(t, boot.tryReconcileEquivocation())
		require.Equal(t, finalNonce, calls.reconciledNonce)
		require.Equal(t, []string{string(localHash)}, calls.blacklisted)
	})

	t.Run("evaluates the authority at most once per round", func(t *testing.T) {
		t.Parallel()

		calls := &reconcileCalls{}
		checker := settlesOnly()
		roundHandler := &mock.RoundHandlerMock{RoundIndex: 7}
		boot := buildBootstrapperWithChecker(nil, calls, checker, roundHandler)

		boot.onEquivocationEvidence(competitorProof, nil)

		require.False(t, boot.tryReconcileEquivocation())
		callsInFirstRound := checker.numCalls
		require.NotZero(t, callsInFirstRound)

		require.False(t, boot.tryReconcileEquivocation())
		require.False(t, boot.tryReconcileEquivocation())
		require.Equal(t, callsInFirstRound, checker.numCalls)

		roundHandler.RoundIndex = 8
		require.False(t, boot.tryReconcileEquivocation())
		require.Greater(t, checker.numCalls, callsInFirstRound)
	})
}

type settlementCheckerStub struct {
	isSettledCalled              func(nonce uint64, headerHash []byte) bool
	deadCrossNotarizedMetaCalled func() (data.HeaderHandler, []byte, bool)
	numCalls                     int
}

func (stub *settlementCheckerStub) deadCrossNotarizedMeta() (data.HeaderHandler, []byte, bool) {
	if stub.deadCrossNotarizedMetaCalled != nil {
		return stub.deadCrossNotarizedMetaCalled()
	}

	return nil, nil, false
}

func (stub *settlementCheckerStub) isSettled(nonce uint64, headerHash []byte) bool {
	stub.numCalls++
	if stub.isSettledCalled != nil {
		return stub.isSettledCalled(nonce, headerHash)
	}

	return false
}

func TestBaseBootstrap_SelectNonBlackListedHash(t *testing.T) {
	t.Parallel()

	nonce := uint64(7)
	cleanHash, deadHash, deadSibling := []byte("cleanHash"), []byte("deadHash"), []byte("deadSibling")

	buildBootstrapper := func(blacklisted []string, siblingProofs []data.HeaderProofHandler, swept *bool) *baseBootstrap {
		return &baseBootstrap{
			shardCoordinator: mock.NewOneShardCoordinatorMock(),
			blackListHandler: &testscommon.TimeCacheStub{
				HasCalled: func(key string) bool {
					for _, blackListedKey := range blacklisted {
						if key == blackListedKey {
							return true
						}
					}
					return false
				},
				SweepCalled: func() {
					if swept != nil {
						*swept = true
					}
				},
			},
			proofs: &testscommonDataRetriever.ProofsPoolMock{
				GetProofsByNonceCalled: func(headerNonce uint64, shardID uint32) ([]data.HeaderProofHandler, error) {
					if len(siblingProofs) == 0 {
						return nil, errors.New("no proofs at nonce")
					}
					return siblingProofs, nil
				},
			},
		}
	}

	t.Run("empty hash is returned as is", func(t *testing.T) {
		t.Parallel()

		boot := buildBootstrapper([]string{string(deadHash)}, nil, nil)
		require.Nil(t, boot.selectNonBlackListedHash(nil, nonce))
	})

	t.Run("non-blacklisted hash is returned unchanged after a sweep", func(t *testing.T) {
		t.Parallel()

		swept := false
		boot := buildBootstrapper([]string{string(deadHash)}, nil, &swept)
		require.Equal(t, cleanHash, boot.selectNonBlackListedHash(cleanHash, nonce))
		require.True(t, swept)
	})

	t.Run("blacklisted hash is replaced by the first non-blacklisted proofed sibling", func(t *testing.T) {
		t.Parallel()

		siblingProofs := []data.HeaderProofHandler{
			&block.HeaderProof{HeaderHash: deadSibling, HeaderNonce: nonce},
			&block.HeaderProof{HeaderHash: cleanHash, HeaderNonce: nonce},
		}
		boot := buildBootstrapper([]string{string(deadHash), string(deadSibling)}, siblingProofs, nil)
		require.Equal(t, cleanHash, boot.selectNonBlackListedHash(deadHash, nonce))
	})

	t.Run("blacklisted hash with no proofs at the nonce returns nil", func(t *testing.T) {
		t.Parallel()

		boot := buildBootstrapper([]string{string(deadHash)}, nil, nil)
		require.Nil(t, boot.selectNonBlackListedHash(deadHash, nonce))
	})

	t.Run("blacklisted hash with only blacklisted siblings returns nil", func(t *testing.T) {
		t.Parallel()

		siblingProofs := []data.HeaderProofHandler{
			&block.HeaderProof{HeaderHash: deadSibling, HeaderNonce: nonce},
		}
		boot := buildBootstrapper([]string{string(deadHash), string(deadSibling)}, siblingProofs, nil)
		require.Nil(t, boot.selectNonBlackListedHash(deadHash, nonce))
	})
}

func TestBaseBootstrap_GetHeaderWithNonceRequestingIfMissingRefusesBlackListedHeader(t *testing.T) {
	t.Parallel()

	nonce := uint64(7)
	deadHash := []byte("deadHash")
	header := &block.HeaderV3{Nonce: nonce}

	requested := false
	chRcvHdrNonce := make(chan bool, 1)
	boot := &baseBootstrap{
		chRcvHdrNonce:    chRcvHdrNonce,
		shardCoordinator: mock.NewOneShardCoordinatorMock(),
		roundHandler:     &mock.RoundHandlerMock{RoundTimeDuration: 100 * time.Millisecond},
		forkDetector:     &mock.ForkDetectorMock{},
		headers: &mock.HeadersCacherStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
				return []data.HeaderHandler{header}, [][]byte{deadHash}, nil
			},
		},
		proofs: &testscommonDataRetriever.ProofsPoolMock{
			GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
				return &block.HeaderProof{HeaderHash: deadHash, HeaderNonce: nonce}, nil
			},
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return true
			},
		},
		blackListHandler: &testscommon.TimeCacheStub{
			HasCalled: func(key string) bool {
				return key == string(deadHash)
			},
		},
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		},
		requestHandler: &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, requestedNonce uint64) {
				requested = true
				// simulate the network answering so the wait returns immediately
				chRcvHdrNonce <- true
			},
		},
	}

	hdr, hash, err := boot.getHeaderWithNonceRequestingIfMissing(nonce)
	require.Nil(t, hdr)
	require.Nil(t, hash)
	require.Equal(t, process.ErrHeaderIsBlackListed, err)
	// the pool hit was ignored (first guard) and the header re-requested; the answer
	// still being blacklisted is then refused by the post-wait guard
	require.True(t, requested)
}

type epochStartDisarmerStub struct {
	disarmCalled func(epoch uint32, deadEpochStartHash []byte) bool
}

func (stub *epochStartDisarmerStub) DisarmDeadEpochStartActivation(epoch uint32, deadEpochStartHash []byte) bool {
	if stub.disarmCalled != nil {
		return stub.disarmCalled(epoch, deadEpochStartHash)
	}

	return false
}

func TestBaseBootstrap_ReconcileDivergence(t *testing.T) {
	t.Parallel()

	deadMetaHash := []byte("deadMetaHash")
	aliveMetaHash := []byte("aliveMetaHash")
	deadMeta := &block.MetaBlock{Nonce: 30}

	headHash, pointerHash := []byte("ownHash12"), []byte("ownHash11")
	pointerHeader := &block.HeaderV3{Nonce: 11, MetaBlockHashes: [][]byte{aliveMetaHash, deadMetaHash}}
	headHeader := &block.HeaderV3{Nonce: 12, PrevHash: pointerHash}

	type divergenceCalls struct {
		reconciledBelowNonce uint64
		rollBackNonce        uint64
		blacklisted          []string
		numReconcileBelow    int
	}

	buildBootstrapper := func(
		calls *divergenceCalls,
		checker settlementChecker,
		roundHandler *mock.RoundHandlerMock,
		reconcileBelowResult bool,
	) *baseBootstrap {
		return &baseBootstrap{
			settlementChecker:        checker,
			roundHandler:             roundHandler,
			divergenceEvaluatedRound: -1,
			chainHandler: &testscommon.ChainHandlerStub{
				GetCurrentBlockHeaderHashCalled: func() []byte {
					return headHash
				},
			},
			blockBootstrapper: &blockBootstrapperStub{
				getCurrHeaderCalled: func() (data.HeaderHandler, error) {
					return headHeader, nil
				},
				getPrevHeaderCalled: func(header data.HeaderHandler, _ storage.Storer) (data.HeaderHandler, error) {
					if header.GetNonce() == 12 {
						return pointerHeader, nil
					}
					return nil, errors.New("no previous header")
				},
			},
			forkDetector: &mock.ForkDetectorMock{
				GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
					return 5, nil
				},
				ReconcileFinalCheckpointBelowCalled: func(nonce uint64) bool {
					calls.reconciledBelowNonce = nonce
					calls.numReconcileBelow++
					return reconcileBelowResult
				},
				SetRollBackNonceCalled: func(nonce uint64) {
					calls.rollBackNonce = nonce
				},
			},
			blackListHandler: &testscommon.TimeCacheStub{
				AddCalled: func(key string) error {
					calls.blacklisted = append(calls.blacklisted, key)
					return nil
				},
			},
			statusHandler: &statusHandlerMock.AppStatusHandlerStub{},
		}
	}

	deadChecker := func() *settlementCheckerStub {
		return &settlementCheckerStub{
			deadCrossNotarizedMetaCalled: func() (data.HeaderHandler, []byte, bool) {
				return deadMeta, deadMetaHash, true
			},
		}
	}

	t.Run("fires once per round and arms the forced rollback below the earliest dead block", func(t *testing.T) {
		t.Parallel()

		calls := &divergenceCalls{}
		roundHandler := &mock.RoundHandlerMock{RoundIndex: 7}
		boot := buildBootstrapper(calls, deadChecker(), roundHandler, true)

		require.True(t, boot.tryReconcileDivergence())
		require.Equal(t, uint64(11), calls.reconciledBelowNonce)
		require.Equal(t, uint64(11), calls.rollBackNonce)
		require.Equal(t, []string{string(headHash), string(pointerHash)}, calls.blacklisted)

		// gated within the same round, fires again in the next one
		require.False(t, boot.tryReconcileDivergence())
		require.Equal(t, 1, calls.numReconcileBelow)
		roundHandler.RoundIndex = 8
		require.True(t, boot.tryReconcileDivergence())
		require.Equal(t, 2, calls.numReconcileBelow)
	})

	t.Run("no verdict from the authority is a no-op", func(t *testing.T) {
		t.Parallel()

		calls := &divergenceCalls{}
		boot := buildBootstrapper(calls, &settlementCheckerStub{}, &mock.RoundHandlerMock{RoundIndex: 7}, true)

		require.False(t, boot.tryReconcileDivergence())
		require.Zero(t, calls.numReconcileBelow)
		require.Empty(t, calls.blacklisted)
	})

	t.Run("aborts when the pointer block does not reference the dead meta", func(t *testing.T) {
		t.Parallel()

		calls := &divergenceCalls{}
		boot := buildBootstrapper(calls, deadChecker(), &mock.RoundHandlerMock{RoundIndex: 7}, true)
		boot.blockBootstrapper = &blockBootstrapperStub{
			getCurrHeaderCalled: func() (data.HeaderHandler, error) {
				return headHeader, nil
			},
			getPrevHeaderCalled: func(header data.HeaderHandler, _ storage.Storer) (data.HeaderHandler, error) {
				return &block.HeaderV3{Nonce: 11, MetaBlockHashes: [][]byte{aliveMetaHash}}, nil
			},
		}

		require.False(t, boot.tryReconcileDivergence())
		require.Zero(t, calls.numReconcileBelow)
		require.Empty(t, calls.blacklisted)
	})

	t.Run("aborts at the settled floor without a pointer block", func(t *testing.T) {
		t.Parallel()

		calls := &divergenceCalls{}
		boot := buildBootstrapper(calls, deadChecker(), &mock.RoundHandlerMock{RoundIndex: 7}, true)
		boot.forkDetector = &mock.ForkDetectorMock{
			GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
				return 11, nil
			},
			ReconcileFinalCheckpointBelowCalled: func(nonce uint64) bool {
				calls.numReconcileBelow++
				return true
			},
		}

		require.False(t, boot.tryReconcileDivergence())
		require.Zero(t, calls.numReconcileBelow)
	})

	t.Run("a refused finality regression arms nothing", func(t *testing.T) {
		t.Parallel()

		calls := &divergenceCalls{}
		boot := buildBootstrapper(calls, deadChecker(), &mock.RoundHandlerMock{RoundIndex: 7}, false)

		require.False(t, boot.tryReconcileDivergence())
		require.Equal(t, 1, calls.numReconcileBelow)
		require.Zero(t, calls.rollBackNonce)
		require.Empty(t, calls.blacklisted)
	})

	t.Run("disarms a dead epoch start activation", func(t *testing.T) {
		t.Parallel()

		deadEpochStartMeta := &block.MetaBlock{
			Nonce: 30,
			Epoch: 3,
			EpochStart: block.EpochStart{
				LastFinalizedHeaders: []block.EpochStartShardData{{ShardID: 0}},
			},
		}
		checker := &settlementCheckerStub{
			deadCrossNotarizedMetaCalled: func() (data.HeaderHandler, []byte, bool) {
				return deadEpochStartMeta, deadMetaHash, true
			},
		}

		var disarmedEpoch uint32
		var disarmedHash []byte
		calls := &divergenceCalls{}
		boot := buildBootstrapper(calls, checker, &mock.RoundHandlerMock{RoundIndex: 7}, true)
		boot.epochStartDisarmer = &epochStartDisarmerStub{
			disarmCalled: func(epoch uint32, deadEpochStartHash []byte) bool {
				disarmedEpoch = epoch
				disarmedHash = deadEpochStartHash
				return true
			},
		}

		require.True(t, boot.tryReconcileDivergence())
		require.Equal(t, uint32(3), disarmedEpoch)
		require.Equal(t, deadMetaHash, disarmedHash)
	})
}

func TestBaseBootstrap_RollBackOneBlockV3RevertsEpochStartTrigger(t *testing.T) {
	t.Parallel()

	prevHeader := &block.HeaderV3{Nonce: 11}
	currHeader := &block.HeaderV3{Nonce: 12}
	currHash, prevHash := []byte("currHash"), []byte("prevHash")

	buildBootstrapper := func(reverted *[]data.HeaderHandler, revertErr error) *baseBootstrap {
		return &baseBootstrap{
			chainHandler: &testscommon.ChainHandlerStub{},
			epochStartTrigger: &testscommon.EpochStartTriggerStub{
				RevertStateToBlockCalled: func(header data.HeaderHandler) error {
					*reverted = append(*reverted, header)
					return revertErr
				},
			},
			executionManager:     &processMocks.ExecutionManagerMock{},
			blockBootstrapper:    &blockBootstrapperStub{},
			blockProcessor:       &testscommon.BlockProcessorStub{},
			headers:              &mock.HeadersCacherStub{},
			forkDetector:         &mock.ForkDetectorMock{},
			marshalizer:          &mock.MarshalizerMock{},
			hasher:               &hashingMocks.HasherMock{},
			uint64Converter:      &mock.Uint64ByteSliceConverterMock{},
			headerNonceHashStore: &storageStubs.StorerStub{},
		}
	}

	t.Run("reverts the trigger to the new head on every rolled back block", func(t *testing.T) {
		t.Parallel()

		reverted := make([]data.HeaderHandler, 0)
		boot := buildBootstrapper(&reverted, nil)

		_, err := boot.rollBackOneBlockV3(currHash, currHeader, prevHash, prevHeader)
		require.Nil(t, err)
		require.Equal(t, []data.HeaderHandler{prevHeader}, reverted)
	})

	t.Run("a failing trigger revert aborts the roll back and restores the head", func(t *testing.T) {
		t.Parallel()

		expectedRevertErr := errors.New("revert error")
		reverted := make([]data.HeaderHandler, 0)
		boot := buildBootstrapper(&reverted, expectedRevertErr)

		restoredHashes := make([][]byte, 0)
		boot.chainHandler = &testscommon.ChainHandlerStub{
			SetCurrentBlockHeaderAndHashCalled: func(hash []byte, header data.HeaderHandler) error {
				restoredHashes = append(restoredHashes, hash)
				return nil
			},
		}

		_, err := boot.rollBackOneBlockV3(currHash, currHeader, prevHash, prevHeader)
		require.Equal(t, expectedRevertErr, err)
		require.Equal(t, [][]byte{prevHash, currHash}, restoredHashes)
	})
}
