package sync

import (
	"bytes"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	testscommonDataRetriever "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

type recoveryRequestHandlerStub struct {
	testscommon.RequestHandlerStub
	requestInterval time.Duration
}

func (stub *recoveryRequestHandlerStub) RequestInterval() time.Duration {
	return stub.requestInterval
}

func newRecoveryBootstrap(
	roundHandler *mock.RoundHandlerMock,
	currentHeader data.HeaderHandler,
	probableNonce *uint64,
	requestHandler process.RequestHandler,
) *baseBootstrap {
	return &baseBootstrap{
		roundHandler: roundHandler,
		chainHandler: &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled:      func() data.HeaderHandler { return &block.Header{} },
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler { return currentHeader },
		},
		forkDetector: &mock.ForkDetectorMock{
			ProbableHighestNonceCalled: func() uint64 { return *probableNonce },
		},
		shardCoordinator: &mock.ShardCoordinatorStub{
			SelfIdCalled: func() uint32 { return 1 },
		},
		networkWatcher: &mock.NetworkConnectionWatcherStub{
			IsConnectedToTheNetworkCalled: func() bool { return true },
		},
		enableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, _ uint32) bool {
				return flag == common.AndromedaFlag
			},
		},
		requestHandler: requestHandler,
		proofs: &testscommonDataRetriever.ProofsPoolMock{
			HasProofCalled: func(_ uint32, _ []byte) bool { return false },
		},
		headers:     &mock.HeadersCacherStub{},
		marshalizer: &marshal.GogoProtoMarshalizer{},
		hasher:      &hashingMocks.HasherMock{},
		forkInfo:    process.NewForkInfo(),
		processConfigsHandler: &testscommon.ProcessConfigsHandlerStub{
			GetMaxRoundsWithoutNewBlockReceivedByRoundCalled: func(_ uint64) uint32 { return 100 },
			GetRoundModulusTriggerWhenSyncIsStuckCalled:      func(_ uint64) uint32 { return 200 },
		},
	}
}

func TestResyncRecovery_StableParentRequestsHeaderThenProof(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	probableNonce := uint64(5)
	currentHeader := &block.Header{ShardID: 1, Nonce: probableNonce, Round: 5, Epoch: 6}
	parent := &block.Header{ShardID: 1, Nonce: 9, Round: 9, Epoch: 6}
	parentHash, err := core.CalculateHash(&marshal.GogoProtoMarshalizer{}, &hashingMocks.HasherMock{}, parent)
	require.NoError(t, err)
	child := &block.Header{ShardID: 1, Nonce: 10, Round: 10, Epoch: 7, PrevHash: parentHash}
	_ = child.SetEpochStartMetaHash([]byte("epoch start"))

	var requestedHeaderHash []byte
	var requestedHeaderEpoch uint32
	var requestedProofHash []byte
	requestHandler := &recoveryRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestShardHeaderForEpochCalled: func(_ uint32, hash []byte, epoch uint32) {
				requestedHeaderHash = append([]byte(nil), hash...)
				requestedHeaderEpoch = epoch
			},
			RequestEquivalentProofByHashForEpochCalled: func(_ uint32, hash []byte, _ uint32) {
				requestedProofHash = append([]byte(nil), hash...)
			},
		},
	}
	boot := newRecoveryBootstrap(roundHandler, currentHeader, &probableNonce, requestHandler)
	boot.headers = &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(_ []byte) (data.HeaderHandler, error) {
			return nil, errors.New("missing")
		},
	}

	boot.observeRecoveryHeader(child)
	boot.evaluateFastRecovery(roundHandler.RoundIndex)
	require.Empty(t, requestedHeaderHash)

	roundHandler.RoundIndex++
	boot.observeRecoveryHeader(child)
	boot.evaluateFastRecovery(roundHandler.RoundIndex)
	require.True(t, bytes.Equal(parentHash, requestedHeaderHash))
	require.Equal(t, uint32(6), requestedHeaderEpoch)
	require.Empty(t, requestedProofHash)

	boot.headers = &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			require.True(t, bytes.Equal(parentHash, hash))
			return parent, nil
		},
	}
	roundHandler.RoundIndex++
	boot.evaluateFastRecovery(roundHandler.RoundIndex)
	require.True(t, bytes.Equal(parentHash, requestedProofHash))
}

func TestResyncRecovery_RotatingParentsDoNotArmCandidate(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	probableNonce := uint64(5)
	currentHeader := &block.Header{ShardID: 1, Nonce: probableNonce, Round: 5, Epoch: 6}
	numRequests := 0
	requestHandler := &recoveryRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestShardHeaderForEpochCalled: func(_ uint32, _ []byte, _ uint32) { numRequests++ },
		},
	}
	boot := newRecoveryBootstrap(roundHandler, currentHeader, &probableNonce, requestHandler)

	for idx := 0; idx < maxResyncRecoveryCandidates+1; idx++ {
		child := &block.Header{
			ShardID:  1,
			Nonce:    10,
			Epoch:    6,
			PrevHash: []byte{byte(idx + 1)},
		}
		boot.observeRecoveryHeader(child)
		boot.evaluateFastRecovery(roundHandler.RoundIndex)
		roundHandler.RoundIndex++
	}

	require.Zero(t, numRequests)
	activeCandidates := 0
	for idx := range boot.recoveryState.candidates {
		if boot.recoveryState.candidates[idx].active {
			activeCandidates++
		}
	}
	require.Equal(t, maxResyncRecoveryCandidates, activeCandidates)
}

func TestResyncRecovery_StaleEvaluationRoundDoesNotResetFresherRecoveryState(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 101}
	probableNonce := uint64(5)
	currentHeader := &block.Header{ShardID: 1, Nonce: probableNonce, Round: 5, Epoch: 6}
	boot := newRecoveryBootstrap(roundHandler, currentHeader, &probableNonce, &recoveryRequestHandlerStub{})
	boot.recoveryState.chronologySet = true
	boot.recoveryState.lastChronologyRound = 101
	boot.recoveryState.candidates[0] = resyncRecoveryCandidate{
		active:         true,
		firstRound:     101,
		observations:   1,
		committedNonce: probableNonce,
		probableNonce:  probableNonce,
	}
	boot.recoveryState.bypass = postBootstrapWatchdogBypass{
		armed:          true,
		generation:     1,
		armedRound:     101,
		committedNonce: probableNonce,
		probableNonce:  probableNonce,
	}
	boot.recoveryActive.Store(true)
	boot.recoveryBypass.Store(true)

	boot.evaluateFastRecovery(100)

	require.True(t, boot.recoveryState.candidates[0].active)
	require.True(t, boot.recoveryState.bypass.armed)
	require.True(t, boot.recoveryActive.Load())
	require.True(t, boot.recoveryBypass.Load())
	require.Equal(t, int64(101), boot.recoveryState.lastChronologyRound)
}

func TestResyncRecovery_StaleActionDoesNotRequestAfterClose(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	probableNonce := uint64(5)
	currentHeader := &block.Header{ShardID: 1, Nonce: probableNonce, Round: 5, Epoch: 6}
	child := &block.Header{ShardID: 1, Nonce: 10, Round: 10, Epoch: 6, PrevHash: []byte("parent")}
	numRequests := 0
	requestHandler := &recoveryRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestShardHeaderForEpochCalled: func(_ uint32, _ []byte, _ uint32) { numRequests++ },
		},
	}
	boot := newRecoveryBootstrap(roundHandler, currentHeader, &probableNonce, requestHandler)
	boot.headers = &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(_ []byte) (data.HeaderHandler, error) {
			boot.closeRecovery()
			return nil, errors.New("missing")
		},
	}

	boot.observeRecoveryHeader(child)
	roundHandler.RoundIndex++
	boot.observeRecoveryHeader(child)
	boot.evaluateFastRecovery(roundHandler.RoundIndex)

	require.Zero(t, numRequests)
}

func TestResyncRecovery_InvalidParentIsExpiredAndCooledDown(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	probableNonce := uint64(5)
	currentHeader := &block.Header{ShardID: 1, Nonce: probableNonce, Round: 5, Epoch: 6}
	numProofRequests := 0
	requestHandler := &recoveryRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestEquivalentProofByHashForEpochCalled: func(_ uint32, _ []byte, _ uint32) { numProofRequests++ },
		},
	}
	boot := newRecoveryBootstrap(roundHandler, currentHeader, &probableNonce, requestHandler)
	boot.headers = &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(_ []byte) (data.HeaderHandler, error) {
			return &block.Header{ShardID: 1, Nonce: 9, Epoch: 6}, nil
		},
	}
	action := resyncRecoveryAction{
		generation: 1,
		parentHash: []byte("different hash"),
		childNonce: 10,
		childEpoch: 6,
	}
	boot.recoveryState.candidates[0] = resyncRecoveryCandidate{
		active:       true,
		generation:   action.generation,
		parentHash:   action.parentHash,
		childNonce:   action.childNonce,
		childEpoch:   action.childEpoch,
		firstRound:   roundHandler.RoundIndex,
		observations: 2,
	}
	boot.recoveryActive.Store(true)

	boot.executeFastRecoveryAction(action, roundHandler.RoundIndex)

	require.Zero(t, numProofRequests)
	require.False(t, boot.recoveryState.candidates[0].active)
	boot.mutRecovery.Lock()
	isCoolingDown := boot.isRecoveryHashCoolingDownLocked(action.parentHash, roundHandler.RoundIndex)
	boot.mutRecovery.Unlock()
	require.True(t, isCoolingDown)
}

func TestExpectedRecoveryParentEpoch(t *testing.T) {
	t.Parallel()

	epoch, ok := expectedRecoveryParentEpoch(7, false)
	require.True(t, ok)
	require.Equal(t, uint32(7), epoch)

	epoch, ok = expectedRecoveryParentEpoch(7, true)
	require.True(t, ok)
	require.Equal(t, uint32(6), epoch)

	_, ok = expectedRecoveryParentEpoch(0, true)
	require.False(t, ok)
}

func TestResyncRecovery_WatchdogBypassIsBoundedAndStopsOnProgress(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 150}
	probableNonce := uint64(10)
	currentHeader := &block.Header{ShardID: 1, Nonce: probableNonce, Round: 1, Epoch: 6}
	requestHandler := &recoveryRequestHandlerStub{}
	boot := newRecoveryBootstrap(roundHandler, currentHeader, &probableNonce, requestHandler)
	boot.isNodeSynchronized = true

	boot.armPostBootstrapWatchdogBypass()
	for round := int64(150); round < 156; round++ {
		roundHandler.RoundIndex = round
		shouldRequest, generation := boot.shouldTryToRequestHeaders()
		require.True(t, shouldRequest)
		require.NotZero(t, generation)
	}
	roundHandler.RoundIndex = 156
	shouldRequest, _ := boot.shouldTryToRequestHeaders()
	require.False(t, shouldRequest)

	boot.armPostBootstrapWatchdogBypass()
	probableNonce++
	shouldRequest, _ = boot.shouldTryToRequestHeaders()
	require.False(t, shouldRequest)
}

func TestResyncRecovery_RequestHeadersIfStuckRejectsBackwardRound(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 99}
	probableNonce := uint64(10)
	currentHeader := &block.Header{ShardID: 1, Nonce: probableNonce, Round: 100, Epoch: 6}
	numRequests := 0
	boot := newRecoveryBootstrap(roundHandler, currentHeader, &probableNonce, &recoveryRequestHandlerStub{})
	boot.blockBootstrapper = &blockBootstrapperStub{
		requestProofByNonceCalled: func(_ uint64) { numRequests++ },
	}

	boot.requestHeadersIfSyncIsStuck()
	require.Zero(t, numRequests)

	roundHandler.RoundIndex = 100
	boot.requestHeadersIfSyncIsStuck()
	require.Zero(t, numRequests)
}

func TestBaseBootstrap_RequestByHashRechecksHeaderAfterRegisteringExpectation(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	probableNonce := uint64(5)
	headerHash := []byte("header hash")
	header := &block.Header{ShardID: 1, Nonce: 6, Epoch: 6}
	headerRequests := 0
	proofRequests := 0
	requestHandler := &recoveryRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestShardHeaderCalled: func(_ uint32, _ []byte) { headerRequests++ },
			RequestEquivalentProofByHashForEpochCalled: func(_ uint32, hash []byte, epoch uint32) {
				require.Equal(t, headerHash, hash)
				require.Equal(t, header.GetEpoch(), epoch)
				proofRequests++
			},
		},
	}
	boot := newRecoveryBootstrap(roundHandler, header, &probableNonce, requestHandler)
	boot.chRcvHdrHash = make(chan bool)
	boot.headers = &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			require.Equal(t, headerHash, hash)
			return header, nil
		},
	}

	readyHeader := boot.requestHeaderAndProofByHashIfMissing(headerHash, nil, true, true)

	require.Nil(t, readyHeader)
	require.Zero(t, headerRequests)
	require.Equal(t, 1, proofRequests)
	require.Equal(t, headerHash, boot.requestedHeaderHash())
}

func TestBaseBootstrap_GetHeaderByHashRehydratesStoredHeaderBeforeRequestingProof(t *testing.T) {
	t.Parallel()

	headerHash := []byte("header hash")
	header := &block.Header{ShardID: 1, Nonce: 6, Round: 10, Epoch: 6}
	marshaledHeader, err := (&marshal.GogoProtoMarshalizer{}).Marshal(header)
	require.NoError(t, err)

	proof := &block.HeaderProof{
		HeaderHash:    headerHash,
		HeaderShardId: header.GetShardID(),
		HeaderNonce:   header.GetNonce(),
		HeaderRound:   header.GetRound(),
		HeaderEpoch:   header.GetEpoch(),
	}
	proofAvailable := false
	headerRequests := 0
	proofRequests := 0
	var boot *baseBootstrap
	requestHandler := &recoveryRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestShardHeaderCalled: func(_ uint32, _ []byte) {
				headerRequests++
			},
			RequestEquivalentProofByHashForEpochCalled: func(_ uint32, hash []byte, epoch uint32) {
				require.Equal(t, headerHash, hash)
				require.Equal(t, header.GetEpoch(), epoch)
				proofRequests++
				proofAvailable = true
				boot.processReceivedProof(proof)
			},
		},
	}

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10, RoundTimeDuration: time.Second}
	probableNonce := uint64(5)
	boot = newRecoveryBootstrap(roundHandler, header, &probableNonce, requestHandler)
	boot.chRcvHdrHash = make(chan bool, 1)
	boot.requestMiniBlocks = func(_ data.HeaderHandler) {}
	boot.proofs = &testscommonDataRetriever.ProofsPoolMock{
		GetProofCalled: func(_ uint32, _ []byte) (data.HeaderProofHandler, error) {
			if proofAvailable {
				return proof, nil
			}

			return nil, errors.New("missing proof")
		},
		HasProofCalled: func(_ uint32, _ []byte) bool {
			return proofAvailable
		},
	}

	var pooledHeader data.HeaderHandler
	boot.headers = &mock.HeadersCacherStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			require.Equal(t, headerHash, hash)
			if pooledHeader == nil {
				return nil, errors.New("missing header")
			}

			return pooledHeader, nil
		},
		AddCalled: func(hash []byte, storedHeader data.HeaderHandler) {
			require.Equal(t, headerHash, hash)
			require.Equal(t, headerHash, boot.requestedHeaderHash())
			pooledHeader = storedHeader
			boot.processReceivedHeader(storedHeader, hash)
		},
	}
	boot.store = &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			require.Equal(t, dataRetriever.BlockHeaderUnit, unitType)
			return &storageStubs.StorerStub{
				GetCalled: func(hash []byte) ([]byte, error) {
					require.Equal(t, headerHash, hash)
					return marshaledHeader, nil
				},
			}, nil
		},
	}

	result, err := boot.getHeaderWithHashRequestingIfMissing(headerHash)

	require.NoError(t, err)
	require.Equal(t, header.GetNonce(), result.GetNonce())
	require.Equal(t, 0, headerRequests)
	require.Equal(t, 1, proofRequests)
	require.Nil(t, boot.requestedHeaderHash())
}

func TestBaseBootstrap_RequestByNonceRechecksHeaderAfterRegisteringExpectation(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	probableNonce := uint64(5)
	headerHash := []byte("header hash")
	header := &block.Header{ShardID: 1, Nonce: 6, Epoch: 6}
	headerRequests := 0
	proofRequests := 0
	requestHandler := &recoveryRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(_ uint32, nonce uint64) {
				require.Equal(t, header.GetNonce(), nonce)
				headerRequests++
			},
			RequestEquivalentProofByHashForEpochCalled: func(_ uint32, hash []byte, epoch uint32) {
				require.Equal(t, headerHash, hash)
				require.Equal(t, header.GetEpoch(), epoch)
				proofRequests++
			},
		},
	}
	boot := newRecoveryBootstrap(roundHandler, header, &probableNonce, requestHandler)
	boot.chRcvHdrNonce = make(chan bool)
	boot.blackListHandler = &testscommon.TimeCacheStub{}
	boot.headers = &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(nonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
			require.Equal(t, header.GetNonce(), nonce)
			require.Equal(t, uint32(1), shardID)
			return []data.HeaderHandler{header}, [][]byte{headerHash}, nil
		},
	}

	readyHeader, readyHash := boot.requestHeaderAndProofByNonce(nil, nil, header.GetNonce(), true)

	require.Nil(t, readyHeader)
	require.Nil(t, readyHash)
	require.Equal(t, 1, headerRequests)
	require.Equal(t, 1, proofRequests)
	require.Equal(t, header.GetNonce(), *boot.requestedHeaderNonce())
}

func TestBaseBootstrap_RequestByHashRechecksProofAfterRegisteringExpectation(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	probableNonce := uint64(5)
	headerHash := []byte("header hash")
	header := &block.Header{ShardID: 1, Nonce: 6, Epoch: 6}
	proofRequests := 0
	requestHandler := &recoveryRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestEquivalentProofByHashForEpochCalled: func(_ uint32, _ []byte, _ uint32) {
				proofRequests++
			},
		},
	}
	boot := newRecoveryBootstrap(roundHandler, header, &probableNonce, requestHandler)
	boot.chRcvHdrHash = make(chan bool)
	boot.proofs = &testscommonDataRetriever.ProofsPoolMock{
		HasProofCalled: func(shardID uint32, hash []byte) bool {
			require.Equal(t, uint32(1), shardID)
			require.Equal(t, headerHash, hash)
			return true
		},
	}

	readyHeader := boot.requestHeaderAndProofByHashIfMissing(headerHash, header, false, true)

	require.Same(t, header, readyHeader)
	require.Zero(t, proofRequests)
	require.Nil(t, boot.requestedHeaderHash())
}

func TestBaseBootstrap_RequestByNonceRechecksProofAfterRegisteringExpectation(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 10}
	probableNonce := uint64(5)
	headerHash := []byte("header hash")
	header := &block.Header{ShardID: 1, Nonce: 6, Epoch: 6}
	headerRequests := 0
	proofRequests := 0
	requestHandler := &recoveryRequestHandlerStub{
		RequestHandlerStub: testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(_ uint32, _ uint64) { headerRequests++ },
			RequestEquivalentProofByHashForEpochCalled: func(_ uint32, _ []byte, _ uint32) {
				proofRequests++
			},
		},
	}
	boot := newRecoveryBootstrap(roundHandler, header, &probableNonce, requestHandler)
	boot.chRcvHdrNonce = make(chan bool)
	boot.blackListHandler = &testscommon.TimeCacheStub{}
	boot.proofs = &testscommonDataRetriever.ProofsPoolMock{
		HasProofCalled: func(shardID uint32, hash []byte) bool {
			require.Equal(t, uint32(1), shardID)
			require.Equal(t, headerHash, hash)
			return true
		},
	}

	readyHeader, readyHash := boot.requestHeaderAndProofByNonce(headerHash, header, header.GetNonce(), true)

	require.Same(t, header, readyHeader)
	require.Equal(t, headerHash, readyHash)
	require.Zero(t, headerRequests)
	require.Zero(t, proofRequests)
	require.Nil(t, boot.requestedHeaderNonce())
}

func TestResyncRecovery_StaleWatchdogGenerationDoesNotRequest(t *testing.T) {
	t.Parallel()

	roundHandler := &mock.RoundHandlerMock{RoundIndex: 150}
	probableNonce := uint64(10)
	currentHeader := &block.Header{ShardID: 1, Nonce: probableNonce, Round: 1, Epoch: 6}
	boot := newRecoveryBootstrap(roundHandler, currentHeader, &probableNonce, &recoveryRequestHandlerStub{})
	boot.isNodeSynchronized = true
	numRequests := 0
	boot.blockBootstrapper = &blockBootstrapperStub{
		requestProofByNonceCalled: func(_ uint64) { numRequests++ },
	}

	boot.armPostBootstrapWatchdogBypass()
	shouldRequest, generation := boot.shouldTryToRequestHeaders()
	require.True(t, shouldRequest)
	require.NotZero(t, generation)
	probableNonce++
	boot.clearRecoveryAfterProgress()
	boot.requestHeadersIfSyncIsStuckForGeneration(generation)

	require.Zero(t, numRequests)
}

func TestResyncRecovery_ClearInactiveStateHasNoDependencies(t *testing.T) {
	t.Parallel()

	boot := &baseBootstrap{}
	require.NotPanics(t, boot.clearRecoveryAfterProgress)
}

func TestBaseBootstrap_ShouldTryToRequestHeadersGatesWatchdogOnKnownBacklog(t *testing.T) {
	t.Parallel()

	newUnsyncedBoot := func(committedNonce uint64, probableNonce *uint64) *baseBootstrap {
		currentHeader := &block.Header{ShardID: 1, Nonce: committedNonce, Round: 1, Epoch: 6}
		boot := newRecoveryBootstrap(&mock.RoundHandlerMock{RoundIndex: 150}, currentHeader, probableNonce, &recoveryRequestHandlerStub{})
		boot.isNodeSynchronized = false
		return boot
	}

	t.Run("probable ahead of committed does not start the watchdog", func(t *testing.T) {
		t.Parallel()

		probableNonce := uint64(30)
		boot := newUnsyncedBoot(10, &probableNonce)

		shouldRequest, generation := boot.shouldTryToRequestHeaders()
		require.False(t, shouldRequest)
		require.Zero(t, generation)
	})
	t.Run("probable equal to committed keeps watchdog discovery", func(t *testing.T) {
		t.Parallel()

		probableNonce := uint64(10)
		boot := newUnsyncedBoot(10, &probableNonce)

		shouldRequest, generation := boot.shouldTryToRequestHeaders()
		require.True(t, shouldRequest)
		require.Zero(t, generation)
	})
	t.Run("probable behind committed keeps watchdog discovery", func(t *testing.T) {
		t.Parallel()

		probableNonce := uint64(5)
		boot := newUnsyncedBoot(10, &probableNonce)

		shouldRequest, _ := boot.shouldTryToRequestHeaders()
		require.True(t, shouldRequest)
	})
	t.Run("forced rollback is excluded before the backlog gate", func(t *testing.T) {
		t.Parallel()

		probableNonce := uint64(10)
		boot := newUnsyncedBoot(10, &probableNonce)
		boot.forkInfo.IsDetected = true
		boot.forkInfo.Nonce = math.MaxUint64
		boot.forkInfo.Hash = nil

		shouldRequest, _ := boot.shouldTryToRequestHeaders()
		require.False(t, shouldRequest)

		boot.forkInfo = process.NewForkInfo()
		boot.forkInfo.IsDetected = true
		boot.forkInfo.Round = math.MaxUint64
		boot.forkInfo.Hash = nil

		shouldRequest, _ = boot.shouldTryToRequestHeaders()
		require.False(t, shouldRequest)
	})
	t.Run("synchronized node is unaffected by the gate", func(t *testing.T) {
		t.Parallel()

		probableNonce := uint64(10)
		boot := newUnsyncedBoot(10, &probableNonce)
		boot.isNodeSynchronized = true
		boot.roundHandler = &mock.RoundHandlerMock{RoundIndex: 200}

		shouldRequest, _ := boot.shouldTryToRequestHeaders()
		require.True(t, shouldRequest)

		boot.roundHandler = &mock.RoundHandlerMock{RoundIndex: 201}
		shouldRequest, _ = boot.shouldTryToRequestHeaders()
		require.False(t, shouldRequest)
	})
}

func TestBaseBootstrap_LookaheadRequestsWindowWhileProbableIsAhead(t *testing.T) {
	t.Parallel()

	probableNonce := uint64(13)
	currentHeader := &block.Header{ShardID: 1, Nonce: 10, Round: 1, Epoch: 6}
	boot := newRecoveryBootstrap(&mock.RoundHandlerMock{RoundIndex: 150}, currentHeader, &probableNonce, &recoveryRequestHandlerStub{})
	boot.proofs = &testscommonDataRetriever.ProofsPoolMock{
		GetProofByNonceCalled: func(_ uint64, _ uint32) (data.HeaderProofHandler, error) {
			return nil, errors.New("missing proof")
		},
	}
	requestedProofNonces := make([]uint64, 0)
	boot.blockBootstrapper = &blockBootstrapperStub{
		requestProofByNonceCalled: func(nonce uint64) { requestedProofNonces = append(requestedProofNonces, nonce) },
	}

	boot.requestHeadersFromNonceIfMissing(11)

	require.Equal(t, []uint64{11, 12, 13}, requestedProofNonces)
}
