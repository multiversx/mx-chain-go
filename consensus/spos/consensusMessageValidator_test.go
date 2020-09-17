package spos_test

import (
	"crypto/rand"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/stretchr/testify/assert"
)

func TestCheckConsensusMessageValidity_WrongChainID(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{ChainID: wrongChainID}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidChainID))
}

func TestCheckMessageWithFinalInfoValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{Body: []byte("body")}
	err := wrk.CheckMessageWithFinalInfoValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithFinalInfoValidity_InvalidPubKeyBitmap(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{PubKeysBitmap: []byte("0")}
	err := wrk.CheckMessageWithFinalInfoValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidPublicKeyBitmapSize))
}

func TestCheckMessageWithFinalInfoValidity_InvalidAggregateSignatureSize(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{PubKeysBitmap: []byte("01"), AggregateSignature: []byte("0")}
	err := wrk.CheckMessageWithFinalInfoValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidSignatureSize))
}

func TestCheckMessageWithFinalInfoValidity_InvalidLeaderSignatureSize(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)
	cnsMsg := &consensus.Message{PubKeysBitmap: []byte("01"), AggregateSignature: sig, LeaderSignature: []byte("0")}
	err := wrk.CheckMessageWithFinalInfoValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidSignatureSize))
}

func TestCheckMessageWithFinalInfoValidity_ShouldWork(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)
	cnsMsg := &consensus.Message{PubKeysBitmap: []byte("01"), AggregateSignature: sig, LeaderSignature: sig}
	err := wrk.CheckMessageWithFinalInfoValidity(cnsMsg)
	assert.Nil(t, err)
}

func TestCheckMessageWithSignatureValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{Body: []byte("body")}
	err := wrk.CheckMessageWithSignatureValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithSignatureValidity_InvalidSignatureShareSize(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{SignatureShare: []byte("0")}
	err := wrk.CheckMessageWithSignatureValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidSignatureSize))
}

func TestCheckMessageWithSignatureShareValidity_ShouldWork(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)
	cnsMsg := &consensus.Message{SignatureShare: sig}
	err := wrk.CheckMessageWithSignatureValidity(cnsMsg)
	assert.Nil(t, err)
}

func TestCheckMessageWithBlockHeaderValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{SignatureShare: []byte("0")}
	err := wrk.CheckMessageWithBlockHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithBlockHeaderValidity_InvalidHeaderSize(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{}
	err := wrk.CheckMessageWithBlockHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidHeaderSize))
}

func TestCheckMessageWithBlockHeaderValidity_HeaderTooBig(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, core.MegabyteSize+1)
	_, _ = rand.Read(headerBytes)
	cnsMsg := &consensus.Message{Header: headerBytes}
	err := wrk.CheckMessageWithBlockHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidHeaderSize))
}

func TestCheckMessageWithBlockHeaderValidity_HeaderSizeZero(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 0)
	cnsMsg := &consensus.Message{Header: headerBytes}
	err := wrk.CheckMessageWithBlockHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidHeaderSize))
}

func TestCheckMessageWithBlockHeaderValidity_ShouldWork(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{Header: []byte("header")}
	err := wrk.CheckMessageWithBlockHeaderValidity(cnsMsg)
	assert.Nil(t, err)
}

func TestCheckMessageWithBlockBodyValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	cnsMsg := &consensus.Message{Header: headerBytes}
	err := wrk.CheckMessageWithBlockBodyValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithBlockBodyValidity_InvalidBodySize(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	bodyBytes := make([]byte, core.MegabyteSize+1)
	_, _ = rand.Read(bodyBytes)
	cnsMsg := &consensus.Message{Body: bodyBytes}
	err := wrk.CheckMessageWithBlockBodyValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidBodySize))
}

func TestCheckMessageWithBlockBodyValidity_ShouldWork(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	bodyBytes := make([]byte, 100)
	_, _ = rand.Read(bodyBytes)
	cnsMsg := &consensus.Message{Body: bodyBytes}
	err := wrk.CheckMessageWithBlockBodyValidity(cnsMsg)
	assert.Nil(t, err)
}

func TestCheckMessageWithBlockBodyAndHeaderValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{SignatureShare: []byte("0")}
	err := wrk.CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithBlockBodyAndHeaderValidity_InvalidBodySize(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	bodyBytes := make([]byte, core.MegabyteSize+1)
	_, _ = rand.Read(bodyBytes)
	cnsMsg := &consensus.Message{Body: bodyBytes}
	err := wrk.CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidBodySize))
}

func TestCheckMessageWithBlockBodyAndHeaderValidity_InvalidHeaderSize(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, core.MegabyteSize+1)
	_, _ = rand.Read(headerBytes)
	cnsMsg := &consensus.Message{Header: headerBytes}
	err := wrk.CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidHeaderSize))
}

func TestCheckMessageWithBlockBodyAndHeaderValidity_ShouldWork(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	cnsMsg := &consensus.Message{Header: headerBytes}
	err := wrk.CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	assert.Nil(t, err)
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithBlockBodyAndHeaderInvalid(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockBodyAndHeader), SignatureShare: []byte("1")}
	err := wrk.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithBlockBodyInvalid(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockBody), SignatureShare: []byte("1")}
	err := wrk.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithBlockHeaderInvalid(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockHeader), SignatureShare: []byte("1")}
	err := wrk.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithSignatureInvalid(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtSignature), Header: []byte("1")}
	err := wrk.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithFinalInfoInvalid(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockHeaderFinalInfo), Header: []byte("1")}
	err := wrk.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageUnknownInvalid(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtUnknown), Header: []byte("1")}
	err := wrk.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessageType))
}

func TestIsBlockHeaderHashSizeValid_NotValid(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockBody), BlockHeaderHash: []byte("hash")}
	result := wrk.IsBlockHeaderHashSizeValid(cnsMsg)
	assert.False(t, result)
}

func TestIsBlockHeaderHashSizeValid(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerHash := make([]byte, workerArgs.Hasher.Size())
	_, _ = rand.Read(headerHash)
	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockHeader), BlockHeaderHash: headerHash}
	result := wrk.IsBlockHeaderHashSizeValid(cnsMsg)
	assert.True(t, result)
}

func TestCheckConsensusMessageValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	cnsMsg := &consensus.Message{ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader), SignatureShare: []byte("1")}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidity_InvalidHeaderHashSize(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	cnsMsg := &consensus.Message{ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader), Header: headerBytes}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidHeaderHashSize))
}

func TestCheckConsensusMessageValidity_InvalidPublicKeySize(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, workerArgs.Hasher.Size())
	_, _ = rand.Read(headerHash)
	cnsMsg := &consensus.Message{ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader), Header: headerBytes, BlockHeaderHash: headerHash}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidPublicKeySize))
}

func TestCheckConsensusMessageValidity_InvalidSignatureSize(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, workerArgs.Hasher.Size())
	_, _ = rand.Read(headerHash)
	pubKey := make([]byte, PublicKeySize)
	_, _ = rand.Read(pubKey)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader), Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey,
	}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidSignatureSize))
}

func TestCheckConsensusMessageValidity_NodeIsNotEligible(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, workerArgs.Hasher.Size())
	_, _ = rand.Read(headerHash)
	pubKey := make([]byte, PublicKeySize)
	_, _ = rand.Read(pubKey)
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig,
	}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrNodeIsNotInEligibleList))
}

func TestCheckConsensusMessageValidity_ErrMessageForFutureRound(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, workerArgs.Hasher.Size())
	_, _ = rand.Read(headerHash)
	pubKey := []byte(workerArgs.ConsensusState.ConsensusGroup()[0])
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig, RoundIndex: 10,
	}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrMessageForFutureRound))
}

func TestCheckConsensusMessageValidity_ErrMessageForPastRound(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	workerArgs.ConsensusState.RoundIndex = 100
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, workerArgs.Hasher.Size())
	_, _ = rand.Read(headerHash)
	pubKey := []byte(workerArgs.ConsensusState.ConsensusGroup()[0])
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig, RoundIndex: 10,
	}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrMessageForPastRound))
}

func TestCheckConsensusMessageValidity_ErrMessageTypeLimitReached(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	workerArgs.ConsensusState.RoundIndex = 10
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, workerArgs.Hasher.Size())
	_, _ = rand.Read(headerHash)
	pubKey := []byte(workerArgs.ConsensusState.ConsensusGroup()[0])
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	wrk.AddMessageTypeToPublicKey(pubKey, bls.MtBlockBodyAndHeader)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig, RoundIndex: 10,
	}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrMessageTypeLimitReached))
}

func TestCheckConsensusMessageValidity_InvalidSignature(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local error")
	signer := &mock.SingleSignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return localErr
		},
	}
	workerArgs := createDefaultWorkerArgs()
	workerArgs.PeerSignatureHandler = &mock.PeerSignatureHandler{
		Signer: signer,
	}
	workerArgs.ConsensusState.RoundIndex = 10
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, workerArgs.Hasher.Size())
	_, _ = rand.Read(headerHash)
	pubKey := []byte(workerArgs.ConsensusState.ConsensusGroup()[0])
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig, RoundIndex: 10,
	}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidSignature))
}

func TestCheckConsensusMessageValidity_Ok(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	workerArgs.ConsensusState.RoundIndex = 10
	wrk, _ := spos.NewWorker(workerArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, workerArgs.Hasher.Size())
	_, _ = rand.Read(headerHash)
	pubKey := []byte(workerArgs.ConsensusState.ConsensusGroup()[0])
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig, RoundIndex: 10,
	}
	err := wrk.CheckConsensusMessageValidity(cnsMsg, "")
	assert.Nil(t, err)
}

func TestIsMessageTypeLimitReached_ShouldWork(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	assert.False(t, wrk.IsMessageTypeLimitReached([]byte("pk1"), bls.MtBlockBody))

	wrk.AddMessageTypeToPublicKey([]byte("pk1"), bls.MtBlockHeader)

	assert.False(t, wrk.IsMessageTypeLimitReached([]byte("pk1"), bls.MtBlockBody))
	assert.True(t, wrk.IsMessageTypeLimitReached([]byte("pk1"), bls.MtBlockHeader))
}

func TestAddMessageTypeToPublicKey_ShouldWork(t *testing.T) {
	t.Parallel()

	workerArgs := createDefaultWorkerArgs()
	wrk, _ := spos.NewWorker(workerArgs)

	assert.Equal(t, uint32(0), wrk.GetNumOfMessageTypeForPublicKey([]byte("pk1"), bls.MtBlockBody))

	wrk.AddMessageTypeToPublicKey([]byte("pk1"), bls.MtBlockHeader)

	assert.Equal(t, uint32(0), wrk.GetNumOfMessageTypeForPublicKey([]byte("pk1"), bls.MtBlockBody))

	wrk.AddMessageTypeToPublicKey([]byte("pk1"), bls.MtBlockBody)
	wrk.AddMessageTypeToPublicKey([]byte("pk1"), bls.MtBlockHeader)

	assert.Equal(t, uint32(1), wrk.GetNumOfMessageTypeForPublicKey([]byte("pk1"), bls.MtBlockBody))
	assert.Equal(t, uint32(2), wrk.GetNumOfMessageTypeForPublicKey([]byte("pk1"), bls.MtBlockHeader))

	wrk.AddMessageTypeToPublicKey([]byte("pk2"), bls.MtBlockHeaderFinalInfo)

	assert.Equal(t, uint32(0), wrk.GetNumOfMessageTypeForPublicKey([]byte("pk1"), bls.MtBlockHeaderFinalInfo))
	assert.Equal(t, uint32(1), wrk.GetNumOfMessageTypeForPublicKey([]byte("pk2"), bls.MtBlockHeaderFinalInfo))
}
