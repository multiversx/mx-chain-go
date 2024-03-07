package dataCodec

import (
	"encoding/hex"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-sdk-abi-incubator/golang/abi"
)

type abiSerializer interface {
	Serialize(inputValues []any) (string, error)
	Deserialize(data string, outputValues []any) error
}

// DataCodec holds the
type DataCodec struct {
	Serializer abiSerializer
	Marshaller marshal.Marshalizer
}

// NewDataCodec creates a data codec which is able to serialize/deserialize data from incoming/outgoing operations
func NewDataCodec(marshaller marshal.Marshalizer) (*DataCodec, error) {
	if check.IfNil(marshaller) {
		return nil, errors.ErrNilMarshalizer
	}

	codec := abi.NewDefaultCodec()
	return &DataCodec{
		Serializer: abi.NewSerializer(codec),
		Marshaller: marshaller,
	}, nil
}

// SerializeEventData will receive an event data and serialize it
func (dc *DataCodec) SerializeEventData(eventData sovereign.EventData) ([]byte, error) {
	var gasLimit any
	var function any
	var args any
	if eventData.TransferData != nil {
		gasLimit = abi.U64Value{Value: eventData.GasLimit}
		function = abi.BytesValue{Value: eventData.Function}
		arguments := make([]any, len(eventData.Args))
		for i, arg := range eventData.Args {
			arguments[i] = abi.BytesValue{Value: arg}
		}
		args = abi.InputListValue{
			Items: arguments,
		}
	} else {
		gasLimit = nil
		function = nil
		args = nil
	}

	eventDataStruct := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "tx_nonce",
				Value: abi.U64Value{Value: eventData.Nonce},
			},
			{
				Name: "gas_limit",
				Value: abi.OptionValue{
					Value: gasLimit,
				},
			},
			{
				Name: "function",
				Value: abi.OptionValue{
					Value: function,
				},
			},
			{
				Name: "args",
				Value: abi.OptionValue{
					Value: args,
				},
			},
		},
	}

	encodedOp, err := dc.Serializer.Serialize([]any{eventDataStruct})
	if err != nil {
		return nil, err
	}

	operationBytes, _ := hex.DecodeString(encodedOp)
	return operationBytes, nil
}

// DeserializeEventData will deserialize bytes to an event data structure
func (dc *DataCodec) DeserializeEventData(data []byte) (*sovereign.EventData, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data provided")
	}

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

	err := dc.Serializer.Deserialize(hex.EncodeToString(data), []any{&eventDataStruct})
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

	var td *sovereign.TransferData
	if len(function) == 0 {
		td = nil
	} else {
		td = &sovereign.TransferData{
			GasLimit: gasLimit,
			Function: function,
			Args:     args,
		}
	}

	return &sovereign.EventData{
		Nonce:        nonce,
		TransferData: td,
	}, nil
}

// SerializeTokenData will receive an esdt token data and serialize it
func (dc *DataCodec) SerializeTokenData(tokenData sovereign.EsdtTokenData) ([]byte, error) {
	tokenUris := make([]any, len(tokenData.Uris))
	for i, uri := range tokenData.Uris {
		tokenUris[i] = abi.BytesValue{Value: uri}
	}

	tokenDataStruct := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "token_type",
				Value: abi.EnumValue{Discriminant: uint8(tokenData.TokenType)},
			},
			{
				Name:  "amount",
				Value: abi.BigIntValue{Value: tokenData.Amount},
			},
			{
				Name:  "frozen",
				Value: abi.BoolValue{Value: tokenData.Frozen},
			},
			{
				Name:  "hash",
				Value: abi.BytesValue{Value: tokenData.Hash},
			},
			{
				Name:  "name",
				Value: abi.BytesValue{Value: tokenData.Name},
			},
			{
				Name:  "attributes",
				Value: abi.BytesValue{Value: tokenData.Attributes},
			},
			{
				Name:  "creator",
				Value: abi.AddressValue{Value: tokenData.Creator},
			},
			{
				Name:  "royalties",
				Value: abi.BigIntValue{Value: tokenData.Royalties},
			},
			{
				Name: "uris",
				Value: abi.InputListValue{
					Items: tokenUris,
				},
			},
		},
	}

	encodedTokenData, err := dc.Serializer.Serialize([]any{tokenDataStruct})
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(encodedTokenData)
}

// DeserializeTokenData will deserialize bytes to an esdt token data
func (dc *DataCodec) DeserializeTokenData(data []byte) (*sovereign.EsdtTokenData, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty token data provided")
	}

	tokenDataStruct := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "token_type",
				Value: &abi.EnumValue{},
			},
			{
				Name:  "amount",
				Value: &abi.BigIntValue{},
			},
			{
				Name:  "frozen",
				Value: &abi.BoolValue{},
			},
			{
				Name:  "hash",
				Value: &abi.BytesValue{},
			},
			{
				Name:  "name",
				Value: &abi.BytesValue{},
			},
			{
				Name:  "attributes",
				Value: &abi.BytesValue{},
			},
			{
				Name:  "creator",
				Value: &abi.AddressValue{},
			},
			{
				Name:  "royalties",
				Value: &abi.BigIntValue{},
			},
			{
				Name: "uris",
				Value: &abi.OutputListValue{
					ItemCreator: func() any { return &abi.BytesValue{} },
				},
			},
		},
	}

	err := dc.Serializer.Deserialize(hex.EncodeToString(data), []any{&tokenDataStruct})
	if err != nil {
		return nil, err
	}

	urisArg := tokenDataStruct.Fields[8].Value.(*abi.OutputListValue).Items
	uris := make([][]byte, len(urisArg))
	for i, item := range urisArg {
		uris[i] = item.(*abi.BytesValue).Value
	}

	return &sovereign.EsdtTokenData{
		TokenType:  core.ESDTType(tokenDataStruct.Fields[0].Value.(*abi.EnumValue).Discriminant),
		Amount:     tokenDataStruct.Fields[1].Value.(*abi.BigIntValue).Value,
		Frozen:     tokenDataStruct.Fields[2].Value.(*abi.BoolValue).Value,
		Hash:       tokenDataStruct.Fields[3].Value.(*abi.BytesValue).Value,
		Name:       tokenDataStruct.Fields[4].Value.(*abi.BytesValue).Value,
		Attributes: tokenDataStruct.Fields[5].Value.(*abi.BytesValue).Value,
		Creator:    tokenDataStruct.Fields[6].Value.(*abi.AddressValue).Value,
		Royalties:  tokenDataStruct.Fields[7].Value.(*abi.BigIntValue).Value,
		Uris:       uris,
	}, nil
}

// GetTokenDataBytes will receive token properties, deserialize then into a digital token and return marshalled digital token
func (dc *DataCodec) GetTokenDataBytes(tokenNonce []byte, tokenData []byte) ([]byte, error) {
	if len(tokenData) == 0 {
		return nil, fmt.Errorf("empty token data provided")
	}

	esdtTokenData, err := dc.DeserializeTokenData(tokenData)
	if err != nil {
		return nil, err
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

	if digitalToken.Type == uint32(core.Fungible) {
		return digitalToken.Value.Bytes(), nil
	}

	return dc.Marshaller.Marshal(digitalToken)
}

func bytesToUint64(bytes []byte) uint64 {
	var result uint64
	for _, b := range bytes {
		result = (result << 8) | uint64(b)
	}
	return result
}

// SerializeOperation will receive an operation and serialize it
func (dc *DataCodec) SerializeOperation(operation sovereign.Operation) ([]byte, error) {
	tokens := make([]any, len(operation.Tokens))
	for i, token := range operation.Tokens {
		tokenUris := make([]any, len(token.Data.Uris))
		for i, uri := range token.Data.Uris {
			tokenUris[i] = abi.BytesValue{Value: uri}
		}

		item := abi.StructValue{
			Fields: []abi.Field{
				{
					Name:  "token_identifier",
					Value: abi.BytesValue{Value: token.Identifier},
				},
				{
					Name:  "token_nonce",
					Value: abi.U64Value{Value: token.Nonce},
				},
				{
					Name: "token_data",
					Value: abi.StructValue{
						Fields: []abi.Field{
							{
								Name:  "token_type",
								Value: abi.EnumValue{Discriminant: uint8(token.Data.TokenType)},
							},
							{
								Name:  "amount",
								Value: abi.BigIntValue{Value: token.Data.Amount},
							},
							{
								Name:  "frozen",
								Value: abi.BoolValue{Value: token.Data.Frozen},
							},
							{
								Name:  "hash",
								Value: abi.BytesValue{Value: token.Data.Hash},
							},
							{
								Name:  "name",
								Value: abi.BytesValue{Value: token.Data.Name},
							},
							{
								Name:  "attributes",
								Value: abi.BytesValue{Value: token.Data.Attributes},
							},
							{
								Name:  "creator",
								Value: abi.AddressValue{Value: token.Data.Creator},
							},
							{
								Name:  "royalties",
								Value: abi.BigIntValue{Value: token.Data.Royalties},
							},
							{
								Name: "uris",
								Value: abi.InputListValue{
									Items: tokenUris,
								},
							},
						},
					},
				},
			},
		}

		tokens[i] = item
	}

	var transferData any
	if operation.TransferData == nil {
		transferData = nil
	} else {
		args := make([]any, len(operation.TransferData.Args))
		for i, arg := range operation.TransferData.Args {
			args[i] = abi.BytesValue{Value: arg}
		}

		transferData = abi.StructValue{
			Fields: []abi.Field{
				{
					Name:  "gas_limit",
					Value: abi.U64Value{Value: operation.TransferData.GasLimit},
				},
				{
					Name:  "function",
					Value: abi.BytesValue{Value: operation.TransferData.Function},
				},
				{
					Name: "args",
					Value: abi.InputListValue{
						Items: args,
					},
				},
			},
		}
	}

	operationStruct := abi.StructValue{
		Fields: []abi.Field{
			{
				Name:  "receiver",
				Value: abi.AddressValue{Value: operation.Address},
			},
			{
				Name: "tokens",
				Value: abi.InputListValue{
					Items: tokens,
				},
			},
			{
				Name: "transfer_data",
				Value: abi.OptionValue{
					Value: transferData,
				},
			},
		},
	}

	encodedOp, err := dc.Serializer.Serialize([]any{operationStruct})
	if err != nil {
		return nil, err
	}

	return hex.DecodeString(encodedOp)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (dc *DataCodec) IsInterfaceNil() bool {
	return dc == nil
}
