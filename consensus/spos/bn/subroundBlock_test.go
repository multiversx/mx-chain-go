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
	cnWorkers := initSposWorkers()

	cnWorkers[0].SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].SPoS.Chr.Round().TimeDuration()))

	r := cnWorkers[0].DoBlockJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.Chr.Round().UpdateRound(time.Now(), time.Now())
	cnWorkers[0].SPoS.SetStatus(bn.SrBlock, spos.SsFinished)

	r = cnWorkers[0].DoBlockJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetStatus(bn.SrBlock, spos.SsNotFinished)
	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrBlock, true)

	r = cnWorkers[0].DoBlockJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrBlock, false)
	cnWorkers[0].SPoS.RoundConsensus.SetSelfPubKey(cnWorkers[0].SPoS.RoundConsensus.ConsensusGroup()[1])

	r = cnWorkers[0].DoBlockJob()
	assert.False(t, r)

	cnWorkers[0].SPoS.RoundConsensus.SetSelfPubKey(cnWorkers[0].SPoS.RoundConsensus.ConsensusGroup()[0])

	r = cnWorkers[0].DoBlockJob()
	assert.True(t, r)
	assert.Equal(t, uint64(1), cnWorkers[0].Header.Nonce)

	cnWorkers[0].SPoS.SetJobDone(cnWorkers[0].SPoS.SelfPubKey(), bn.SrBlock, false)
	cnWorkers[0].BlockChain.CurrentBlockHeader = cnWorkers[0].Header

	r = cnWorkers[0].DoBlockJob()
	assert.True(t, r)
	assert.Equal(t, uint64(2), cnWorkers[0].Header.Nonce)
}

func TestWorker_ReceivedBlock(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].SPoS.Chr.Round().TimeDuration()))

	blBody := &block.TxBlockBody{}

	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)

	cnsDta := spos.NewConsensusData(
		nil,
		blBodyStr,
		[]byte(cnWorkers[0].SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].BlockBody = &block.TxBlockBody{}

	r := cnWorkers[0].ReceivedBlockBody(cnsDta)
	assert.False(t, r)

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(blBodyStr))

	hdrStr, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdrHash := mock.HasherMock{}.Compute(string(hdrStr))

	cnsDta = spos.NewConsensusData(
		hdrHash,
		hdrStr,
		[]byte(cnWorkers[0].SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		1,
	)

	cnWorkers[0].Header = nil
	cnWorkers[0].SPoS.Data = nil
	r = cnWorkers[0].ReceivedBlockHeader(cnsDta)
	assert.True(t, r)
}

func TestWorker_ReceivedBlockBodyShouldSetJobDone(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].SPoS.Chr.Round().TimeDuration()))

	blBody := &block.TxBlockBody{}

	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)

	cnsDta := spos.NewConsensusData(
		nil,
		blBodyStr,
		[]byte(cnWorkers[0].SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		1,
	)

	cnWorkers[0].Header = &block.Header{}

	r := cnWorkers[0].ReceivedBlockBody(cnsDta)
	assert.True(t, r)
}

func TestWorker_ReceivedBlockBodyShouldErrProcessBlock(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].SPoS.Chr.Round().UpdateRound(time.Now(), time.Now().Add(cnWorkers[0].SPoS.Chr.Round().TimeDuration()))

	blBody := &block.TxBlockBody{}

	blBodyStr, _ := mock.MarshalizerMock{}.Marshal(blBody)

	cnsDta := spos.NewConsensusData(
		nil,
		blBodyStr,
		[]byte(cnWorkers[0].SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].Header = &block.Header{}

	blProcMock := initMockBlockProcessor()

	blProcMock.ProcessBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
		return process.ErrNilPreviousBlockHash
	}

	cnWorkers[0].BlockProcessor = blProcMock

	r := cnWorkers[0].ReceivedBlockBody(cnsDta)
	assert.False(t, r)
}

func TestWorker_DecodeBlockBody(t *testing.T) {
	cnWorkers := initSposWorkers()

	blk := &block.TxBlockBody{}

	mblks := make([]block.MiniBlock, 0)
	mblks = append(mblks, block.MiniBlock{ShardID: 69})
	blk.MiniBlocks = mblks

	message, err := mock.MarshalizerMock{}.Marshal(blk)

	assert.Nil(t, err)

	dcdBlk := cnWorkers[0].DecodeBlockBody(nil)

	assert.Nil(t, dcdBlk)

	dcdBlk = cnWorkers[0].DecodeBlockBody(message)

	assert.Equal(t, blk, dcdBlk)
	assert.Equal(t, uint32(69), dcdBlk.MiniBlocks[0].ShardID)
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenBodyAndHeaderAreNotSet(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	worker, _ := bn.NewWorker(
		sPoS,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		bootMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	assert.False(t, worker.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockFails(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	err := errors.New("error process block")
	blProcMock.ProcessBlockCalled = func(*blockchain.BlockChain, *block.Header, *block.TxBlockBody, func() time.Duration) error {
		return err
	}

	worker, _ := bn.NewWorker(
		sPoS,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		bootMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	worker.Header = hdr
	worker.BlockBody = blk

	assert.False(t, worker.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsInNextRound(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	worker, _ := bn.NewWorker(
		sPoS,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		bootMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		-1,
	)

	worker.Header = hdr
	worker.BlockBody = blk

	assert.False(t, worker.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnFalseWhenProcessBlockReturnsTooLate(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	worker, _ := bn.NewWorker(
		sPoS,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		bootMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	worker.Header = hdr
	worker.BlockBody = blk

	endTime := getEndTime(worker.SPoS.Chr, bn.SrEndRound)
	worker.SPoS.Chr.SetClockOffset(time.Duration(endTime))

	assert.False(t, worker.ProcessReceivedBlock(cnsDta))
}

func TestWorker_ProcessReceivedBlockShouldReturnTrue(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	worker, _ := bn.NewWorker(
		sPoS,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		bootMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	hdr := &block.Header{}
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	worker.Header = hdr
	worker.BlockBody = blk

	assert.True(t, worker.ProcessReceivedBlock(cnsDta))
}

func TestHaveTime_ShouldReturnNegativeValue(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	worker, _ := bn.NewWorker(
		sPoS,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		bootMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	time.Sleep(roundDuration)

	haveTime := func() time.Duration {
		chr := worker.SPoS.Chr

		roundStartTime := chr.Round().TimeStamp()
		currentTime := chr.SyncTime().CurrentTime(chr.ClockOffset())
		elapsedTime := currentTime.Sub(roundStartTime)
		haveTime := float64(chr.Round().TimeDuration())*float64(0.85) - float64(elapsedTime)

		return time.Duration(haveTime)
	}

	ret := haveTime()

	assert.True(t, ret < 0)
}

func TestWorker_DecodeBlockHeader(t *testing.T) {
	cnWorkers := initSposWorkers()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].SPoS.Chr.RoundTimeStamp()
	hdr.Signature = []byte(cnWorkers[0].SPoS.SelfPubKey())

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	dcdHdr := cnWorkers[0].DecodeBlockHeader(nil)

	assert.Nil(t, dcdHdr)

	dcdHdr = cnWorkers[0].DecodeBlockHeader(message)

	assert.Equal(t, hdr, dcdHdr)
	assert.Equal(t, []byte(cnWorkers[0].SPoS.SelfPubKey()), dcdHdr.Signature)
}

func TestWorker_CheckBlockConsensus(t *testing.T) {
	sPoS := InitSposWorker()

	worker, _ := bn.NewWorker(
		sPoS,
		&blockchain.BlockChain{},
		mock.HasherMock{},
		mock.MarshalizerMock{},
		&mock.BlockProcessorMock{},
		&mock.BootstrapMock{ShouldSyncCalled: func() bool {
			return false
		}},
		&mock.BelNevMock{},
		&mock.KeyGenMock{},
		&mock.PrivateKeyMock{},
		&mock.PublicKeyMock{})

	GenerateSubRoundHandlers(100*time.Millisecond, sPoS, worker)

	sPoS.SetStatus(bn.SrBlock, spos.SsNotFinished)

	ok := worker.CheckBlockConsensus()
	assert.False(t, ok)
	assert.Equal(t, spos.SsNotFinished, sPoS.Status(bn.SrBlock))

	sPoS.SetJobDone("2", bn.SrBlock, true)

	ok = worker.CheckBlockConsensus()
	assert.True(t, ok)
	assert.Equal(t, spos.SsFinished, sPoS.Status(bn.SrBlock))
}

func TestWorker_IsBlockReceived(t *testing.T) {
	sPoS := InitSposWorker()

	worker, _ := bn.NewWorker(
		sPoS,
		&blockchain.BlockChain{},
		mock.HasherMock{},
		mock.MarshalizerMock{},
		&mock.BlockProcessorMock{},
		&mock.BootstrapMock{ShouldSyncCalled: func() bool {
			return false
		}},
		&mock.BelNevMock{},
		&mock.KeyGenMock{},
		&mock.PrivateKeyMock{},
		&mock.PublicKeyMock{})

	GenerateSubRoundHandlers(100*time.Millisecond, sPoS, worker)

	for i := 0; i < len(worker.SPoS.ConsensusGroup()); i++ {
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrBlock, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrCommitmentHash, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrBitmap, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrCommitment, false)
		worker.SPoS.SetJobDone(worker.SPoS.ConsensusGroup()[i], bn.SrSignature, false)
	}

	ok := worker.IsBlockReceived(1)
	assert.False(t, ok)

	worker.SPoS.SetJobDone("1", bn.SrBlock, true)
	isJobDone, _ := worker.SPoS.GetJobDone("1", bn.SrBlock)

	assert.True(t, isJobDone)

	ok = worker.IsBlockReceived(1)
	assert.True(t, ok)

	ok = worker.IsBlockReceived(2)
	assert.False(t, ok)
}

func TestWorker_DoExtendBlockShouldNotSetBlockExtended(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return true
	}}

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	worker, _ := bn.NewWorker(
		sPoS,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		bootMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	worker.ExtendBlock()

	assert.NotEqual(t, spos.SsExtended, worker.SPoS.Status(bn.SrBlock))
}

func TestWorker_DoExtendBlockShouldSetBlockExtended(t *testing.T) {
	blkc := blockchain.BlockChain{}
	keyGenMock, privKeyMock, pubKeyMock := initSingleSigning()
	multisigner := initMultisigner()
	blProcMock := initMockBlockProcessor()
	bootMock := &mock.BootstrapMock{ShouldSyncCalled: func() bool {
		return false
	}}

	consensusGroupSize := 9
	roundDuration := 100 * time.Millisecond
	genesisTime := time.Now()
	// create consensus group list
	consensusGroup := CreateEligibleList(consensusGroupSize)

	sPoS := initSpos(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	worker, _ := bn.NewWorker(
		sPoS,
		&blkc,
		mock.HasherMock{},
		mock.MarshalizerMock{},
		blProcMock,
		bootMock,
		multisigner,
		keyGenMock,
		privKeyMock,
		pubKeyMock,
	)

	worker.ExtendBlock()

	assert.Equal(t, spos.SsExtended, worker.SPoS.Status(bn.SrBlock))
}

func TestWorker_ExtendBlock(t *testing.T) {
	cnWorkers := initSposWorkers()

	cnWorkers[0].ExtendBlock()
	assert.Equal(t, spos.SsExtended, cnWorkers[0].SPoS.Status(bn.SrBlock))
}

func TestWorker_CheckIfBlockIsValid(t *testing.T) {
	cnWorkers := initSposWorkers()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].SPoS.Chr.RoundTimeStamp()

	hdr.PrevHash = []byte("X")

	r := cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.PrevHash = []byte("")

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.True(t, r)

	hdr.Nonce = 2

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 1
	cnWorkers[0].BlockChain.CurrentBlockHeader = hdr

	hdr = &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].SPoS.Chr.RoundTimeStamp()

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("X")

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 3
	hdr.PrevHash = []byte("")

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.False(t, r)

	hdr.Nonce = 2

	prevHeader, _ := mock.MarshalizerMock{}.Marshal(cnWorkers[0].BlockChain.CurrentBlockHeader)
	hdr.PrevHash = mock.HasherMock{}.Compute(string(prevHeader))

	r = cnWorkers[0].CheckIfBlockIsValid(hdr)
	assert.True(t, r)
}
