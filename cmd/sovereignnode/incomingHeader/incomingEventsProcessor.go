package incomingHeader

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/sovereignnode/dataCodec"
)

const (
	minTopicsInTransferEvent  = 5
	numTransferTopics         = 3
	numExecutedBridgeOpTopics = 3
)

const (
	eventIDExecutedOutGoingBridgeOp = "executeBridgeOps"
	eventIDDepositIncomingTransfer  = "deposit"

	topicIDDepositConfirmedOutGoingOperation = "executedBridgeOp"
	topicIDDepositIncomingTransfer           = "executeBridgeOps"
)

type confirmedBridgeOp struct {
	hashOfHashes []byte
	hash         []byte
}

type eventData struct {
	nonce                uint64
	functionCallWithArgs []byte
	gasLimit             uint64
}

type scrInfo struct {
	scr  *smartContractResult.SmartContractResult
	hash []byte
}

type eventsResult struct {
	scrs               []*scrInfo
	confirmedBridgeOps []*confirmedBridgeOp
}

type incomingEventsProcessor struct {
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
	dataCodec  dataCodec.DataCodecProcessor
}

func (iep *incomingEventsProcessor) processIncomingEvents(events []data.EventHandler) (*eventsResult, error) {
	scrs := make([]*scrInfo, 0, len(events))
	confirmedBridgeOps := make([]*confirmedBridgeOp, 0, len(events))

	for idx, event := range events {
		topics := event.GetTopics()

		var scr *scrInfo
		var confirmedOp *confirmedBridgeOp
		var err error
		switch string(event.GetIdentifier()) {
		case eventIDDepositIncomingTransfer:
			scr, err = iep.createSCRInfo(topics, event)
			scrs = append(scrs, scr)
		case eventIDExecutedOutGoingBridgeOp:
			if len(topics) == 0 {
				return nil, errInvalidNumTopicsIncomingEvent
			}

			switch string(topics[0]) {
			case topicIDDepositIncomingTransfer:
				scr, err = iep.createSCRInfo(topics, event)
				scrs = append(scrs, scr)
			case topicIDDepositConfirmedOutGoingOperation:
				confirmedOp, err = iep.getConfirmedBridgeOperation(topics)
				confirmedBridgeOps = append(confirmedBridgeOps, confirmedOp)
			default:
				return nil, errInvalidIncomingTopicIdentifier
			}
		default:
			return nil, errInvalidIncomingEventIdentifier
		}

		if err != nil {
			return nil, fmt.Errorf("%w, event idx = %d", err, idx)
		}
	}

	return &eventsResult{
		scrs:               scrs,
		confirmedBridgeOps: confirmedBridgeOps,
	}, nil
}

func (iep *incomingEventsProcessor) createSCRInfo(topics [][]byte, event data.EventHandler) (*scrInfo, error) {
	err := checkTopicsValidity(topics)
	if err != nil {
		return nil, err
	}

	receivedEventData, err := iep.createEventData(event.GetData())
	if err != nil {
		return nil, err
	}

	scrData, err := iep.createSCRData(topics)
	if err != nil {
		return nil, err
	}

	scrData = append(scrData, receivedEventData.functionCallWithArgs...)
	scr := &smartContractResult.SmartContractResult{
		Nonce:          receivedEventData.nonce,
		OriginalTxHash: nil, // TODO:  Implement this in MX-14321 task
		RcvAddr:        topics[1],
		SndAddr:        core.ESDTSCAddress,
		Data:           scrData,
		Value:          big.NewInt(0),
		GasLimit:       receivedEventData.gasLimit,
	}

	hash, err := core.CalculateHash(iep.marshaller, iep.hasher, scr)
	if err != nil {
		return nil, err
	}

	return &scrInfo{
		scr:  scr,
		hash: hash,
	}, nil
}

func (iep *incomingEventsProcessor) createEventData(data []byte) (*eventData, error) {
	if len(data) == 0 {
		return nil, errEmptyLogData
	}

	evData, err := iep.dataCodec.DeserializeEventData(data)
	if err != nil {
		return nil, err
	}

	gasLimit := uint64(0)
	args := make([]byte, 0)
	if evData.TransferData != nil {
		gasLimit = evData.GasLimit

		args = append(args, []byte("@")...)
		args = append(args, hex.EncodeToString(evData.Function)...)

		if len(evData.Args) > 0 {
			for _, arg := range evData.Args {
				args = append(args, []byte("@")...)
				args = append(args, hex.EncodeToString(arg)...)
			}
		}
	}

	return &eventData{
		nonce:                evData.Nonce,
		functionCallWithArgs: args,
		gasLimit:             gasLimit,
	}, nil
}

func (iep *incomingEventsProcessor) createSCRData(topics [][]byte) ([]byte, error) {
	numTokensToTransfer := len(topics[2:]) / numTransferTopics
	numTokensToTransferBytes := big.NewInt(int64(numTokensToTransfer)).Bytes()

	ret := []byte(core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(numTokensToTransferBytes))

	for idx := 2; idx < 2+len(topics[2:]); idx += 3 {
		tokenData, err := iep.dataCodec.GetTokenDataBytes(topics[idx+1], topics[idx+2])
		if err != nil {
			return nil, err
		}

		transfer := []byte("@" +
			hex.EncodeToString(topics[idx]) + // tokenID
			"@" + hex.EncodeToString(topics[idx+1]) + //nonce
			"@" + hex.EncodeToString(tokenData)) //tokenData

		ret = append(ret, transfer...)
	}

	return ret, nil
}

func (iep *incomingEventsProcessor) getConfirmedBridgeOperation(topics [][]byte) (*confirmedBridgeOp, error) {
	if len(topics) != numExecutedBridgeOpTopics {
		return nil, fmt.Errorf("%w for %s; num topics = %d", errInvalidNumTopicsIncomingEvent, eventIDExecutedOutGoingBridgeOp, len(topics))
	}

	return &confirmedBridgeOp{
		hashOfHashes: topics[0],
		hash:         topics[1],
	}, nil
}

func checkTopicsValidity(topics [][]byte) error {
	// TODO: Check each param validity (e.g. check that topic[0] == valid address)
	if len(topics) < minTopicsInTransferEvent || len(topics[2:])%numTransferTopics != 0 {
		log.Error("incomingHeaderHandler.createIncomingSCRs",
			"error", errInvalidNumTopicsIncomingEvent,
			"num topics", len(topics),
			"topics", topics)
		return fmt.Errorf("%w for %s; num topics = %d", errInvalidNumTopicsIncomingEvent, eventIDDepositIncomingTransfer, len(topics))
	}

	return nil
}
