package spos_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/stretchr/testify/assert"
)

func TestInitReceivedMessages_ShouldInitMap(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.ReceivedMessages = nil

	cnWorker.InitReceivedMessages()

	assert.NotNil(t, cnWorker.ReceivedMessages[spos.MtBlockBody])
}

func TestCleanReceivedMessages_ShouldCleanList(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		cnWorker.GetRoundTime(),
		-1,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.CleanReceivedMessages()

	assert.Equal(t, 0, len(cnWorker.ReceivedMessages[spos.MtBlockBody]))
}

func TestExecuteMessages_ShouldNotExecuteWhenConsensusDataIsNil(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		cnWorker.GetRoundTime(),
		0,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, nil)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.ExecuteMessage(cnsDataList)

	assert.Nil(t, cnWorker.ReceivedMessages[cnsDta.MsgType][0])
}

func TestExecuteMessages_ShouldNotExecuteWhenShouldDropConsensusMessage(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		cnWorker.GetRoundTime(),
		-1,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, cnWorker.ReceivedMessages[cnsDta.MsgType][0])
}

func TestExecuteMessages_ShouldNotExecuteWhenShouldSync(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		cnWorker.GetRoundTime(),
		0,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, cnWorker.ReceivedMessages[cnsDta.MsgType][0])
}

func TestExecuteBlockBodyMessages_ShouldNotExecuteWhenStartRoundIsNotFinished(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		cnWorker.GetRoundTime(),
		0,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, cnWorker.ReceivedMessages[cnsDta.MsgType][0])
}

func TestExecuteBlockHeaderMessages_ShouldNotExecuteWhenStartRoundIsNotFinished(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtBlockHeader,
		cnWorker.GetRoundTime(),
		0,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, cnWorker.ReceivedMessages[cnsDta.MsgType][0])
}

func TestExecuteCommitmentHashMessages_ShouldNotExecuteWhenBlockIsNotFinished(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtCommitmentHash,
		cnWorker.GetRoundTime(),
		0,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, cnWorker.ReceivedMessages[cnsDta.MsgType][0])
}

func TestExecuteBitmapMessages_ShouldNotExecuteWhenBlockIsNotFinished(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtBitmap,
		cnWorker.GetRoundTime(),
		0,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, cnWorker.ReceivedMessages[cnsDta.MsgType][0])
}

func TestExecuteCommitmentMessages_ShouldNotExecuteWhenBitmapIsNotFinished(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtCommitment,
		cnWorker.GetRoundTime(),
		0,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, cnWorker.ReceivedMessages[cnsDta.MsgType][0])
}

func TestExecuteSignatureMessages_ShouldNotExecuteWhenBitmapIsNotFinished(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtSignature,
		cnWorker.GetRoundTime(),
		0,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, cnWorker.ReceivedMessages[cnsDta.MsgType][0])
}

func TestExecuteMessages_ShouldExecute(t *testing.T) {
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
	consensusGroup := CreateConsensusGroup(consensusGroupSize)

	cns := initConsensus(
		genesisTime,
		roundDuration,
		consensusGroup,
		consensusGroupSize,
		0,
	)

	cnWorker, _ := spos.NewConsensusWorker(
		cns,
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

	cnWorker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorker.Cns.ConsensusGroup()[1]),
		[]byte("sig"),
		spos.MtBlockBody,
		cnWorker.GetRoundTime(),
		0,
	)

	cnsDataList := cnWorker.ReceivedMessages[cnsDta.MsgType]
	cnsDataList = append(cnsDataList, cnsDta)
	cnWorker.ReceivedMessages[cnsDta.MsgType] = cnsDataList

	cnWorker.Cns.SetStatus(spos.SrStartRound, spos.SsFinished)

	cnWorker.ExecuteMessage(cnsDataList)

	assert.Nil(t, cnWorker.ReceivedMessages[cnsDta.MsgType][0])
}
