package spos

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func (wrk *Worker) checkConsensusMessageValidity(cnsMsg *consensus.Message, originator core.PeerID) error {
	if !bytes.Equal(cnsMsg.ChainID, wrk.chainID) {
		return fmt.Errorf("%w : received chain ID from consensus topic is invalid: %s",
			ErrInvalidChainID,
			logger.DisplayByteSlice(cnsMsg.ChainID))
	}

	err := wrk.checkConsensusMessageValidityForMessageType(cnsMsg)
	if err != nil {
		return err
	}

	if !wrk.isBlockHeaderHashSizeValid(cnsMsg) {
		return fmt.Errorf("%w : received header hash from consensus topic has an invalid size: %d",
			ErrInvalidHeaderHashSize,
			len(cnsMsg.BlockHeaderHash))
	}

	if len(cnsMsg.PubKey) != wrk.publicKeySize {
		return fmt.Errorf("%w : received public key from consensus topic has an invalid size: %d",
			ErrInvalidPublicKeySize,
			len(cnsMsg.PubKey))
	}

	if len(cnsMsg.Signature) != wrk.signatureSize {
		return fmt.Errorf("%w : received signature from consensus topic has an invalid size: %d",
			ErrInvalidSignatureSize,
			len(cnsMsg.Signature))
	}

	isNodeInEligibleList := wrk.consensusState.IsNodeInEligibleList(string(cnsMsg.PubKey))
	if !isNodeInEligibleList {
		return fmt.Errorf("%w : received message from consensus topic has an invalid public key: %s",
			ErrNodeIsNotInEligibleList,
			logger.DisplayByteSlice(cnsMsg.PubKey))
	}

	msgType := consensus.MessageType(cnsMsg.MsgType)

	if wrk.consensusState.RoundIndex+1 < cnsMsg.RoundIndex {
		log.Trace("received message from consensus topic has a future round",
			"msg type", wrk.consensusService.GetStringValue(msgType),
			"from", cnsMsg.PubKey,
			"header hash", cnsMsg.BlockHeaderHash,
			"msg round", cnsMsg.RoundIndex,
			"round", wrk.consensusState.RoundIndex,
		)

		return fmt.Errorf("%w : received message from consensus topic has a future round: %d",
			ErrMessageForFutureRound,
			cnsMsg.RoundIndex)
	}

	if wrk.consensusState.RoundIndex > cnsMsg.RoundIndex {
		log.Trace("received message from consensus topic has a past round",
			"msg type", wrk.consensusService.GetStringValue(msgType),
			"from", cnsMsg.PubKey,
			"header hash", cnsMsg.BlockHeaderHash,
			"msg round", cnsMsg.RoundIndex,
			"round", wrk.consensusState.RoundIndex,
		)

		return fmt.Errorf("%w : received message from consensus topic has a past round: %d",
			ErrMessageForPastRound,
			cnsMsg.RoundIndex)
	}

	if wrk.isMessageTypeLimitReached(cnsMsg.PubKey, msgType) {
		log.Trace("received message type from consensus topic reached the limit",
			"msg type", wrk.consensusService.GetStringValue(msgType),
			"public key", cnsMsg.PubKey,
		)

		return fmt.Errorf("%w : received message type %s from consensus topic reached the limit for public key: %s",
			ErrMessageTypeLimitReached,
			wrk.consensusService.GetStringValue(msgType),
			logger.DisplayByteSlice(cnsMsg.PubKey))
	}

	err = wrk.peerSignatureHandler.VerifyPeerSignature(cnsMsg.PubKey, core.PeerID(cnsMsg.OriginatorPid), cnsMsg.Signature)
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

	wrk.addMessageTypeToPublicKey(cnsMsg.PubKey, msgType)

	return nil
}

func (wrk *Worker) isBlockHeaderHashSizeValid(cnsMsg *consensus.Message) bool {
	msgType := consensus.MessageType(cnsMsg.MsgType)
	isMessageWithBlockBody := wrk.consensusService.IsMessageWithBlockBody(msgType)

	if isMessageWithBlockBody {
		return cnsMsg.BlockHeaderHash == nil
	}

	return len(cnsMsg.BlockHeaderHash) == wrk.hasher.Size()
}

func (wrk *Worker) checkConsensusMessageValidityForMessageType(cnsMsg *consensus.Message) error {
	msgType := consensus.MessageType(cnsMsg.MsgType)

	if wrk.consensusService.IsMessageWithBlockBodyAndHeader(msgType) {
		return wrk.checkMessageWithBlockBodyAndHeaderValidity(cnsMsg)
	}

	if wrk.consensusService.IsMessageWithBlockBody(msgType) {
		return wrk.checkMessageWithBlockBodyValidity(cnsMsg)
	}

	if wrk.consensusService.IsMessageWithBlockHeader(msgType) {
		return wrk.checkMessageWithBlockHeaderValidity(cnsMsg)
	}

	if wrk.consensusService.IsMessageWithSignature(msgType) {
		return wrk.checkMessageWithSignatureValidity(cnsMsg)
	}

	if wrk.consensusService.IsMessageWithFinalInfo(msgType) {
		return wrk.checkMessageWithFinalInfoValidity(cnsMsg)
	}

	return fmt.Errorf("%w : received message type from consensus topic is invalid: %d",
		ErrInvalidMessageType,
		msgType)
}

func (wrk *Worker) checkMessageWithBlockBodyAndHeaderValidity(cnsMsg *consensus.Message) error {
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

func (wrk *Worker) checkMessageWithBlockBodyValidity(cnsMsg *consensus.Message) error {
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

func (wrk *Worker) checkMessageWithBlockHeaderValidity(cnsMsg *consensus.Message) error {
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

func (wrk *Worker) checkMessageWithSignatureValidity(cnsMsg *consensus.Message) error {
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

	if len(cnsMsg.SignatureShare) != wrk.signatureSize {
		return fmt.Errorf("%w : received signature share from consensus topic has an invalid size: %d",
			ErrInvalidSignatureSize,
			len(cnsMsg.SignatureShare))
	}

	return nil
}

func (wrk *Worker) checkMessageWithFinalInfoValidity(cnsMsg *consensus.Message) error {
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

	if len(cnsMsg.PubKeysBitmap) != wrk.publicKeyBitmapSize {
		return fmt.Errorf("%w : received public key bitmap from consensus topic has an invalid size: %d",
			ErrInvalidPublicKeyBitmapSize,
			len(cnsMsg.PubKeysBitmap))
	}

	if len(cnsMsg.AggregateSignature) != wrk.signatureSize {
		return fmt.Errorf("%w : received aggregate signature from consensus topic has an invalid size: %d",
			ErrInvalidSignatureSize,
			len(cnsMsg.AggregateSignature))
	}

	if len(cnsMsg.LeaderSignature) != wrk.signatureSize {
		return fmt.Errorf("%w : received leader signature from consensus topic has an invalid size: %d",
			ErrInvalidSignatureSize,
			len(cnsMsg.LeaderSignature))
	}

	return nil
}

func (wrk *Worker) isMessageTypeLimitReached(pk []byte, msgType consensus.MessageType) bool {
	wrk.mutPkConsensusMessages.RLock()
	defer wrk.mutPkConsensusMessages.RUnlock()

	mapMsgType, ok := wrk.mapPkConsensusMessages[string(pk)]
	if !ok {
		return false
	}

	numMsgType, ok := mapMsgType[msgType]
	if !ok {
		return false
	}

	return numMsgType >= MaxNumOfMessageTypeAccepted
}

func (wrk *Worker) addMessageTypeToPublicKey(pk []byte, msgType consensus.MessageType) {
	wrk.mutPkConsensusMessages.Lock()
	defer wrk.mutPkConsensusMessages.Unlock()

	mapMsgType, ok := wrk.mapPkConsensusMessages[string(pk)]
	if !ok {
		mapMsgType = make(map[consensus.MessageType]uint32)
		wrk.mapPkConsensusMessages[string(pk)] = mapMsgType
	}

	mapMsgType[msgType]++
}
