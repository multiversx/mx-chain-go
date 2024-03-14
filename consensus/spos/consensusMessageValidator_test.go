package spos_test

import (
	"crypto/rand"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/assert"
)

func createDefaultConsensusMessageValidatorArgs() spos.ArgsConsensusMessageValidator {
	consensusState := initConsensusState()
	blsService, _ := bls.NewConsensusService()
	singleSignerMock := &mock.SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return []byte("signed"), nil
		},
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return nil
		},
	}
	keyGeneratorMock, _, _ := mock.InitKeys()
	peerSigHandler := &mock.PeerSignatureHandler{Signer: singleSignerMock, KeyGen: keyGeneratorMock}
	hasher := &hashingMocks.HasherMock{}

	return spos.ArgsConsensusMessageValidator{
		ConsensusState:       consensusState,
		ConsensusService:     blsService,
		PeerSignatureHandler: peerSigHandler,
		EnableEpochsHandler:  enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		Marshaller:           &marshallerMock.MarshalizerStub{},
		ShardCoordinator:     &testscommon.ShardsCoordinatorMock{},
		SignatureSize:        SignatureSize,
		PublicKeySize:        PublicKeySize,
		HeaderHashSize:       hasher.Size(),
		ChainID:              chainID,
	}
}

func TestNewConsensusMessageValidator(t *testing.T) {
	t.Parallel()

	t.Run("nil ConsensusService", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.ConsensusService = nil
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrNilConsensusService, err)
	})
	t.Run("nil PeerSignatureHandler", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.PeerSignatureHandler = nil
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrNilPeerSignatureHandler, err)
	})
	t.Run("nil EnableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.EnableEpochsHandler = nil
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrNilEnableEpochsHandler, err)
	})
	t.Run("nil Marshaller", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.Marshaller = nil
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrNilMarshalizer, err)
	})
	t.Run("nil ShardCoordinator", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.ShardCoordinator = nil
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrNilShardCoordinator, err)
	})
	t.Run("nil ConsensusState", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.ConsensusState = nil
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrNilConsensusState, err)
	})
	t.Run("nil chain ID", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.ChainID = nil
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrInvalidChainID, err)
	})
	t.Run("empty chain ID", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.ChainID = make([]byte, 0)
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrInvalidChainID, err)
	})
	t.Run("invalid header hash size", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.HeaderHashSize = 0
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrInvalidHeaderHashSize, err)
	})
	t.Run("invalid public key size", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.PublicKeySize = 0
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrInvalidPublicKeySize, err)
	})
	t.Run("invalid signature size", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		args.SignatureSize = 0
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.Nil(t, validator)
		assert.Equal(t, spos.ErrInvalidSignatureSize, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createDefaultConsensusMessageValidatorArgs()
		validator, err := spos.NewConsensusMessageValidator(args)

		assert.NotNil(t, validator)
		assert.Nil(t, err)
	})
}

func TestCheckConsensusMessageValidity_WrongChainID(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{ChainID: wrongChainID}
	err := cmv.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidChainID))
}

func TestCheckMessageWithFinalInfoValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{Body: []byte("body")}
	err := cmv.CheckMessageWithFinalInfoValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithFinalInfoValidity_InvalidPubKeyBitmap(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{PubKeysBitmap: []byte("0")}
	err := cmv.CheckMessageWithFinalInfoValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidPublicKeyBitmapSize))
}

func TestCheckMessageWithFinalInfoValidity_InvalidAggregateSignatureSize(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{PubKeysBitmap: []byte("01"), AggregateSignature: []byte("0")}
	err := cmv.CheckMessageWithFinalInfoValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidSignatureSize))
}

func TestCheckMessageWithFinalInfoValidity_InvalidLeaderSignatureSize(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)
	cnsMsg := &consensus.Message{PubKeysBitmap: []byte("01"), AggregateSignature: sig, LeaderSignature: []byte("0")}
	err := cmv.CheckMessageWithFinalInfoValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidSignatureSize))
}

func TestCheckMessageWithFinalInfoValidity_ShouldWork(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)
	cnsMsg := &consensus.Message{PubKeysBitmap: []byte("01"), AggregateSignature: sig, LeaderSignature: sig}
	err := cmv.CheckMessageWithFinalInfoValidity(cnsMsg)
	assert.Nil(t, err)
}

func TestCheckMessageWithSignatureValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{Body: []byte("body")}
	err := cmv.CheckMessageWithSignatureValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithSignatureValidity_InvalidSignatureShareSize(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{SignatureShare: []byte("0")}
	err := cmv.CheckMessageWithSignatureValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidSignatureSize))
}

func TestCheckMessageWithSignatureShareValidity_ShouldWork(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)
	cnsMsg := &consensus.Message{SignatureShare: sig}
	err := cmv.CheckMessageWithSignatureValidity(cnsMsg)
	assert.Nil(t, err)
}

func TestCheckMessageWithBlockHeaderValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{SignatureShare: []byte("0")}
	err := cmv.CheckMessageWithBlockHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithBlockHeaderValidity_InvalidHeaderSize(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{}
	err := cmv.CheckMessageWithBlockHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidHeaderSize))
}

func TestCheckMessageWithBlockHeaderValidity_HeaderTooBig(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, core.MegabyteSize+1)
	_, _ = rand.Read(headerBytes)
	cnsMsg := &consensus.Message{Header: headerBytes}
	err := cmv.CheckMessageWithBlockHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidHeaderSize))
}

func TestCheckMessageWithBlockHeaderValidity_HeaderSizeZero(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 0)
	cnsMsg := &consensus.Message{Header: headerBytes}
	err := cmv.CheckMessageWithBlockHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidHeaderSize))
}

func TestCheckMessageWithBlockHeaderValidity_ShouldWork(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{Header: []byte("header")}
	err := cmv.CheckMessageWithBlockHeaderValidity(cnsMsg)
	assert.Nil(t, err)
}

func TestCheckMessageWithBlockBodyValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	cnsMsg := &consensus.Message{Header: headerBytes}
	err := cmv.CheckMessageWithBlockBodyValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithBlockBodyValidity_InvalidBodySize(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	bodyBytes := make([]byte, core.MegabyteSize+1)
	_, _ = rand.Read(bodyBytes)
	cnsMsg := &consensus.Message{Body: bodyBytes}
	err := cmv.CheckMessageWithBlockBodyValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidBodySize))
}

func TestCheckMessageWithBlockBodyValidity_ShouldWork(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	bodyBytes := make([]byte, 100)
	_, _ = rand.Read(bodyBytes)
	cnsMsg := &consensus.Message{Body: bodyBytes}
	err := cmv.CheckMessageWithBlockBodyValidity(cnsMsg)
	assert.Nil(t, err)
}

func TestCheckMessageWithBlockBodyAndHeaderValidity_NilSigShareAfterActivation(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	consensusMessageValidatorArgs.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.EquivalentMessagesFlag
		},
	}
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{SignatureShare: nil}
	err := cmv.CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithBlockBodyAndHeaderValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{SignatureShare: []byte("0")}
	err := cmv.CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckMessageWithBlockBodyAndHeaderValidity_InvalidBodySize(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	bodyBytes := make([]byte, core.MegabyteSize+1)
	_, _ = rand.Read(bodyBytes)
	cnsMsg := &consensus.Message{Body: bodyBytes}
	err := cmv.CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidBodySize))
}

func TestCheckMessageWithBlockBodyAndHeaderValidity_InvalidHeaderSize(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, core.MegabyteSize+1)
	_, _ = rand.Read(headerBytes)
	cnsMsg := &consensus.Message{Header: headerBytes}
	err := cmv.CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidHeaderSize))
}

func TestCheckMessageWithBlockBodyAndHeaderValidity_ShouldWork(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	cnsMsg := &consensus.Message{Header: headerBytes}
	err := cmv.CheckMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	assert.Nil(t, err)
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithBlockBodyAndHeaderInvalid(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockBodyAndHeader), SignatureShare: []byte("1")}
	err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithBlockBodyInvalid(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockBody), SignatureShare: []byte("1")}
	err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithBlockHeaderInvalid(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockHeader), SignatureShare: []byte("1")}
	err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithBlockHeaderInvalidAfterFlag(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	consensusMessageValidatorArgs.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.EquivalentMessagesFlag
		},
	}
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockHeader), SignatureShare: nil}
	err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithSignatureInvalid(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtSignature), Header: []byte("1")}
	err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithFinalInfoInvalid(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockHeaderFinalInfo), Header: []byte("1")}
	err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidityForMessageType_MessageWithInvalidSigners(t *testing.T) {
	t.Parallel()

	t.Run("message with header, invalid err", func(t *testing.T) {
		t.Parallel()

		consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
		cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

		cnsMsg := &consensus.Message{MsgType: int64(bls.MtInvalidSigners), Header: []byte("1")}
		err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
		assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
	})

	t.Run("message with body, invalid err", func(t *testing.T) {
		t.Parallel()

		consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
		cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

		cnsMsg := &consensus.Message{MsgType: int64(bls.MtInvalidSigners), Body: []byte("1")}
		err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
		assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
	})

	t.Run("message with pubKeys bitmap, invalid err", func(t *testing.T) {
		t.Parallel()

		consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
		cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

		cnsMsg := &consensus.Message{MsgType: int64(bls.MtInvalidSigners), PubKeysBitmap: []byte("1")}
		err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
		assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
	})

	t.Run("message with aggregated signature, invalid err", func(t *testing.T) {
		t.Parallel()

		consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
		cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

		cnsMsg := &consensus.Message{MsgType: int64(bls.MtInvalidSigners), AggregateSignature: []byte("1")}
		err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
		assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
	})

	t.Run("message with leader signature, invalid err", func(t *testing.T) {
		t.Parallel()

		consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
		cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

		cnsMsg := &consensus.Message{MsgType: int64(bls.MtInvalidSigners), LeaderSignature: []byte("1")}
		err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
		assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
	})

	t.Run("valid message, should work", func(t *testing.T) {
		t.Parallel()

		consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
		cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

		cnsMsg := &consensus.Message{MsgType: int64(bls.MtInvalidSigners), InvalidSigners: []byte("1")}
		err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
		assert.Nil(t, err)
	})
}

func TestCheckConsensusMessageValidityForMessageType_MessageUnknownInvalid(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtUnknown), Header: []byte("1")}
	err := cmv.CheckConsensusMessageValidityForMessageType(cnsMsg)
	assert.True(t, errors.Is(err, spos.ErrInvalidMessageType))
}

func TestIsBlockHeaderHashSizeValid_NotValid(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockBody), BlockHeaderHash: []byte("hash")}
	result := cmv.IsBlockHeaderHashSizeValid(cnsMsg)
	assert.False(t, result)
}

func TestIsBlockHeaderHashSizeValid(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerHash := make([]byte, consensusMessageValidatorArgs.HeaderHashSize)
	_, _ = rand.Read(headerHash)
	cnsMsg := &consensus.Message{MsgType: int64(bls.MtBlockHeader), BlockHeaderHash: headerHash}
	result := cmv.IsBlockHeaderHashSizeValid(cnsMsg)
	assert.True(t, result)
}

func TestCheckConsensusMessageValidity_InvalidMessage(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cnsMsg := &consensus.Message{ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader), SignatureShare: []byte("1")}
	err := cmv.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidMessage))
}

func TestCheckConsensusMessageValidity_InvalidHeaderHashSize(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	cnsMsg := &consensus.Message{ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader), Header: headerBytes}
	err := cmv.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidHeaderHashSize))
}

func TestCheckConsensusMessageValidity_InvalidPublicKeySize(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, consensusMessageValidatorArgs.HeaderHashSize)
	_, _ = rand.Read(headerHash)
	cnsMsg := &consensus.Message{ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader), Header: headerBytes, BlockHeaderHash: headerHash}
	err := cmv.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidPublicKeySize))
}

func TestCheckConsensusMessageValidity_InvalidSignatureSize(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, consensusMessageValidatorArgs.HeaderHashSize)
	_, _ = rand.Read(headerHash)
	pubKey := make([]byte, PublicKeySize)
	_, _ = rand.Read(pubKey)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader), Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey,
	}
	err := cmv.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidSignatureSize))
}

func TestCheckConsensusMessageValidity_NodeIsNotEligible(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, consensusMessageValidatorArgs.HeaderHashSize)
	_, _ = rand.Read(headerHash)
	pubKey := make([]byte, PublicKeySize)
	_, _ = rand.Read(pubKey)
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig,
	}
	err := cmv.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrNodeIsNotInEligibleList))
}

func TestCheckConsensusMessageValidity_ErrMessageForFutureRound(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, consensusMessageValidatorArgs.HeaderHashSize)
	_, _ = rand.Read(headerHash)
	pubKey := []byte(consensusMessageValidatorArgs.ConsensusState.ConsensusGroup()[0])
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig, RoundIndex: 10,
	}
	err := cmv.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrMessageForFutureRound))
}

func TestCheckConsensusMessageValidity_ErrMessageForPastRound(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	consensusMessageValidatorArgs.ConsensusState.RoundIndex = 100
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, consensusMessageValidatorArgs.HeaderHashSize)
	_, _ = rand.Read(headerHash)
	pubKey := []byte(consensusMessageValidatorArgs.ConsensusState.ConsensusGroup()[0])
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig, RoundIndex: 10,
	}
	err := cmv.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrMessageForPastRound))
}

func TestCheckConsensusMessageValidity_ErrMessageTypeLimitReached(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	consensusMessageValidatorArgs.ConsensusState.RoundIndex = 10

	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)
	pubKey := []byte(consensusMessageValidatorArgs.ConsensusState.ConsensusGroup()[0])

	cnsMsgBlockBodyAndHeader := createMockConsensusMessage(consensusMessageValidatorArgs, pubKey, bls.MtBlockBodyAndHeader)
	cnsMsgBlockBodyAndHeader.Header = createDummyByteSlice(100)

	cnsMsgSignature := createMockConsensusMessage(consensusMessageValidatorArgs, pubKey, bls.MtSignature)
	cnsMsgSignature.SignatureShare = createDummyByteSlice(SignatureSize)

	// no message received
	err := cmv.CheckConsensusMessageValidity(cnsMsgBlockBodyAndHeader, "")
	assert.Nil(t, err)

	err = cmv.CheckConsensusMessageValidity(cnsMsgSignature, "")
	assert.Nil(t, err)

	// last checks added messages in the maps, let's test again
	err = cmv.CheckConsensusMessageValidity(cnsMsgBlockBodyAndHeader, "")
	assert.True(t, errors.Is(err, spos.ErrMessageTypeLimitReached))

	err = cmv.CheckConsensusMessageValidity(cnsMsgSignature, "")
	assert.Nil(t, err)

	// and another round of tests
	err = cmv.CheckConsensusMessageValidity(cnsMsgBlockBodyAndHeader, "")
	assert.True(t, errors.Is(err, spos.ErrMessageTypeLimitReached))

	err = cmv.CheckConsensusMessageValidity(cnsMsgSignature, "")
	assert.True(t, errors.Is(err, spos.ErrMessageTypeLimitReached))
}

func createDummyByteSlice(size int) []byte {
	buff := make([]byte, size)
	_, _ = rand.Read(buff)

	return buff
}

func createMockConsensusMessage(args spos.ArgsConsensusMessageValidator, pubKey []byte, msgType consensus.MessageType) *consensus.Message {
	return &consensus.Message{
		ChainID:         chainID,
		MsgType:         int64(msgType),
		PubKey:          pubKey,
		Signature:       createDummyByteSlice(SignatureSize),
		RoundIndex:      args.ConsensusState.RoundIndex,
		BlockHeaderHash: createDummyByteSlice(args.HeaderHashSize),
	}
}

func TestCheckConsensusMessageValidity_InvalidSignature(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local error")
	signer := &mock.SingleSignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return localErr
		},
	}

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	consensusMessageValidatorArgs.PeerSignatureHandler = &mock.PeerSignatureHandler{
		Signer: signer,
	}
	consensusMessageValidatorArgs.ConsensusState.RoundIndex = 10
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, consensusMessageValidatorArgs.HeaderHashSize)
	_, _ = rand.Read(headerHash)
	pubKey := []byte(consensusMessageValidatorArgs.ConsensusState.ConsensusGroup()[0])
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig, RoundIndex: 10,
	}
	err := cmv.CheckConsensusMessageValidity(cnsMsg, "")
	assert.True(t, errors.Is(err, spos.ErrInvalidSignature))
}

func TestCheckConsensusMessageValidity_Ok(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	consensusMessageValidatorArgs.ConsensusState.RoundIndex = 10
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	headerBytes := make([]byte, 100)
	_, _ = rand.Read(headerBytes)
	headerHash := make([]byte, consensusMessageValidatorArgs.HeaderHashSize)
	_, _ = rand.Read(headerHash)
	pubKey := []byte(consensusMessageValidatorArgs.ConsensusState.ConsensusGroup()[0])
	sig := make([]byte, SignatureSize)
	_, _ = rand.Read(sig)

	cnsMsg := &consensus.Message{
		ChainID: chainID, MsgType: int64(bls.MtBlockBodyAndHeader),
		Header: headerBytes, BlockHeaderHash: headerHash, PubKey: pubKey, Signature: sig, RoundIndex: 10,
	}
	err := cmv.CheckConsensusMessageValidity(cnsMsg, "")
	assert.Nil(t, err)
}

func TestIsMessageTypeLimitReached_ShouldWork(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	assert.False(t, cmv.IsMessageTypeLimitReached([]byte("pk1"), 1, bls.MtBlockBody))

	cmv.AddMessageTypeToPublicKey([]byte("pk1"), 1, bls.MtBlockHeader)

	assert.False(t, cmv.IsMessageTypeLimitReached([]byte("pk1"), 1, bls.MtBlockBody))
	assert.True(t, cmv.IsMessageTypeLimitReached([]byte("pk1"), 1, bls.MtBlockHeader))
	assert.False(t, cmv.IsMessageTypeLimitReached([]byte("pk1"), 2, bls.MtBlockHeader))
}

func TestAddMessageTypeToPublicKey_ShouldWork(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	assert.Equal(t, uint32(0), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 1, bls.MtBlockBody))

	cmv.AddMessageTypeToPublicKey([]byte("pk1"), 1, bls.MtBlockHeader)

	assert.Equal(t, uint32(0), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 1, bls.MtBlockBody))
	assert.Equal(t, uint32(0), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 2, bls.MtBlockHeader))

	cmv.AddMessageTypeToPublicKey([]byte("pk1"), 1, bls.MtBlockBody)
	cmv.AddMessageTypeToPublicKey([]byte("pk1"), 1, bls.MtBlockHeader)

	assert.Equal(t, uint32(1), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 1, bls.MtBlockBody))
	assert.Equal(t, uint32(2), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 1, bls.MtBlockHeader))
	assert.Equal(t, uint32(0), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 2, bls.MtBlockHeader))

	cmv.AddMessageTypeToPublicKey([]byte("pk2"), 1, bls.MtBlockHeaderFinalInfo)

	assert.Equal(t, uint32(0), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 1, bls.MtBlockHeaderFinalInfo))
	assert.Equal(t, uint32(0), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk2"), 2, bls.MtBlockHeaderFinalInfo))
	assert.Equal(t, uint32(1), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk2"), 1, bls.MtBlockHeaderFinalInfo))
}

func TestResetConsensusMessages_ShouldWork(t *testing.T) {
	t.Parallel()

	consensusMessageValidatorArgs := createDefaultConsensusMessageValidatorArgs()
	cmv, _ := spos.NewConsensusMessageValidator(consensusMessageValidatorArgs)

	cmv.AddMessageTypeToPublicKey([]byte("pk1"), 1, bls.MtBlockBody)
	cmv.AddMessageTypeToPublicKey([]byte("pk1"), 1, bls.MtBlockBody)
	cmv.AddMessageTypeToPublicKey([]byte("pk1"), 2, bls.MtBlockBody)
	cmv.AddMessageTypeToPublicKey([]byte("pk1"), 1, bls.MtBlockHeader)
	cmv.AddMessageTypeToPublicKey([]byte("pk2"), 1, bls.MtBlockHeaderFinalInfo)

	assert.Equal(t, uint32(2), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 1, bls.MtBlockBody))
	assert.Equal(t, uint32(1), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 1, bls.MtBlockHeader))
	assert.Equal(t, uint32(1), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 2, bls.MtBlockBody))
	assert.Equal(t, uint32(1), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk2"), 1, bls.MtBlockHeaderFinalInfo))

	cmv.ResetConsensusMessages()

	assert.Equal(t, uint32(0), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 1, bls.MtBlockBody))
	assert.Equal(t, uint32(0), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 1, bls.MtBlockHeader))
	assert.Equal(t, uint32(0), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk1"), 2, bls.MtBlockBody))
	assert.Equal(t, uint32(0), cmv.GetNumOfMessageTypeForPublicKey([]byte("pk2"), 1, bls.MtBlockHeaderFinalInfo))
}
