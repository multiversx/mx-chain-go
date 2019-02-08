package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func initSubroundBlock() bn.SubroundBlock {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	//validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrStartRound),
		int(bn.SrBlock),
		int(bn.SrCommitmentHash),
		int64(5*roundTimeDuration/100),
		int64(25*roundTimeDuration/100),
		"(BLOCK)",
		ch,
	)

	srBlock, _ := bn.NewSubroundBlock(
		sr,
		&blockChain,
		blockProcessorMock,
		bootstraperMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	return srBlock
}

func TestWorker_SendBlock(t *testing.T) {
	sr := *initSubroundBlock()

	sr.Rounder().UpdateRound(time.Now(), time.Now().Add(sr.Rounder().TimeDuration()))

	r := sr.DoBlockJob()
	assert.False(t, r)

	sr.Rounder().UpdateRound(time.Now(), time.Now())
	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)

	r = sr.DoBlockJob()
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsNotFinished)
	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrBlock, true)

	r = sr.DoBlockJob()
	assert.False(t, r)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrBlock, false)
	sr.ConsensusState().RoundConsensus.SetSelfPubKey(sr.ConsensusState().RoundConsensus.ConsensusGroup()[1])

	r = sr.DoBlockJob()
	assert.False(t, r)

	sr.ConsensusState().RoundConsensus.SetSelfPubKey(sr.ConsensusState().RoundConsensus.ConsensusGroup()[0])

	r = sr.DoBlockJob()
	assert.True(t, r)
	assert.Equal(t, uint64(1), sr.ConsensusState().Header.Nonce)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrBlock, false)
	sr.BlockChain().CurrentBlockHeader = sr.ConsensusState().Header

	r = sr.DoBlockJob()
	assert.True(t, r)
	assert.Equal(t, uint64(2), sr.ConsensusState().Header.Nonce)
}

func TestWorker_ReceivedBlock(t *testing.T) {
	sr := *initSubroundBlock()

	sr.Rounder().UpdateRound(time.Now(), time.Now().Add(sr.Rounder().TimeDuration()))

	blBody := &block.TxBlockBody{}

	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)

	cnsDta := spos.NewConsensusData(
		nil,
		blBodyStr,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.ConsensusState().BlockBody = &block.TxBlockBody{}

	r := sr.ReceivedBlockBody(cnsDta)
	assert.False(t, r)

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(blBodyStr))

	hdrStr, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdrHash := mock.HasherMock{}.Compute(string(hdrStr))

	cnsDta = spos.NewConsensusData(
		hdrHash,
		hdrStr,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(sr.Rounder().TimeStamp().Unix()),
		1,
	)

	sr.ConsensusState().Header = nil
	sr.ConsensusState().Data = nil
	r = sr.ReceivedBlockHeader(cnsDta)
	assert.True(t, r)
}

func TestWorker_ReceivedBlockBodyShouldSetJobDone(t *testing.T) {
	sr := *initSubroundBlock()

	sr.Rounder().UpdateRound(time.Now(), time.Now().Add(sr.Rounder().TimeDuration()))

	blBody := &block.TxBlockBody{}

	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)

	cnsDta := spos.NewConsensusData(
		nil,
		blBodyStr,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		1,
	)

	sr.ConsensusState().Header = &block.Header{}

	r := sr.ReceivedBlockBody(cnsDta)
	assert.True(t, r)
}

func TestWorker_ReceivedBlockBodyShouldErrProcessBlock(t *testing.T) {
	sr := *initSubroundBlock()

	sr.Rounder().UpdateRound(time.Now(), time.Now().Add(sr.Rounder().TimeDuration()))

	blBody := &block.TxBlockBody{}

	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)

	cnsDta := spos.NewConsensusData(
		nil,
		blBodyStr,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.ConsensusState().Header = &block.Header{}

	blProcMock := initBlockProcessorMock()

	blProcMock.ProcessBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
		return process.ErrNilPreviousBlockHash
	}

	sr.SetBlockProcessor(blProcMock)

	r := sr.ReceivedBlockBody(cnsDta)
	assert.False(t, r)
}

func TestWorker_DecodeBlockBody(t *testing.T) {
	sr := *initSubroundBlock()

	blk := &block.TxBlockBody{}

	mblks := make([]block.MiniBlock, 0)
	mblks = append(mblks, block.MiniBlock{ShardID: 69})
	blk.MiniBlocks = mblks

	message, err := mock.MarshalizerMock{}.Marshal(blk)

	assert.Nil(t, err)

	dcdBlk := sr.DecodeBlockBody(nil)

	assert.Nil(t, dcdBlk)

	dcdBlk = sr.DecodeBlockBody(message)

	assert.Equal(t, blk, dcdBlk)
	assert.Equal(t, uint32(69), dcdBlk.MiniBlocks[0].ShardID)
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenBodyAndHeaderAreNotSet(t *testing.T) {
	sr := *initSubroundBlock()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	assert.False(t, sr.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockFails(t *testing.T) {
	sr := *initSubroundBlock()

	blProcMock := initBlockProcessorMock()

	err := errors.New("error process block")
	blProcMock.ProcessBlockCalled = func(*blockchain.BlockChain, *block.Header, *block.TxBlockBody, func() time.Duration) error {
		return err
	}

	sr.SetBlockProcessor(blProcMock)

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.ConsensusState().Header = hdr
	sr.ConsensusState().BlockBody = blk

	assert.False(t, sr.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsInNextRound(t *testing.T) {
	sr := *initSubroundBlock()

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.ConsensusState().Header = hdr
	sr.ConsensusState().BlockBody = blk

	sr.SetSyncTimer(mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration)
	}})

	assert.False(t, sr.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnTrue(t *testing.T) {
	sr := *initSubroundBlock()

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)

	sr.ConsensusState().Header = hdr
	sr.ConsensusState().BlockBody = blk

	assert.True(t, sr.ProcessReceivedBlock(cnsDta))
}

func TestHaveTime_ShouldReturnNegativeValue(t *testing.T) {
	sr := *initSubroundBlock()

	sr.SetSyncTimer(mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration)
	}})

	haveTime := func() time.Duration {
		roundStartTime := sr.Rounder().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		haveTime := float64(sr.Rounder().TimeDuration())*float64(0.85) - float64(elapsedTime)

		return time.Duration(haveTime)
	}

	ret := haveTime()

	assert.True(t, ret < 0)
}

func TestWorker_DecodeBlockHeader(t *testing.T) {
	sr := *initSubroundBlock()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(sr.Rounder().TimeStamp().Unix())
	hdr.Signature = []byte(sr.ConsensusState().SelfPubKey())

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	dcdHdr := sr.DecodeBlockHeader(nil)

	assert.Nil(t, dcdHdr)

	dcdHdr = sr.DecodeBlockHeader(message)

	assert.Equal(t, hdr, dcdHdr)
	assert.Equal(t, []byte(sr.ConsensusState().SelfPubKey()), dcdHdr.Signature)
}

func TestWorker_CheckBlockConsensus(t *testing.T) {
	sr := *initSubroundBlock()

	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsNotFinished)

	ok := sr.DoBlockConsensusCheck()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sr.ConsensusState().Status(bn.SrBlock))

	sr.ConsensusState().SetJobDone("B", bn.SrBlock, true)

	ok = sr.DoBlockConsensusCheck()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sr.ConsensusState().Status(bn.SrBlock))
}

func TestWorker_IsBlockReceived(t *testing.T) {
	sr := *initSubroundBlock()

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBlock, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitmentHash, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBitmap, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrCommitment, false)
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := sr.IsBlockReceived(1)
	assert.False(t, ok)

	sr.ConsensusState().SetJobDone("A", bn.SrBlock, true)
	isJobDone, _ := sr.ConsensusState().GetJobDone("A", bn.SrBlock)

	assert.True(t, isJobDone)

	ok = sr.IsBlockReceived(1)
	assert.True(t, ok)

	ok = sr.IsBlockReceived(2)
	assert.False(t, ok)
}

func TestWorker_CheckIfBlockIsValid(t *testing.T) {
	sr := *initSubroundBlock()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(sr.Rounder().TimeStamp().Unix())

	hdr.PrevHash = []byte("X")

	r := sr.CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.PrevHash = []byte("")

	r = sr.CheckIfBlockIsValid(hdr)
	assert.True(t, r)

	hdr.Nonce = 2

	r = sr.CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 1
	sr.BlockChain().CurrentBlockHeader = hdr

	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(sr.Rounder().TimeStamp().Unix())

	r = sr.CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("X")

	r = sr.CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 3
	hdr.PrevHash = []byte("")

	r = sr.CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 2

	prevHeader, _ := mock.MarshalizerMock{}.Marshal(sr.BlockChain().CurrentBlockHeader)
	hdr.PrevHash = mock.HasherMock{}.Compute(string(prevHeader))

	r = sr.CheckIfBlockIsValid(hdr)
	assert.True(t, r)
}
