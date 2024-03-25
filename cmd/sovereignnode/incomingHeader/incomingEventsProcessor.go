package incomingHeader

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
)

const (
	minTopicsInTransferEvent  = 4
	numTransferTopics         = 3
	numExecutedBridgeOpTopics = 2
	minNumEventDataTokens     = 4
)

const (
	topicIDExecutedBridgeOp = "executedBridgeOp"
	topicIDDeposit          = "deposit"
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
		case topicIDDeposit:
			scr, err = iep.createSCRInfo(topics, event)
			scrs = append(scrs, scr)
		case topicIDExecutedBridgeOp:
			confirmedOp, err = iep.getConfirmedBridgeOperation(topics)
			confirmedBridgeOps = append(confirmedBridgeOps, confirmedOp)
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
	// TODO: Check each param validity (e.g. check that topic[0] == valid address)
	if len(topics) < minTopicsInTransferEvent || len(topics[1:])%numTransferTopics != 0 {
		log.Error("incomingHeaderHandler.createIncomingSCRs",
			"error", errInvalidNumTopicsIncomingEvent,
			"num topics", len(topics),
			"topics", topics)
		return nil, fmt.Errorf("%w; num topics = %d",
			errInvalidNumTopicsIncomingEvent, len(topics))
	}

	receivedEventData, err := getEventData(event.GetData())
	if err != nil {
		return nil, err
	}

	scrData := createSCRData(topics)
	scrData = append(scrData, receivedEventData.functionCallWithArgs...)
	scr := &smartContractResult.SmartContractResult{
		Nonce:          receivedEventData.nonce,
		OriginalTxHash: nil, // TODO:  Implement this in MX-14321 task
		RcvAddr:        topics[0],
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

func getEventData(data []byte) (*eventData, error) {
	if len(data) == 0 {
		return nil, errEmptyLogData
	}

	tokens := strings.Split(string(data), "@")
	numTokens := len(tokens)

	// TODO: WITHOUT FUNCTION CALL
	if numTokens < minNumEventDataTokens {
		return nil, fmt.Errorf("%w, expected min num tokens: %d, received num tokens: %d",
			errInvalidNumTokensOnLogData, minNumEventDataTokens, numTokens)
	}

	// TODO: Add validity checks
	eventNonce := big.NewInt(0).SetBytes([]byte(tokens[0]))
	gasLimit := big.NewInt(0).SetBytes([]byte(tokens[numTokens-1]))

	functionCallWithArgs := []byte("@" + tokens[1])
	for i := 2; i < numTokens-1; i++ {
		functionCallWithArgs = append(functionCallWithArgs, []byte("@"+tokens[i])...)
	}

	return &eventData{
		nonce:                eventNonce.Uint64(),
		gasLimit:             gasLimit.Uint64(),
		functionCallWithArgs: functionCallWithArgs,
	}, nil
}

func createSCRData(topics [][]byte) []byte {
	numTokensToTransfer := len(topics[1:]) / numTransferTopics
	numTokensToTransferBytes := big.NewInt(int64(numTokensToTransfer)).Bytes()

	ret := []byte(core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(numTokensToTransferBytes))

	for idx := 1; idx < len(topics[1:]); idx += 3 {
		transfer := []byte("@" +
			hex.EncodeToString(topics[idx]) + // tokenID
			"@" + hex.EncodeToString(topics[idx+1]) + //nonce
			"@" + hex.EncodeToString(topics[idx+2])) //value

		ret = append(ret, transfer...)
	}

	return ret
}

func (iep *incomingEventsProcessor) getConfirmedBridgeOperation(topics [][]byte) (*confirmedBridgeOp, error) {
	if len(topics) != numExecutedBridgeOpTopics {
		return nil, fmt.Errorf("%w for %s", errInvalidNumTopicsIncomingEvent, topicIDExecutedBridgeOp)
	}

	return &confirmedBridgeOp{
		hashOfHashes: topics[0],
		hash:         topics[1],
	}, nil
}
