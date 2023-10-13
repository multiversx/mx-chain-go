package sovereign

import (
	"bytes"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const bridgeOpPrefix = "bridgeOps"
const numTransferTopics = 3

var log = logger.GetOrCreate("outgoing-operations")

// SubscribedEvent contains a subscribed event from the main chain that the op is watching
type SubscribedEvent struct {
	Identifier []byte
	Addresses  map[string]string
}

type outgoingOperations struct {
	subscribedEvents []SubscribedEvent
}

func NewOutgoingOperationsCreator(subscribedEvents []SubscribedEvent) (*outgoingOperations, error) {
	err := checkEvents(subscribedEvents)
	if err != nil {
		return nil, err
	}

	return &outgoingOperations{
		subscribedEvents: subscribedEvents,
	}, nil
}

func checkEvents(events []SubscribedEvent) error {
	if len(events) == 0 {
		return errNoSubscribedEvent
	}

	log.Debug("sovereign op received config", "num subscribed events", len(events))
	for idx, event := range events {
		if len(event.Identifier) == 0 {
			return fmt.Errorf("%w at event index = %d", errNoSubscribedIdentifier, idx)
		}

		log.Debug("sovereign op", "subscribed event identifier", string(event.Identifier))

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

		log.Debug("sovereign op", "subscribed address", encodedAddr)
	}

	return nil
}

func (op *outgoingOperations) CreateOutgoingTxData(logs []*data.LogData) []byte {
	outgoingEvents := op.createIncomingEvents(logs)
	if len(outgoingEvents) == 0 {
		return make([]byte, 0)
	}

	txData := []byte(bridgeOpPrefix)
	for _, ev := range outgoingEvents {
		txData = append(txData, byte('@'))
		txData = append(txData, createSCRData(ev.GetTopics())...)
		txData = append(txData, byte('@'))
		txData = append(txData, ev.GetData()...)
	}

	return txData
}

func createSCRData(topics [][]byte) []byte {
	ret := topics[0]
	for idx := 1; idx < len(topics[1:]); idx += 1 {
		transfer := []byte("@")
		transfer = append(transfer, topics[idx]...)

		ret = append(ret, transfer...)
	}

	return ret
}

func (op *outgoingOperations) createIncomingEvents(logs []*data.LogData) []data.EventHandler {
	incomingEvents := make([]data.EventHandler, 0)

	for _, logData := range logs {
		eventsFromLog := op.getIncomingEvents(logData)
		incomingEvents = append(incomingEvents, eventsFromLog...)
	}

	return incomingEvents
}

func (op *outgoingOperations) getIncomingEvents(logData *data.LogData) []data.EventHandler {
	incomingEvents := make([]data.EventHandler, 0)

	for _, event := range logData.GetLogEvents() {
		if !op.isSubscribed(event, logData.TxHash) {
			continue
		}

		incomingEvents = append(incomingEvents, event)
	}

	return incomingEvents
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

func (op *outgoingOperations) IsInterfaceNil() bool {
	return op == nil
}
