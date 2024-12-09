package dataCodec

import (
	"encoding/hex"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-sdk-abi-go/abi"
)

type dataCodec struct {
	serializer AbiSerializer
}

// TODO MX-15286 split data codec in multiple components

// NewDataCodec creates a data codec which is able to serialize/deserialize data from incoming/outgoing operations
func NewDataCodec(serializer AbiSerializer) (*dataCodec, error) {
	if serializer == nil {
		return nil, errors.ErrNilSerializer
	}

	return &dataCodec{
		serializer: serializer,
	}, nil
}

// SerializeEventData will receive an event data and serialize it
func (dc *dataCodec) SerializeEventData(eventData sovereign.EventData) ([]byte, error) {
	eventDataStruct := getEventDataStruct(eventData)

	encodedOp, err := dc.serializer.Serialize([]any{eventDataStruct})
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(encodedOp)
}

func getEventDataStruct(eventData sovereign.EventData) *abi.StructValue {
	transferData := getTransferDataInAbiFormat(eventData.TransferData)

	return &abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "op_nonce",
				Value: &abi.U64Value{Value: eventData.Nonce},
			},
			{
				Name:  "op_sender",
				Value: &abi.AddressValue{Value: eventData.Sender},
			},
			{
				Name: "opt_transfer_data",
				Value: &abi.OptionValue{
					Value: transferData,
				},
			},
		},
	}
}

// DeserializeEventData will deserialize bytes to an event data structure
func (dc *dataCodec) DeserializeEventData(data []byte) (*sovereign.EventData, error) {
	if len(data) == 0 {
		return nil, errEmptyData
	}

	nonce := &abi.U64Value{}
	sender := &abi.AddressValue{}
	gasLimit := &abi.U64Value{}
	function := &abi.BytesValue{}
	args := &abi.ListValue{
		ItemCreator: func() abi.SingleValue {
			return &abi.BytesValue{}
		},
	}

	eventDataStruct := &abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "op_nonce",
				Value: nonce,
			},
			{
				Name:  "op_sender",
				Value: sender,
			},
			{
				Name: "opt_transfer_data",
				Value: &abi.OptionValue{
					Value: &abi.StructValue{
						Fields: []abi.Field{
							{
								Name:  "gas_limit",
								Value: gasLimit,
							},
							{
								Name:  "function",
								Value: function,
							},
							{
								Name:  "args",
								Value: args,
							},
						},
					},
				},
			},
		},
	}

	err := dc.serializer.Deserialize(hex.EncodeToString(data), []any{eventDataStruct})
	if err != nil {
		return nil, err
	}

	arguments, err := getTransferDataArguments(args.Items)
	if err != nil {
		return nil, err
	}

	transferData := getTransferData(gasLimit.Value, function.Value, arguments)

	return &sovereign.EventData{
		Nonce:        nonce.Value,
		Sender:       sender.Value,
		TransferData: transferData,
	}, nil
}

// SerializeTokenData will receive an esdt token data and serialize it
func (dc *dataCodec) SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error) {
	tokenDataStruct := getTokenDataStruct(tokenData)

	encodedTokenData, err := dc.serializer.Serialize([]any{tokenDataStruct})
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(encodedTokenData)
}

func getTokenDataStruct(tokenData sovereign.EsdtTokenData) *abi.StructValue {
	uris := getUrisInAbiFormat(tokenData.Uris)

	return &abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "token_type",
				Value: &abi.EnumValue{Discriminant: uint8(tokenData.TokenType)},
			},
			{
				Name:  "amount",
				Value: &abi.BigUIntValue{Value: tokenData.Amount},
			},
			{
				Name:  "frozen",
				Value: &abi.BoolValue{Value: tokenData.Frozen},
			},
			{
				Name:  "hash",
				Value: &abi.BytesValue{Value: tokenData.Hash},
			},
			{
				Name:  "name",
				Value: &abi.BytesValue{Value: tokenData.Name},
			},
			{
				Name:  "attributes",
				Value: &abi.BytesValue{Value: tokenData.Attributes},
			},
			{
				Name:  "creator",
				Value: &abi.AddressValue{Value: tokenData.Creator},
			},
			{
				Name:  "royalties",
				Value: &abi.BigUIntValue{Value: tokenData.Royalties},
			},
			{
				Name: "uris",
				Value: &abi.ListValue{
					Items: uris,
				},
			},
		},
	}
}

// DeserializeTokenData will deserialize bytes to an esdt token data
func (dc *dataCodec) DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error) {
	if len(data) == 0 {
		return nil, errEmptyTokenData
	}

	enum := &abi.EnumValue{
		FieldsProvider: func(discriminant uint8) []abi.Field {
			return nil
		},
	}
	amount := &abi.BigUIntValue{}
	frozen := &abi.BoolValue{}
	hash := &abi.BytesValue{}
	name := &abi.BytesValue{}
	attributes := &abi.BytesValue{}
	creator := &abi.AddressValue{}
	royalties := &abi.BigUIntValue{}
	uris := &abi.ListValue{
		ItemCreator: func() abi.SingleValue {
			return &abi.BytesValue{}
		},
	}

	tokenDataStruct := &abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "token_type",
				Value: enum,
			},
			{
				Name:  "amount",
				Value: amount,
			},
			{
				Name:  "frozen",
				Value: frozen,
			},
			{
				Name:  "hash",
				Value: hash,
			},
			{
				Name:  "name",
				Value: name,
			},
			{
				Name:  "attributes",
				Value: attributes,
			},
			{
				Name:  "creator",
				Value: creator,
			},
			{
				Name:  "royalties",
				Value: royalties,
			},
			{
				Name:  "uris",
				Value: uris,
			},
		},
	}

	err := dc.serializer.Deserialize(hex.EncodeToString(data), []any{tokenDataStruct})
	if err != nil {
		return nil, err
	}

	tokenUris, err := getTokenDataUris(uris.Items)
	if err != nil {
		return nil, err
	}

	return &sovereign.EsdtTokenData{
		TokenType:  core.ESDTType(enum.Discriminant),
		Amount:     amount.Value,
		Frozen:     frozen.Value,
		Hash:       hash.Value,
		Name:       name.Value,
		Attributes: attributes.Value,
		Creator:    creator.Value,
		Royalties:  royalties.Value,
		Uris:       tokenUris,
	}, nil
}

// SerializeOperation will receive an operation and serialize it
func (dc *dataCodec) SerializeOperation(operation sovereign.Operation) ([]byte, error) {
	operationStruct := getOperationStruct(operation)

	encodedOp, err := dc.serializer.Serialize([]any{operationStruct})
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(encodedOp)
}

func getOperationStruct(operation sovereign.Operation) *abi.StructValue {
	tokens := getOperationTokens(operation.Tokens)
	operationData := getOperationData(*operation.Data)

	return &abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "to",
				Value: &abi.AddressValue{Value: operation.Address},
			},
			{
				Name: "tokens",
				Value: &abi.ListValue{
					Items: tokens,
				},
			},
			{
				Name:  "data",
				Value: operationData,
			},
		},
	}
}

func createTransferData(transferData sovereign.TransferData) *abi.StructValue {
	arguments := make([]abi.SingleValue, len(transferData.Args))
	for i, arg := range transferData.Args {
		arguments[i] = &abi.BytesValue{Value: arg}
	}

	return &abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "gas_limit",
				Value: &abi.U64Value{Value: transferData.GasLimit},
			},
			{
				Name:  "function",
				Value: &abi.BytesValue{Value: transferData.Function},
			},
			{
				Name: "args",
				Value: &abi.ListValue{
					Items: arguments,
				},
			},
		},
	}
}

func getTransferDataInAbiFormat(transferData *sovereign.TransferData) abi.SingleValue {
	if transferData != nil {
		return createTransferData(*transferData)
	}

	return nil
}

func getTransferDataArguments(items []abi.SingleValue) ([][]byte, error) {
	arguments := make([][]byte, 0)
	for _, item := range items {
		arg, ok := item.(*abi.BytesValue)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		arguments = append(arguments, arg.Value)
	}

	return arguments, nil
}

func getTransferData(gasLimit uint64, function []byte, arguments [][]byte) *sovereign.TransferData {
	if len(function) == 0 {
		return nil
	}

	return &sovereign.TransferData{
		GasLimit: gasLimit,
		Function: function,
		Args:     arguments,
	}
}

func getUrisInAbiFormat(items [][]byte) []abi.SingleValue {
	uris := make([]abi.SingleValue, len(items))
	for i, uri := range items {
		uris[i] = &abi.BytesValue{Value: uri}
	}

	return uris
}

func getTokenDataUris(items []abi.SingleValue) ([][]byte, error) {
	uris := make([][]byte, 0)
	for _, item := range items {
		uri, ok := item.(*abi.BytesValue)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		uris = append(uris, uri.Value)
	}

	return uris, nil
}

func getOperationTokens(tokens []sovereign.EsdtToken) []abi.SingleValue {
	operationTokens := make([]abi.SingleValue, 0)
	for _, token := range tokens {
		tokenStruct := getTokenStruct(token)
		operationTokens = append(operationTokens, tokenStruct)
	}

	return operationTokens
}

func getTokenStruct(token sovereign.EsdtToken) *abi.StructValue {
	uris := getUrisInAbiFormat(token.Data.Uris)

	return &abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "token_identifier",
				Value: &abi.BytesValue{Value: token.Identifier},
			},
			{
				Name:  "token_nonce",
				Value: &abi.U64Value{Value: token.Nonce},
			},
			{
				Name: "token_data",
				Value: &abi.StructValue{
					Fields: []abi.Field{
						{
							Name:  "token_type",
							Value: &abi.EnumValue{Discriminant: uint8(token.Data.TokenType)},
						},
						{
							Name:  "amount",
							Value: &abi.BigUIntValue{Value: token.Data.Amount},
						},
						{
							Name:  "frozen",
							Value: &abi.BoolValue{Value: token.Data.Frozen},
						},
						{
							Name:  "hash",
							Value: &abi.BytesValue{Value: token.Data.Hash},
						},
						{
							Name:  "name",
							Value: &abi.BytesValue{Value: token.Data.Name},
						},
						{
							Name:  "attributes",
							Value: &abi.BytesValue{Value: token.Data.Attributes},
						},
						{
							Name:  "creator",
							Value: &abi.AddressValue{Value: token.Data.Creator},
						},
						{
							Name:  "royalties",
							Value: &abi.BigUIntValue{Value: token.Data.Royalties},
						},
						{
							Name: "uris",
							Value: &abi.ListValue{
								Items: uris,
							},
						},
					},
				},
			},
		},
	}
}

func getOperationData(data sovereign.EventData) *abi.StructValue {
	var transferData abi.SingleValue
	if data.TransferData == nil {
		transferData = nil
	} else {
		transferData = getTransferDataInAbiFormat(data.TransferData)
	}

	return &abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "op_nonce",
				Value: &abi.U64Value{Value: data.Nonce},
			},
			{
				Name:  "op_sender",
				Value: &abi.AddressValue{Value: data.Sender},
			},
			{
				Name: "opt_transfer_data",
				Value: &abi.OptionValue{
					Value: transferData,
				},
			},
		},
	}
}

// IsInterfaceNil checks if the underlying pointer is nil
func (dc *dataCodec) IsInterfaceNil() bool {
	return dc == nil
}
