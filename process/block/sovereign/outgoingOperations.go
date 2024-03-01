package sovereign

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

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

type operationData struct {
	address      []byte
	tokens       []esdtTokenPayment
	transferData transferData
}

type esdtTokenPayment struct {
	tokenIdentifier []byte
	nonce           uint64
	amount          *big.Int
}

type transferData struct {
	gasLimit uint64
	function []byte
	args     [][]byte
}

type eventData struct {
	nonce uint64
	*transferData
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

	txsData := make([]byte, 0)
	for _, event := range outgoingEvents {
		txsData = append(txsData, getOperationData(event)...)
	}

	// TODO: Check gas limit here and split tx data in multiple batches if required
	// Task: MX-14720
	return [][]byte{txsData}
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

func getOperationData(event data.EventHandler) []byte {
	opData := newOperationData(event.GetTopics())
	evData, _ := getEventData(event.GetData())
	opData.transferData = transferData{
		gasLimit: evData.gasLimit,
		function: evData.function,
		args:     evData.args,
	}

	return serializeOperationData(opData)
}

func newOperationData(topics [][]byte) *operationData {
	receiver := topics[1]

	var tokens []esdtTokenPayment
	for i := 2; i < len(topics); i += 3 {
		payment := esdtTokenPayment{
			tokenIdentifier: topics[i],
			nonce:           byteSliceToUint64(topics[i+1]),
			amount:          byteSliceToBigInt(topics[i+2]),
		}
		tokens = append(tokens, payment)
	}

	return &operationData{
		address: receiver,
		tokens:  tokens,
	}
}

func byteSliceToUint64(byteSlice []byte) uint64 {
	var result uint64
	for _, b := range byteSlice {
		result = (result << 8) | uint64(b)
	}
	return result
}

func byteSliceToBigInt(byteSlice []byte) *big.Int {
	return new(big.Int).SetBytes(byteSlice)
}

func getEventData(data []byte) (*eventData, error) {
	codec := abi.NewDefaultCodec()
	serializer := abi.NewSerializer(codec)

	eventDataStruct := abi.StructValue{
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

	err := serializer.Deserialize(hex.EncodeToString(data), []any{&eventDataStruct})
	if err != nil {
		return nil, err
	}

	nonce := eventDataStruct.Fields[0].Value.(*abi.U64Value).Value

	gasLimit := uint64(0)
	optionGas := eventDataStruct.Fields[1].Value.(*abi.OptionValue).Value
	if optionGas != nil {
		gasLimit = optionGas.(*abi.U64Value).Value
	}

	function := make([]byte, 0)
	optionFunc := eventDataStruct.Fields[2].Value.(*abi.OptionValue).Value
	if optionFunc != nil {
		function = optionFunc.(*abi.BytesValue).Value
	}

	args := make([][]byte, 0)
	optionArgs := eventDataStruct.Fields[3].Value.(*abi.OptionValue).Value
	if optionArgs != nil {
		items := optionArgs.(*abi.OutputListValue).Items
		if items != nil && len(items) > 0 {
			for _, item := range items {
				arg := item.(*abi.BytesValue).Value
				args = append(args, arg)
			}
		}
	}

	return &eventData{
		nonce: nonce,
		transferData: &transferData{
			gasLimit: gasLimit,
			function: function,
			args:     args,
		},
	}, nil
}

func serializeOperationData(txData *operationData) []byte {
	codec := abi.NewDefaultCodec()
	serializer := abi.NewSerializer(codec)

	var tokens []abi.StructValue
	for _, token := range txData.tokens {
		item := abi.StructValue{
			Fields: []abi.Field{
				{
					Name:  "token_identifier",
					Value: abi.BytesValue{Value: token.tokenIdentifier},
				},
				{
					Name:  "nonce",
					Value: abi.U64Value{Value: token.nonce},
				},
				{
					Name:  "amount",
					Value: abi.BigIntValue{Value: token.amount},
				},
			},
		}
		tokens = append(tokens, item)
	}

	var args []abi.BytesValue
	for _, arg := range txData.transferData.args {
		item := abi.BytesValue{Value: arg}
		args = append(args, item)
	}

	txDataStruct :=
		abi.StructValue{
			Fields: []abi.Field{
				{
					Name:  "receiver",
					Value: abi.BytesValue{Value: txData.address},
				},
				{
					Name: "tokens",
					Value: abi.InputListValue{
						Items: []any{tokens},
					},
				},
				{
					Name: "transfer_data",
					Value: abi.OptionValue{
						Value: abi.StructValue{
							Fields: []abi.Field{
								{
									Name:  "gas_limit",
									Value: abi.U64Value{Value: txData.transferData.gasLimit},
								},
								{
									Name:  "function",
									Value: abi.BytesValue{Value: txData.transferData.function},
								},
								{
									Name: "args",
									Value: abi.InputListValue{
										Items: []any{args},
									},
								},
							},
						},
					},
				},
			},
		}

	encoded, _ := serializer.Serialize([]any{txDataStruct})
	encodedBytes, _ := hex.DecodeString(encoded)
	return encodedBytes
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *outgoingOperations) IsInterfaceNil() bool {
	return op == nil
}
