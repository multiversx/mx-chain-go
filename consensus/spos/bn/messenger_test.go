package bn_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestWorker_InitReceivedMessagesShouldInitMap(t *testing.T) {
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

	worker.ReceivedMessages = nil

	worker.InitReceivedMessages()

	assert.NotNil(t, worker.ReceivedMessages[bn.MtBlockBody])
}

func TestWorker_CleanReceivedMessagesShouldCleanList(t *testing.T) {
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

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		-1,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.CleanReceivedMessages()

	assert.Equal(t, 0, len(worker.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ShouldDropConsensusMessageShouldReturnTrueWhenMessageReceivedIsFromThePastRounds(t *testing.T) {
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
		-1,
	)

	assert.True(t, worker.ShouldDropConsensusMessage(cnsDta))
}

func TestWorker_ShouldDropConsensusMessageShouldReturnTrueWhenMessageIsReceivedAfterEndRound(t *testing.T) {
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

	endTime := getEndTime(worker.SPoS.Chr, bn.SrEndRound)
	worker.SPoS.Chr.SetClockOffset(time.Duration(endTime))

	assert.True(t, worker.ShouldDropConsensusMessage(cnsDta))
}

func TestWorker_ShouldDropConsensusMessageShouldReturnFalse(t *testing.T) {
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

	assert.False(t, worker.ShouldDropConsensusMessage(cnsDta))
}

func TestWorker_ReceivedMessageTxBlockBody(t *testing.T) {
	cnWorkers := initSposWorkers()

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorkers[0].SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].ReceivedMessage(string(consensusTopic), cnsDta, nil)
}

func TestWorker_ReceivedMessageUnknown(t *testing.T) {
	cnWorkers := initSposWorkers()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].SPoS.Chr.RoundTimeStamp()

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)
	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))
	message, _ = mock.MarshalizerMock{}.Marshal(hdr)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorkers[0].SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtUnknown),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].ReceivedMessage(string(consensusTopic), cnsDta, nil)
}

func TestWorker_ReceivedMessageShouldReturnWhenIsCanceled(t *testing.T) {
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
		[]byte(worker.SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	worker.SPoS.Chr.SetSelfSubround(-1)
	worker.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	assert.Equal(t, 0, len(worker.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldReturnWhenDataReceivedIsInvalid(t *testing.T) {
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

	worker.ReceivedMessage(string(consensusTopic), nil, nil)

	assert.Equal(t, 0, len(worker.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldReturnWhenNodeIsNotInTheConsensusGroup(t *testing.T) {
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
		[]byte("X"),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	worker.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	assert.Equal(t, 0, len(worker.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldReturnWhenShouldDropMessage(t *testing.T) {
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
		[]byte(worker.SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		-1,
	)

	worker.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	assert.Equal(t, 0, len(worker.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldReturnWhenReceivedMessageIsFromSelf(t *testing.T) {
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
		[]byte(worker.SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	worker.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	assert.Equal(t, 0, len(worker.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldReturnWhenSignatureIsInvalid(t *testing.T) {
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
		nil,
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	worker.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	assert.Equal(t, 0, len(worker.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_ReceivedMessageShouldSendReceivedMesageOnChannel(t *testing.T) {
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

	worker.ReceivedMessage(string(consensusTopic), cnsDta, nil)

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, len(worker.ReceivedMessages[bn.MtBlockBody]))
}

func TestWorker_CheckSignatureShouldReturnErrNilConsensusData(t *testing.T) {
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

	err := worker.CheckSignature(nil)

	assert.Equal(t, spos.ErrNilConsensusData, err)
}

func TestWorker_CheckSignatureShouldReturnErrNilPublicKey(t *testing.T) {
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
		nil,
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	err := worker.CheckSignature(cnsDta)

	assert.Equal(t, spos.ErrNilPublicKey, err)
}

func TestWorker_CheckSignatureShouldReturnErrNilSignature(t *testing.T) {
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
		nil,
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	err := worker.CheckSignature(cnsDta)

	assert.Equal(t, spos.ErrNilSignature, err)
}

func TestWorker_CheckSignatureShouldReturnPublicKeyFromByteArrayErr(t *testing.T) {
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

	err := errors.New("error public key from byte array")
	keyGenMock.PublicKeyFromByteArrayMock = func(b []byte) (crypto.PublicKey, error) {
		return nil, err
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

	err2 := worker.CheckSignature(cnsDta)

	assert.Equal(t, err, err2)
}

func TestWorker_CheckSignatureShouldReturnNilErr(t *testing.T) {
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

	err := worker.CheckSignature(cnsDta)

	assert.Nil(t, err)
}

func TestWorker_ExecuteMessagesShouldNotExecuteWhenConsensusDataIsNil(t *testing.T) {
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

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, nil)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.ExecuteMessage(cnsDataList)

	assert.Nil(t, worker.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteMessagesShouldNotExecuteWhenShouldDropConsensusMessage(t *testing.T) {
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

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		-1,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, worker.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteMessagesShouldNotExecuteWhenShouldSync(t *testing.T) {
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

	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, worker.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteBlockBodyMessagesShouldNotExecuteWhenStartRoundIsNotFinished(t *testing.T) {
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

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, worker.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteBlockHeaderMessagesShouldNotExecuteWhenStartRoundIsNotFinished(t *testing.T) {
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

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, worker.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteCommitmentHashMessagesShouldNotExecuteWhenBlockIsNotFinished(t *testing.T) {
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

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, worker.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteBitmapMessagesShouldNotExecuteWhenBlockIsNotFinished(t *testing.T) {
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

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtBitmap),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, worker.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteCommitmentMessagesShouldNotExecuteWhenBitmapIsNotFinished(t *testing.T) {
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

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[1]),
		[]byte("sig"),
		int(bn.MtCommitment),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, worker.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteSignatureMessagesShouldNotExecuteWhenBitmapIsNotFinished(t *testing.T) {
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

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtSignature),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.ExecuteMessage(cnsDataList)

	assert.NotNil(t, worker.ReceivedMessages[msgType][0])
}

func TestWorker_ExecuteMessagesShouldExecute(t *testing.T) {
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

	worker.InitReceivedMessages()

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(worker.SPoS.ConsensusGroup()[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		worker.SPoS.Chr.RoundTimeStamp(),
		0,
	)

	msgType := bn.MessageType(cnsDta.MsgType)

	cnsDataList := worker.ReceivedMessages[msgType]
	cnsDataList = append(cnsDataList, cnsDta)
	worker.ReceivedMessages[msgType] = cnsDataList

	worker.SPoS.SetStatus(bn.SrStartRound, spos.SsFinished)

	worker.ExecuteMessage(cnsDataList)

	assert.Nil(t, worker.ReceivedMessages[msgType][0])
}

func TestWorker_CheckChannelTxBlockBody(t *testing.T) {
	cnWorkers := initSposWorkers()
	cnWorkers[0].Header = nil
	rnd := cnWorkers[0].SPoS.Chr.Round()
	roundDuration := rnd.TimeDuration()
	rnd.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	cnsGroup := cnWorkers[0].SPoS.ConsensusGroup()
	rcns := cnWorkers[0].SPoS.RoundConsensus

	// BLOCK BODY
	blk := &block.TxBlockBody{}
	message, _ := mock.MarshalizerMock{}.Marshal(blk)

	cnsDta := spos.NewConsensusData(
		nil,
		message,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].MessageChannels[bn.MtBlockBody] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	isBlockJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrBlock)

	assert.Nil(t, err)
	// Not done since header is missing
	assert.False(t, isBlockJobDone)
}

func TestWorker_CheckChannelBlockHeader(t *testing.T) {
	cnWorkers := initSposWorkers()

	rnd := cnWorkers[0].SPoS.Chr.Round()
	roundDuration := rnd.TimeDuration()
	rnd.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	cnsGroup := cnWorkers[0].SPoS.ConsensusGroup()
	rcns := cnWorkers[0].SPoS.RoundConsensus

	// BLOCK HEADER
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].SPoS.Chr.RoundTimeStamp()

	message, _ := mock.MarshalizerMock{}.Marshal(hdr)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	cnsDta := spos.NewConsensusData(
		nil,
		message,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtBlockBody),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		1,
	)

	cnWorkers[0].MessageChannels[bn.MtBlockBody] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	cnsDta.MsgType = int(bn.MtBlockHeader)
	cnWorkers[0].MessageChannels[bn.MtBlockHeader] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	isBlockJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrBlock)

	assert.Nil(t, err)
	assert.True(t, isBlockJobDone)
}

func TestWorker_CheckChannelsCommitmentHash(t *testing.T) {
	cnWorkers := initSposWorkers()

	rnd := cnWorkers[0].SPoS.Chr.Round()
	roundDuration := rnd.TimeDuration()
	rnd.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	rcns := cnWorkers[0].SPoS.RoundConsensus
	cnsGroup := cnWorkers[0].SPoS.ConsensusGroup()

	commitmentHash := []byte("commitmentHash")

	// COMMITMENT_HASH
	cnsDta := spos.NewConsensusData(
		cnWorkers[0].SPoS.Data,
		commitmentHash,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].MessageChannels[bn.MtCommitmentHash] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	isCommitmentHashJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrCommitmentHash)

	assert.Nil(t, err)
	assert.True(t, isCommitmentHashJobDone)
}

func TestWorker_CheckChannelsBitmap(t *testing.T) {
	cnWorkers := initSposWorkers()

	rnd := cnWorkers[0].SPoS.Chr.Round()
	roundDuration := rnd.TimeDuration()
	rnd.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	rcns := cnWorkers[0].SPoS.RoundConsensus
	cnsGroup := cnWorkers[0].SPoS.ConsensusGroup()

	bitmap := make([]byte, len(cnsGroup)/8+1)

	for i := 0; i < len(cnsGroup); i++ {
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	// BITMAP
	cnsDta := spos.NewConsensusData(
		cnWorkers[0].SPoS.Data,
		bitmap,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtBitmap),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].MessageChannels[bn.MtBitmap] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < len(cnsGroup); i++ {
		isBitmapJobDone, err := rcns.GetJobDone(cnsGroup[i], bn.SrBitmap)
		assert.Nil(t, err)
		assert.True(t, isBitmapJobDone)
	}
}

func TestWorker_CheckChannelsCommitment(t *testing.T) {
	cnWorkers := initSposWorkers()

	rnd := cnWorkers[0].SPoS.Chr.Round()
	roundDuration := rnd.TimeDuration()
	rnd.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	rcns := cnWorkers[0].SPoS.RoundConsensus
	cnsGroup := cnWorkers[0].SPoS.ConsensusGroup()

	bitmap := make([]byte, len(cnsGroup)/8+1)

	for i := 0; i < len(cnsGroup); i++ {
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	commHash := mock.HasherMock{}.Compute("commitment")

	// Commitment Hash
	cnsDta := spos.NewConsensusData(
		cnWorkers[0].SPoS.Data,
		commHash,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtCommitmentHash),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].MessageChannels[bn.MtCommitmentHash] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	// Bitmap
	cnsDta.MsgType = int(bn.MtBitmap)
	cnsDta.SubRoundData = bitmap
	cnWorkers[0].MessageChannels[bn.MtBitmap] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	// Commitment
	cnsDta.MsgType = int(bn.MtCommitment)
	cnsDta.SubRoundData = []byte("commitment")
	cnWorkers[0].MessageChannels[bn.MtCommitment] <- cnsDta
	time.Sleep(10 * time.Millisecond)

	isCommitmentJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrCommitment)

	assert.Nil(t, err)
	assert.True(t, isCommitmentJobDone)
}

func TestWorker_CheckChannelsSignature(t *testing.T) {
	cnWorkers := initSposWorkers()

	rnd := cnWorkers[0].SPoS.Chr.Round()
	roundDuration := rnd.TimeDuration()
	rnd.UpdateRound(time.Now(), time.Now().Add(roundDuration))

	rcns := cnWorkers[0].SPoS.RoundConsensus
	cnsGroup := cnWorkers[0].SPoS.ConsensusGroup()

	bitmap := make([]byte, len(cnsGroup)/8+1)

	for i := 0; i < len(cnsGroup); i++ {
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	// Bitmap
	cnsDta := spos.NewConsensusData(
		cnWorkers[0].SPoS.Data,
		bitmap,
		[]byte(cnsGroup[0]),
		[]byte("sig"),
		int(bn.MtBitmap),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].MessageChannels[bn.MtBitmap] <- cnsDta

	// Signature
	cnsDta.MsgType = int(bn.MtSignature)
	cnsDta.SubRoundData = []byte("signature")
	time.Sleep(10 * time.Millisecond)

	cnWorkers[0].MessageChannels[bn.MtSignature] <- cnsDta
	time.Sleep(10 * time.Millisecond)
	isSigJobDone, err := rcns.GetJobDone(cnsGroup[0], bn.SrSignature)

	assert.Nil(t, err)
	assert.True(t, isSigJobDone)
}

func TestWorker_SendConsensusMessage(t *testing.T) {
	cnWorkers := initSposWorkers()

	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.TimeStamp = cnWorkers[0].SPoS.Chr.RoundTimeStamp()

	message, err := mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	hdr.BlockBodyHash = mock.HasherMock{}.Compute(string(message))

	message, err = mock.MarshalizerMock{}.Marshal(hdr)

	assert.Nil(t, err)

	cnsDta := spos.NewConsensusData(
		message,
		nil,
		[]byte(cnWorkers[0].SPoS.SelfPubKey()),
		[]byte("sig"),
		int(bn.MtBlockHeader),
		cnWorkers[0].SPoS.Chr.RoundTimeStamp(),
		0,
	)

	cnWorkers[0].SendMessage = nil
	r := cnWorkers[0].SendConsensusMessage(cnsDta)
	assert.False(t, r)

	cnWorkers[0].SendMessage = SendMessage
	r = cnWorkers[0].SendConsensusMessage(cnsDta)

	assert.True(t, r)
}
