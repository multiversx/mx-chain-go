package spos

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
)

func (wrk *Worker) checkConsensusMessageValidity(cnsMsg *consensus.Message) error {
	if !bytes.Equal(cnsMsg.ChainID, wrk.chainID) {
		return fmt.Errorf("%w : received chain ID from consensus topic is invalid: %s",
			ErrInvalidChainID,
			logger.DisplayByteSlice(cnsMsg.ChainID))
	}

	err := wrk.checkConsensusMessageValidityForMessageType(cnsMsg)
	if err != nil {
		return err
	}

	if len(cnsMsg.BlockHeaderHash) != core.HashSizeInBytes {
		return fmt.Errorf("%w : received header hash from consensus topic has an invalid size: %d",
			ErrInvalidHeaderHashSize,
			len(cnsMsg.BlockHeaderHash))
	}

	if len(cnsMsg.PubKey) != core.PublicKeySizeInBytes {
		return fmt.Errorf("%w : received public key from consensus topic has an invalid size: %d",
			ErrInvalidPublicKeySize,
			len(cnsMsg.PubKey))
	}

	if len(cnsMsg.Signature) != core.SignatureSizeInBytes {
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

		return ErrMessageForFutureRound
	}

	if wrk.consensusState.RoundIndex > cnsMsg.RoundIndex {
		log.Trace("received message from consensus topic has a past round",
			"msg type", wrk.consensusService.GetStringValue(msgType),
			"from", cnsMsg.PubKey,
			"header hash", cnsMsg.BlockHeaderHash,
			"msg round", cnsMsg.RoundIndex,
			"round", wrk.consensusState.RoundIndex,
		)

		return ErrMessageForPastRound
	}

	err = wrk.checkSignature(cnsMsg)
	if err != nil {
		return fmt.Errorf("%w : verify signature for received message from consensus topic failed: %s",
			ErrInvalidSignature,
			err.Error())
	}

	return nil
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
		return fmt.Errorf("%w : received message from consensus topic is invalid: "+
			"SignatureShare = %s PubKeysBitmap = %s AggregateSignature = %s LeaderSignature = %s",
			ErrInvalidMessage,
			logger.DisplayByteSlice(cnsMsg.SignatureShare),
			logger.DisplayByteSlice(cnsMsg.PubKeysBitmap),
			logger.DisplayByteSlice(cnsMsg.AggregateSignature),
			logger.DisplayByteSlice(cnsMsg.LeaderSignature))
	}

	if len(cnsMsg.Body) > core.MegabyteSize {
		return fmt.Errorf("%w : received body from consensus topic has an invalid size: %d",
			ErrInvalidBodySize,
			len(cnsMsg.Body))
	}

	if cnsMsg.Header == nil || len(cnsMsg.Header) > core.MegabyteSize {
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
		return fmt.Errorf("%w : received message from consensus topic is invalid: "+
			"len(Header) = %d SignatureShare = %s PubKeysBitmap = %s AggregateSignature = %s LeaderSignature = %s",
			ErrInvalidMessage,
			len(cnsMsg.Header),
			logger.DisplayByteSlice(cnsMsg.SignatureShare),
			logger.DisplayByteSlice(cnsMsg.PubKeysBitmap),
			logger.DisplayByteSlice(cnsMsg.AggregateSignature),
			logger.DisplayByteSlice(cnsMsg.LeaderSignature))
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
		return fmt.Errorf("%w : received message from consensus topic is invalid: "+
			"len(Body) = %d SignatureShare = %s PubKeysBitmap = %s AggregateSignature = %s LeaderSignature = %s",
			ErrInvalidMessage,
			len(cnsMsg.Body),
			logger.DisplayByteSlice(cnsMsg.SignatureShare),
			logger.DisplayByteSlice(cnsMsg.PubKeysBitmap),
			logger.DisplayByteSlice(cnsMsg.AggregateSignature),
			logger.DisplayByteSlice(cnsMsg.LeaderSignature))
	}

	if cnsMsg.Header == nil || len(cnsMsg.Header) > core.MegabyteSize {
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
		return fmt.Errorf("%w : received message from consensus topic is invalid: "+
			"len(Body) = %d len(Header) = %d PubKeysBitmap = %s AggregateSignature = %s LeaderSignature = %s",
			ErrInvalidMessage,
			len(cnsMsg.Body),
			len(cnsMsg.Header),
			logger.DisplayByteSlice(cnsMsg.PubKeysBitmap),
			logger.DisplayByteSlice(cnsMsg.AggregateSignature),
			logger.DisplayByteSlice(cnsMsg.LeaderSignature))
	}

	if len(cnsMsg.SignatureShare) != core.SignatureSizeInBytes {
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
		return fmt.Errorf("%w : received message from consensus topic is invalid: "+
			"len(Body) = %d len(Header) = %d SignatureShare = %s",
			ErrInvalidMessage,
			len(cnsMsg.Body),
			len(cnsMsg.Header),
			logger.DisplayByteSlice(cnsMsg.SignatureShare))
	}

	if len(cnsMsg.PubKeysBitmap) != wrk.getPubKeyBitmapSize() {
		return fmt.Errorf("%w : received public key bitmap from consensus topic has an invalid size: %d",
			ErrInvalidPublicKeyBitmapSize,
			len(cnsMsg.PubKeysBitmap))
	}

	if len(cnsMsg.AggregateSignature) != core.SignatureSizeInBytes {
		return fmt.Errorf("%w : received aggregate signature from consensus topic has an invalid size: %d",
			ErrInvalidSignatureSize,
			len(cnsMsg.AggregateSignature))
	}

	if len(cnsMsg.LeaderSignature) != core.SignatureSizeInBytes {
		return fmt.Errorf("%w : received leader signature from consensus topic has an invalid size: %d",
			ErrInvalidSignatureSize,
			len(cnsMsg.LeaderSignature))
	}

	return nil
}

func (wrk *Worker) getPubKeyBitmapSize() int {
	sizeConsensus := wrk.consensusState.consensusGroupSize
	bitmapSize := sizeConsensus / 8
	if sizeConsensus%8 != 0 {
		bitmapSize++
	}

	return bitmapSize
}
