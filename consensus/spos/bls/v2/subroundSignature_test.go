package v2_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	v2 "github.com/multiversx/mx-chain-go/consensus/spos/bls/v2"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/consensus/initializers"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

const setThresholdJobsDone = "threshold"

func initSubroundSignatureWithContainer(container *spos.ConsensusCore) v2.SubroundSignature {
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.7,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	srSignature, _ := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
		v2.NewSignatureEvidenceStore(nil),
	)

	return srSignature
}

func initSubroundSignature() v2.SubroundSignature {
	container := consensusMocks.InitConsensusCore()
	return initSubroundSignatureWithContainer(container)
}

func TestNewSubroundSignature(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.7,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	t.Run("nil subround should error", func(t *testing.T) {
		t.Parallel()

		srSignature, err := v2.NewSubroundSignature(
			nil,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		assert.Nil(t, srSignature)
		assert.Equal(t, spos.ErrNilSubround, err)
	})
	t.Run("nil worker should error", func(t *testing.T) {
		t.Parallel()

		srSignature, err := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			nil,
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		assert.Nil(t, srSignature)
		assert.Equal(t, spos.ErrNilWorker, err)
	})
	t.Run("nil app status handler should error", func(t *testing.T) {
		t.Parallel()

		srSignature, err := v2.NewSubroundSignature(
			sr,
			nil,
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		assert.Nil(t, srSignature)
		assert.Equal(t, spos.ErrNilAppStatusHandler, err)
	})
	t.Run("nil sent signatures tracker should error", func(t *testing.T) {
		t.Parallel()

		srSignature, err := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			nil,
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		assert.Nil(t, srSignature)
		assert.Equal(t, v2.ErrNilSentSignatureTracker, err)
	})

	t.Run("nil signatureThrottler should error", func(t *testing.T) {
		t.Parallel()

		srSignature, err := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			nil,
			v2.NewSignatureEvidenceStore(nil),
		)

		assert.Nil(t, srSignature)
		assert.Equal(t, spos.ErrNilThrottler, err)
	})

	t.Run("nil signature evidence should error", func(t *testing.T) {
		t.Parallel()

		srSignature, err := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			nil,
		)

		assert.Nil(t, srSignature)
		assert.Equal(t, v2.ErrNilSignatureEvidence, err)
	})
}

func TestSubroundSignature_NewSubroundSignatureNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.7,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	sr.ConsensusStateHandler = nil
	srSignature, err := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
		v2.NewSignatureEvidenceStore(nil),
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundSignature_NewSubroundSignatureNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.7,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetHasher(nil)
	srSignature, err := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
		v2.NewSignatureEvidenceStore(nil),
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubroundSignature_NewSubroundSignatureNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.7,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetMultiSignerContainer(nil)
	srSignature, err := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
		v2.NewSignatureEvidenceStore(nil),
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestSubroundSignature_NewSubroundSignatureNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.7,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetRoundHandler(nil)

	srSignature, err := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
		v2.NewSignatureEvidenceStore(nil),
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestSubroundSignature_NewSubroundSignatureNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.7,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetSyncTimer(nil)
	srSignature, err := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
		v2.NewSignatureEvidenceStore(nil),
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundSignature_NewSubroundSignatureNilAppStatusHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.7,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	srSignature, err := v2.NewSubroundSignature(
		sr,
		nil,
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
		v2.NewSignatureEvidenceStore(nil),
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilAppStatusHandler, err)
}

func TestSubroundSignature_NewSubroundSignatureShouldWork(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.7,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	srSignature, err := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
		v2.NewSignatureEvidenceStore(nil),
	)

	assert.False(t, check.IfNil(srSignature))
	assert.Nil(t, err)
}

func TestSubroundSignature_DoSignatureJobRefusesSecondHashInSameRound(t *testing.T) {
	t.Parallel()

	container := consensusMocks.InitConsensusCore()
	consensusState := initializers.InitConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		roundTimeDuration,
		0.7,
		0.85,
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	broadcastCalled := false
	container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
			broadcastCalled = true
			return nil
		},
	})

	srSignature, _ := v2.NewSubroundSignature(
		sr,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{
			ReserveSignatureInRoundCalled: func(pkBytes []byte, roundIndex int64, headerHash []byte) bool {
				return false
			},
		},
		&consensusMocks.SposWorkerMock{},
		&dataRetrieverMock.ThrottlerStub{},
		v2.NewSignatureEvidenceStore(nil),
	)

	srSignature.SetHeader(&block.Header{})
	leader, err := srSignature.GetLeader()
	assert.Nil(t, err)
	srSignature.SetSelfPubKey(leader)

	r := srSignature.DoSignatureJob()
	assert.False(t, r)
	assert.False(t, broadcastCalled)
}

func TestSubroundSignature_DoSignatureJob(t *testing.T) {
	t.Parallel()

	t.Run("job done should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundSignatureWithContainer(container)
		sr.SetStatus(bls.SrSignature, spos.SsFinished)

		r := sr.DoSignatureJob()
		assert.False(t, r)
	})
	t.Run("nil header should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundSignatureWithContainer(container)
		sr.SetHeader(nil)

		r := sr.DoSignatureJob()
		assert.False(t, r)
	})
	t.Run("proof already received should return true", func(t *testing.T) {
		t.Parallel()

		providedHash := []byte("providedHash")
		container := consensusMocks.InitConsensusCore()
		container.SetEquivalentProofsPool(&dataRetriever.ProofsPoolMock{
			HasProofCalled: func(shardID uint32, headerHash []byte) bool {
				return string(headerHash) == string(providedHash)
			},
		})
		sr := initSubroundSignatureWithContainer(container)
		sr.SetData(providedHash)
		sr.SetHeader(&block.Header{})

		r := sr.DoSignatureJob()
		assert.True(t, r)
	})
	t.Run("single key error should return false", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundSignatureWithContainer(container)

		sr.SetHeader(&block.Header{})
		leader, err := sr.GetLeader()
		assert.Nil(t, err)
		sr.SetSelfPubKey(leader)
		container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return expectedErr
			},
		})
		r := sr.DoSignatureJob()
		assert.False(t, r)
	})
	t.Run("single key mode should work", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundSignatureWithContainer(container)

		sr.SetHeader(&block.Header{})
		leader, err := sr.GetLeader()
		assert.Nil(t, err)
		sr.SetSelfPubKey(leader)
		container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				if string(message.PubKey) != leader || message.MsgType != int64(bls.MtSignature) {
					assert.Fail(t, "should have not been called")
				}
				return nil
			},
		})
		r := sr.DoSignatureJob()
		assert.True(t, r)

		assert.False(t, sr.GetRoundCanceled())
		assert.Nil(t, err)
		leaderJobDone, err := sr.JobDone(leader, bls.SrSignature)
		assert.NoError(t, err)
		assert.True(t, leaderJobDone)
		assert.True(t, sr.IsSubroundFinished(bls.SrSignature))
	})
	t.Run("multikey mode should work", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		signingHandler := &consensusMocks.SigningHandlerStub{
			CreateSignatureShareForPublicKeyCalled: func(_ context.Context, msg []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
				return []byte("SIG"), nil
			},
		}
		container.SetSigningHandler(signingHandler)
		consensusState := initializers.InitConsensusStateWithKeysHandler(
			&testscommon.KeysHandlerStub{
				IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
					return true
				},
			},
		)
		ch := make(chan bool, 1)

		sr, _ := spos.NewSubround(
			bls.SrBlock,
			bls.SrSignature,
			bls.SrEndRound,
			roundTimeDuration,
			0.7,
			0.85,
			"(SIGNATURE)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)

		signatureSentForPks := make(map[string]struct{})
		mutex := sync.Mutex{}
		srSignature, _ := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{
				SignatureSentCalled: func(pkBytes []byte) {
					mutex.Lock()
					signatureSentForPks[string(pkBytes)] = struct{}{}
					mutex.Unlock()
				},
			},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		sr.SetHeader(&block.Header{})
		signaturesBroadcast := make(map[string]int)
		container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				mutex.Lock()
				signaturesBroadcast[string(message.PubKey)]++
				mutex.Unlock()
				return nil
			},
		})

		sr.SetSelfPubKey("OTHER")

		r := srSignature.DoSignatureJob()
		assert.True(t, r)

		assert.False(t, sr.GetRoundCanceled())
		assert.True(t, sr.IsSubroundFinished(bls.SrSignature))

		for _, pk := range sr.ConsensusGroup() {
			isJobDone, err := sr.JobDone(pk, bls.SrSignature)
			assert.NoError(t, err)
			assert.True(t, isJobDone)
		}

		expectedMap := map[string]struct{}{"A": {}, "B": {}, "C": {}, "D": {}, "E": {}, "F": {}, "G": {}, "H": {}, "I": {}}
		// leader also sends his signature
		expectedBroadcastMap := map[string]int{"A": 1, "B": 1, "C": 1, "D": 1, "E": 1, "F": 1, "G": 1, "H": 1, "I": 1}

		mutex.Lock()
		assert.Equal(t, expectedMap, signatureSentForPks)
		assert.Equal(t, expectedBroadcastMap, signaturesBroadcast)
		mutex.Unlock()
	})
}

func TestSubroundSignature_SendSignature(t *testing.T) {
	t.Parallel()

	t.Run("sendSignatureForManagedKey will return false because of error", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()

		container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
			CreateSignatureShareForPublicKeyCalled: func(_ context.Context, message []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
				return make([]byte, 0), expErr
			},
		})
		consensusState := initializers.InitConsensusStateWithKeysHandler(
			&testscommon.KeysHandlerStub{
				IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
					return true
				},
			},
		)

		ch := make(chan bool, 1)

		sr, _ := spos.NewSubround(
			bls.SrBlock,
			bls.SrSignature,
			bls.SrEndRound,
			roundTimeDuration,
			0.7,
			0.85,
			"(SIGNATURE)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetHeader(&block.Header{})

		signatureSentForPks := make(map[string]struct{})
		srSignature, _ := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{
				SignatureSentCalled: func(pkBytes []byte) {
					signatureSentForPks[string(pkBytes)] = struct{}{}
				},
			},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		r := srSignature.SendSignatureForManagedKey(context.Background(), 0, "a")

		assert.False(t, r)
	})

	t.Run("sendSignatureForManagedKey should be false", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
			CreateSignatureShareForPublicKeyCalled: func(_ context.Context, message []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
				return []byte("SIG"), nil
			},
		})

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)

		container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return fmt.Errorf("error")
			},
		})
		consensusState := initializers.InitConsensusStateWithKeysHandler(
			&testscommon.KeysHandlerStub{
				IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
					return true
				},
			},
		)

		ch := make(chan bool, 1)

		sr, _ := spos.NewSubround(
			bls.SrBlock,
			bls.SrSignature,
			bls.SrEndRound,
			roundTimeDuration,
			0.7,
			0.85,
			"(SIGNATURE)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetHeader(&block.Header{})

		signatureSentForPks := make(map[string]struct{})
		srSignature, _ := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{
				SignatureSentCalled: func(pkBytes []byte) {
					signatureSentForPks[string(pkBytes)] = struct{}{}
				},
			},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		r := srSignature.SendSignatureForManagedKey(context.Background(), 1, "a")

		assert.False(t, r)
	})

	t.Run("SentSignature should be called", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
			CreateSignatureShareForPublicKeyCalled: func(_ context.Context, message []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
				return []byte("SIG"), nil
			},
		})

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)

		container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return nil
			},
		})
		consensusState := initializers.InitConsensusStateWithKeysHandler(
			&testscommon.KeysHandlerStub{
				IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
					return true
				},
			},
		)

		ch := make(chan bool, 1)

		sr, _ := spos.NewSubround(
			bls.SrBlock,
			bls.SrSignature,
			bls.SrEndRound,
			roundTimeDuration,
			0.7,
			0.85,
			"(SIGNATURE)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetHeader(&block.Header{})

		signatureSentForPks := make(map[string]struct{})
		varCalled := false
		srSignature, _ := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{
				SignatureSentCalled: func(pkBytes []byte) {
					signatureSentForPks[string(pkBytes)] = struct{}{}
					varCalled = true
				},
			},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		_ = srSignature.SendSignatureForManagedKey(context.Background(), 1, "a")

		assert.True(t, varCalled)
	})

	t.Run("if sig share already available, should not create it", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		container.SetSigningHandler(&consensusMocks.SigningHandlerStub{
			CreateSignatureShareForPublicKeyCalled: func(_ context.Context, message []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
				assert.Fail(t, "should not have been called")
				return []byte(""), nil
			},
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return []byte("SIG"), nil
			},
		})

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)

		container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				return nil
			},
		})
		consensusState := initializers.InitConsensusStateWithKeysHandler(
			&testscommon.KeysHandlerStub{
				IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
					return true
				},
			},
		)

		ch := make(chan bool, 1)

		sr, _ := spos.NewSubround(
			bls.SrBlock,
			bls.SrSignature,
			bls.SrEndRound,
			roundTimeDuration,
			0.7,
			0.85,
			"(SIGNATURE)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)
		sr.SetHeader(&block.Header{})

		signatureSentForPks := make(map[string]struct{})
		varCalled := false
		srSignature, _ := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{
				SignatureSentCalled: func(pkBytes []byte) {
					signatureSentForPks[string(pkBytes)] = struct{}{}
					varCalled = true
				},
			},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		_ = srSignature.SendSignatureForManagedKey(context.Background(), 1, "a")

		assert.True(t, varCalled)
	})
}

func TestSubroundSignature_DoSignatureJobForManagedKeys(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)

		signingHandler := &consensusMocks.SigningHandlerStub{
			CreateSignatureShareForPublicKeyCalled: func(_ context.Context, msg []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
				return []byte("SIG"), nil
			},
		}
		container.SetSigningHandler(signingHandler)
		consensusState := initializers.InitConsensusStateWithKeysHandler(
			&testscommon.KeysHandlerStub{
				IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
					return true
				},
			},
		)
		ch := make(chan bool, 1)

		sr, _ := spos.NewSubround(
			bls.SrBlock,
			bls.SrSignature,
			bls.SrEndRound,
			roundTimeDuration,
			0.7,
			0.85,
			"(SIGNATURE)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)

		signatureSentForPks := make(map[string]struct{})
		mutex := sync.Mutex{}
		srSignature, _ := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{
				SignatureSentCalled: func(pkBytes []byte) {
					mutex.Lock()
					signatureSentForPks[string(pkBytes)] = struct{}{}
					mutex.Unlock()
				},
			},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		sr.SetHeader(&block.Header{})
		signaturesBroadcast := make(map[string]int)
		container.SetBroadcastMessenger(&consensusMocks.BroadcastMessengerMock{
			BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
				mutex.Lock()
				signaturesBroadcast[string(message.PubKey)]++
				mutex.Unlock()
				return nil
			},
		})

		sr.SetSelfPubKey("OTHER")

		r := srSignature.DoSignatureJobForManagedKeys(context.TODO())
		assert.True(t, r)

		for _, pk := range sr.ConsensusGroup() {
			isJobDone, err := sr.JobDone(pk, bls.SrSignature)
			assert.NoError(t, err)
			assert.True(t, isJobDone)
		}

		expectedMap := map[string]struct{}{"A": {}, "B": {}, "C": {}, "D": {}, "E": {}, "F": {}, "G": {}, "H": {}, "I": {}}
		expectedBroadcastMap := map[string]int{"A": 1, "B": 1, "C": 1, "D": 1, "E": 1, "F": 1, "G": 1, "H": 1, "I": 1}

		mutex.Lock()
		assert.Equal(t, expectedMap, signatureSentForPks)
		assert.Equal(t, expectedBroadcastMap, signaturesBroadcast)
		mutex.Unlock()
	})

	t.Run("context done should return early", func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)

		consensusState := initializers.InitConsensusStateWithKeysHandler(
			&testscommon.KeysHandlerStub{
				IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
					return true
				},
			},
		)
		ch := make(chan bool, 1)

		sr, _ := spos.NewSubround(
			bls.SrBlock,
			bls.SrSignature,
			bls.SrEndRound,
			roundTimeDuration,
			0.7,
			0.85,
			"(SIGNATURE)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)

		srSignature, _ := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{},
			v2.NewSignatureEvidenceStore(nil),
		)

		sr.SetHeader(&block.Header{})
		sr.SetSelfPubKey("OTHER")

		ctx, cancel := context.WithCancel(context.TODO())
		cancel()

		r := srSignature.DoSignatureJobForManagedKeys(ctx)
		assert.False(t, r)

		for _, pk := range sr.ConsensusGroup() {
			isJobDone, err := sr.JobDone(pk, bls.SrSignature)
			assert.NoError(t, err)
			assert.False(t, isJobDone)
		}
	})

	t.Run("should fail", func(t *testing.T) {
		t.Parallel()
		container := consensusMocks.InitConsensusCore()
		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}
		container.SetEnableEpochsHandler(enableEpochsHandler)

		consensusState := initializers.InitConsensusStateWithKeysHandler(
			&testscommon.KeysHandlerStub{
				IsKeyManagedByCurrentNodeCalled: func(pkBytes []byte) bool {
					return true
				},
			},
		)
		ch := make(chan bool, 1)

		sr, _ := spos.NewSubround(
			bls.SrBlock,
			bls.SrSignature,
			bls.SrEndRound,
			roundTimeDuration,
			0.7,
			0.85,
			"(SIGNATURE)",
			consensusState,
			ch,
			executeStoredMessages,
			container,
			chainID,
			currentPid,
			&statusHandler.AppStatusHandlerStub{},
		)

		srSignature, _ := v2.NewSubroundSignature(
			sr,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
			&consensusMocks.SposWorkerMock{},
			&dataRetrieverMock.ThrottlerStub{
				CanProcessCalled: func() bool {
					return false
				},
			},
			v2.NewSignatureEvidenceStore(nil),
		)

		sr.SetHeader(&block.Header{})
		ctx, cancel := context.WithCancel(context.TODO())
		cancel()
		r := srSignature.DoSignatureJobForManagedKeys(ctx)
		assert.False(t, r)
	})
}

func TestSubroundSignature_DoSignatureConsensusCheck(t *testing.T) {
	t.Parallel()

	t.Run("round canceled should return false", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()
		sr.SetRoundCanceled(true)
		assert.False(t, sr.DoSignatureConsensusCheck())
	})
	t.Run("subround already finished should return true", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()
		sr.SetStatus(bls.SrSignature, spos.SsFinished)
		assert.True(t, sr.DoSignatureConsensusCheck())
	})
	t.Run("sig collection done should return true", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()

		for i := 0; i < sr.Threshold(bls.SrSignature); i++ {
			_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
		}

		sr.SetHeader(&block.HeaderV2{})
		assert.True(t, sr.DoSignatureConsensusCheck())
	})
	t.Run("sig collection failed should return false", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()
		sr.SetHeader(&block.HeaderV2{Header: createDefaultHeader()})
		assert.False(t, sr.DoSignatureConsensusCheck())
	})
	t.Run("not all sig collected in time should return false", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()
		sr.SetHeader(&block.HeaderV2{Header: createDefaultHeader()})
		assert.False(t, sr.DoSignatureConsensusCheck())
	})
	t.Run("nil header should return false", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()
		sr.SetHeader(nil)
		assert.False(t, sr.DoSignatureConsensusCheck())
	})
	t.Run("node not in consensus group should return true", func(t *testing.T) {
		t.Parallel()

		sr := initSubroundSignature()
		sr.SetHeader(&block.HeaderV2{Header: createDefaultHeader()})
		sr.SetSelfPubKey("X")
		assert.True(t, sr.DoSignatureConsensusCheck())
	})
}

func TestSubroundSignature_DoSignatureConsensusCheckAllSignaturesCollected(t *testing.T) {
	t.Parallel()
	t.Run("with flag active, should return true", testSubroundSignatureDoSignatureConsensusCheck(argTestSubroundSignatureDoSignatureConsensusCheck{
		flagActive:     true,
		jobsDone:       "all",
		expectedResult: true,
	}))
}

func TestSubroundSignature_DoSignatureConsensusCheckEnoughButNotAllSignaturesCollectedAndTimeIsOut(t *testing.T) {
	t.Parallel()

	t.Run("with flag active, should return true", testSubroundSignatureDoSignatureConsensusCheck(argTestSubroundSignatureDoSignatureConsensusCheck{
		flagActive:     true,
		jobsDone:       setThresholdJobsDone,
		expectedResult: true,
	}))
}

type argTestSubroundSignatureDoSignatureConsensusCheck struct {
	flagActive     bool
	jobsDone       string
	expectedResult bool
}

func testSubroundSignatureDoSignatureConsensusCheck(args argTestSubroundSignatureDoSignatureConsensusCheck) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		container := consensusMocks.InitConsensusCore()
		sr := initSubroundSignatureWithContainer(container)

		if !args.flagActive {
			leader, err := sr.GetLeader()
			assert.Nil(t, err)
			sr.SetSelfPubKey(leader)
		}

		numberOfJobsDone := sr.ConsensusGroupSize()
		if args.jobsDone == setThresholdJobsDone {
			numberOfJobsDone = sr.Threshold(bls.SrSignature)
		}
		for i := 0; i < numberOfJobsDone; i++ {
			_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
		}

		sr.SetHeader(&block.HeaderV2{})
		assert.Equal(t, args.expectedResult, sr.DoSignatureConsensusCheck())
	}
}
