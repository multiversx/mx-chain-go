package incomingHeader

import (
	"encoding/hex"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/sovereignnode/dataCodec"
)

const (
	minTopicsInTransferEvent  = 5
	numTransferTopics         = 3
	numExecutedBridgeOpTopics = 3
	tokensIndex               = 2
)

const (
	eventIDExecutedOutGoingBridgeOp = "executed"
	eventIDDepositIncomingTransfer  = "deposit"

	topicIDConfirmedOutGoingOperation = "executedBridgeOp"
	topicIDDepositIncomingTransfer    = "deposit"
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
	marshaller    marshal.Marshalizer
	hasher        hashing.Hasher
	dataCodec     dataCodec.SovereignDataDecoder
	topicsChecker TopicsChecker
}

// TODO refactor this to work with processors that assign tasks based on event id
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
				return nil, fmt.Errorf("%w for %s; num topics = %d", errInvalidNumTopicsIncomingEvent, eventIDExecutedOutGoingBridgeOp, len(topics))
			}

			switch string(topics[0]) {
			case topicIDDepositIncomingTransfer:
				scr, err = iep.createSCRInfo(topics, event)
				scrs = append(scrs, scr)
			case topicIDConfirmedOutGoingOperation:
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
	err := iep.topicsChecker.CheckValidity(topics)
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

	d := string(scr.Data)
	_ = d

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
	evData, err := iep.dataCodec.DeserializeEventData(data)
	if err != nil {
		return nil, err
	}

	gasLimit, args := extractTransferData(evData.TransferData)

	return &eventData{
		nonce:                evData.Nonce,
		functionCallWithArgs: args,
		gasLimit:             gasLimit,
	}, nil
}

func extractTransferData(transferData *sovereign.TransferData) (uint64, []byte) {
	gasLimit := uint64(0)
	args := make([]byte, 0)
	if transferData != nil {
		gasLimit = transferData.GasLimit

		args = append(args, []byte("@")...)
		args = append(args, hex.EncodeToString(transferData.Function)...)
		args = append(args, extractArguments(transferData.Args)...)
	}

	return gasLimit, args
}

func extractArguments(arguments [][]byte) []byte {
	args := make([]byte, 0)
	if len(arguments) > 0 {
		for _, arg := range arguments {
			args = append(args, []byte("@")...)
			args = append(args, hex.EncodeToString(arg)...)
		}
	}

	return args
}

func (iep *incomingEventsProcessor) createSCRData(topics [][]byte) ([]byte, error) {
	numTokensToTransfer := len(topics[2:]) / numTransferTopics
	numTokensToTransferBytes := big.NewInt(int64(numTokensToTransfer)).Bytes()

	ret := []byte(core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(numTokensToTransferBytes))

	for idx := tokensIndex; idx < tokensIndex+len(topics[tokensIndex:]); idx += numTransferTopics {
		tokenData, err := iep.getTokenDataBytes(topics[idx+1], topics[idx+2])
		if err != nil {
			return nil, err
		}

		transfer := []byte("@" +
			hex.EncodeToString(topics[idx]) + // tokenID
			"@" + hex.EncodeToString(topics[idx+1]) + // nonce
			"@" + hex.EncodeToString(tokenData)) // value/tokenData

		ret = append(ret, transfer...)
	}

	return ret, nil
}

func (iep *incomingEventsProcessor) getTokenDataBytes(tokenNonce []byte, tokenData []byte) ([]byte, error) {
	esdtTokenData, err := iep.dataCodec.DeserializeTokenData(tokenData)
	if err != nil {
		return nil, err
	}

	if esdtTokenData.TokenType == core.Fungible {
		return esdtTokenData.Amount.Bytes(), nil
	}

	digitalToken := &esdt.ESDigitalToken{
		Type:  uint32(esdtTokenData.TokenType),
		Value: esdtTokenData.Amount,
		TokenMetaData: &esdt.MetaData{
			Nonce:      bytesToUint64(tokenNonce),
			Name:       esdtTokenData.Name,
			Creator:    esdtTokenData.Creator,
			Royalties:  uint32(esdtTokenData.Royalties.Uint64()),
			Hash:       esdtTokenData.Hash,
			URIs:       esdtTokenData.Uris,
			Attributes: esdtTokenData.Attributes,
		},
	}

	return iep.marshaller.Marshal(digitalToken)
}

func bytesToUint64(bytes []byte) uint64 {
	var result uint64
	for _, b := range bytes {
		result = (result << 8) | uint64(b)
	}
	return result
}

func (iep *incomingEventsProcessor) getConfirmedBridgeOperation(topics [][]byte) (*confirmedBridgeOp, error) {
	if len(topics) != numExecutedBridgeOpTopics {
		return nil, fmt.Errorf("%w for %s; num topics = %d", errInvalidNumTopicsIncomingEvent, eventIDExecutedOutGoingBridgeOp, len(topics))
	}

	return &confirmedBridgeOp{
		hashOfHashes: topics[1],
		hash:         topics[2],
	}, nil
}
