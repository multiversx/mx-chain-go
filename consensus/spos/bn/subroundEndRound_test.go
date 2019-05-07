package bn_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/stretchr/testify/assert"
)

func initSubroundEndRoundWithContainer(container *mock.ConsensusCoreMock) bn.SubroundEndRound {
	ch := make(chan bool, 1)

	consensusState := initConsensusState()

	sr, _ := bn.NewSubround(
		int(bn.SrSignature),
		int(bn.SrEndRound),
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		container,
	)

	srEndRound, _ := bn.NewSubroundEndRound(
		sr,
		broadcastBlock,
		extend,
	)

	return srEndRound
}

func initSubroundEndRound() bn.SubroundEndRound {
	container := mock.InitConsensusCore()
	return initSubroundEndRoundWithContainer(container)
}

func TestSubroundEndRound_NewSubroundEndRoundNilSubroundShouldFail(t *testing.T) {
	t.Parallel()
	srEndRound, err := bn.NewSubroundEndRound(
		nil,
		broadcastBlock,
		extend,
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilBlockChainShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrSignature),
		int(bn.SrEndRound),
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetBlockchain(nil)
	srEndRound, err := bn.NewSubroundEndRound(
		sr,
		broadcastBlock,
		extend,
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrSignature),
		int(bn.SrEndRound),
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetBlockProcessor(nil)
	srEndRound, err := bn.NewSubroundEndRound(
		sr,
		broadcastBlock,
		extend,
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrSignature),
		int(bn.SrEndRound),
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		container,
	)

	sr.ConsensusState = nil
	srEndRound, err := bn.NewSubroundEndRound(
		sr,
		broadcastBlock,
		extend,
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilMultisignerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrSignature),
		int(bn.SrEndRound),
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetMultiSigner(nil)
	srEndRound, err := bn.NewSubroundEndRound(
		sr,
		broadcastBlock,
		extend,
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrSignature),
		int(bn.SrEndRound),
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetRounder(nil)
	srEndRound, err := bn.NewSubroundEndRound(
		sr,
		broadcastBlock,
		extend,
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrSignature),
		int(bn.SrEndRound),
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		container,
	)
	container.SetSyncTimer(nil)
	srEndRound, err := bn.NewSubroundEndRound(
		sr,
		broadcastBlock,
		extend,
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundEndRound_NewSubroundEndRoundNilBroadcastBlockFunctionShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrSignature),
		int(bn.SrEndRound),
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		container,
	)

	srEndRound, err := bn.NewSubroundEndRound(
		sr,
		nil,
		extend,
	)

	assert.Nil(t, srEndRound)
	assert.Equal(t, spos.ErrNilBroadcastBlockFunction, err)
}

func TestSubroundEndRound_NewSubroundEndRoundShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrSignature),
		int(bn.SrEndRound),
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		consensusState,
		ch,
		container,
	)

	srEndRound, err := bn.NewSubroundEndRound(
		sr,
		broadcastBlock,
		extend,
	)

	assert.NotNil(t, srEndRound)
	assert.Nil(t, err)
}

func TestSubroundEndRound_DoEndRoundJobErrAggregatingSigShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container)

	multiSignerMock := mock.InitMultiSignerMock()

	multiSignerMock.AggregateSigsMock = func(bitmap []byte) ([]byte, error) {
		return nil, crypto.ErrNilHasher
	}

	container.SetMultiSigner(multiSignerMock)
	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestSubroundEndRound_DoEndRoundJobErrCommitBlockShouldFail(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container)

	blProcMock := mock.InitBlockProcessorMock()

	blProcMock.CommitBlockCalled = func(
		blockChain data.ChainHandler,
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

func TestSubroundEndRound_DoEndRoundJobErrBroadcastBlockOK(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound()

	sr.SetBroadcastBlock(func(data.BodyHandler, data.HeaderHandler) error {
		return spos.ErrNilBroadcastBlockFunction
	})

	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestSubroundEndRound_DoEndRoundJobAllOK(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound()

	sr.Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound()

	sr.RoundCanceled = true

	ok := sr.DoEndRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnTrueWhenRoundIsFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound()

	sr.SetStatus(bn.SrEndRound, spos.SsFinished)

	ok := sr.DoEndRoundConsensusCheck()
	assert.True(t, ok)
}

func TestSubroundEndRound_DoEndRoundConsensusCheckShouldReturnFalseWhenRoundIsNotFinished(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound()

	ok := sr.DoEndRoundConsensusCheck()
	assert.False(t, ok)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldErrNilSignature(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound()

	err := sr.CheckSignaturesValidity([]byte(string(2)))
	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldErrIndexOutOfBounds(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container)

	_, _ = sr.MultiSigner().Create(nil, 0)

	sr.SetJobDone(sr.ConsensusGroup()[0], bn.SrSignature, true)

	multiSignerMock := mock.InitMultiSignerMock()
	multiSignerMock.SignatureShareMock = func(index uint16) ([]byte, error) {
		return nil, crypto.ErrIndexOutOfBounds
	}

	container.SetMultiSigner(multiSignerMock)

	err := sr.CheckSignaturesValidity([]byte(string(1)))
	assert.Equal(t, crypto.ErrIndexOutOfBounds, err)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldErrInvalidSignatureShare(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundEndRoundWithContainer(container)

	multiSignerMock := mock.InitMultiSignerMock()

	err := errors.New("invalid signature share")
	multiSignerMock.VerifySignatureShareMock = func(index uint16, sig []byte, message []byte, bitmap []byte) error {
		return err
	}

	container.SetMultiSigner(multiSignerMock)

	sr.SetJobDone(sr.ConsensusGroup()[0], bn.SrSignature, true)

	err2 := sr.CheckSignaturesValidity([]byte(string(1)))
	assert.Equal(t, err, err2)
}

func TestSubroundEndRound_CheckSignaturesValidityShouldRetunNil(t *testing.T) {
	t.Parallel()

	sr := *initSubroundEndRound()

	sr.SetJobDone(sr.ConsensusGroup()[0], bn.SrSignature, true)

	err := sr.CheckSignaturesValidity([]byte(string(1)))
	assert.Equal(t, nil, err)
}
