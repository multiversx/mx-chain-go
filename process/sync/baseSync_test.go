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
