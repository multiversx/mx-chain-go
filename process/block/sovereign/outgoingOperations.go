package sovereign

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-go/errors"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("outgoing-operations")

// SubscribedEvent contains a subscribed event from the sovereign chain needed to be transferred to the main chain
type SubscribedEvent struct {
	Identifier []byte
	Addresses  map[string]string
}

type ArgsOutgoingOperations struct {
	SubscribedEvents []SubscribedEvent
	DataCodec        DataCodecProcessor
	TopicsChecker    TopicsChecker
}

type outgoingOperations struct {
	subscribedEvents []SubscribedEvent
	dataCodec        DataCodecProcessor
	topicsChecker    TopicsChecker
}

// TODO: We should create a common base functionality from this component. Similar behavior is also found in
// mx-chain-sovereign-notifier-go in the sovereignNotifier.go file. This applies for the factory as well
// Task: MX-14721

// NewOutgoingOperationsFormatter creates an outgoing operations formatter
func NewOutgoingOperationsFormatter(args ArgsOutgoingOperations) (*outgoingOperations, error) {
	err := checkEvents(args.SubscribedEvents)
	if err != nil {
		return nil, err
	}
	if check.IfNil(args.DataCodec) {
		return nil, errors.ErrNilDataCodec
	}
	if check.IfNil(args.TopicsChecker) {
		return nil, errors.ErrNilTopicsChecker
	}

	return &outgoingOperations{
		subscribedEvents: args.SubscribedEvents,
		dataCodec:        args.DataCodec,
		topicsChecker:    args.TopicsChecker,
	}, nil
}

func checkEvents(events []SubscribedEvent) error {
	if len(events) == 0 {
		return errNoSubscribedEvent
	}

	log.Debug("sovereign outgoing operations creator: received config", "num subscribed events", len(events))
	for idx, event := range events {
		if len(event.Identifier) == 0 {
			return fmt.Errorf("%w at event index = %d", errNoSubscribedIdentifier, idx)
		}

		log.Debug("sovereign outgoing operations creator", "subscribed event identifier", string(event.Identifier))

		err := checkEmptyAddresses(event.Addresses)
		if err != nil {
			return fmt.Errorf("%w at event index = %d", err, idx)
		}
	}

	return nil
}

func checkEmptyAddresses(addresses map[string]string) error {
	if len(addresses) == 0 {
		return errNoSubscribedAddresses
	}

	for decodedAddr, encodedAddr := range addresses {
		if len(decodedAddr) == 0 || len(encodedAddr) == 0 {
			return errNoSubscribedAddresses
		}

		log.Debug("sovereign outgoing operations creator", "subscribed address", encodedAddr)
	}

	return nil
}

// CreateOutgoingTxsData collects relevant outgoing events(based on subscribed addresses and topics) for bridge from the
// logs and creates outgoing data that needs to be signed by validators to bridge tokens
func (op *outgoingOperations) CreateOutgoingTxsData(logs []*data.LogData) ([][]byte, error) {
	outgoingEvents := op.createOutgoingEvents(logs)
	if len(outgoingEvents) == 0 {
		return make([][]byte, 0), nil
	}

	txsData := make([][]byte, 0)
	for i, event := range outgoingEvents {
		operation, err := op.getOperationData(event)
		if err != nil {
			log.Error("outgoingOperations.CreateOutgoingTxsData error",
				"tx hash", logs[i].TxHash,
				"event", string(event.GetIdentifier()),
				"error", err)

			return nil, err
		}

		txsData = append(txsData, operation)
	}

	// TODO: Check gas limit here and split tx data in multiple batches if required
	// Task: MX-14720
	return txsData, nil
}

func (op *outgoingOperations) createOutgoingEvents(logs []*data.LogData) []data.EventHandler {
	events := make([]data.EventHandler, 0)

	for _, logData := range logs {
		eventsFromLog := op.createOutgoingEvent(logData)
		events = append(events, eventsFromLog...)
	}

	return events
}

func (op *outgoingOperations) createOutgoingEvent(logData *data.LogData) []data.EventHandler {
	events := make([]data.EventHandler, 0)

	for _, event := range logData.GetLogEvents() {
		if !op.isSubscribed(event, logData.TxHash) {
			continue
		}

		events = append(events, event)
	}

	return events
}

func (op *outgoingOperations) isSubscribed(event data.EventHandler, txHash string) bool {
	for _, subEvent := range op.subscribedEvents {
		if !bytes.Equal(event.GetIdentifier(), subEvent.Identifier) {
			continue
		}

		receiver := event.GetAddress()
		encodedAddr, found := subEvent.Addresses[string(receiver)]
		if !found {
			continue
		}

		log.Trace("found outgoing event", "original tx hash", txHash, "receiver", encodedAddr)
		return true
	}

	return false
}

func (op *outgoingOperations) getOperationData(event data.EventHandler) ([]byte, error) {
	operation, err := op.createOperationData(event.GetTopics())
	if err != nil {
		return nil, err
	}

	evData, err := op.dataCodec.DeserializeEventData(event.GetData())
	if err != nil {
		return nil, err
	}

	operation.Data = evData

	operationBytes, err := op.dataCodec.SerializeOperation(*operation)
	if err != nil {
		return nil, err
	}

	return operationBytes, nil
}

func (op *outgoingOperations) createOperationData(topics [][]byte) (*sovereign.Operation, error) {
	err := op.topicsChecker.CheckValidity(topics)
	if err != nil {
		return nil, err
	}

	receiver := topics[1]

	var tokens []sovereign.EsdtToken
	for i := 2; i < len(topics); i += 3 {
		tokenIdentifier := topics[i]
		tokenNonce := byteSliceToUint64(topics[i+1])
		tokenData, err := op.dataCodec.DeserializeTokenData(topics[i+2])
		if err != nil {
			return nil, err
		}

		payment := sovereign.EsdtToken{
			Identifier: tokenIdentifier,
			Nonce:      tokenNonce,
			Data:       *tokenData,
		}
		tokens = append(tokens, payment)
	}

	return &sovereign.Operation{
		Address: receiver,
		Tokens:  tokens,
	}, nil
}

func byteSliceToUint64(byteSlice []byte) uint64 {
	var result uint64
	for _, b := range byteSlice {
		result = (result << 8) | uint64(b)
	}
	return result
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *outgoingOperations) IsInterfaceNil() bool {
	return op == nil
}
