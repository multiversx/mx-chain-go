package sovereign

import (
	"bytes"
	"encoding/hex"

	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const bridgeOpPrefix = "bridgeOps"
const numTransferTopics = 3

var log = logger.GetOrCreate("outgoing-operations")

// SubscribedEvent contains a subscribed event from the main chain that the notifier is watching
type SubscribedEvent struct {
	Identifier []byte
	Addresses  map[string]string
}

type outgoingOperations struct {
	subscribedEvents []SubscribedEvent
}

func (notifier *outgoingOperations) CreateOutgoingTxData(logs []*data.LogData) []byte {
	outgoingEvents := notifier.createIncomingEvents(logs)
	if len(outgoingEvents) == 0 {
		return make([]byte, 0)
	}

	txData := []byte(bridgeOpPrefix)
	for _, ev := range outgoingEvents {
		txData = append(txData, byte('@'))
		txData = append(txData, createSCRData(ev.GetTopics())...)
		txData = append(txData, ev.GetData()...)
	}

	return txData
}

func createSCRData(topics [][]byte) []byte {
	ret := topics[0]
	for idx := 1; idx < len(topics[1:]); idx += 3 {
		transfer := []byte("@" +
			hex.EncodeToString(topics[idx]) + // tokenID
			"@" + hex.EncodeToString(topics[idx+1]) + //nonce
			"@" + hex.EncodeToString(topics[idx+2])) //value

		ret = append(ret, transfer...)
	}

	return ret
}

func (notifier *outgoingOperations) createIncomingEvents(logs []*data.LogData) []data.EventHandler {
	incomingEvents := make([]data.EventHandler, 0)

	for _, logData := range logs {
		eventsFromLog := notifier.getIncomingEvents(logData)
		incomingEvents = append(incomingEvents, eventsFromLog...)
	}

	return incomingEvents
}

func (notifier *outgoingOperations) getIncomingEvents(logData *data.LogData) []data.EventHandler {
	incomingEvents := make([]data.EventHandler, 0)

	for _, event := range logData.GetLogEvents() {
		if !notifier.isSubscribed(event, logData.TxHash) {
			continue
		}

		incomingEvents = append(incomingEvents, event)
	}

	return incomingEvents
}

func (notifier *outgoingOperations) isSubscribed(event data.EventHandler, txHash string) bool {
	for _, subEvent := range notifier.subscribedEvents {
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
