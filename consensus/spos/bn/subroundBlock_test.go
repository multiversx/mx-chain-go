package bn_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func initSubroundBlock(blockChain data.ChainHandler, container *mock.ConsensusCoreMock) bn.SubroundBlock {
	if blockChain == nil {
		blockChain = &mock.BlockChainMock{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.Header{}
			},
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					Nonce:     uint64(0),
					Signature: []byte("genesis signature"),
					RandSeed:  []byte{0},
				}
			},
			GetGenesisHeaderHashCalled: func() []byte {
				return []byte("genesis header hash")
			},
		}
	}

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	container.SetBlockchain(blockChain)

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	srBlock, _ := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	return srBlock
}

func initSubroundBlockWithBlockProcessor(bp *mock.BlockProcessorMock, container *mock.ConsensusCoreMock) bn.
	SubroundBlock {
	blockChain := &mock.BlockChainMock{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Nonce:     uint64(0),
				Signature: []byte("genesis signature"),
			}
		},
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("genesis header hash")
		},
	}
	blockProcessorMock := bp

	container.SetBlockchain(blockChain)
	container.SetBlockProcessor(blockProcessorMock)
	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	srBlock, _ := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	return srBlock
}

func TestSubroundBlock_NewSubroundBlockNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srBlock, err := bn.NewSubroundBlock(
		nil,
		sendConsensusMessage,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestSubroundBlock_NewSubroundBlockNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)

	container.SetBlockchain(nil)

	srBlock, err := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundBlock_NewSubroundBlockNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)

	container.SetBlockProcessor(nil)

	srBlock, err := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubroundBlock_NewSubroundBlockNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)

	sr.ConsensusState = nil

	srBlock, err := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundBlock_NewSubroundBlockNilHasherShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)

	container.SetHasher(nil)
	srBlock, err := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubroundBlock_NewSubroundBlockNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)

	container.SetMarshalizer(nil)
	srBlock, err := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestSubroundBlock_NewSubroundBlockNilMultisignerShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)

	container.SetMultiSigner(nil)
	srBlock, err := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestSubroundBlock_NewSubroundBlockNilRounderShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)

	container.SetRounder(nil)
	srBlock, err := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestSubroundBlock_NewSubroundBlockNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)

	container.SetShardCoordinator(nil)
	srBlock, err := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestSubroundBlock_NewSubroundBlockNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)

	container.SetSyncTimer(nil)
	srBlock, err := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundBlock_NewSubroundBlockNilSendConsensusMessageFunctionShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	srBlock, err := bn.NewSubroundBlock(
		sr,
		nil,
		extend,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilSendConsensusMessageFunction, err)
}

func TestSubroundBlock_NewSubroundBlockShouldWork(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		container,
	)
	srBlock, err := bn.NewSubroundBlock(
		sr,
		sendConsensusMessage,
		extend,
	)
	assert.NotNil(t, srBlock)
	assert.Nil(t, err)
}

func TestSubroundBlock_DoBlockJob(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	r := sr.DoBlockJob()
	assert.False(t, r)

	sr.SetSelfPubKey(sr.ConsensusGroup()[0])
	sr.SetJobDone(sr.SelfPubKey(), bn.SrBlock, true)
	r = sr.DoBlockJob()
	assert.False(t, r)

	sr.SetJobDone(sr.SelfPubKey(), bn.SrBlock, false)
	sr.SetStatus(bn.SrBlock, spos.SsFinished)
	r = sr.DoBlockJob()
	assert.False(t, r)

	sr.SetStatus(bn.SrBlock, spos.SsNotFinished)
	bpm := &mock.BlockProcessorMock{}
	err := errors.New("error")
	bpm.CreateBlockCalled = func(round int32, remainingTime func() bool) (data.BodyHandler, error) {
		return nil, err
	}
	container.SetBlockProcessor(bpm)

	r = sr.DoBlockJob()
	assert.False(t, r)

	bpm = mock.InitBlockProcessorMock()
	container.SetBlockProcessor(bpm)
	r = sr.DoBlockJob()
	assert.True(t, r)
	assert.Equal(t, uint64(1), sr.Header.GetNonce())
}

func TestSubroundBlock_ReceivedBlock(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	blockProcessorMock := mock.InitBlockProcessorMock()
	blBody := make(block.Body, 0)
	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		blBodyStr,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)
	sr.BlockBody = make(block.Body, 0)
	r := sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	sr.BlockBody = nil
	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[1])
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[0])
	sr.SetStatus(bn.SrBlock, spos.SsFinished)
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	sr.SetStatus(bn.SrBlock, spos.SsNotFinished)
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	hdr := &block.Header{}
	hdr.Nonce = 2
	hdrStr, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdrHash := mock.HasherMock{}.Compute(string(hdrStr))
	cnsMsg = consensus.NewConsensusMessage(
		hdrHash,
		hdrStr,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	sr.Data = nil
	sr.Header = hdr
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	sr.Header = nil
	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[1])
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[0])
	sr.SetStatus(bn.SrBlock, spos.SsFinished)
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	sr.SetStatus(bn.SrBlock, spos.SsNotFinished)
	container.SetBlockProcessor(blockProcessorMock)
	sr.Data = nil
	sr.Header = nil
	hdr = &block.Header{}
	hdr.Nonce = 1
	hdrStr, _ = mock.MarshalizerMock{}.Marshal(hdr)
	hdrHash = mock.HasherMock{}.Compute(string(hdrStr))
	cnsMsg.BlockHeaderHash = hdrHash
	cnsMsg.SubRoundData = hdrStr
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.True(t, r)
}

func TestSubroundBlock_DecodeBlockBody(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	body := make(block.Body, 0)
	body = append(body, &block.MiniBlock{ReceiverShardID: 69})
	message, err := mock.MarshalizerMock{}.Marshal(body)
	assert.Nil(t, err)

	dcdBlk := sr.DecodeBlockBody(nil)
	assert.Nil(t, dcdBlk)

	dcdBlk = sr.DecodeBlockBody(message)
	assert.Equal(t, body, dcdBlk)
	assert.Equal(t, uint32(69), body[0].ReceiverShardID)
}

func TestSubroundBlock_DecodeBlockHeader(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(sr.Rounder().TimeStamp().Unix())
	hdr.Signature = []byte(sr.SelfPubKey())
	message, err := mock.MarshalizerMock{}.Marshal(hdr)
	assert.Nil(t, err)

	message, err = mock.MarshalizerMock{}.Marshal(hdr)
	assert.Nil(t, err)

	dcdHdr := sr.DecodeBlockHeader(nil)
	assert.Nil(t, dcdHdr)

	dcdHdr = sr.DecodeBlockHeader(message)
	assert.Equal(t, hdr, dcdHdr)
	assert.Equal(t, []byte(sr.SelfPubKey()), dcdHdr.Signature)
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenBodyAndHeaderAreNotSet(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockFails(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	blProcMock := mock.InitBlockProcessorMock()
	err := errors.New("error process block")
	blProcMock.ProcessBlockCalled = func(data.ChainHandler, data.HeaderHandler, data.BodyHandler, func() time.Duration) error {
		return err
	}
	container.SetBlockProcessor(blProcMock)
	hdr := &block.Header{}
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)
	sr.Header = hdr
	sr.BlockBody = blk
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsInNextRound(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	hdr := &block.Header{}
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)
	sr.Header = hdr
	sr.BlockBody = blk
	blockProcessorMock := mock.InitBlockProcessorMock()
	blockProcessorMock.ProcessBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return errors.New("error")
	}
	container.SetBlockProcessor(blockProcessorMock)
	container.SetRounder(&mock.RounderMock{RoundIndex: 1})
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnTrue(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	hdr := &block.Header{}
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)
	sr.Header = hdr
	sr.BlockBody = blk
	assert.True(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_RemainingTimeShouldReturnNegativeValue(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	rounderMock := initRounderMock()
	container.SetRounder(rounderMock)

	sr := *initSubroundBlock(nil, container)
	remainingTimeInThisRound := func() time.Duration {
		roundStartTime := sr.Rounder().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		remainingTime := sr.Rounder().TimeDuration()*85/100 - elapsedTime

		return time.Duration(remainingTime)
	}
	container.SetSyncTimer(mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 84 / 100)
	}})
	ret := remainingTimeInThisRound()
	assert.True(t, ret > 0)

	container.SetSyncTimer(mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 85 / 100)
	}})
	ret = remainingTimeInThisRound()
	assert.True(t, ret == 0)

	container.SetSyncTimer(mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 86 / 100)
	}})
	ret = remainingTimeInThisRound()
	assert.True(t, ret < 0)
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	sr.RoundCanceled = true
	assert.False(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	sr.SetStatus(bn.SrBlock, spos.SsFinished)
	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenBlockIsReceivedReturnTrue(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	for i := 0; i < sr.Threshold(bn.SrBlock); i++ {
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBlock, true)
	}
	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnFalseWhenBlockIsReceivedReturnFalse(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	assert.False(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_IsBlockReceived(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBlock, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrBitmap, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrCommitment, false)
		sr.SetJobDone(sr.ConsensusGroup()[i], bn.SrSignature, false)
	}
	ok := sr.IsBlockReceived(1)
	assert.False(t, ok)

	sr.SetJobDone("A", bn.SrBlock, true)
	isJobDone, _ := sr.JobDone("A", bn.SrBlock)
	assert.True(t, isJobDone)

	ok = sr.IsBlockReceived(1)
	assert.True(t, ok)

	ok = sr.IsBlockReceived(2)
	assert.False(t, ok)
}

func TestSubroundBlock_HaveTimeInCurrentSubroundShouldReturnTrue(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	haveTimeInCurrentSubound := func() bool {
		roundStartTime := sr.Rounder().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		remainingTime := sr.EndTime() - int64(elapsedTime)

		return time.Duration(remainingTime) > 0
	}
	rounderMock := &mock.RounderMock{}
	rounderMock.RoundTimeDuration = time.Duration(4000 * time.Millisecond)
	rounderMock.RoundTimeStamp = time.Unix(0, 0)
	syncTimerMock := &mock.SyncTimerMock{}
	timeElapsed := int64(sr.EndTime() - 1)
	syncTimerMock.CurrentTimeCalled = func() time.Time {
		return time.Unix(0, timeElapsed)
	}
	container.SetRounder(rounderMock)
	container.SetSyncTimer(syncTimerMock)

	assert.True(t, haveTimeInCurrentSubound())
}

func TestSubroundBlock_HaveTimeInCurrentSuboundShouldReturnFalse(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	haveTimeInCurrentSubound := func() bool {
		roundStartTime := sr.Rounder().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		remainingTime := sr.EndTime() - int64(elapsedTime)

		return time.Duration(remainingTime) > 0
	}
	rounderMock := &mock.RounderMock{}
	rounderMock.RoundTimeDuration = time.Duration(4000 * time.Millisecond)
	rounderMock.RoundTimeStamp = time.Unix(0, 0)
	syncTimerMock := &mock.SyncTimerMock{}
	timeElapsed := int64(sr.EndTime() + 1)
	syncTimerMock.CurrentTimeCalled = func() time.Time {
		return time.Unix(0, timeElapsed)
	}
	container.SetRounder(rounderMock)
	container.SetSyncTimer(syncTimerMock)

	assert.False(t, haveTimeInCurrentSubound())
}

func TestSubroundBlock_CreateHeaderNilCurrentHeader(t *testing.T) {
	blockChain := &mock.BlockChainMock{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return nil
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Nonce:     uint64(0),
				Signature: []byte("genesis signature"),
				RandSeed:  []byte{0},
			}
		},
		GetGenesisHeaderHashCalled: func() []byte {
			return []byte("genesis header hash")
		},
	}
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(blockChain, container)
	sr.BlockChain().SetCurrentBlockHeader(nil)
	header, _ := sr.CreateHeader()
	expectedHeader := &block.Header{
		Round:            uint32(sr.Rounder().Index()),
		TimeStamp:        uint64(sr.Rounder().TimeStamp().Unix()),
		RootHash:         []byte{},
		Nonce:            uint64(1),
		PrevHash:         sr.BlockChain().GetGenesisHeaderHash(),
		PrevRandSeed:     sr.BlockChain().GetGenesisHeader().GetSignature(),
		RandSeed:         []byte{0},
		MiniBlockHeaders: header.(*block.Header).MiniBlockHeaders,
	}

	assert.Equal(t, expectedHeader, header)
}

func TestSubroundBlock_CreateHeaderNotNilCurrentHeader(t *testing.T) {
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	sr.BlockChain().SetCurrentBlockHeader(&block.Header{
		Nonce: 1,
	})
	header, _ := sr.CreateHeader()
	expectedHeader := &block.Header{
		Round:            uint32(sr.Rounder().Index()),
		TimeStamp:        uint64(sr.Rounder().TimeStamp().Unix()),
		RootHash:         []byte{},
		Nonce:            uint64(sr.BlockChain().GetCurrentBlockHeader().GetNonce() + 1),
		PrevHash:         sr.BlockChain().GetCurrentBlockHeaderHash(),
		RandSeed:         []byte{0},
		MiniBlockHeaders: header.(*block.Header).MiniBlockHeaders,
	}

	assert.Equal(t, expectedHeader, header)
}

func TestSubroundBlock_CreateHeaderMultipleMiniBlocks(t *testing.T) {
	mbHeaders := []block.MiniBlockHeader{
		{Hash: []byte("mb1"), SenderShardID: 1, ReceiverShardID: 1},
		{Hash: []byte("mb2"), SenderShardID: 1, ReceiverShardID: 2},
		{Hash: []byte("mb3"), SenderShardID: 2, ReceiverShardID: 3},
	}
	blockChainMock := mock.BlockChainMock{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Nonce: 1,
			}
		},
	}
	bp := mock.InitBlockProcessorMock()
	bp.CreateBlockHeaderCalled = func(body data.BodyHandler, round int32, haveTime func() bool) (header data.HeaderHandler, e error) {
		return &block.Header{MiniBlockHeaders: mbHeaders, RootHash: []byte{}}, nil
	}
	container := mock.InitConsensusCore()
	sr := *initSubroundBlockWithBlockProcessor(bp, container)
	container.SetBlockchain(&blockChainMock)
	header, _ := sr.CreateHeader()
	expectedHeader := &block.Header{
		Round:            uint32(sr.Rounder().Index()),
		TimeStamp:        uint64(sr.Rounder().TimeStamp().Unix()),
		RootHash:         []byte{},
		Nonce:            uint64(sr.BlockChain().GetCurrentBlockHeader().GetNonce() + 1),
		PrevHash:         sr.BlockChain().GetCurrentBlockHeaderHash(),
		RandSeed:         []byte{0},
		MiniBlockHeaders: mbHeaders,
	}

	assert.Equal(t, expectedHeader, header)
}

func TestSubroundBlock_CreateHeaderNilMiniBlocks(t *testing.T) {
	expectedErr := errors.New("nil mini blocks")
	bp := mock.InitBlockProcessorMock()
	bp.CreateBlockHeaderCalled = func(body data.BodyHandler, round int32, haveTime func() bool) (header data.HeaderHandler, e error) {
		return nil, expectedErr
	}
	container := mock.InitConsensusCore()
	sr := *initSubroundBlockWithBlockProcessor(bp, container)
	sr.BlockChain().SetCurrentBlockHeader(&block.Header{
		Nonce: 1,
	})
	header, err := sr.CreateHeader()
	assert.Nil(t, header)
	assert.Equal(t, expectedErr, err)
}

func TestSubroundBlock_CallFuncRemainingTimeWithStructShouldWork(t *testing.T) {
	roundStartTime := time.Now()
	maxTime := time.Duration(100 * time.Millisecond)
	newRoundStartTime := time.Time{}
	newRoundStartTime = roundStartTime
	remainingTimeInCurrentRound := func() time.Duration {
		return RemainingTimeWithStruct(newRoundStartTime, maxTime)
	}
	assert.True(t, remainingTimeInCurrentRound() > 0)

	time.Sleep(200 * time.Millisecond)
	assert.True(t, remainingTimeInCurrentRound() < 0)

	roundStartTime = roundStartTime.Add(500 * time.Millisecond)
	assert.True(t, remainingTimeInCurrentRound() < 0)
}

func TestSubroundBlock_CallFuncRemainingTimeWithStructShouldNotWork(t *testing.T) {
	roundStartTime := time.Now()
	maxTime := time.Duration(100 * time.Millisecond)
	remainingTimeInCurrentRound := func() time.Duration {
		return RemainingTimeWithStruct(roundStartTime, maxTime)
	}
	assert.True(t, remainingTimeInCurrentRound() > 0)

	time.Sleep(200 * time.Millisecond)
	assert.True(t, remainingTimeInCurrentRound() < 0)

	roundStartTime = roundStartTime.Add(500 * time.Millisecond)
	assert.False(t, remainingTimeInCurrentRound() < 0)
}

func RemainingTimeWithStruct(startTime time.Time, maxTime time.Duration) time.Duration {
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(startTime)
	remainingTime := maxTime - elapsedTime
	return remainingTime
}
