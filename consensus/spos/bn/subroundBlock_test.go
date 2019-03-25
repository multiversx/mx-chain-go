package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func initSubroundBlock(blockChain data.ChainHandler) bn.SubroundBlock {
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
		blockChain,
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

func initSubroundBlockWithBlockProcessor(bp *mock.BlockProcessorMock) bn.SubroundBlock {
	blockChain := mock.BlockChainMock{
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
	blockChain := mock.BlockChainMock{}
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
	t.Parallel()
	sr := *initSubroundBlock(nil)
	r := sr.DoBlockJob()
	assert.False(t, r)

	sr.ConsensusState().SetSelfPubKey(sr.ConsensusState().ConsensusGroup()[0])
	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrBlock, true)
	r = sr.DoBlockJob()
	assert.False(t, r)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().SelfPubKey(), bn.SrBlock, false)
	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	r = sr.DoBlockJob()
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsNotFinished)
	bpm := &mock.BlockProcessorMock{
		GetRootHashCalled: func() []byte {
			return []byte{}
		},
	}
	err := errors.New("error")
	bpm.CreateBlockCalled = func(round int32, remainingTime func() bool) (data.BodyHandler, error) {
		return nil, err
	}
	sr.SetBlockProcessor(bpm)
	r = sr.DoBlockJob()
	assert.False(t, r)

	bpm = initBlockProcessorMock()
	sr.SetBlockProcessor(bpm)
	r = sr.DoBlockJob()
	assert.True(t, r)
	assert.Equal(t, uint64(1), sr.ConsensusState().Header.GetNonce())
}

func TestSubroundBlock_ReceivedBlock(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	blockProcessorMock := initBlockProcessorMock()
	blBody := make(block.Body, 0)
	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)
	cnsMsg := spos.NewConsensusMessage(
		nil,
		blBodyStr,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)
	sr.ConsensusState().BlockBody = make(block.Body, 0)
	r := sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().BlockBody = nil
	cnsMsg.PubKey = []byte(sr.ConsensusState().ConsensusGroup()[1])
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusState().ConsensusGroup()[0])
	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsNotFinished)
	r = sr.ReceivedBlockBody(cnsMsg)
	assert.False(t, r)

	hdr := &block.Header{}
	hdr.Nonce = 2
	hdrStr, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdrHash := mock.HasherMock{}.Compute(string(hdrStr))
	cnsMsg = spos.NewConsensusMessage(
		hdrHash,
		hdrStr,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().Data = nil
	sr.ConsensusState().Header = hdr
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	sr.ConsensusState().Header = nil
	cnsMsg.PubKey = []byte(sr.ConsensusState().ConsensusGroup()[1])
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	cnsMsg.PubKey = []byte(sr.ConsensusState().ConsensusGroup()[0])
	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	blockProcessorMock.CheckBlockValidityCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) bool {
		return false
	}
	sr.SetBlockProcessor(blockProcessorMock)
	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsNotFinished)
	r = sr.ReceivedBlockHeader(cnsMsg)
	assert.False(t, r)

	blockProcessorMock.CheckBlockValidityCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) bool {
		return true
	}
	sr.SetBlockProcessor(blockProcessorMock)
	sr.ConsensusState().Data = nil
	sr.ConsensusState().Header = nil
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
	sr := *initSubroundBlock(nil)
	body := make(block.Body, 0)
	body = append(body, &block.MiniBlock{ShardID: 69})
	message, err := mock.MarshalizerMock{}.Marshal(body)
	assert.Nil(t, err)

	dcdBlk := sr.DecodeBlockBody(nil)
	assert.Nil(t, dcdBlk)

	dcdBlk = sr.DecodeBlockBody(message)
	assert.Equal(t, body, dcdBlk)
	assert.Equal(t, uint32(69), body[0].ShardID)
}

func TestSubroundBlock_DecodeBlockHeader(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = uint64(sr.Rounder().TimeStamp().Unix())
	hdr.Signature = []byte(sr.ConsensusState().SelfPubKey())
	message, err := mock.MarshalizerMock{}.Marshal(hdr)
	assert.Nil(t, err)

	message, err = mock.MarshalizerMock{}.Marshal(hdr)
	assert.Nil(t, err)

	dcdHdr := sr.DecodeBlockHeader(nil)
	assert.Nil(t, dcdHdr)

	dcdHdr = sr.DecodeBlockHeader(message)
	assert.Equal(t, hdr, dcdHdr)
	assert.Equal(t, []byte(sr.ConsensusState().SelfPubKey()), dcdHdr.Signature)
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenBodyAndHeaderAreNotSet(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := spos.NewConsensusMessage(
		message,
		nil,
		[]byte(sr.ConsensusState().ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		0,
	)
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockFails(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	blProcMock := initBlockProcessorMock()
	err := errors.New("error process block")
	blProcMock.ProcessBlockCalled = func(data.ChainHandler, data.HeaderHandler, data.BodyHandler, func() time.Duration) error {
		return err
	}
	sr.SetBlockProcessor(blProcMock)
	hdr := &block.Header{}
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := spos.NewConsensusMessage(
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
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsInNextRound(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	hdr := &block.Header{}
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := spos.NewConsensusMessage(
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
	blockProcessorMock := initBlockProcessorMock()
	blockProcessorMock.ProcessBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return errors.New("error")
	}
	sr.SetBlockProcessor(blockProcessorMock)
	sr.SetRounder(&mock.RounderMock{RoundIndex: 1})
	assert.False(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_ProcessReceivedBlockShouldReturnTrue(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	hdr := &block.Header{}
	blk := make(block.Body, 0)
	message, _ := mock.MarshalizerMock{}.Marshal(blk)
	cnsMsg := spos.NewConsensusMessage(
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
	assert.True(t, sr.ProcessReceivedBlock(cnsMsg))
}

func TestSubroundBlock_RemainingTimeShouldReturnNegativeValue(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	remainingTimeInThisRound := func() time.Duration {
		roundStartTime := sr.Rounder().TimeStamp()
		currentTime := sr.SyncTimer().CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		remainingTime := sr.Rounder().TimeDuration()*85/100 - elapsedTime

		return time.Duration(remainingTime)
	}
	sr.SetSyncTimer(mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 84 / 100)
	}})
	ret := remainingTimeInThisRound()
	assert.True(t, ret > 0)

	sr.SetSyncTimer(mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 85 / 100)
	}})
	ret = remainingTimeInThisRound()
	assert.True(t, ret == 0)

	sr.SetSyncTimer(mock.SyncTimerMock{CurrentTimeCalled: func() time.Time {
		return time.Unix(0, 0).Add(roundTimeDuration * 86 / 100)
	}})
	ret = remainingTimeInThisRound()
	assert.True(t, ret < 0)
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnFalseWhenRoundIsCanceled(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	sr.ConsensusState().RoundCanceled = true
	assert.False(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenSubroundIsFinished(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnTrueWhenBlockIsReceivedReturnTrue(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	for i := 0; i < sr.ConsensusState().Threshold(bn.SrBlock); i++ {
		sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[i], bn.SrBlock, true)
	}
	assert.True(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_DoBlockConsensusCheckShouldReturnFalseWhenBlockIsReceivedReturnFalse(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
	assert.False(t, sr.DoBlockConsensusCheck())
}

func TestSubroundBlock_IsBlockReceived(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
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
	isJobDone, _ := sr.ConsensusState().JobDone("A", bn.SrBlock)
	assert.True(t, isJobDone)

	ok = sr.IsBlockReceived(1)
	assert.True(t, ok)

	ok = sr.IsBlockReceived(2)
	assert.False(t, ok)
}

func TestSubroundBlock_HaveTimeInCurrentSubroundShouldReturnTrue(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
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
	sr.SetRounder(rounderMock)
	sr.SetSyncTimer(syncTimerMock)

	assert.True(t, haveTimeInCurrentSubound())
}

func TestSubroundBlock_HaveTimeInCurrentSuboundShouldReturnFalse(t *testing.T) {
	t.Parallel()
	sr := *initSubroundBlock(nil)
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
	sr.SetRounder(rounderMock)
	sr.SetSyncTimer(syncTimerMock)

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
	sr := *initSubroundBlock(blockChain)
	sr.BlockChain().SetCurrentBlockHeader(nil)
	header, _ := sr.CreateHeader()
	bp := initBlockProcessorMock()
	expectedHeader := &block.Header{
		Round:            uint32(sr.Rounder().Index()),
		TimeStamp:        uint64(sr.Rounder().TimeStamp().Unix()),
		RootHash:         bp.GetRootHash(),
		Nonce:            uint64(1),
		PrevHash:         sr.BlockChain().GetGenesisHeaderHash(),
		PrevRandSeed:     sr.BlockChain().GetGenesisHeader().GetSignature(),
		RandSeed:         []byte{0},
		MiniBlockHeaders: header.(*block.Header).MiniBlockHeaders,
	}

	assert.Equal(t, expectedHeader, header)
}

func TestSubroundBlock_CreateHeaderNotNilCurrentHeader(t *testing.T) {
	sr := *initSubroundBlock(nil)
	sr.BlockChain().SetCurrentBlockHeader(&block.Header{
		Nonce: 1,
	})
	header, _ := sr.CreateHeader()
	bp := initBlockProcessorMock()
	expectedHeader := &block.Header{
		Round:            uint32(sr.Rounder().Index()),
		TimeStamp:        uint64(sr.Rounder().TimeStamp().Unix()),
		RootHash:         bp.GetRootHash(),
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
	bp := initBlockProcessorMock()
	bp.CreateBlockHeaderCalled = func(body data.BodyHandler) (header data.HeaderHandler, e error) {
		return &block.Header{MiniBlockHeaders: mbHeaders, RootHash: bp.GetRootHash()}, nil
	}
	sr := *initSubroundBlockWithBlockProcessor(bp)
	sr.SetBlockChain(&blockChainMock)
	header, _ := sr.CreateHeader()
	expectedHeader := &block.Header{
		Round:            uint32(sr.Rounder().Index()),
		TimeStamp:        uint64(sr.Rounder().TimeStamp().Unix()),
		RootHash:         bp.GetRootHash(),
		Nonce:            uint64(sr.BlockChain().GetCurrentBlockHeader().GetNonce() + 1),
		PrevHash:         sr.BlockChain().GetCurrentBlockHeaderHash(),
		RandSeed:         []byte{0},
		MiniBlockHeaders: mbHeaders,
	}

	assert.Equal(t, expectedHeader, header)
}

func TestSubroundBlock_CreateHeaderNilMiniBlocks(t *testing.T) {
	expectedErr := errors.New("nil mini blocks")
	bp := initBlockProcessorMock()
	bp.CreateBlockHeaderCalled = func(body data.BodyHandler) (header data.HeaderHandler, e error) {
		return nil, expectedErr
	}
	sr := *initSubroundBlockWithBlockProcessor(bp)
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
