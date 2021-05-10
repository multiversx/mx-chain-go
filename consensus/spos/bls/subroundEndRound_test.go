package bls_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
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
	)

	return srEndRound
}

func initSubroundEndRound(appStatusHandler core.AppStatusHandler) bls.SubroundEndRound {
	container := mock.InitConsensusCore()
	return initSubroundEndRoundWithContainer(container, appStatusHandler)
}

func TestSubroundEndRound_NewSubroundEndRoundNilSubroundShouldFail(t *testing.T) {
	t.Parallel()
	srEndRound, err := bls.NewSubroundEndRound(
		nil,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&mock.AppStatusHandlerStub{},
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilSubround, err)
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
		&mock.AppStatusHandlerStub{},
	)
	container.SetBlockchain(nil)
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&mock.AppStatusHandlerStub{},
	)

	assert.Nil(t, srEndRound)
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
		&mock.AppStatusHandlerStub{},
	)
	container.SetBlockProcessor(nil)
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&mock.AppStatusHandlerStub{},
	)

	assert.Nil(t, srEndRound)
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
		&mock.AppStatusHandlerStub{},
	)

	sr.ConsensusState = nil
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&mock.AppStatusHandlerStub{},
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilMultisignerShouldFail(t *testing.T) {
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
		&mock.AppStatusHandlerStub{},
	)
	container.SetMultiSigner(nil)
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&mock.AppStatusHandlerStub{},
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
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
		&mock.AppStatusHandlerStub{},
	)
	container.SetRoundHandler(nil)
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&mock.AppStatusHandlerStub{},
	)

	assert.Nil(t, srEndRound)
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
		&mock.AppStatusHandlerStub{},
	)
	container.SetSyncTimer(nil)
	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&mock.AppStatusHandlerStub{},
	)

	assert.Nil(t, srEndRound)
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
		&mock.AppStatusHandlerStub{},
	)

	srEndRound, err := bls.NewSubroundEndRound(
		sr,
		extend,
		bls.ProcessingThresholdPercent,
		displayStatistics,
		&mock.AppStatusHandlerStub{},
	)

	assert.NotNil(t, srEndRound)
	assert.Nil(t, err)
}

func TestSubroundEndRound_DoEndRoundJobErrAggregatingSigShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
	multiSignerMock := mock.InitMultiSignerMock()
	multiSignerMock.AggregateSigsMock = func(bitmap []byte) ([]byte, error) {
		return nil, crypto.ErrNilHasher
	}

	container.SetMultiSigner(multiSignerMock)
	sr.Header = &block.Header{}

	sr.SetSelfPubKey("A")

	assert.True(t, sr.IsSelfLeaderInCurrentRound())
	r := sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestSubroundEndRound_DoEndRoundJobErrCommitBlockShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	blProcMock := mock.InitBlockProcessorMock()
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
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
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
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestSubroundEndRound_DoEndRoundJobErrMarshalizedDataToBroadcastOK(t *testing.T) {
	t.Parallel()

	err := errors.New("")
	container := mock.InitConsensusCore()

	bpm := mock.InitBlockProcessorMock()
	bpm.MarshalizedDataToBroadcastCalled = func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
		err = errors.New("error marshalized data to broadcast")
		return make(map[uint32][]byte), make(map[string][][]byte), err
	}
	container.SetBlockProcessor(bpm)

	bm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			return nil
		},
		BroadcastMiniBlocksCalled: func(bytes map[uint32][]byte) error {
			return nil
		},
		BroadcastTransactionsCalled: func(bytes map[string][][]byte) error {
			return nil
		},
	}
	container.SetBroadcastMessenger(bm)
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
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

	bpm := mock.InitBlockProcessorMock()
	bpm.MarshalizedDataToBroadcastCalled = func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
		return make(map[uint32][]byte), make(map[string][][]byte), nil
	}
	container.SetBlockProcessor(bpm)

	bm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			return nil
		},
		BroadcastMiniBlocksCalled: func(bytes map[uint32][]byte) error {
			err = errors.New("error broadcast miniblocks")
			return err
		},
		BroadcastTransactionsCalled: func(bytes map[string][][]byte) error {
			return nil
		},
	}
	container.SetBroadcastMessenger(bm)
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
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

	bpm := mock.InitBlockProcessorMock()
	bpm.MarshalizedDataToBroadcastCalled = func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
		return make(map[uint32][]byte), make(map[string][][]byte), nil
	}
	container.SetBlockProcessor(bpm)

	bm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			return nil
		},
		BroadcastMiniBlocksCalled: func(bytes map[uint32][]byte) error {
			return nil
		},
		BroadcastTransactionsCalled: func(bytes map[string][][]byte) error {
			err = errors.New("error broadcast transactions")
			return err
		},
	}
	container.SetBroadcastMessenger(bm)
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
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
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestSubroundEndRound_CheckIfSignatureIsFilled(t *testing.T) {
	t.Parallel()

	expectedSignature := []byte("signature")
	container := mock.InitConsensusCore()
	singleSigner := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			var receivedHdr block.Header
			_ = container.Marshalizer().Unmarshal(&receivedHdr, msg)
			return expectedSignature, nil
		},
	}
	container.SetSingleSigner(singleSigner)
	bm := &mock.BroadcastMessengerMock{
		BroadcastBlockCalled: func(handler data.BodyHandler, handler2 data.HeaderHandler) error {
			return errors.New("error")
		},
	}
	container.SetBroadcastMessenger(bm)
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
	sr.SetSelfPubKey("A")

	sr.Header = &block.Header{Nonce: 5}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
	assert.Equal(t, expectedSignature, sr.Header.GetLeaderSignature())
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
	sr.RoundCanceled = true

	ok := sr.DoEndRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnTrueWhenRoundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
	sr.SetStatus(bls.SrEndRound, spos.SsFinished)

	ok := sr.DoEndRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnFalseWhenRoundIsNotFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})

	ok := sr.DoEndRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldErrNilSignature(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})

	err := sr.CheckSignaturesValidity([]byte{2})
	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldErrIndexOutOfBounds(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
	_, _ = sr.MultiSigner().Create(nil, 0)
	_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)

	multiSignerMock := mock.InitMultiSignerMock()
	multiSignerMock.SignatureShareMock = func(index uint16) ([]byte, error) {
		return nil, crypto.ErrIndexOutOfBounds
	}
	container.SetMultiSigner(multiSignerMock)

	err := sr.CheckSignaturesValidity([]byte{1})
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldErrInvalidSignatureShare(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
	multiSignerMock := mock.InitMultiSignerMock()
	err := errors.New("invalid signature share")
	multiSignerMock.VerifySignatureShareMock = func(index uint16, sig []byte, msg []byte, bitmap []byte) error {
		return err
	}
	container.SetMultiSigner(multiSignerMock)

	_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)

	err2 := sr.CheckSignaturesValidity([]byte{1})
	assert.Equal(t, err, err2)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldReturnNil(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})

	_ = sr.SetJobDone(sr.ConsensusGroup()[0], bls.SrSignature, true)

	err := sr.CheckSignaturesValidity([]byte{1})
	assert.Equal(t, nil, err)
}

func TestSubroundEndRound_DoEndRoundJobByParticipant_RoundCanceledShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
	sr.RoundCanceled = true

	cnsData := consensus.Message{}
	res := sr.DoEndRoundJobByParticipant(&cnsData)
	assert.False(t, res)
}

func TestSubroundEndRound_DoEndRoundJobByParticipant_ConsensusDataNotSetShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
	sr.Data = nil

	cnsData := consensus.Message{}
	res := sr.DoEndRoundJobByParticipant(&cnsData)
	assert.False(t, res)
}

func TestSubroundEndRound_DoEndRoundJobByParticipant_PreviousSubroundNotFinishedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
	sr.SetStatus(2, spos.SsNotFinished)
	cnsData := consensus.Message{}
	res := sr.DoEndRoundJobByParticipant(&cnsData)
	assert.False(t, res)
}

func TestSubroundEndRound_DoEndRoundJobByParticipant_CurrentSubroundFinishedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})

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

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})

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
	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
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
	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
	sr.Header = hdr

	res, retHdr := sr.IsConsensusHeaderReceived()
	assert.False(t, res)
	assert.Nil(t, retHdr)
}

func TestSubroundEndRound_IsConsensusHeaderReceived_HeaderNotReceivedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 37}
	hdrToSearchFor := &block.Header{Nonce: 38}
	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
	sr.AddReceivedHeader(hdr)
	sr.Header = hdrToSearchFor

	res, retHdr := sr.IsConsensusHeaderReceived()
	assert.False(t, res)
	assert.Nil(t, retHdr)
}

func TestSubroundEndRound_IsConsensusHeaderReceivedShouldReturnTrue(t *testing.T) {
	t.Parallel()

	hdr := &block.Header{Nonce: 37}
	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
	sr.Header = hdr
	sr.AddReceivedHeader(hdr)

	res, retHdr := sr.IsConsensusHeaderReceived()
	assert.True(t, res)
	assert.Equal(t, hdr, retHdr)
}

func TestSubroundEndRound_HaveConsensusHeaderWithFullInfoNilHdrShouldNotWork(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})

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
	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
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
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
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
	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})
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
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
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

	sr := *initSubroundEndRound(&mock.AppStatusHandlerStub{})

	res := sr.IsOutOfTime()
	assert.False(t, res)
}

func TestSubroundEndRound_IsOutOfTimeShouldReturnTrue(t *testing.T) {
	t.Parallel()

	// update roundHandler's mock so it will calculate for real the duration
	container := mock.InitConsensusCore()
	roundHandler := mock.RoundHandlerMock{RemainingTimeCalled: func(startTime time.Time, maxTime time.Duration) time.Duration {
		currentTime := time.Now()
		elapsedTime := currentTime.Sub(startTime)
		remainingTime := maxTime - elapsedTime

		return remainingTime
	}}
	container.SetRoundHandler(&roundHandler)
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})

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
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
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
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
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
	sr := *initSubroundEndRoundWithContainer(container, &mock.AppStatusHandlerStub{})
	cnsDta := &consensus.Message{}
	sr.Header = &block.Header{}
	isValid := sr.IsBlockHeaderFinalInfoValid(cnsDta)
	assert.True(t, isValid)
}
