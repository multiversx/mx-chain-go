package spos

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type consensusMessageValidator struct {
	consensusState       *ConsensusState
	consensusService     ConsensusService
	peerSignatureHandler crypto.PeerSignatureHandler

	signatureSize       int
	publicKeySize       int
	publicKeyBitmapSize int
	hasherSize          int
	chainID             []byte

	mutPkConsensusMessages sync.RWMutex
	mapPkConsensusMessages map[string]map[consensus.MessageType]uint32
}

// ArgsConsensusMessageValidator holds the consensus message validator arguments
type ArgsConsensusMessageValidator struct {
	ConsensusState       *ConsensusState
	ConsensusService     ConsensusService
	PeerSignatureHandler crypto.PeerSignatureHandler
	SignatureSize        int
	PublicKeySize        int
	HasherSize           int
	ChainID              []byte
}

// NewConsensusMessageValidator creates a new consensusMessageValidator object
func NewConsensusMessageValidator(args *ArgsConsensusMessageValidator) (*consensusMessageValidator, error) {
	cmv := &consensusMessageValidator{
		consensusState:       args.ConsensusState,
		consensusService:     args.ConsensusService,
		peerSignatureHandler: args.PeerSignatureHandler,
		signatureSize:        args.SignatureSize,
		publicKeySize:        args.PublicKeySize,
		chainID:              args.ChainID,
		hasherSize:           args.HasherSize,
	}

	cmv.publicKeyBitmapSize = cmv.getPublicKeyBitmapSize()
	cmv.mapPkConsensusMessages = make(map[string]map[consensus.MessageType]uint32)

	return cmv, nil
}

func (cmv *consensusMessageValidator) getPublicKeyBitmapSize() int {
	sizeConsensus := cmv.consensusState.consensusGroupSize
	bitmapSize := sizeConsensus / 8
	if sizeConsensus%8 != 0 {
		bitmapSize++
	}

	return bitmapSize
}

func (cmv *consensusMessageValidator) checkConsensusMessageValidity(cnsMsg *consensus.Message, originator core.PeerID) error {
	if !bytes.Equal(cnsMsg.ChainID, cmv.chainID) {
		return fmt.Errorf("%w : received chain ID from consensus topic is invalid: %s",
			ErrInvalidChainID,
			logger.DisplayByteSlice(cnsMsg.ChainID))
	}

	err := cmv.checkConsensusMessageValidityForMessageType(cnsMsg)
	if err != nil {
		return err
	}

	if !cmv.isBlockHeaderHashSizeValid(cnsMsg) {
		return fmt.Errorf("%w : received header hash from consensus topic has an invalid size: %d",
			ErrInvalidHeaderHashSize,
			len(cnsMsg.BlockHeaderHash))
	}

	if len(cnsMsg.PubKey) != cmv.publicKeySize {
		return fmt.Errorf("%w : received public key from consensus topic has an invalid size: %d",
			ErrInvalidPublicKeySize,
			len(cnsMsg.PubKey))
	}

	if len(cnsMsg.Signature) != cmv.signatureSize {
		return fmt.Errorf("%w : received signature from consensus topic has an invalid size: %d",
			ErrInvalidSignatureSize,
			len(cnsMsg.Signature))
	}

	isNodeInEligibleList := cmv.consensusState.IsNodeInEligibleList(string(cnsMsg.PubKey))
	if !isNodeInEligibleList {
		return fmt.Errorf("%w : received message from consensus topic has an invalid public key: %s",
			ErrNodeIsNotInEligibleList,
			logger.DisplayByteSlice(cnsMsg.PubKey))
	}

	msgType := consensus.MessageType(cnsMsg.MsgType)

	if cmv.consensusState.RoundIndex+1 < cnsMsg.RoundIndex {
		log.Trace("received message from consensus topic has a future round",
			"msg type", cmv.consensusService.GetStringValue(msgType),
			"from", cnsMsg.PubKey,
			"header hash", cnsMsg.BlockHeaderHash,
			"msg round", cnsMsg.RoundIndex,
			"round", cmv.consensusState.RoundIndex,
		)

		return fmt.Errorf("%w : received message from consensus topic has a future round: %d",
			ErrMessageForFutureRound,
			cnsMsg.RoundIndex)
	}

	if cmv.consensusState.RoundIndex > cnsMsg.RoundIndex {
		log.Trace("received message from consensus topic has a past round",
			"msg type", cmv.consensusService.GetStringValue(msgType),
			"from", cnsMsg.PubKey,
			"header hash", cnsMsg.BlockHeaderHash,
			"msg round", cnsMsg.RoundIndex,
			"round", cmv.consensusState.RoundIndex,
		)

		return fmt.Errorf("%w : received message from consensus topic has a past round: %d",
			ErrMessageForPastRound,
			cnsMsg.RoundIndex)
	}

	if cmv.isMessageTypeLimitReached(cnsMsg.PubKey, cnsMsg.RoundIndex, msgType) {
		log.Trace("received message type from consensus topic reached the limit",
			"msg type", cmv.consensusService.GetStringValue(msgType),
			"public key", cnsMsg.PubKey,
		)

		return fmt.Errorf("%w : received message type %s from consensus topic reached the limit for public key: %s",
			ErrMessageTypeLimitReached,
			cmv.consensusService.GetStringValue(msgType),
			logger.DisplayByteSlice(cnsMsg.PubKey))
	}

	err = cmv.peerSignatureHandler.VerifyPeerSignature(cnsMsg.PubKey, core.PeerID(cnsMsg.OriginatorPid), cnsMsg.Signature)
	if err != nil {
		return fmt.Errorf("%w : verify signature for received message from consensus topic failed: %s",
			ErrInvalidSignature,
			err.Error())
	}

	cnsMsgOriginator := core.PeerID(cnsMsg.OriginatorPid)
	if cnsMsgOriginator != originator {
		return fmt.Errorf("%w : pubsub originator pid: %s, cnsMsg.OriginatorPid: %s",
			ErrOriginatorMismatch, p2p.PeerIdToShortString(originator), p2p.PeerIdToShortString(cnsMsgOriginator))
	}

	cmv.addMessageTypeToPublicKey(cnsMsg.PubKey, cnsMsg.RoundIndex, msgType)

	return nil
}

func (cmv *consensusMessageValidator) isBlockHeaderHashSizeValid(cnsMsg *consensus.Message) bool {
	msgType := consensus.MessageType(cnsMsg.MsgType)
	isMessageWithBlockBody := cmv.consensusService.IsMessageWithBlockBody(msgType)

	if isMessageWithBlockBody {
		return cnsMsg.BlockHeaderHash == nil
	}

	return len(cnsMsg.BlockHeaderHash) == cmv.hasherSize
}

func (cmv *consensusMessageValidator) checkConsensusMessageValidityForMessageType(cnsMsg *consensus.Message) error {
	msgType := consensus.MessageType(cnsMsg.MsgType)

	if cmv.consensusService.IsMessageWithBlockBodyAndHeader(msgType) {
		return cmv.checkMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	}

	if cmv.consensusService.IsMessageWithBlockBody(msgType) {
		return cmv.checkMessageWithBlockBodyValidity(cnsMsg)
	}

	if cmv.consensusService.IsMessageWithBlockHeader(msgType) {
		return cmv.checkMessageWithBlockHeaderValidity(cnsMsg)
	}

	if cmv.consensusService.IsMessageWithSignature(msgType) {
		return cmv.checkMessageWithSignatureValidity(cnsMsg)
	}

	if cmv.consensusService.IsMessageWithFinalInfo(msgType) {
		return cmv.checkMessageWithFinalInfoValidity(cnsMsg)
	}

	return fmt.Errorf("%w : received message type from consensus topic is invalid: %d",
		ErrInvalidMessageType,
		msgType)
}

func (cmv *consensusMessageValidator) checkMessageWithBlockBodyAndHeaderValidity(cnsMsg *consensus.Message) error {
	isMessageInvalid := cnsMsg.SignatureShare != nil ||
		cnsMsg.PubKeysBitmap != nil ||
		cnsMsg.AggregateSignature != nil ||
		cnsMsg.LeaderSignature != nil

	if isMessageInvalid {
		log.Trace("received message from consensus topic is invalid",
			"SignatureShare", cnsMsg.SignatureShare,
			"PubKeysBitmap", cnsMsg.PubKeysBitmap,
			"AggregateSignature", cnsMsg.AggregateSignature,
			"LeaderSignature", cnsMsg.LeaderSignature)

		return fmt.Errorf("%w : received message from public key: %s from consensus topic is invalid",
			ErrInvalidMessage,
			logger.DisplayByteSlice(cnsMsg.PubKey))
	}

	if len(cnsMsg.Body) > core.MegabyteSize {
		return fmt.Errorf("%w : received body from consensus topic has an invalid size: %d",
			ErrInvalidBodySize,
			len(cnsMsg.Body))
	}

	headerLen := len(cnsMsg.Header)
	if headerLen == 0 || headerLen > core.MegabyteSize {
		return fmt.Errorf("%w : received header from consensus topic has an invalid size: %d",
			ErrInvalidHeaderSize,
			len(cnsMsg.Header))
	}

	return nil
}

func (cmv *consensusMessageValidator) checkMessageWithBlockBodyValidity(cnsMsg *consensus.Message) error {
	isMessageInvalid := cnsMsg.Header != nil ||
		cnsMsg.SignatureShare != nil ||
		cnsMsg.PubKeysBitmap != nil ||
		cnsMsg.AggregateSignature != nil ||
		cnsMsg.LeaderSignature != nil

	if isMessageInvalid {
		log.Trace("received message from consensus topic is invalid",
			"header len", len(cnsMsg.Header),
			"SignatureShare", cnsMsg.SignatureShare,
			"PubKeysBitmap", cnsMsg.PubKeysBitmap,
			"AggregateSignature", cnsMsg.AggregateSignature,
			"LeaderSignature", cnsMsg.LeaderSignature)

		return fmt.Errorf("%w : received message from public key: %s from consensus topic is invalid",
			ErrInvalidMessage,
			logger.DisplayByteSlice(cnsMsg.PubKey))
	}

	if len(cnsMsg.Body) > core.MegabyteSize {
		return fmt.Errorf("%w : received body from consensus topic has an invalid size: %d",
			ErrInvalidBodySize,
			len(cnsMsg.Body))
	}

	return nil
}

func (cmv *consensusMessageValidator) checkMessageWithBlockHeaderValidity(cnsMsg *consensus.Message) error {
	isMessageInvalid := cnsMsg.Body != nil ||
		cnsMsg.SignatureShare != nil ||
		cnsMsg.PubKeysBitmap != nil ||
		cnsMsg.AggregateSignature != nil ||
		cnsMsg.LeaderSignature != nil

	if isMessageInvalid {
		log.Trace("received message from consensus topic is invalid",
			"body len", len(cnsMsg.Body),
			"SignatureShare", cnsMsg.SignatureShare,
			"PubKeysBitmap", cnsMsg.PubKeysBitmap,
			"AggregateSignature", cnsMsg.AggregateSignature,
			"LeaderSignature", cnsMsg.LeaderSignature)

		return fmt.Errorf("%w : received message from public key: %s from consensus topic is invalid",
			ErrInvalidMessage,
			logger.DisplayByteSlice(cnsMsg.PubKey))
	}

	headerLen := len(cnsMsg.Header)
	if headerLen == 0 || headerLen > core.MegabyteSize {
		return fmt.Errorf("%w : received header from consensus topic has an invalid size: %d",
			ErrInvalidHeaderSize,
			len(cnsMsg.Header))
	}

	return nil
}

func (cmv *consensusMessageValidator) checkMessageWithSignatureValidity(cnsMsg *consensus.Message) error {
	isMessageInvalid := cnsMsg.Body != nil ||
		cnsMsg.Header != nil ||
		cnsMsg.PubKeysBitmap != nil ||
		cnsMsg.AggregateSignature != nil ||
		cnsMsg.LeaderSignature != nil

	if isMessageInvalid {
		log.Trace("received message from consensus topic is invalid",
			"body len", len(cnsMsg.Body),
			"header len", len(cnsMsg.Header),
			"PubKeysBitmap", cnsMsg.PubKeysBitmap,
			"AggregateSignature", cnsMsg.AggregateSignature,
			"LeaderSignature", cnsMsg.LeaderSignature)

		return fmt.Errorf("%w : received message from public key: %s from consensus topic is invalid",
			ErrInvalidMessage,
			logger.DisplayByteSlice(cnsMsg.PubKey))
	}

	if len(cnsMsg.SignatureShare) != cmv.signatureSize {
		return fmt.Errorf("%w : received signature share from consensus topic has an invalid size: %d",
			ErrInvalidSignatureSize,
			len(cnsMsg.SignatureShare))
	}

	return nil
}

func (cmv *consensusMessageValidator) checkMessageWithFinalInfoValidity(cnsMsg *consensus.Message) error {
	isMessageInvalid := cnsMsg.Body != nil ||
		cnsMsg.Header != nil ||
		cnsMsg.SignatureShare != nil

	if isMessageInvalid {
		log.Trace("received message from consensus topic is invalid",
			"body len", len(cnsMsg.Body),
			"header len", len(cnsMsg.Header),
			"SignatureShare", cnsMsg.SignatureShare)

		return fmt.Errorf("%w : received message from public key: %s from consensus topic is invalid",
			ErrInvalidMessage,
			logger.DisplayByteSlice(cnsMsg.PubKey))
	}

	if len(cnsMsg.PubKeysBitmap) != cmv.publicKeyBitmapSize {
		return fmt.Errorf("%w : received public key bitmap from consensus topic has an invalid size: %d",
			ErrInvalidPublicKeyBitmapSize,
			len(cnsMsg.PubKeysBitmap))
	}

	if len(cnsMsg.AggregateSignature) != cmv.signatureSize {
		return fmt.Errorf("%w : received aggregate signature from consensus topic has an invalid size: %d",
			ErrInvalidSignatureSize,
			len(cnsMsg.AggregateSignature))
	}

	if len(cnsMsg.LeaderSignature) != cmv.signatureSize {
		return fmt.Errorf("%w : received leader signature from consensus topic has an invalid size: %d",
			ErrInvalidSignatureSize,
			len(cnsMsg.LeaderSignature))
	}

	return nil
}

func (cmv *consensusMessageValidator) isMessageTypeLimitReached(pk []byte, round int64, msgType consensus.MessageType) bool {
	cmv.mutPkConsensusMessages.RLock()
	defer cmv.mutPkConsensusMessages.RUnlock()

	key := fmt.Sprintf("%s_%d", string(pk), round)

	mapMsgType, ok := cmv.mapPkConsensusMessages[key]
	if !ok {
		return false
	}

	numMsgType, ok := mapMsgType[msgType]
	if !ok {
		return false
	}

	return numMsgType >= MaxNumOfMessageTypeAccepted
}

func (cmv *consensusMessageValidator) addMessageTypeToPublicKey(pk []byte, round int64, msgType consensus.MessageType) {
	cmv.mutPkConsensusMessages.Lock()
	defer cmv.mutPkConsensusMessages.Unlock()

	key := fmt.Sprintf("%s_%d", string(pk), round)

	mapMsgType, ok := cmv.mapPkConsensusMessages[key]
	if !ok {
		mapMsgType = make(map[consensus.MessageType]uint32)
		cmv.mapPkConsensusMessages[key] = mapMsgType
	}

	mapMsgType[msgType]++
}

func (cmv *consensusMessageValidator) resetConsensusMessages() {
	cmv.mutPkConsensusMessages.Lock()
	cmv.mapPkConsensusMessages = make(map[string]map[consensus.MessageType]uint32)
	cmv.mutPkConsensusMessages.Unlock()
}
