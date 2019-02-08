package bn_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
)

func TestWorker_DoEndRoundJobNotFinished(t *testing.T) {
	wrk := initWorker()

	wrk.Header = &block.Header{}

	r := wrk.DoEndRoundJob()
	assert.False(t, r)
}

func TestWorker_DoEndRoundJobErrAggregatingSigShouldFail(t *testing.T) {
	wrk := initWorker()

	multiSignerMock := initMultisigner()

	multiSignerMock.AggregateSigsMock = func(bitmap []byte) ([]byte, error) {
		return nil, crypto.ErrNilHasher
	}

	wrk.SetMultiSigner(multiSignerMock)

	wrk.SendMessage = sendMessage
	wrk.BroadcastBlockBody = broadcastMessage
	wrk.BroadcastHeader = broadcastMessage

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	wrk.Header = &block.Header{}

	r := wrk.DoEndRoundJob()
	assert.False(t, r)
}

func TestWorker_DoEndRoundJobErrCommitBlockShouldFail(t *testing.T) {
	wrk := initWorker()

	blProcMock := initMockBlockProcessor()

	blProcMock.CommitBlockCalled = func(
		blockChain *blockchain.BlockChain,
		header *block.Header,
		block *block.TxBlockBody,
	) error {
		return blockchain.ErrHeaderUnitNil
	}

	wrk.SetBlockProcessor(blProcMock)

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	wrk.Header = &block.Header{}

	r := wrk.DoEndRoundJob()
	assert.False(t, r)
}

func TestWorker_DoEndRoundJobErrRemBlockTxOK(t *testing.T) {
	wrk := initWorker()

	blProcMock := initMockBlockProcessor()

	blProcMock.RemoveBlockTxsFromPoolCalled = func(body *block.TxBlockBody) error {
		return process.ErrNilBlockBodyPool
	}

	wrk.SetBlockProcessor(blProcMock)

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	wrk.Header = &block.Header{}

	r := wrk.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_DoEndRoundJobErrBroadcastTxBlockBodyOK(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	wrk.BroadcastBlockBody = nil

	wrk.Header = &block.Header{}

	r := wrk.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_DoEndRoundJobErrBroadcastHeaderOK(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	wrk.BroadcastHeader = nil

	wrk.Header = &block.Header{}

	r := wrk.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_DoEndRoundJobAllOK(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	wrk.Header = &block.Header{}

	r := wrk.DoEndRoundJob()
	assert.True(t, r)
}

func TestWorker_CheckEndRoundConsensus(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitmentHash, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrBitmap, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrCommitment, spos.SsFinished)
	wrk.SPoS.SetStatus(bn.SrSignature, spos.SsFinished)

	ok := wrk.CheckEndRoundConsensus()
	assert.True(t, ok)
}

func TestWorker_ExtendEndRound(t *testing.T) {
	wrk := initWorker()

	wrk.ExtendEndRound()
}

func TestWorker_CheckSignaturesValidityShouldErrNilSignature(t *testing.T) {
	wrk := initWorker()

	err := wrk.CheckSignaturesValidity([]byte(string(2)))
	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestWorker_CheckSignaturesValidityShouldErrInvalidIndex(t *testing.T) {
	wrk := initWorker()

	wrk.MultiSigner().Reset(nil, 0)

	wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrSignature, true)

	err := wrk.CheckSignaturesValidity([]byte(string(1)))
	assert.Equal(t, crypto.ErrInvalidIndex, err)
}

func TestWorker_CheckSignaturesValidityShouldRetunNil(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[0], bn.SrSignature, true)

	err := wrk.CheckSignaturesValidity([]byte(string(1)))
	assert.Equal(t, nil, err)
}
