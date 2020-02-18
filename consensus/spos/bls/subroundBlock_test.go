package bls_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
)

func defaultSubroundForSRBlock(consensusState *spos.ConsensusState, ch chan bool,
	container *mock.ConsensusCoreMock) (*spos.Subround, error) {
	return spos.NewSubround(
		SrStartRound,
		SrBlock,
		SrCommitmentHash,
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		consensusState,
		ch,
		executeStoredMessages,
		container,
		chainID,
	)
}

func defaultSubroundBlockFromSubround(sr *spos.Subround) (bls.SubroundBlock, error) {
	srBlock, err := bls.NewSubroundBlock(
		sr,
		extend,
		processingThresholdPercent,
	)

	return srBlock, err
}

func defaultSubroundBlockWithoutErrorFromSubround(sr *spos.Subround) bls.SubroundBlock {
	srBlock, _ := bls.NewSubroundBlock(
		sr,
		extend,
		processingThresholdPercent,
	)

	return srBlock
}

func initSubroundBlock(blockChain data.ChainHandler, container *mock.ConsensusCoreMock) bls.SubroundBlock {
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

	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)
	srBlock, _ := defaultSubroundBlockFromSubround(sr)
	return srBlock
}

func initSubroundBlockWithBlockProcessor(
	bp *mock.BlockProcessorMock,
	container *mock.ConsensusCoreMock,
) bls.SubroundBlock {
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

	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)
	srBlock, _ := defaultSubroundBlockFromSubround(sr)
	return srBlock
}

func TestSubroundBlock_NewSubroundBlockNilSubroundShouldFail(t *testing.T) {
	t.Parallel()

	srBlock, err := bls.NewSubroundBlock(
		nil,
		extend,
		processingThresholdPercent,
	)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilSubround, err)
}

func TestSubroundBlock_NewSubroundBlockNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)

	container.SetBlockchain(nil)

	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestSubroundBlock_NewSubroundBlockNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)

	container.SetBlockProcessor(nil)

	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestSubroundBlock_NewSubroundBlockNilConsensusStateShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)

	sr.ConsensusState = nil

	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilConsensusState, err)
}

func TestSubroundBlock_NewSubroundBlockNilHasherShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)

	container.SetHasher(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestSubroundBlock_NewSubroundBlockNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)

	container.SetMarshalizer(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestSubroundBlock_NewSubroundBlockNilMultisignerShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)

	container.SetMultiSigner(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestSubroundBlock_NewSubroundBlockNilRounderShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)

	container.SetRounder(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilRounder, err)
}

func TestSubroundBlock_NewSubroundBlockNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)

	container.SetShardCoordinator(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestSubroundBlock_NewSubroundBlockNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()

	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)

	container.SetSyncTimer(nil)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
	assert.Nil(t, srBlock)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestSubroundBlock_NewSubroundBlockShouldWork(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)
	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)
	srBlock, err := defaultSubroundBlockFromSubround(sr)
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
	_ = sr.SetJobDone(sr.SelfPubKey(), SrBlock, true)
	r = sr.DoBlockJob()
	assert.False(t, r)

	_ = sr.SetJobDone(sr.SelfPubKey(), SrBlock, false)
	sr.SetStatus(SrBlock, spos.SsFinished)
	r = sr.DoBlockJob()
	assert.False(t, r)

	sr.SetStatus(SrBlock, spos.SsNotFinished)
	bpm := &mock.BlockProcessorMock{}
	err := errors.New("error")
	bpm.CreateBlockCalled = func(header data.HeaderHandler, remainingTime func() bool) (data.BodyHandler, error) {
		return nil, err
	}
	container.SetBlockProcessor(bpm)
	r = sr.DoBlockJob()
	assert.False(t, r)

	bpm = mock.InitBlockProcessorMock()
	container.SetBlockProcessor(bpm)
	bm := &mock.BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
			return nil
		},
	}
	container.SetBroadcastMessenger(bm)
	container.SetRounder(&mock.RounderMock{
		RoundIndex: 1,
	})
	r = sr.DoBlockJob()
	assert.True(t, r)
	assert.Equal(t, uint64(1), sr.Header.GetNonce())
}

func TestSubroundBlock_ReceivedBlock(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	blockProcessorMock := mock.InitBlockProcessorMock()
	blBody := &block.Body{}
	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)
	cnsMsg := consensus.NewConsensusMessage(
		nil,
		blBodyStr,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		MtBlockBody,
		0,
		chainID,
		nil,
		nil,
		nil,
	)
	sr.Body = &block.Body{}
	r := sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	sr.Body = nil
	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[1])
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusGroup()[0])
	sr.SetStatus(SrBlock, spos.SsFinished)
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	sr.SetStatus(SrBlock, spos.SsNotFinished)
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
		MtBlockHeader,
		0,
		chainID,
		nil,
		nil,
		nil,
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
	sr.SetStatus(SrBlock, spos.SsFinished)
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	sr.SetStatus(SrBlock, spos.SsNotFinished)
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

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenBodyAndHeaderAreNotSet(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	blk := &block.Body{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		MtBlockBody,
		0,
		chainID,
		nil,
		nil,
		nil,
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
	blk := &block.Body{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		MtBlockBody,
		0,
		chainID,
		nil,
		nil,
		nil,
	)
	sr.Header = hdr
	sr.Body = blk
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsInNextRound(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	hdr := &block.Header{}
	blk := &block.Body{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		MtBlockBody,
		0,
		chainID,
		nil,
		nil,
		nil,
	)
	sr.Header = hdr
	sr.Body = blk
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
	blk := &block.Body{
		MiniBlocks: []*block.MiniBlock{},
	}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		MtBlockBody,
		0,
		chainID,
		nil,
		nil,
		nil,
	)
	sr.Header = hdr
	sr.Body = blk
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

		return remainingTime
	}
	container.SetSyncTimer(&mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 84 / 100)
	}})
	ret := remainingTimeInThisRound()
	assert.True(t, ret > 0)

	container.SetSyncTimer(&mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 85 / 100)
	}})
	ret = remainingTimeInThisRound()
	assert.True(t, ret == 0)

	container.SetSyncTimer(&mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
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
	sr.SetStatus(SrBlock, spos.SsFinished)
	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenBlockIsReceivedReturnTrue(t *testing.T) {
	t.Parallel()
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	for i := 0; i < sr.Threshold(SrBlock); i++ {
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], SrBlock, true)
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
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], SrBlock, false)
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], SrCommitmentHash, false)
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], SrBitmap, false)
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], SrCommitment, false)
		_ = sr.SetJobDone(sr.ConsensusGroup()[i], SrSignature, false)
	}
	ok := sr.IsBlockReceived(1)
	assert.False(t, ok)

	_ = sr.SetJobDone("A", SrBlock, true)
	isJobDone, _ := sr.JobDone("A", SrBlock)
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
	rounderMock.TimeDurationCalled = func() time.Duration {
		return 4000 * time.Millisecond
	}
	rounderMock.TimeStampCalled = func() time.Time {
		return time.Unix(0, 0)
	}
	syncTimerMock := &mock.SyncTimerMock{}
	timeElapsed := sr.EndTime() - 1
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
	rounderMock.TimeDurationCalled = func() time.Duration {
		return 4000 * time.Millisecond
	}
	rounderMock.TimeStampCalled = func() time.Time {
		return time.Unix(0, 0)
	}
	syncTimerMock := &mock.SyncTimerMock{}
	timeElapsed := sr.EndTime() + 1
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
	_ = sr.BlockChain().SetCurrentBlockHeader(nil)
	header, _ := sr.CreateHeader()
	body, _ := sr.CreateBody(header)
	_, _ = sr.BlockProcessor().ApplyBodyToHeader(header, body)
	marshalizedBody, _ := sr.Marshalizer().Marshal(body)
	marshalizedHeader, _ := sr.Marshalizer().Marshal(header)
	_ = sr.SendBlockBody(body, marshalizedBody)
	_ = sr.SendBlockHeader(header, marshalizedHeader)

	oldRand := sr.BlockChain().GetGenesisHeader().GetRandSeed()
	newRand, _ := sr.SingleSigner().Sign(sr.PrivateKey(), oldRand)
	expectedHeader := &block.Header{
		Round:            uint64(sr.Rounder().Index()),
		TimeStamp:        uint64(sr.Rounder().TimeStamp().Unix()),
		RootHash:         []byte{},
		Nonce:            uint64(1),
		PrevHash:         sr.BlockChain().GetGenesisHeaderHash(),
		PrevRandSeed:     sr.BlockChain().GetGenesisHeader().GetRandSeed(),
		RandSeed:         newRand,
		MiniBlockHeaders: header.(*block.Header).MiniBlockHeaders,
		ChainID:          chainID,
	}

	assert.Equal(t, expectedHeader, header)
}

func TestSubroundBlock_CreateHeaderNotNilCurrentHeader(t *testing.T) {
	container := mock.InitConsensusCore()
	sr := *initSubroundBlock(nil, container)
	_ = sr.BlockChain().SetCurrentBlockHeader(&block.Header{
		Nonce: 1,
	})

	header, _ := sr.CreateHeader()
	body, _ := sr.CreateBody(header)
	_, _ = sr.BlockProcessor().ApplyBodyToHeader(header, body)
	marshalizedBody, _ := sr.Marshalizer().Marshal(body)
	marshalizedHeader, _ := sr.Marshalizer().Marshal(header)
	_ = sr.SendBlockBody(body, marshalizedBody)
	_ = sr.SendBlockHeader(header, marshalizedHeader)

	oldRand := sr.BlockChain().GetGenesisHeader().GetRandSeed()
	newRand, _ := sr.SingleSigner().Sign(sr.PrivateKey(), oldRand)

	expectedHeader := &block.Header{
		Round:            uint64(sr.Rounder().Index()),
		TimeStamp:        uint64(sr.Rounder().TimeStamp().Unix()),
		RootHash:         []byte{},
		Nonce:            sr.BlockChain().GetCurrentBlockHeader().GetNonce() + 1,
		PrevHash:         sr.BlockChain().GetCurrentBlockHeaderHash(),
		RandSeed:         newRand,
		MiniBlockHeaders: header.(*block.Header).MiniBlockHeaders,
		ChainID:          chainID,
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
	bp.ApplyBodyToHeaderCalled = func(header data.HeaderHandler, body data.BodyHandler) (data.BodyHandler, error) {
		shardHeader, _ := header.(*block.Header)
		shardHeader.MiniBlockHeaders = mbHeaders
		shardHeader.RootHash = []byte{}

		return body, nil
	}
	container := mock.InitConsensusCore()
	sr := *initSubroundBlockWithBlockProcessor(bp, container)
	container.SetBlockchain(&blockChainMock)

	header, _ := sr.CreateHeader()
	body, _ := sr.CreateBody(header)
	_, _ = sr.BlockProcessor().ApplyBodyToHeader(header, body)
	marshalizedBody, _ := sr.Marshalizer().Marshal(body)
	marshalizedHeader, _ := sr.Marshalizer().Marshal(header)
	_ = sr.SendBlockBody(body, marshalizedBody)
	_ = sr.SendBlockHeader(header, marshalizedHeader)

	oldRand := sr.BlockChain().GetCurrentBlockHeader().GetRandSeed()
	newRand, _ := sr.SingleSigner().Sign(sr.PrivateKey(), oldRand)
	expectedHeader := &block.Header{
		Round:            uint64(sr.Rounder().Index()),
		TimeStamp:        uint64(sr.Rounder().TimeStamp().Unix()),
		RootHash:         []byte{},
		Nonce:            sr.BlockChain().GetCurrentBlockHeader().GetNonce() + 1,
		PrevHash:         sr.BlockChain().GetCurrentBlockHeaderHash(),
		RandSeed:         newRand,
		MiniBlockHeaders: mbHeaders,
		ChainID:          chainID,
	}

	assert.Equal(t, expectedHeader, header)
}

func TestSubroundBlock_CreateHeaderNilMiniBlocks(t *testing.T) {
	expectedErr := errors.New("nil mini blocks")
	bp := mock.InitBlockProcessorMock()
	bp.ApplyBodyToHeaderCalled = func(header data.HeaderHandler, body data.BodyHandler) (data.BodyHandler, error) {
		return body, expectedErr
	}
	container := mock.InitConsensusCore()
	sr := *initSubroundBlockWithBlockProcessor(bp, container)
	_ = sr.BlockChain().SetCurrentBlockHeader(&block.Header{
		Nonce: 1,
	})
	header, _ := sr.CreateHeader()
	body, _ := sr.CreateBody(header)

	_, err := sr.BlockProcessor().ApplyBodyToHeader(header, body)
	assert.Equal(t, expectedErr, err)
}

func TestSubroundBlock_CallFuncRemainingTimeWithStructShouldWork(t *testing.T) {
	roundStartTime := time.Now()
	maxTime := 100 * time.Millisecond
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
	maxTime := 100 * time.Millisecond
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

func TestSubroundBlock_ReceivedBlockComputeProcessDuration(t *testing.T) {
	t.Parallel()

	srStartTime := int64(5 * roundTimeDuration / 100)
	srEndTime := int64(25 * roundTimeDuration / 100)
	srDuration := srEndTime - srStartTime
	delay := srDuration * 430 / 1000

	container := mock.InitConsensusCore()
	container.SetBlockProcessor(&mock.BlockProcessorMock{
		ProcessBlockCalled: func(_ data.ChainHandler, _ data.HeaderHandler, _ data.BodyHandler, _ func() time.Duration) error {
			time.Sleep(time.Duration(delay))
			return nil
		},
	})
	sr := *initSubroundBlock(nil, container)
	hdr := &block.Header{}
	blk := &block.Body{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsMsg := consensus.NewConsensusMessage(
		message,
		nil,
		[]byte(sr.ConsensusGroup()[0]),
		[]byte("sig"),
		MtBlockBody,
		0,
		chainID,
		nil,
		nil,
		nil,
	)
	sr.Header = hdr
	sr.Body = blk
	receivedValue := uint64(0)
	_ = sr.SetAppStatusHandler(&mock.AppStatusHandlerStub{
		SetUInt64ValueHandler: func(key string, value uint64) {
			receivedValue = value
		},
	})

	minimumExpectedValue := uint64(delay * 100 / srDuration)
	_ = sr.ProcessReceivedBlock(cnsMsg)

	assert.True(t,
		receivedValue >= minimumExpectedValue,
		fmt.Sprintf("minimum expected was %d, got %d", minimumExpectedValue, receivedValue),
	)
}

func TestSubroundBlock_ReceivedBlockComputeProcessDurationWithZeroDurationShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not have paniced", r)
		}
	}()

	container := mock.InitConsensusCore()

	consensusState := initConsensusState()
	ch := make(chan bool, 1)

	sr, _ := defaultSubroundForSRBlock(consensusState, ch, container)
	srBlock := *defaultSubroundBlockWithoutErrorFromSubround(sr)

	srBlock.ComputeSubroundProcessingMetric(time.Now(), "dummy")
}
