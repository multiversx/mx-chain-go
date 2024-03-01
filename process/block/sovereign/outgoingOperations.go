package sovereign

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/errors"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-abi-incubator/golang/abi"
)

var log = logger.GetOrCreate("outgoing-operations")

// SubscribedEvent contains a subscribed event from the sovereign chain needed to be transferred to the main chain
type SubscribedEvent struct {
	Identifier []byte
	Addresses  map[string]string
}

type outgoingOperations struct {
	subscribedEvents []SubscribedEvent
	roundHandler     RoundHandler
}

// TODO: We should create a common base functionality from this component. Similar behavior is also found in
// mx-chain-sovereign-notifier-go in the sovereignNotifier.go file. This applies for the factory as well
// Task: MX-14721

// NewOutgoingOperationsFormatter creates an outgoing operations formatter
func NewOutgoingOperationsFormatter(subscribedEvents []SubscribedEvent, roundHandler RoundHandler) (*outgoingOperations, error) {
	err := checkEvents(subscribedEvents)
	if err != nil {
		return nil, err
	}
	if check.IfNil(roundHandler) {
		return nil, errors.ErrNilRoundHandler
	}

	return &outgoingOperations{
		subscribedEvents: subscribedEvents,
		roundHandler:     roundHandler,
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
func (op *outgoingOperations) CreateOutgoingTxsData(logs []*data.LogData) [][]byte {
	outgoingEvents := op.createOutgoingEvents(logs)
	if len(outgoingEvents) == 0 {
		return make([][]byte, 0)
	}

	txData := make([]byte, 0)
	for _, ev := range outgoingEvents {
		_, eventArgs, _ := getEventData(ev.GetData())

		txData = append(txData, byte('@'))
		txData = append(txData, createSCRData(ev.GetTopics())...)
		txData = append(txData, eventArgs...)
	}

	// TODO: Check gas limit here and split tx data in multiple batches if required
	// Task: MX-14720
	return [][]byte{txData}
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

func createSCRData(topics [][]byte) []byte {
	ret := make([]byte, 0)
	ret = append(ret, hex.EncodeToString(topics[1])...)

	for idx := 2; idx < len(topics); idx += 1 {
		ret = append(ret, []byte("@")...)
		ret = append(ret, hex.EncodeToString(topics[idx])...)
	}

	return ret
}

func getEventData(data []byte) (uint64, []byte, error) {
	codec := abi.NewDefaultCodec()
	serializer := abi.NewSerializer(codec)

	transferData := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "tx_nonce",
				Value: &abi.U64Value{},
			},
			{
				Name: "gas_limit",
				Value: &abi.OptionValue{
					Value: &abi.U64Value{},
				},
			},
			{
				Name: "function",
				Value: &abi.OptionValue{
					Value: &abi.BytesValue{},
				},
			},
			{
				Name: "args",
				Value: &abi.OptionValue{
					Value: &abi.OutputListValue{
						ItemCreator: func() any { return &abi.BytesValue{} },
					},
				},
			},
		},
	}

	err := serializer.Deserialize(hex.EncodeToString(data), []any{&transferData})
	if err != nil {
		return 0, nil, err
	}

	nonce := transferData.Fields[0].Value.(*abi.U64Value).Value
	args := make([]byte, 0)

	optionFunc := transferData.Fields[2].Value.(*abi.OptionValue).Value
	if optionFunc != nil {
		function := optionFunc.(*abi.BytesValue).Value
		args = append(args, []byte("@")...)
		args = append(args, hex.EncodeToString(function)...)

		optionArgs := transferData.Fields[3].Value.(*abi.OptionValue).Value
		if optionArgs != nil {
			items := optionArgs.(*abi.OutputListValue).Items
			if items != nil && len(items) > 0 {
				for _, item := range items {
					arg := item.(*abi.BytesValue).Value
					args = append(args, []byte("@")...)
					args = append(args, hex.EncodeToString(arg)...)
				}
			}
		}

		gasLimit := uint64(0)
		optionGas := transferData.Fields[1].Value.(*abi.OptionValue).Value
		if optionGas != nil {
			gasLimit = optionGas.(*abi.U64Value).Value
		}

		gasHex := append([]byte("@"), valueToHexString(gasLimit)...)
		args = append(gasHex, args...)
	}

	return nonce, args, nil
}

func valueToHexString(value uint64) string {
	hexString := strconv.FormatUint(value, 16)
	paddedHexString := fmt.Sprintf("%016s", hexString)

	return paddedHexString
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *outgoingOperations) IsInterfaceNil() bool {
	return op == nil
}
