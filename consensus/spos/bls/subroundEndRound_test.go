package bls_test

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	mxErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/p2p/factory"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initSubroundEndRoundWithContainer(
	container *mock.ConsensusCoreMock,
	appStatusHandler core.AppStatusHandler,
) bls.SubroundEndRound {
	ch := make(chan bool, 1)
	consensusState := initConsensusState()
	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		appStatusHandler,
	)

	srEndRound, _ := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		appStatusHandler,
		&testscommon.SentSignatureTrackerStub{},
	)

	return srEndRound
}

func initSubroundEndRound(appStatusHandler core.AppStatusHandler) bls.SubroundEndRound {
	container := mock.InitConsensusCore()
	return initSubroundEndRoundWithContainer(container, appStatusHandler)
}

func TestNewSubroundEndRound(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
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

		srEndRound, err := bls.NewSubroundEndRound(
			nil,
			extend,
			bls.ProcessingThresholdPercent,
			displayStatistics,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
		)

		assert.Nil(t, srEndRound)
		assert.Equal(t, spos.ErrNilSubround, err)
	})
	t.Run("nil extend function handler should error", func(t *testing.T) {
		t.Parallel()

		srEndRound, err := bls.NewSubroundEndRound(
			sr,
			nil,
			bls.ProcessingThresholdPercent,
			displayStatistics,
			&statusHandler.AppStatusHandlerStub{},
			&testscommon.SentSignatureTrackerStub{},
		)

		assert.Nil(t, srEndRound)
		assert.ErrorIs(t, err, spos.ErrNilFunctionHandler)
	})
	t.Run("nil app status handler should error", func(t *testing.T) {
		t.Parallel()

		srEndRound, err := bls.NewSubroundEndRound(
			sr,
			extend,
			bls.ProcessingThresholdPercent,
			displayStatistics,
			nil,
			&testscommon.SentSignatureTrackerStub{},
		)

		assert.Nil(t, srEndRound)
		assert.Equal(t, spos.ErrNilAppStatusHandler, err)
	})
	t.Run("nil sent signatures tracker should error", func(t *testing.T) {
		t.Parallel()

		srEndRound, err := bls.NewSubroundEndRound(
			sr,
			extend,
			bls.ProcessingThresholdPercent,
			displayStatistics,
			&statusHandler.AppStatusHandlerStub{},
			nil,
		)

		assert.Nil(t, srEndRound)
		assert.Equal(t, mxErrors.ErrNilSentSignatureTracker, err)
	})
}

func TestSubroundEndRound_NewSubroundEndRoundNilBlockChainShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetBlockchain(nil)
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetBlockProcessor(nil)
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	sr.ConsensusState = nil
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetMultiSignerContainer(nil)
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetRoundHandler(nil)
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)
	container.SetSyncTimer(nil)
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
	)

	assert.True(t, check.IfNil(srEndRound))
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundEndRound_NewSubroundEndRoundShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
	)

	assert.False(t, check.IfNil(srEndRound))
	assert.Nil(t, err)
}

func TestSubroundEndRound_DoEndRoundJobErrAggregatingSigShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

	signingHandler := &consensusMocks.SigningHandlerStub{
		AggregateSigsCalled: func(bitmap []byte, epoch uint32) ([]byte, error) {
			return nil, crypto.ErrNilHasher
		},
	}
	container.SetSigningHandler(signingHandler)

	sr.Header = &block.Header{}

	sr.SetSelfPubKey("A")

	assert.True(t, sr.IsSelfLeaderInCurrentRound())
	r := sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestSubroundEndRound_DoEndRoundJobErrCommitBlockShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	blProcMock := mock.InitBlockProcessorMock(container.Marshalizer())
	blProcMock.CommitBlockCalled = func(
		header data.HeaderHandler,
		body data.BodyHandler,
	) error {
		return blockchain.ErrHeaderUnitNil
	}

	container.SetBlockProcessor(blProcMock)
	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestSubroundEndRound_DoEndRoundJobErrTimeIsOutShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	remainingTime := time.Millisecond
	roundHandlerMock := &mock.RoundHandlerMock{
		RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
			return remainingTime
		},
	}

	container.SetRoundHandler(roundHandlerMock)
	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)

	remainingTime = -time.Millisecond

	r = sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestSubroundEndRound_DoEndRoundJobErrBroadcastBlockOK(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	bm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			return errors.New("error")
		},
	}
	container.SetBroadcastMessenger(bm)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestSubroundEndRound_DoEndRoundJobErrMarshalizedDataToBroadcastOK(t *testing.T) {
	t.Parallel()

	err := errors.New("")
	container := mock.InitConsensusCore()

	bpm := mock.InitBlockProcessorMock(container.Marshalizer())
	bpm.MarshalizedDataToBroadcastCalled = func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
		err = errors.New("error marshalized data to broadcast")
		return make(map[uint32][]byte), make(map[string][][]byte), err
	}
	container.SetBlockProcessor(bpm)

	bm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			return nil
		},
		BroadcastMiniBlocksCalled: func(bytes map[uint32][]byte, pkBytes []byte) error {
			return nil
		},
		BroadcastTransactionsCalled: func(bytes map[string][][]byte, pkBytes []byte) error {
			return nil
		},
	}
	container.SetBroadcastMessenger(bm)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
	assert.Equal(t, errors.New("error marshalized data to broadcast"), err)
}

func TestSubroundEndRound_DoEndRoundJobErrBroadcastMiniBlocksOK(t *testing.T) {
	t.Parallel()

	err := errors.New("")
	container := mock.InitConsensusCore()

	bpm := mock.InitBlockProcessorMock(container.Marshalizer())
	bpm.MarshalizedDataToBroadcastCalled = func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
		return make(map[uint32][]byte), make(map[string][][]byte), nil
	}
	container.SetBlockProcessor(bpm)

	bm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			return nil
		},
		BroadcastMiniBlocksCalled: func(bytes map[uint32][]byte, pkBytes []byte) error {
			err = errors.New("error broadcast miniblocks")
			return err
		},
		BroadcastTransactionsCalled: func(bytes map[string][][]byte, pkBytes []byte) error {
			return nil
		},
	}
	container.SetBroadcastMessenger(bm)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
	// no error as broadcast is delayed
	assert.Equal(t, errors.New("error broadcast miniblocks"), err)
}

func TestSubroundEndRound_DoEndRoundJobErrBroadcastTransactionsOK(t *testing.T) {
	t.Parallel()

	err := errors.New("")
	container := mock.InitConsensusCore()

	bpm := mock.InitBlockProcessorMock(container.Marshalizer())
	bpm.MarshalizedDataToBroadcastCalled = func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
		return make(map[uint32][]byte), make(map[string][][]byte), nil
	}
	container.SetBlockProcessor(bpm)

	bm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			return nil
		},
		BroadcastMiniBlocksCalled: func(bytes map[uint32][]byte, pkBytes []byte) error {
			return nil
		},
		BroadcastTransactionsCalled: func(bytes map[string][][]byte, pkBytes []byte) error {
			err = errors.New("error broadcast transactions")
			return err
		},
	}
	container.SetBroadcastMessenger(bm)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
	// no error as broadcast is delayed
	assert.Equal(t, errors.New("error broadcast transactions"), err)
}

func TestSubroundEndRound_DoEndRoundJobAllOK(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	bm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			return errors.New("error")
		},
	}
	container.SetBroadcastMessenger(bm)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestSubroundEndRound_CheckIfSignatureIsFilled(t *testing.T) {
	t.Parallel()

	expectedSignature := []byte("signature")
	container := mock.InitConsensusCore()
	signingHandler := &consensusMocks.SigningHandlerStub{
		CreateSignatureForPublicKeyCalled: func(publicKeyBytes []byte, msg []byte) ([]byte, error) {
			var receivedHdr block.Header
			_ = container.Marshalizer().Unmarshal(&receivedHdr, msg)
			return expectedSignature, nil
		},
	}
	container.SetSigningHandler(signingHandler)
	bm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			return errors.New("error")
		},
	}
	container.SetBroadcastMessenger(bm)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	sr.Header = &block.Header{Nonce: 5}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
	assert.Equal(t, expectedSignature, sr.Header.GetLeaderSignature())
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.RoundCanceled = true

	ok := sr.DoEndRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnTrueWhenRoundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.SetStatus(bls.SrEndRound, spos.SsFinished)

	ok := sr.DoEndRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnFalseWhenRoundIsNotFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	ok := sr.DoEndRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldErrNilSignature(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	err := sr.CheckSignaturesValidity([]byte{2})
	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldReturnNil(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)

	err := sr.CheckSignaturesValidity([]byte{1})
	assert.Equal(t, nil, err)
}

func TestSubroundEndRound_DoEndRoundJobByParticipant_RoundCanceledShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.RoundCanceled = true

	cnsData := consensus.Message{}
	res := sr.DoEndRoundJobByParticipant(&cnsData)
	assert.False(t, res)
}

func TestSubroundEndRound_DoEndRoundJobByParticipant_ConsensusDataNotSetShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.Data = nil

	cnsData := consensus.Message{}
	res := sr.DoEndRoundJobByParticipant(&cnsData)
	assert.False(t, res)
}

func TestSubroundEndRound_DoEndRoundJobByParticipant_PreviousSubroundNotFinishedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.SetStatus(2, spos.SsNotFinished)
	cnsData := consensus.Message{}
	res := sr.DoEndRoundJobByParticipant(&cnsData)
	assert.False(t, res)
}

func TestSubroundEndRound_DoEndRoundJobByParticipant_CurrentSubroundFinishedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	// set previous as finished
	sr.SetStatus(2, spos.SsFinished)

	// set current as finished
	sr.SetStatus(3, spos.SsFinished)

	cnsData := consensus.Message{}
	res := sr.DoEndRoundJobByParticipant(&cnsData)
	assert.False(t, res)
}

func TestSubroundEndRound_DoEndRoundJobByParticipant_ConsensusHeaderNotReceivedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	// set previous as finished
	sr.SetStatus(2, spos.SsFinished)

	// set current as not finished
	sr.SetStatus(3, spos.SsNotFinished)

	cnsData := consensus.Message{}
	res := sr.DoEndRoundJobByParticipant(&cnsData)
	assert.False(t, res)
}

func TestSubroundEndRound_DoEndRoundJobByParticipant_ShouldReturnTrue(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 37}
	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.Header = hdr
	sr.AddReceivedHeader(hdr)

	// set previous as finished
	sr.SetStatus(2, spos.SsFinished)

	// set current as not finished
	sr.SetStatus(3, spos.SsNotFinished)

	cnsData := consensus.Message{}
	res := sr.DoEndRoundJobByParticipant(&cnsData)
	assert.True(t, res)
}

func TestSubroundEndRound_IsConsensusHeaderReceived_NoReceivedHeadersShouldReturnFalse(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 37}
	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.Header = hdr

	res, retHdr := sr.IsConsensusHeaderReceived()
	assert.False(t, res)
	assert.Nil(t, retHdr)
}

func TestSubroundEndRound_IsConsensusHeaderReceived_HeaderNotReceivedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 37}
	hdrToSearchFor := &block.Header{Nonce: 38}
	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.AddReceivedHeader(hdr)
	sr.Header = hdrToSearchFor

	res, retHdr := sr.IsConsensusHeaderReceived()
	assert.False(t, res)
	assert.Nil(t, retHdr)
}

func TestSubroundEndRound_IsConsensusHeaderReceivedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 37}
	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.Header = hdr
	sr.AddReceivedHeader(hdr)

	res, retHdr := sr.IsConsensusHeaderReceived()
	assert.True(t, res)
	assert.Equal(t, hdr, retHdr)
}

func TestSubroundEndRound_HaveConsensusHeaderWithFullInfoNilHdrShouldNotWork(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	cnsData := consensus.Message{}

	haveHdr, hdr := sr.HaveConsensusHeaderWithFullInfo(&cnsData)
	assert.False(t, haveHdr)
	assert.Nil(t, hdr)
}

func TestSubroundEndRound_HaveConsensusHeaderWithFullInfoShouldWork(t *testing.T) {
	t.Parallel()

	originalPubKeyBitMap := []byte{0, 1, 2}
	newPubKeyBitMap := []byte{3, 4, 5}
	originalLeaderSig := []byte{6, 7, 8}
	newLeaderSig := []byte{9, 10, 11}
	originalSig := []byte{12, 13, 14}
	newSig := []byte{15, 16, 17}
	hdr := block.Header{
		PubKeysBitmap:   originalPubKeyBitMap,
		Signature:       originalSig,
		LeaderSignature: originalLeaderSig,
	}
	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.Header = &hdr

	cnsData := consensus.Message{
		PubKeysBitmap:      newPubKeyBitMap,
		LeaderSignature:    newLeaderSig,
		AggregateSignature: newSig,
	}
	haveHdr, newHdr := sr.HaveConsensusHeaderWithFullInfo(&cnsData)
	assert.True(t, haveHdr)
	require.NotNil(t, newHdr)
	assert.Equal(t, newPubKeyBitMap, newHdr.GetPubKeysBitmap())
	assert.Equal(t, newLeaderSig, newHdr.GetLeaderSignature())
	assert.Equal(t, newSig, newHdr.GetSignature())
}

func TestSubroundEndRound_CreateAndBroadcastHeaderFinalInfoBroadcastShouldBeCalled(t *testing.T) {
	t.Parallel()

	chanRcv := make(chan bool, 1)
	leaderSigInHdr := []byte("leader sig")
	container := mock.InitConsensusCore()
	messenger := &mock.BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
			chanRcv <- true
			assert.Equal(t, message.LeaderSignature, leaderSigInHdr)
			return nil
		},
	}
	container.SetBroadcastMessenger(messenger)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	sr.Header = &block.Header{LeaderSignature: leaderSigInHdr}

	sr.CreateAndBroadcastHeaderFinalInfo()

	select {
	case <-chanRcv:
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "broadcast not called")
	}
}

func TestSubroundEndRound_ReceivedBlockHeaderFinalInfoShouldWork(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 37}
	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})
	sr.Header = hdr
	sr.AddReceivedHeader(hdr)

	sr.SetStatus(2, spos.SsFinished)
	sr.SetStatus(3, spos.SsNotFinished)

	cnsData := consensus.Message{
		// apply the data which is mocked in consensus state so the checks will pass
		BlockHeaderHash: []byte("X"),
		PubKey:          []byte("A"),
	}

	res := sr.ReceivedBlockHeaderFinalInfo(&cnsData)
	assert.True(t, res)
}

func TestSubroundEndRound_ReceivedBlockHeaderFinalInfoShouldReturnFalseWhenFinalInfoIsNotValid(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	headerSigVerifier := &mock.HeaderSigVerifierStub{
		VerifyLeaderSignatureCalled: func(header data.HeaderHandler) error {
			return errors.New("error")
		},
		VerifySignatureCalled: func(header data.HeaderHandler) error {
			return errors.New("error")
		},
	}

	container.SetHeaderSigVerifier(headerSigVerifier)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	cnsData := consensus.Message{
		BlockHeaderHash: []byte("X"),
		PubKey:          []byte("A"),
	}
	sr.Header = &block.Header{}
	res := sr.ReceivedBlockHeaderFinalInfo(&cnsData)
	assert.False(t, res)
}

func TestSubroundEndRound_IsOutOfTimeShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&statusHandler.AppStatusHandlerStub{})

	res := sr.IsOutOfTime()
	assert.False(t, res)
}

func TestSubroundEndRound_IsOutOfTimeShouldReturnTrue(t *testing.T) {
	t.Parallel()

	// update roundHandler's mock, so it will calculate for real the duration
	container := mock.InitConsensusCore()
	roundHandler := mock.RoundHandlerMock{RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
		currentTime := time.Now()
		elapsedTime := currentTime.Sub(startTime)
		remainingTime := maxTime - elapsedTime

		return remainingTime
	}}
	container.SetRoundHandler(&roundHandler)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

	sr.RoundTimeStamp = time.Now().AddDate(0, 0, -1)

	res := sr.IsOutOfTime()
	assert.True(t, res)
}

func TestSubroundEndRound_IsBlockHeaderFinalInfoValidShouldReturnFalseWhenVerifyLeaderSignatureFails(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	headerSigVerifier := &mock.HeaderSigVerifierStub{
		VerifyLeaderSignatureCalled: func(header data.HeaderHandler) error {
			return errors.New("error")
		},
		VerifySignatureCalled: func(header data.HeaderHandler) error {
			return nil
		},
	}

	container.SetHeaderSigVerifier(headerSigVerifier)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	cnsDta := &consensus.Message{}
	sr.Header = &block.Header{}
	isValid := sr.IsBlockHeaderFinalInfoValid(cnsDta)
	assert.False(t, isValid)
}

func TestSubroundEndRound_IsBlockHeaderFinalInfoValidShouldReturnFalseWhenVerifySignatureFails(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	headerSigVerifier := &mock.HeaderSigVerifierStub{
		VerifyLeaderSignatureCalled: func(header data.HeaderHandler) error {
			return nil
		},
		VerifySignatureCalled: func(header data.HeaderHandler) error {
			return errors.New("error")
		},
	}

	container.SetHeaderSigVerifier(headerSigVerifier)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	cnsDta := &consensus.Message{}
	sr.Header = &block.Header{}
	isValid := sr.IsBlockHeaderFinalInfoValid(cnsDta)
	assert.False(t, isValid)
}

func TestSubroundEndRound_IsBlockHeaderFinalInfoValidShouldReturnTrue(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	headerSigVerifier := &mock.HeaderSigVerifierStub{
		VerifyLeaderSignatureCalled: func(header data.HeaderHandler) error {
			return nil
		},
		VerifySignatureCalled: func(header data.HeaderHandler) error {
			return nil
		},
	}

	container.SetHeaderSigVerifier(headerSigVerifier)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
	cnsDta := &consensus.Message{}
	sr.Header = &block.Header{}
	isValid := sr.IsBlockHeaderFinalInfoValid(cnsDta)
	assert.True(t, isValid)
}

func TestVerifyNodesOnAggSigVerificationFail(t *testing.T) {
	t.Parallel()

	t.Run("fail to get signature share", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		expectedErr := errors.New("exptected error")
		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, expectedErr
			},
		}

		container.SetSigningHandler(signingHandler)

		sr.Header = &block.Header{}
		_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)

		_, err := sr.VerifyNodesOnAggSigFail()
		require.Equal(t, expectedErr, err)
	})

	t.Run("fail to verify signature share, job done will be set to false", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		expectedErr := errors.New("exptected error")
		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, nil
			},
			VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
				return expectedErr
			},
		}

		sr.Header = &block.Header{}
		_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)
		container.SetSigningHandler(signingHandler)

		_, err := sr.VerifyNodesOnAggSigFail()
		require.Nil(t, err)

		isJobDone, err := sr.JobDone(sr.ConsensusGroup()[0], bls.SrSignature)
		require.Nil(t, err)
		require.False(t, isJobDone)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, nil
			},
			VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
				return nil
			},
			VerifyCalled: func(msg, bitmap []byte, epoch uint32) error {
				return nil
			},
		}
		container.SetSigningHandler(signingHandler)

		sr.Header = &block.Header{}
		_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[1], bls.SrSignature, true)

		invalidSigners, err := sr.VerifyNodesOnAggSigFail()
		require.Nil(t, err)
		require.NotNil(t, invalidSigners)
	})
}

func TestComputeAddSigOnValidNodes(t *testing.T) {
	t.Parallel()

	t.Run("invalid number of valid sig shares", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.Header = &block.Header{}
		sr.SetThreshold(bls.SrEndRound, 2)

		_, _, err := sr.ComputeAggSigOnValidNodes()
		require.True(t, errors.Is(err, spos.ErrInvalidNumSigShares))
	})

	t.Run("fail to created aggregated sig", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		expectedErr := errors.New("exptected error")
		signingHandler := &consensusMocks.SigningHandlerStub{
			AggregateSigsCalled: func(bitmap []byte, epoch uint32) ([]byte, error) {
				return nil, expectedErr
			},
		}
		container.SetSigningHandler(signingHandler)

		sr.Header = &block.Header{}
		_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)

		_, _, err := sr.ComputeAggSigOnValidNodes()
		require.Equal(t, expectedErr, err)
	})

	t.Run("fail to set aggregated sig", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		expectedErr := errors.New("exptected error")
		signingHandler := &consensusMocks.SigningHandlerStub{
			SetAggregatedSigCalled: func(_ []byte) error {
				return expectedErr
			},
		}
		container.SetSigningHandler(signingHandler)
		sr.Header = &block.Header{}
		_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)

		_, _, err := sr.ComputeAggSigOnValidNodes()
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.Header = &block.Header{}
		_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)

		bitmap, sig, err := sr.ComputeAggSigOnValidNodes()
		require.NotNil(t, bitmap)
		require.NotNil(t, sig)
		require.Nil(t, err)
	})
}

func TestSubroundEndRound_DoEndRoundJobByLeaderVerificationFail(t *testing.T) {
	t.Parallel()

	t.Run("not enough valid signature shares", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		verifySigShareNumCalls := 0
		verifyFirstCall := true
		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, nil
			},
			VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
				if verifySigShareNumCalls == 0 {
					verifySigShareNumCalls++
					return errors.New("expected error")
				}

				verifySigShareNumCalls++
				return nil
			},
			VerifyCalled: func(msg, bitmap []byte, epoch uint32) error {
				if verifyFirstCall {
					verifyFirstCall = false
					return errors.New("expected error")
				}

				return nil
			},
		}

		container.SetSigningHandler(signingHandler)

		sr.SetThreshold(bls.SrEndRound, 2)

		_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[1], bls.SrSignature, true)

		sr.Header = &block.Header{}

		r := sr.DoEndRoundJobByLeader()
		require.False(t, r)

		assert.False(t, verifyFirstCall)
		assert.Equal(t, 2, verifySigShareNumCalls)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()
		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		verifySigShareNumCalls := 0
		verifyFirstCall := true
		signingHandler := &consensusMocks.SigningHandlerStub{
			SignatureShareCalled: func(index uint16) ([]byte, error) {
				return nil, nil
			},
			VerifySignatureShareCalled: func(index uint16, sig, msg []byte, epoch uint32) error {
				if verifySigShareNumCalls == 0 {
					verifySigShareNumCalls++
					return errors.New("expected error")
				}

				verifySigShareNumCalls++
				return nil
			},
			VerifyCalled: func(msg, bitmap []byte, epoch uint32) error {
				if verifyFirstCall {
					verifyFirstCall = false
					return errors.New("expected error")
				}

				return nil
			},
		}

		container.SetSigningHandler(signingHandler)

		sr.SetThreshold(bls.SrEndRound, 2)

		_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[1], bls.SrSignature, true)
		_ = sr.SetJobDone(sr.ConsensusGroup()[2], bls.SrSignature, true)

		sr.Header = &block.Header{}

		r := sr.DoEndRoundJobByLeader()
		require.True(t, r)

		assert.False(t, verifyFirstCall)
		assert.Equal(t, 3, verifySigShareNumCalls)
	})
}

func TestSubroundEndRound_ReceivedInvalidSignersInfo(t *testing.T) {
	t.Parallel()

	t.Run("consensus data is not set", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.ConsensusState.Data = nil

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})

	t.Run("received message node is not leader in current round", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("other node"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})

	t.Run("received message for self leader", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.SetSelfPubKey("A")

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})

	t.Run("received hash does not match the hash from current consensus state", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("Y"),
			PubKey:          []byte("A"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})

	t.Run("process received message verification failed, different round index", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
			RoundIndex:      1,
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})

	t.Run("empty invalid signers", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
			InvalidSigners:  []byte{},
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})

	t.Run("invalid signers data", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		messageSigningHandler := &mock.MessageSigningHandlerStub{
			DeserializeCalled: func(messagesBytes []byte) ([]p2p.MessageP2P, error) {
				return nil, expectedErr
			},
		}

		container := mock.InitConsensusCore()
		container.SetMessageSigningHandler(messageSigningHandler)

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
			InvalidSigners:  []byte("invalid data"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.False(t, res)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		cnsData := consensus.Message{
			BlockHeaderHash: []byte("X"),
			PubKey:          []byte("A"),
			InvalidSigners:  []byte("invalidSignersData"),
		}

		res := sr.ReceivedInvalidSignersInfo(&cnsData)
		assert.True(t, res)
	})
}

func TestVerifyInvalidSigners(t *testing.T) {
	t.Parallel()

	t.Run("failed to deserialize invalidSigners field, should error", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		expectedErr := errors.New("expected err")
		messageSigningHandler := &mock.MessageSigningHandlerStub{
			DeserializeCalled: func(messagesBytes []byte) ([]p2p.MessageP2P, error) {
				return nil, expectedErr
			},
		}

		container.SetMessageSigningHandler(messageSigningHandler)

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		err := sr.VerifyInvalidSigners([]byte{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("failed to verify low level p2p message, should error", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		invalidSigners := []p2p.MessageP2P{&factory.Message{
			FromField: []byte("from"),
		}}
		invalidSignersBytes, _ := container.Marshalizer().Marshal(invalidSigners)

		expectedErr := errors.New("expected err")
		messageSigningHandler := &mock.MessageSigningHandlerStub{
			DeserializeCalled: func(messagesBytes []byte) ([]p2p.MessageP2P, error) {
				require.Equal(t, invalidSignersBytes, messagesBytes)
				return invalidSigners, nil
			},
			VerifyCalled: func(message p2p.MessageP2P) error {
				return expectedErr
			},
		}

		container.SetMessageSigningHandler(messageSigningHandler)

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		err := sr.VerifyInvalidSigners(invalidSignersBytes)
		require.Equal(t, expectedErr, err)
	})

	t.Run("failed to verify signature share", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		pubKey := []byte("A") // it's in consensus

		consensusMsg := &consensus.Message{
			PubKey: pubKey,
		}
		consensusMsgBytes, _ := container.Marshalizer().Marshal(consensusMsg)

		invalidSigners := []p2p.MessageP2P{&factory.Message{
			FromField: []byte("from"),
			DataField: consensusMsgBytes,
		}}
		invalidSignersBytes, _ := container.Marshalizer().Marshal(invalidSigners)

		messageSigningHandler := &mock.MessageSigningHandlerStub{
			DeserializeCalled: func(messagesBytes []byte) ([]p2p.MessageP2P, error) {
				require.Equal(t, invalidSignersBytes, messagesBytes)
				return invalidSigners, nil
			},
		}

		wasCalled := false
		signingHandler := &consensusMocks.SigningHandlerStub{
			VerifySingleSignatureCalled: func(publicKeyBytes []byte, message []byte, signature []byte) error {
				wasCalled = true
				return errors.New("expected err")
			},
		}

		container.SetSigningHandler(signingHandler)
		container.SetMessageSigningHandler(messageSigningHandler)

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		err := sr.VerifyInvalidSigners(invalidSignersBytes)
		require.Nil(t, err)
		require.True(t, wasCalled)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		pubKey := []byte("A") // it's in consensus

		consensusMsg := &consensus.Message{
			PubKey: pubKey,
		}
		consensusMsgBytes, _ := container.Marshalizer().Marshal(consensusMsg)

		invalidSigners := []p2p.MessageP2P{&factory.Message{
			FromField: []byte("from"),
			DataField: consensusMsgBytes,
		}}
		invalidSignersBytes, _ := container.Marshalizer().Marshal(invalidSigners)

		messageSigningHandler := &mock.MessageSignerMock{}
		container.SetMessageSigningHandler(messageSigningHandler)

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

		err := sr.VerifyInvalidSigners(invalidSignersBytes)
		require.Nil(t, err)
	})
}

func TestSubroundEndRound_CreateAndBroadcastInvalidSigners(t *testing.T) {
	t.Parallel()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	expectedInvalidSigners := []byte("invalid signers")

	wasCalled := false
	container := mock.InitConsensusCore()
	messenger := &mock.BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
			wg.Done()
			assert.Equal(t, expectedInvalidSigners, message.InvalidSigners)
			wasCalled = true
			return nil
		},
	}
	container.SetBroadcastMessenger(messenger)
	sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})

	sr.CreateAndBroadcastInvalidSigners(expectedInvalidSigners)

	wg.Wait()

	require.True(t, wasCalled)
}

func TestGetFullMessagesForInvalidSigners(t *testing.T) {
	t.Parallel()

	t.Run("empty p2p messages slice if not in state", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		messageSigningHandler := &mock.MessageSigningHandlerStub{
			SerializeCalled: func(messages []p2p.MessageP2P) ([]byte, error) {
				require.Equal(t, 0, len(messages))

				return []byte{}, nil
			},
		}

		container.SetMessageSigningHandler(messageSigningHandler)

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		invalidSigners := []string{"B", "C"}

		invalidSignersBytes, err := sr.GetFullMessagesForInvalidSigners(invalidSigners)
		require.Nil(t, err)
		require.Equal(t, []byte{}, invalidSignersBytes)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		container := mock.InitConsensusCore()

		expectedInvalidSigners := []byte("expectedInvalidSigners")

		messageSigningHandler := &mock.MessageSigningHandlerStub{
			SerializeCalled: func(messages []p2p.MessageP2P) ([]byte, error) {
				require.Equal(t, 2, len(messages))

				return expectedInvalidSigners, nil
			},
		}

		container.SetMessageSigningHandler(messageSigningHandler)

		sr := *initSubroundEndRoundWithContainer(container, &statusHandler.AppStatusHandlerStub{})
		sr.AddMessageWithSignature("B", &p2pmocks.P2PMessageMock{})
		sr.AddMessageWithSignature("C", &p2pmocks.P2PMessageMock{})

		invalidSigners := []string{"B", "C"}

		invalidSignersBytes, err := sr.GetFullMessagesForInvalidSigners(invalidSigners)
		require.Nil(t, err)
		require.Equal(t, expectedInvalidSigners, invalidSignersBytes)
	})
}

func TestSubroundEndRound_getMinConsensusGroupIndexOfManagedKeys(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	keysHandler := &testscommon.KeysHandlerStub{}
	ch := make(chan bool, 1)
	consensusState := initConsensusStateWithKeysHandler(keysHandler)
	sr, _ := spos.NewSubround(
		bls.SrSignature,
		bls.SrEndRound,
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
		currentPid,
		&statusHandler.AppStatusHandlerStub{},
	)

	srEndRound, _ := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&statusHandler.AppStatusHandlerStub{},
		&testscommon.SentSignatureTrackerStub{},
	)

	t.Run("no managed keys from consensus group", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return false
		}

		assert.Equal(t, 9, srEndRound.GetMinConsensusGroupIndexOfManagedKeys())
	})
	t.Run("first managed key in consensus group should return 0", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte("A"), pkBytes)
		}

		assert.Equal(t, 0, srEndRound.GetMinConsensusGroupIndexOfManagedKeys())
	})
	t.Run("third managed key in consensus group should return 2", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte("C"), pkBytes)
		}

		assert.Equal(t, 2, srEndRound.GetMinConsensusGroupIndexOfManagedKeys())
	})
	t.Run("last managed key in consensus group should return 8", func(t *testing.T) {
		keysHandler.IsKeyManagedByCurrentNodeCalled = func(pkBytes []byte) bool {
			return bytes.Equal([]byte("I"), pkBytes)
		}

		assert.Equal(t, 8, srEndRound.GetMinConsensusGroupIndexOfManagedKeys())
	})
}
