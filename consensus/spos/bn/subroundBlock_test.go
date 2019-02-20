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
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

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

func TestSubroundBlock_NewSubroundBlockNilSubroundShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

	srBlock, err := bn.NewSubroundBlock(
		nil,
		&blockChain,
		blockProcessorMock,
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

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilSubround)
}

func TestSubroundBlock_NewSubroundBlockNilBlockchainShouldFail(t *testing.T) {
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		nil,
		blockProcessorMock,
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

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilBlockChain)
}

func TestSubroundBlock_NewSubroundBlockNilBlockProcessorShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		&blockChain,
		nil,
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

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilBlockProcessor)
}

func TestSubroundBlock_NewSubroundBlockNilConsensusStateShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		&blockChain,
		blockProcessorMock,
		nil,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilConsensusState)
}

func TestSubroundBlock_NewSubroundBlockNilHasherShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		&blockChain,
		blockProcessorMock,
		consensusState,
		nil,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilHasher)
}

func TestSubroundBlock_NewSubroundBlockNilMarshalizerShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		&blockChain,
		blockProcessorMock,
		consensusState,
		hasherMock,
		nil,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilMarshalizer)
}

func TestSubroundBlock_NewSubroundBlockNilMultisignerShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		&blockChain,
		blockProcessorMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		nil,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilMultiSigner)
}

func TestSubroundBlock_NewSubroundBlockNilRounderShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		&blockChain,
		blockProcessorMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		nil,
		shardCoordinatorMock,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilRounder)
}

func TestSubroundBlock_NewSubroundBlockNilShardCoordinatorShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	syncTimerMock := mock.SyncTimerMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		&blockChain,
		blockProcessorMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		nil,
		syncTimerMock,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilShardCoordinator)
}

func TestSubroundBlock_NewSubroundBlockNilSyncTimerShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		&blockChain,
		blockProcessorMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		nil,
		sendConsensusMessage,
		extend,
	)

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilSyncTimer)
}

func TestSubroundBlock_NewSubroundBlockNilSendConsensusMessageFunctionShouldFail(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		&blockChain,
		blockProcessorMock,
		consensusState,
		hasherMock,
		marshalizerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		nil,
		extend,
	)

	assert.Nil(t, srBlock)
	assert.Equal(t, err, spos.ErrNilSendConsensusMessageFunction)
}

func TestSubroundBlock_NewSubroundBlockShouldWork(t *testing.T) {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	consensusState := initConsensusState()
	hasherMock := mock.HasherMock{}
	marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}

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

	srBlock, err := bn.NewSubroundBlock(
		sr,
		&blockChain,
		blockProcessorMock,
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

	assert.NotNil(t, srBlock)
	assert.Nil(t, err)
}

func TestSubroundBlock_DoBlockJob(t *testing.T) {
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

func TestSubroundBlock_ReceivedBlock(t *testing.T) {
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

func TestSubroundBlock_ReceivedBlockBodyShouldSetJobDone(t *testing.T) {
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

func TestSubroundBlock_ReceivedBlockBodyShouldErrProcessBlock(t *testing.T) {
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

func TestSubroundBlock_DecodeBlockBody(t *testing.T) {
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

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenBodyAndHeaderAreNotSet(t *testing.T) {
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

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockFails(t *testing.T) {
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

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsInNextRound(t *testing.T) {
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

func TestSubroundBlock_ProcessReceivedBlockShouldReturnTrue(t *testing.T) {
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

func TestSubroundBlock_HaveTimeShouldReturnNegativeValue(t *testing.T) {
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

func TestSubroundBlock_DecodeBlockHeader(t *testing.T) {
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

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	sr := *initSubroundBlock()
	sr.ConsensusState().RoundCanceled = true
	assert.False(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	sr := *initSubroundBlock()
	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenBlockIsReceivedReturnTrue(t *testing.T) {
	sr := *initSubroundBlock()

	for i := 0; i < sr.ConsensusState().Threshold(bn.SrBlock); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBlock, true)
	}

	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnFalseWhenBlockIsReceivedReturnFalse(t *testing.T) {
	sr := *initSubroundBlock()
	assert.False(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_IsBlockReceived(t *testing.T) {
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

func TestSubroundBlock_CheckIfBlockIsValid(t *testing.T) {
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
