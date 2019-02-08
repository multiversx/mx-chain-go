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

func TestWorker_SendBlock(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(wrk.SPoS.Chr.Round().TimeDuration()))

	r := wrk.DoBlockJob()
	assert.False(t, r)

	wrk.SPoS.Chr.Round().UpdateRound(time.Now(), time.Now())
	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsFinished)

	r = wrk.DoBlockJob()
	assert.False(t, r)

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsNotFinished)
	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBlock, true)

	r = wrk.DoBlockJob()
	assert.False(t, r)

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBlock, false)
	wrk.SPoS.RoundConsensus.SetSelfPubKey(wrk.SPoS.RoundConsensus.ConsensusGroup()[1])

	r = wrk.DoBlockJob()
	assert.False(t, r)

	wrk.SPoS.RoundConsensus.SetSelfPubKey(wrk.SPoS.RoundConsensus.ConsensusGroup()[0])

	r = wrk.DoBlockJob()
	assert.True(t, r)
	assert.Equal(t, uint64(1), wrk.Header.Nonce)

	wrk.SPoS.SetJobDone(wrk.SPoS.SelfPubKey(), bn.SrBlock, false)
	wrk.BlockChain.CurrentBlockHeader = wrk.Header

	r = wrk.DoBlockJob()
	assert.True(t, r)
	assert.Equal(t, uint64(2), wrk.Header.Nonce)
}

func TestWorker_ReceivedBlock(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(wrk.SPoS.Chr.Round().TimeDuration()))

	blBody := &block.TxBlockBody{}

	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)

	cnsDta := spos.NewConsensusData(
		nil,
		blBodyStr,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.BlockBody = &block.TxBlockBody{}

	r := wrk.ReceivedBlockBody(cnsDta)
	assert.False(t, r)

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(blBodyStr))

	hdrStr, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdrHash := mock.HasherMock{}.Compute(string(hdrStr))

	cnsDta = spos.NewConsensusData(
		hdrHash,
		hdrStr,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		wrk.SPoS.Chr.RoundTimeStamp(),
		1,
	)

	wrk.Header = nil
	wrk.SPoS.Data = nil
	r = wrk.ReceivedBlockHeader(cnsDta)
	assert.True(t, r)
}

func TestWorker_ReceivedBlockBodyShouldSetJobDone(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(wrk.SPoS.Chr.Round().TimeDuration()))

	blBody := &block.TxBlockBody{}

	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)

	cnsDta := spos.NewConsensusData(
		nil,
		blBodyStr,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		1,
	)

	wrk.Header = &block.Header{}

	r := wrk.ReceivedBlockBody(cnsDta)
	assert.True(t, r)
}

func TestWorker_ReceivedBlockBodyShouldErrProcessBlock(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(wrk.SPoS.Chr.Round().TimeDuration()))

	blBody := &block.TxBlockBody{}

	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)

	cnsDta := spos.NewConsensusData(
		nil,
		blBodyStr,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.Header = &block.Header{}

	blProcMock := initMockBlockProcessor()

	blProcMock.ProcessBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
		return process.ErrNilPreviousBlockHash
	}

	wrk.SetBlockProcessor(blProcMock)

	r := wrk.ReceivedBlockBody(cnsDta)
	assert.False(t, r)
}

func TestWorker_DecodeBlockBody(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}

	mblks := make([]block.MiniBlock, 0)
	mblks = append(mblks, block.MiniBlock{ShardID: 69})
	blk.MiniBlocks = mblks

	message, err := mock.MarshalizerMock{}.Marshal(blk)

	assert.Nil(t, err)

	dcdBlk := wrk.DecodeBlockBody(nil)

	assert.Nil(t, dcdBlk)

	dcdBlk = wrk.DecodeBlockBody(message)

	assert.Equal(t, blk, dcdBlk)
	assert.Equal(t, uint32(69), dcdBlk.MiniBlocks[0].ShardID)
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenBodyAndHeaderAreNotSet(t *testing.T) {
	wrk := initWorker()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	assert.False(t, wrk.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockFails(t *testing.T) {
	wrk := initWorker()

	blProcMock := initMockBlockProcessor()

	err := errors.New("error process block")
	blProcMock.ProcessBlockCalled = func(*blockchain.BlockChain, *block.Header, *block.TxBlockBody, func() time.Duration) error {
		return err
	}

	wrk.SetBlockProcessor(blProcMock)

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.Header = hdr
	wrk.BlockBody = blk

	assert.False(t, wrk.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsInNextRound(t *testing.T) {
	wrk := initWorker()

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		-1,
	)

	wrk.Header = hdr
	wrk.BlockBody = blk

	assert.False(t, wrk.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsTooLate(t *testing.T) {
	wrk := initWorker()

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.Header = hdr
	wrk.BlockBody = blk

	endTime := getEndTime(wrk.SPoS.Chr, bn.SrEndRound)
	wrk.SPoS.Chr.SetClockOffset(time.Duration(endTime))

	assert.False(t, wrk.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnTrue(t *testing.T) {
	wrk := initWorker()

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(wrk.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	wrk.Header = hdr
	wrk.BlockBody = blk

	assert.True(t, wrk.ProcessReceivedBlock(cnsDta))
}

func TestHaveTime_ShouldReturnNegativeValue(t *testing.T) {
	wrk := initWorker()

	time.Sleep(wrk.SPoS.Chr.Round().TimeDuration())

	haveTime := func() time.Duration {
		chr := wrk.SPoS.Chr

		roundStartTime := chr.Round().TimeStamp()
		currentTime := chr.SyncTimer().CurrentTime(chr.ClockOffset())
		elapsedTime := currentTime.Sub(roundStartTime)
		haveTime := float64(chr.Round().TimeDuration())*float64(0.85) - float64(elapsedTime)

		return time.Duration(haveTime)
	}

	ret := haveTime()

	assert.True(t, ret < 0)
}

func TestWorker_DecodeBlockHeader(t *testing.T) {
	wrk := initWorker()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = wrk.SPoS.Chr.RoundTimeStamp()
	hdr.Signature = []byte(wrk.SPoS.SelfPubKey())

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	dcdHdr := wrk.DecodeBlockHeader(nil)

	assert.Nil(t, dcdHdr)

	dcdHdr = wrk.DecodeBlockHeader(message)

	assert.Equal(t, hdr, dcdHdr)
	assert.Equal(t, []byte(wrk.SPoS.SelfPubKey()), dcdHdr.Signature)
}

func TestWorker_CheckBlockConsensus(t *testing.T) {
	wrk := initWorker()

	wrk.SPoS.SetStatus(bn.SrBlock, spos.SsNotFinished)

	ok := wrk.CheckBlockConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, wrk.SPoS.Status(bn.SrBlock))

	wrk.SPoS.SetJobDone("B", bn.SrBlock, true)

	ok = wrk.CheckBlockConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, wrk.SPoS.Status(bn.SrBlock))
}

func TestWorker_IsBlockReceived(t *testing.T) {
	wrk := initWorker()

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBlock, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrBitmap, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrCommitment, false)
		wrk.SPoS.SetJobDone(wrk.SPoS.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := wrk.IsBlockReceived(1)
	assert.False(t, ok)

	wrk.SPoS.SetJobDone("A", bn.SrBlock, true)
	isJobDone, _ := wrk.SPoS.GetJobDone("A", bn.SrBlock)

	assert.True(t, isJobDone)

	ok = wrk.IsBlockReceived(1)
	assert.True(t, ok)

	ok = wrk.IsBlockReceived(2)
	assert.False(t, ok)
}

func TestWorker_DoExtendBlockShouldNotSetBlockExtended(t *testing.T) {
	wrk := initWorker()

	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return true
	}}

	wrk.SetBootstraper(bootMock)

	wrk.ExtendBlock()

	assert.NotEqual(t, spos.SsExtended, wrk.SPoS.Status(bn.SrBlock))
}

func TestWorker_DoExtendBlockShouldSetBlockExtended(t *testing.T) {
	wrk := initWorker()

	wrk.ExtendBlock()

	assert.Equal(t, spos.SsExtended, wrk.SPoS.Status(bn.SrBlock))
}

func TestWorker_ExtendBlock(t *testing.T) {
	wrk := initWorker()

	wrk.ExtendBlock()
	assert.Equal(t, spos.SsExtended, wrk.SPoS.Status(bn.SrBlock))
}

func TestWorker_CheckIfBlockIsValid(t *testing.T) {
	wrk := initWorker()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = wrk.SPoS.Chr.RoundTimeStamp()

	hdr.PrevHash = []byte("X")

	r := wrk.CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.PrevHash = []byte("")

	r = wrk.CheckIfBlockIsValid(hdr)
	assert.True(t, r)

	hdr.Nonce = 2

	r = wrk.CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 1
	wrk.BlockChain.CurrentBlockHeader = hdr

	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = wrk.SPoS.Chr.RoundTimeStamp()

	r = wrk.CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("X")

	r = wrk.CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 3
	hdr.PrevHash = []byte("")

	r = wrk.CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 2

	prevHeader, _ := mock.MarshalizerMock{}.Marshal(wrk.BlockChain.CurrentBlockHeader)
	hdr.PrevHash = mock.HasherMock{}.Compute(string(prevHeader))

	r = wrk.CheckIfBlockIsValid(hdr)
	assert.True(t, r)
}
