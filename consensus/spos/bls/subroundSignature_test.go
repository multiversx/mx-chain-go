package bls_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/multiversx/mx-chain-go/testscommon/subRounds"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initSubroundSignatureWithExtraSigners(extraSigners bls.SubRoundSignatureExtraSignersHolder) bls.SubroundSignature {
	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		initConsensusState(),
		make(chan bool, 1),
		executeStoredMessages,
		mock.InitConsensusCore(),
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)

	srSignature, _ := bls.NewSubroundSignature(
		sr,
		extend,
		&statusHandler.AppStatusHandlerStub{},
		extraSigners,
		&mock.SentSignatureTrackerStub{},
	)

	return srSignature
}

func initSubroundSignatureWithContainer(container *mock.ConsensusCoreMock, enableEpochHandler common.EnableEpochsHandler) bls.SubroundSignature {
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		enableEpochHandler,
	)

	srSignature, _ := bls.NewSubroundSignature(
		sr,
		extend,
		&statusHandler.AppStatusHandlerStub{},
		&subRounds.SubRoundSignatureExtraSignersHolderMock{},
		&mock.SentSignatureTrackerStub{},
	)

	return srSignature
}

func initSubroundSignature() bls.SubroundSignature {
	container := mock.InitConsensusCore()
	return initSubroundSignatureWithContainer(container, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
}

func TestNewSubroundSignature(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	)

	t.Run("nil subround should error", func(t *testing.T) {
		t.Parallel()

		srSignature, err := bls.NewSubroundSignature(
			nil,
			extend,
			&statusHandler.AppStatusHandlerStub{},
			&subRounds.SubRoundSignatureExtraSignersHolderMock{},
			&mock.SentSignatureTrackerStub{},
		)

		assert.Nil(t, srSignature)
		assert.Equal(t, spos.ErrNilSubround, err)
	})
	t.Run("nil extend function handler should error", func(t *testing.T) {
		t.Parallel()

		srSignature, err := bls.NewSubroundSignature(
			sr,
			nil,
			&statusHandler.AppStatusHandlerStub{},
			&subRounds.SubRoundSignatureExtraSignersHolderMock{},
			&mock.SentSignatureTrackerStub{},
		)

		assert.Nil(t, srSignature)
		assert.ErrorIs(t, err, spos.ErrNilFunctionHandler)
	})
	t.Run("nil app status handler should error", func(t *testing.T) {
		t.Parallel()

		srSignature, err := bls.NewSubroundSignature(
			sr,
			extend,
			nil,
			&subRounds.SubRoundSignatureExtraSignersHolderMock{},
			&mock.SentSignatureTrackerStub{},
		)

		assert.Nil(t, srSignature)
		assert.Equal(t, spos.ErrNilAppStatusHandler, err)
	})
	t.Run("nil sent signatures tracker should error", func(t *testing.T) {
		t.Parallel()

		srSignature, err := bls.NewSubroundSignature(
			sr,
			extend,
			&statusHandler.AppStatusHandlerStub{},
			&subRounds.SubRoundSignatureExtraSignersHolderMock{},
			nil,
		)

		assert.Nil(t, srSignature)
		assert.Equal(t, spos.ErrNilSentSignatureTracker, err)
	})
}

func TestSubroundSignature_NewSubroundSignatureNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)

	sr.ConsensusState = nil
	srSignature, err := bls.NewSubroundSignature(
		sr,
		extend,
		&statusHandler.AppStatusHandlerStub{},
		&subRounds.SubRoundSignatureExtraSignersHolderMock{},
		&mock.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundSignature_NewSubroundSignatureNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)
	container.SetHasher(nil)
	srSignature, err := bls.NewSubroundSignature(
		sr,
		extend,
		&statusHandler.AppStatusHandlerStub{},
		&subRounds.SubRoundSignatureExtraSignersHolderMock{},
		&mock.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubroundSignature_NewSubroundSignatureNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)
	container.SetMultiSignerContainer(nil)
	srSignature, err := bls.NewSubroundSignature(
		sr,
		extend,
		&statusHandler.AppStatusHandlerStub{},
		&subRounds.SubRoundSignatureExtraSignersHolderMock{},
		&mock.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestSubroundSignature_NewSubroundSignatureNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)
	container.SetRoundHandler(nil)

	srSignature, err := bls.NewSubroundSignature(
		sr,
		extend,
		&statusHandler.AppStatusHandlerStub{},
		&subRounds.SubRoundSignatureExtraSignersHolderMock{},
		&mock.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestSubroundSignature_NewSubroundSignatureNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)
	container.SetSyncTimer(nil)
	srSignature, err := bls.NewSubroundSignature(
		sr,
		extend,
		&statusHandler.AppStatusHandlerStub{},
		&subRounds.SubRoundSignatureExtraSignersHolderMock{},
		&mock.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srSignature))
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundSignature_NewSubroundSignatureNilExtraSignersHolderShouldFail(t *testing.T) {
	t.Parallel()

	sr, _ := defaultSubround(initConsensusState(), make(chan bool, 1), mock.InitConsensusCore())
	srSignature, err := bls.NewSubroundSignature(sr, extend, &statusHandler.AppStatusHandlerStub{}, nil, &mock.SentSignatureTrackerStub{})
	require.True(t, check.IfNil(srSignature))
	require.Equal(t, errorsMx.ErrNilSignatureRoundExtraSignersHolder, err)
}

func TestSubroundSignature_NewSubroundSignatureShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrBlock,
		bls.SrSignature,
		bls.SrEndRound,
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		&enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	)

	srSignature, err := bls.NewSubroundSignature(
		sr,
		extend,
		&statusHandler.AppStatusHandlerStub{},
		&subRounds.SubRoundSignatureExtraSignersHolderMock{},
		&mock.SentSignatureTrackerStub{},
	)

	assert.False(t, check.IfNil(srSignature))
	assert.Nil(t, err)
}

func TestSubroundSignature_DoSignatureJob(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundSignatureWithContainer(container, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})

	sr.Header = &block.Header{}
	sr.Data = nil
	r := sr.DoSignatureJob()
	assert.False(t, r)

	sr.Data = []byte("X")

	err := errors.New("create signature share error")
	signingHandler := &consensusMocks.SigningHandlerStub{
		CreateSignatureShareForPublicKeyCalled: func(msg []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
			return nil, err
		},
	}
	container.SetSigningHandler(signingHandler)

	r = sr.DoSignatureJob()
	assert.False(t, r)

	signingHandler = &consensusMocks.SigningHandlerStub{
		CreateSignatureShareForPublicKeyCalled: func(msg []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
			return []byte("SIG"), nil
		},
	}
	container.SetSigningHandler(signingHandler)

	r = sr.DoSignatureJob()
	assert.True(t, r)

	_ = sr.SetJobDone(sr.SelfPubKey(), bls.SrSignature, false)
	sr.RoundCanceled = false
	sr.SetSelfPubKey(sr.ConsensusGroup()[0])
	r = sr.DoSignatureJob()
	assert.True(t, r)
	assert.False(t, sr.RoundCanceled)
}

func TestSubroundSignature_DoSignatureJobWithExtraSigners(t *testing.T) {
	t.Parallel()

	wasExtraSigAdded := false
	extraSigs := map[string][]byte{
		"id1": []byte("sig1"),
	}
	extraSigners := &subRounds.SubRoundSignatureExtraSignersHolderMock{
		CreateExtraSignatureSharesCalled: func(header data.HeaderHandler, selfIndex uint16, selfPubKey []byte) (map[string][]byte, error) {
			return extraSigs, nil
		},
		AddExtraSigSharesToConsensusMessageCalled: func(extraSigShares map[string][]byte, cnsMsg *consensus.Message) error {
			require.Equal(t, extraSigs, extraSigShares)

			wasExtraSigAdded = true
			return nil
		},
	}
	sr := *initSubroundSignatureWithExtraSigners(extraSigners)

	sr.Header = &block.Header{}
	sr.Data = []byte("data")

	_ = sr.SetJobDone(sr.SelfPubKey(), bls.SrSignature, false)
	jobDone := sr.DoSignatureJob()
	require.True(t, jobDone)
	require.False(t, sr.RoundCanceled)
	require.True(t, wasExtraSigAdded)
}

func TestSubroundSignature_DoSignatureJobWithMultikey(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusStateWithKeysHandler(
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
		int64(70*roundTimeDuration/100),
		int64(85*roundTimeDuration/100),
		"(SIGNATURE)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
		enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	)

	signatureSentForPks := make(map[string]struct{})
	srSignature, _ := bls.NewSubroundSignature(
		sr,
		extend,
		&statusHandler.AppStatusHandlerStub{},
		&subRounds.SubRoundSignatureExtraSignersHolderMock{},
		&mock.SentSignatureTrackerStub{
			SignatureSentCalled: func(pkBytes []byte) {
				signatureSentForPks[string(pkBytes)] = struct{}{}
			},
		},
	)

	srSignature.Header = &block.Header{}
	srSignature.Data = nil
	r := srSignature.DoSignatureJob()
	assert.False(t, r)

	sr.Data = []byte("X")

	err := errors.New("create signature share error")
	signingHandler := &consensusMocks.SigningHandlerStub{
		CreateSignatureShareForPublicKeyCalled: func(msg []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
			return nil, err
		},
	}
	container.SetSigningHandler(signingHandler)

	r = srSignature.DoSignatureJob()
	assert.False(t, r)

	signingHandler = &consensusMocks.SigningHandlerStub{
		CreateSignatureShareForPublicKeyCalled: func(msg []byte, index uint16, epoch uint32, publicKeyBytes []byte) ([]byte, error) {
			return []byte("SIG"), nil
		},
	}
	container.SetSigningHandler(signingHandler)

	r = srSignature.DoSignatureJob()
	assert.True(t, r)

	_ = sr.SetJobDone(sr.SelfPubKey(), bls.SrSignature, false)
	sr.RoundCanceled = false
	sr.SetSelfPubKey(sr.ConsensusGroup()[0])
	r = srSignature.DoSignatureJob()
	assert.True(t, r)
	assert.False(t, sr.RoundCanceled)
	expectedMap := map[string]struct{}{
		"A": {},
		"B": {},
		"C": {},
		"D": {},
		"E": {},
		"F": {},
		"G": {},
		"H": {},
		"I": {},
	}
	assert.Equal(t, expectedMap, signatureSentForPks)
}

func TestSubroundSignature_ReceivedSignature(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	signature := []byte("signature")
	cnsMsg := consensus.NewConsensusMessage(
		sr.Data,
		signature,
		nil,
		nil,
		[]byte(sr.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bls.MtSignature),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
		nil,
	)

	sr.Header = &block.Header{}
	sr.Data = nil
	r := sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.Data = []byte("Y")
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.Data = []byte("X")
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])

	cnsMsg.PubKey = []byte("X")
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[1])
	maxCount := len(sr.ConsensusGroup()) * 2 / 3
	count := 0
	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		if sr.ConsensusGroup()[i] != string(cnsMsg.PubKey) {
			_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
			count++
			if count == maxCount {
				break
			}
		}
	}
	r = sr.ReceivedSignature(cnsMsg)
	assert.True(t, r)
}

func TestSubroundSignature_ReceivedSignatureWithExtraSigners(t *testing.T) {
	t.Parallel()

	var cnsMsg *consensus.Message

	wasSigStored := false
	expectedIdx := uint16(1)
	extraSigners := &subRounds.SubRoundSignatureExtraSignersHolderMock{
		StoreExtraSignatureShareCalled: func(index uint16, receivedMsg *consensus.Message) error {
			require.Equal(t, cnsMsg, receivedMsg)
			require.Equal(t, expectedIdx, index)

			wasSigStored = true
			return nil
		},
	}
	sr := *initSubroundSignatureWithExtraSigners(extraSigners)

	cnsMsg = consensus.NewConsensusMessage(
		sr.Data,
		[]byte("signature"),
		nil,
		nil,
		[]byte(sr.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bls.MtSignature),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
		nil,
	)

	sr.Header = &block.Header{}
	sr.Data = []byte("X")

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])
	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[expectedIdx])

	jobDone := sr.ReceivedSignature(cnsMsg)
	require.True(t, jobDone)
	require.True(t, wasSigStored)
}
func TestSubroundSignature_ReceivedSignatureStoreShareFailed(t *testing.T) {
	t.Parallel()

	errStore := errors.New("signature share store failed")
	storeSigShareCalled := false
	signingHandler := &consensusMocks.SigningHandlerStub{
		VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
			return nil
		},
		StoreSignatureShareCalled: func(index uint16, sig []byte) error {
			storeSigShareCalled = true
			return errStore
		},
	}

	container := mock.InitConsensusCore()
	container.SetSigningHandler(signingHandler)
	sr := *initSubroundSignatureWithContainer(container, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	sr.Header = &block.Header{}

	signature := []byte("signature")
	cnsMsg := consensus.NewConsensusMessage(
		sr.Data,
		signature,
		nil,
		nil,
		[]byte(sr.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bls.MtSignature),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
		nil,
	)

	sr.Data = nil
	r := sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.Data = []byte("Y")
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.Data = []byte("X")
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])

	cnsMsg.PubKey = []byte("X")
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[1])
	maxCount := len(sr.ConsensusGroup()) * 2 / 3
	count := 0
	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		if sr.ConsensusGroup()[i] != string(cnsMsg.PubKey) {
			_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
			count++
			if count == maxCount {
				break
			}
		}
	}
	r = sr.ReceivedSignature(cnsMsg)
	assert.False(t, r)
	assert.True(t, storeSigShareCalled)
}

func TestSubroundSignature_SignaturesCollected(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrBlock, false)
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, false)
	}

	ok, n := sr.AreSignaturesCollected(2)
	assert.False(t, ok)
	assert.Equal(t, 0, n)

	ok, _ = sr.AreSignaturesCollected(2)
	assert.False(t, ok)

	_ = sr.SetJobDone("B", bls.SrSignature, true)
	isJobDone, _ := sr.JobDone("B", bls.SrSignature)
	assert.True(t, isJobDone)

	ok, _ = sr.AreSignaturesCollected(2)
	assert.False(t, ok)

	_ = sr.SetJobDone("C", bls.SrSignature, true)
	ok, _ = sr.AreSignaturesCollected(2)
	assert.True(t, ok)
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	sr.RoundCanceled = true
	assert.False(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	sr.SetStatus(bls.SrSignature, spos.SsFinished)
	assert.True(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnTrueWhenSignaturesCollectedReturnTrue(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	for i := 0; i < sr.Threshold(bls.SrSignature); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
	}

	assert.True(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnFalseWhenSignaturesCollectedReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()
	assert.False(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnFalseWhenNotAllSignaturesCollectedAndTimeIsNotOut(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundSignatureWithContainer(container, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	sr.WaitingAllSignaturesTimeOut = false

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])

	for i := 0; i < sr.Threshold(bls.SrSignature); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
	}

	assert.False(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnTrueWhenAllSignaturesCollected(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundSignatureWithContainer(container, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	sr.WaitingAllSignaturesTimeOut = false

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])

	for i := 0; i < sr.ConsensusGroupSize(); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
	}

	assert.True(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnTrueWhenEnoughButNotAllSignaturesCollectedAndTimeIsOut(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundSignatureWithContainer(container, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	sr.WaitingAllSignaturesTimeOut = true

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])

	for i := 0; i < sr.Threshold(bls.SrSignature); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
	}

	assert.True(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnFalseWhenFallbackThresholdCouldNotBeApplied(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	container.SetFallbackHeaderValidator(&testscommon.FallBackHeaderValidatorStub{
		ShouldApplyFallbackValidationCalled: func(headerHandler data.HeaderHandler) bool {
			return false
		},
	})
	sr := *initSubroundSignatureWithContainer(container, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	sr.WaitingAllSignaturesTimeOut = false

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])

	for i := 0; i < sr.FallbackThreshold(bls.SrSignature); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
	}

	assert.False(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_DoSignatureConsensusCheckShouldReturnTrueWhenFallbackThresholdCouldBeApplied(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	container.SetFallbackHeaderValidator(&testscommon.FallBackHeaderValidatorStub{
		ShouldApplyFallbackValidationCalled: func(headerHandler data.HeaderHandler) bool {
			return true
		},
	})
	sr := *initSubroundSignatureWithContainer(container, &enableEpochsHandlerMock.EnableEpochsHandlerStub{})
	sr.WaitingAllSignaturesTimeOut = true

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])

	for i := 0; i < sr.FallbackThreshold(bls.SrSignature); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], bls.SrSignature, true)
	}

	assert.True(t, sr.DoSignatureConsensusCheck())
}

func TestSubroundSignature_ReceivedSignatureReturnFalseWhenConsensusDataIsNotEqual(t *testing.T) {
	t.Parallel()

	sr := *initSubroundSignature()

	cnsMsg := consensus.NewConsensusMessage(
		append(sr.Data, []byte("X")...),
		[]byte("signature"),
		nil,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bls.MtSignature),
		0,
		chainID,
		nil,
		nil,
		nil,
		currentPid,
		nil,
		nil,
	)

	assert.False(t, sr.ReceivedSignature(cnsMsg))
}

func TestSubroundEndRound_GetProcessedHeaderHashInSubroundSignatureShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("get processed header hash in subround signature with consensus model V1 should work", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		enableEpochHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
		sr := *initSubroundSignatureWithContainer(container, enableEpochHandler)

		sr.Data = []byte("X")
		hdrHash := sr.GetProcessedHeaderHash()
		assert.Nil(t, hdrHash)
	})

	t.Run("get processed header hash in subround signature with consensus model V2 should work", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		enableEpochHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ConsensusModelV2Flag)
		sr := *initSubroundSignatureWithContainer(container, enableEpochHandler)

		sr.Data = []byte("X")
		hdrHash := sr.GetProcessedHeaderHash()
		assert.Equal(t, sr.Data, hdrHash)
	})
}
