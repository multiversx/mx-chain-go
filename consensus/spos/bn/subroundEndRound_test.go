package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
)

func initSubroundEndRound() bn.SubroundEndRound {
	blockChain := blockchain.BlockChain{}
	blockProcessorMock := initBlockProcessorMock()
	//bootstraperMock := &mock.BootstraperMock{ShouldSyncCalled: func() bool {
	//	return false
	//}}

	consensusState := initConsensusState()
	//hasherMock := mock.HasherMock{}
	//marshalizerMock := mock.MarshalizerMock{}
	multiSignerMock := initMultiSignerMock()
	rounderMock := initRounderMock()
	//shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := mock.SyncTimerMock{}
	//validatorGroupSelector := mock.ValidatorGroupSelectorMock{}

	ch := make(chan bool, 1)

	sr, _ := bn.NewSubround(
		int(bn.SrSignature),
		int(bn.SrEndRound),
		-1,
		int64(85*roundTimeDuration/100),
		int64(95*roundTimeDuration/100),
		"(END_ROUND)",
		ch,
	)

	srEndRound, _ := bn.NewSubroundEndRound(
		sr,
		&blockChain,
		blockProcessorMock,
		consensusState,
		multiSignerMock,
		rounderMock,
		syncTimerMock,
		broadcastTxBlockBody,
		broadcastHeader,
		extend,
	)

	return srEndRound
}

func TestWorker_DoEndRoundJobErrAggregatingSigShouldFail(t *testing.T) {
	sr := *initSubroundEndRound()

	multiSignerMock := initMultiSignerMock()

	multiSignerMock.AggregateSigsMock = func(bitmap []byte) ([]byte, error) {
		return nil, crypto.ErrNilHasher
	}

	sr.SetMultiSigner(multiSignerMock)

	//sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	sr.ConsensusState().Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestWorker_DoEndRoundJobErrCommitBlockShouldFail(t *testing.T) {
	sr := *initSubroundEndRound()

	blProcMock := initBlockProcessorMock()

	blProcMock.CommitBlockCalled = func(
		blockChain *blockchain.BlockChain,
		header *block.Header,
		block *block.TxBlockBody,
	) error {
		return blockchain.ErrHeaderUnitNil
	}

	sr.SetBlockProcessor(blProcMock)

	//sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	sr.ConsensusState().Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.False(t, r)
}

func TestWorker_DoEndRoundJobErrRemBlockTxOK(t *testing.T) {
	sr := *initSubroundEndRound()

	blProcMock := initBlockProcessorMock()

	blProcMock.RemoveBlockTxsFromPoolCalled = func(body *block.TxBlockBody) error {
		return process.ErrNilBlockBodyPool
	}

	sr.SetBlockProcessor(blProcMock)

	//sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	sr.ConsensusState().Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_DoEndRoundJobErrBroadcastTxBlockBodyOK(t *testing.T) {
	sr := *initSubroundEndRound()

	//sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	sr.SetBroadcastTxBlockBody(func(txBlockBody *block.TxBlockBody) error {
		return spos.ErrNilBroadcastTxBlockBodyFunction
	})

	sr.ConsensusState().Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_DoEndRoundJobErrBroadcastHeaderOK(t *testing.T) {
	sr := *initSubroundEndRound()

	//sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	sr.SetBroadcastHeader(func(header *block.Header) error {
		return spos.ErrNilBroadcastHeaderFunction
	})

	sr.ConsensusState().Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_DoEndRoundJobAllOK(t *testing.T) {
	sr := *initSubroundEndRound()

	//sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	sr.ConsensusState().Header = &block.Header{}

	r := sr.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_CheckEndRoundConsensus(t *testing.T) {
	sr := *initSubroundEndRound()

	//sr.ConsensusState().SetStatus(bn.SrBlock, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrBitmap, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrCommitment, spos.SsFinished)
	//sr.ConsensusState().SetStatus(bn.SrSignature, spos.SsFinished)

	ok := sr.DoEndRoundConsensusCheck()
	assert.True(t, ok)
}

func TestWorker_CheckSignaturesValidityShouldErrNilSignature(t *testing.T) {
	sr := *initSubroundEndRound()

	err := sr.CheckSignaturesValidity([]byte(string(2)))
	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestWorker_CheckSignaturesValidityShouldErrInvalidIndex(t *testing.T) {
	sr := *initSubroundEndRound()

	sr.MultiSigner().Reset(nil, 0)

	sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrSignature, true)

	err := sr.CheckSignaturesValidity([]byte(string(1)))
	assert.Equal(t, crypto.ErrInvalidIndex, err)
}

func TestWorker_CheckSignaturesValidityShouldRetunNil(t *testing.T) {
	sr := *initSubroundEndRound()

	sr.ConsensusState().SetJobDone(sr.ConsensusState().ConsensusGroup()[0], bn.SrSignature, true)

	err := sr.CheckSignaturesValidity([]byte(string(1)))
	assert.Equal(t, nil, err)
}
